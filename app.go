package fynekit

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	build                      func(*App) fyne.CanvasObject
	destroy                    func()
	runner                     *Runner
	fyneApp                    fyne.App
	window                     fyne.Window
	appStarted                 chan struct{}
	appStopped                 chan struct{}
	startedListeners           []func()
	stoppedListeners           []func()
	enteredForegroundListeners []func()
	exitedForegroundListeners  []func()
	runWasCalled               int32
}

func NewApp(appID, windowTitle string, build func(*App) fyne.CanvasObject, destroy func()) *App {
	a := &App{
		build:      build,
		destroy:    destroy,
		runner:     NewRunner(),
		appStarted: make(chan struct{}),
		appStopped: make(chan struct{}),
	}

	a.fyneApp = app.NewWithID(appID)
	a.fyneApp.Lifecycle().SetOnStarted(func() {
		close(a.appStarted)
		for _, fn := range a.startedListeners {
			fn()
		}
	})
	a.fyneApp.Lifecycle().SetOnStopped(func() {
		close(a.appStopped)
		for _, fn := range a.stoppedListeners {
			fn()
		}
	})
	a.fyneApp.Lifecycle().SetOnEnteredForeground(func() {
		for _, fn := range a.enteredForegroundListeners {
			fn()
		}
	})
	a.fyneApp.Lifecycle().SetOnExitedForeground(func() {
		for _, fn := range a.exitedForegroundListeners {
			fn()
		}
	})

	a.window = a.fyneApp.NewWindow(windowTitle)

	return a
}

func (a *App) Runner() *Runner {
	return a.runner
}

func (a *App) FyneApp() fyne.App {
	return a.fyneApp
}

func (a *App) Window() fyne.Window {
	return a.window
}

func (a *App) Run() {
	if !atomic.CompareAndSwapInt32(&a.runWasCalled, 0, 1) {
		panic("App.Run() was already called")
	}
	var stopOnce sync.Once
	a.Refresh()
	a.window.SetCloseIntercept(func() {
		go func() {
			stopOnce.Do(a.stop)
			fyne.DoAndWait(a.window.Close)
		}()
	})
	a.window.ShowAndRun()
	stopOnce.Do(a.stop)
}

func (a *App) Refresh() {
	a.window.SetContent(a.build(a))
}

func (a *App) stop() {
	a.runner.Stop()
	a.destroy()
}

func (a *App) dialogMinSize() fyne.Size {
	s := a.window.Canvas().Size()
	if fyne.IsVertical(fyne.CurrentDevice().Orientation()) {
		s.Width *= 0.75
		s.Height = 0
	} else {
		s.Width *= 0.25
		s.Height = 0
	}
	s = s.Max(fyne.NewSize(200, 0))
	return s
}

func (a *App) AppStarted() <-chan struct{} {
	return a.appStarted
}

func (a *App) AppStopped() <-chan struct{} {
	return a.appStopped
}

func (a *App) AddStartedListener(fn func()) {
	a.startedListeners = append(a.startedListeners, fn)
}

func (a *App) AddStoppedListener(fn func()) {
	a.stoppedListeners = append(a.stoppedListeners, fn)
}

func (a *App) AddEnteredForegroundListener(fn func()) {
	a.enteredForegroundListeners = append(a.enteredForegroundListeners, fn)
}

func (a *App) AddExitedForegroundListener(fn func()) {
	a.exitedForegroundListeners = append(a.exitedForegroundListeners, fn)
}

func (a *App) Go(fn func(ctx context.Context)) (done <-chan struct{}, err error) {
	if a.runWasCalled == 0 {
		panic("App.Run() was not called")
	}
	d := make(chan struct{})
	err = a.runner.DoAsync(func(ctx context.Context) {
		defer close(d)
		select {
		case <-ctx.Done():
			return
		case <-a.appStarted:
		}
		fn(ctx)
	})
	if err != nil {
		close(d)
	}
	return d, err
}

func (a *App) Do(fn func()) (done <-chan struct{}, err error) {
	return a.Go(func(context.Context) {
		fyne.DoAndWait(fn)
	})
}

func (a *App) DoWhenNoOverlay(fn func()) (done <-chan struct{}, err error) {
	return a.Go(func(ctx context.Context) {
		for finished := false; !finished; {
			fyne.DoAndWait(func() {
				if a.window.Canvas().Overlays().Top() != nil {
					return
				}
				fn()
				finished = true
			})
			if finished {
				continue
			}
			select {
			case <-ctx.Done():
				finished = true
			case <-time.After(time.Second / 64):
			}
		}
	})
}

func (a *App) Execute(ctx context.Context, diag dialog.Dialog, fn func(ctx context.Context) (finalize func())) {
	a.Do(func() {
		ctx, cancel := context.WithCancel(ctx)
		if diag == nil {
			diag = dialog.NewCustomWithoutButtons(lang.L("Please wait..."),
				WrapWithMinSize(widget.NewProgressBarInfinite(), a.dialogMinSize(), nil),
				a.window,
			)
		}
		diag.SetOnClosed(cancel)
		diag.Show()
		if _, e := a.Go(func(ctx2 context.Context) {
			defer cancel()
			go func() {
				select {
				case <-ctx.Done():
				case <-ctx2.Done():
					cancel()
				}
			}()
			finalize := fn(ctx)
			fyne.DoAndWait(func() {
				diag.Dismiss()
				if finalize != nil {
					finalize()
				}
			})
		}); e != nil {
			diag.Dismiss()
			cancel()
		}
	})
}

func (a *App) CreateBasicDialog(title, dismiss, message string, icon fyne.Resource) dialog.Dialog {
	diag := dialog.NewCustom(title, dismiss,
		WrapWithMinSize(widget.NewLabel(message), a.dialogMinSize(), func(o fyne.CanvasObject) {
			w := o.(*widget.Label)
			w.Alignment = fyne.TextAlignCenter
			w.Wrapping = fyne.TextWrapWord
		}),
		a.window,
	)
	if icon != nil {
		diag.SetIcon(icon)
	}
	return diag
}

func (a *App) ShowInformationDialog(message string, onClosed func()) {
	a.DoWhenNoOverlay(func() {
		diag := a.CreateBasicDialog(lang.L("Information"), lang.L("OK"), message, theme.InfoIcon())
		if onClosed != nil {
			diag.SetOnClosed(onClosed)
		}
		diag.Show()
	})
}

func (a *App) ShowErrorDialog(message string, onClosed func()) {
	a.DoWhenNoOverlay(func() {
		diag := a.CreateBasicDialog(lang.L("Error"), lang.L("OK"), message, theme.ErrorIcon())
		if onClosed != nil {
			diag.SetOnClosed(onClosed)
		}
		diag.Show()
	})
}
