package fynekit

import (
	"context"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/orkunkaraduman/goinsane"
)

type App struct {
	build                      func(*App) fyne.CanvasObject
	destroy                    func()
	contextRunner              *goinsane.ContextRunner
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
		build:         build,
		destroy:       destroy,
		contextRunner: goinsane.NewContextRunner(nil),
		appStarted:    make(chan struct{}),
		appStopped:    make(chan struct{}),
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
	a.Rebuild()
	a.window.SetCloseIntercept(func() {
		go func() {
			a.contextRunner.Stop()
			fyne.DoAndWait(a.window.Close)
		}()
	})
	a.window.ShowAndRun()
	a.contextRunner.Stop()
	if a.destroy != nil {
		a.destroy()
	}
}

func (a *App) Rebuild() {
	a.window.SetContent(a.build(a))
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

func (a *App) Go(fn func(ctx context.Context) error) (done <-chan error) {
	if a.runWasCalled == 0 {
		panic("App.Run() was not called")
	}
	d := make(chan error, 1)
	err := a.contextRunner.RunAsync(func(ctx context.Context) {
		defer close(d)
		select {
		case <-ctx.Done():
			d <- ctx.Err()
		case <-a.appStarted:
			d <- fn(ctx)
		}
	})
	if err != nil {
		d <- err
		close(d)
	}
	return d
}

func (a *App) Do(fn func()) (done <-chan error) {
	return a.Go(func(context.Context) error {
		fyne.DoAndWait(fn)
		return nil
	})
}

func (a *App) DoWhenNoOverlay(fn func()) (done <-chan error) {
	return a.Go(func(ctx context.Context) (err error) {
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
				err = ctx.Err()
				finished = true
			case <-time.After(time.Second / 16):
			}
		}
		return
	})
}

func (a *App) Execute(ctx context.Context, diag dialog.Dialog, fn func(ctx context.Context) (finalize func())) {
	a.Do(func() {
		ctx, cancel := context.WithCancel(ctx)
		var activity *widget.Activity
		if diag == nil {
			s := a.getDialogMinSize()
			/*diag = dialog.NewCustomWithoutButtons(lang.L("Please wait..."),
				WrapWithMinSize(widget.NewProgressBarInfinite(), s, nil),
				a.window,
			)*/
			activity = widget.NewActivity()
			activity.Start()
			diag = dialog.NewCustomWithoutButtons(lang.L("Please wait..."),
				WrapWithMinSize(activity, fyne.NewSize(s.Width, s.Width/2), nil),
				a.window,
			)
		}
		diag.SetOnClosed(cancel)
		diag.Show()
		done := a.Go(func(ctx2 context.Context) error {
			go func() {
				select {
				case <-ctx.Done():
				case <-ctx2.Done():
					cancel()
				}
			}()
			finalize := fn(ctx)
			if finalize != nil {
				fyne.DoAndWait(finalize)
			}
			return nil
		})
		go func() {
			<-done
			cancel()
			fyne.DoAndWait(func() {
				diag.Dismiss()
				if activity != nil {
					activity.Stop()
				}
			})
		}()
	})
}

func (a *App) CreateCustomDialog(title, dismiss string, content fyne.CanvasObject, icon fyne.Resource) *dialog.CustomDialog {
	diag := dialog.NewCustom(title, dismiss,
		WrapWithMinSize(content, a.getDialogMinSize(), nil),
		a.window,
	)
	if icon != nil {
		diag.SetIcon(icon)
	}
	return diag
}

func (a *App) CreateBasicDialog(title, dismiss, message string, icon fyne.Resource) *dialog.CustomDialog {
	return a.CreateCustomDialog(
		title,
		dismiss,
		container.NewVBox(
			Wrap(widget.NewLabel(message), func(o fyne.CanvasObject) {
				w := o.(*widget.Label)
				w.Alignment = fyne.TextAlignCenter
				w.Wrapping = fyne.TextWrapWord
			}),
			Wrap(NewFiller(""), func(o fyne.CanvasObject) {
				w := o.(*Filler)
				w.SetMinSize(fyne.NewSize(0, theme.CurrentForWidget(w).Size(theme.SizeNameInnerPadding)))
			}),
		),
		icon,
	)
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

func (a *App) ShowYesNoDialog(title, message string, onYes func(), onNo func()) {
	a.ShowCustomConfirmDialog(title, message, lang.L("Yes"), lang.L("No"), onYes, onNo)
}

func (a *App) ShowConfirmDialog(title, message string, onConfirm func(), onCancel func()) {
	a.ShowCustomConfirmDialog(title, message, lang.L("Confirm"), lang.L("Cancel"), onConfirm, onCancel)
}

func (a *App) ShowCustomConfirmDialog(title, message, confirm, cancel string, onConfirm func(), onCancel func()) {
	a.DoWhenNoOverlay(func() {
		diag := a.CreateBasicDialog(title, lang.L("OK"), message, nil)

		confirmed := false

		confirmBtn := widget.NewButtonWithIcon(confirm, theme.ConfirmIcon(), func() {
			confirmed = true
			diag.Hide()
		})
		confirmBtn.Importance = widget.HighImportance

		cancelBtn := widget.NewButtonWithIcon(cancel, theme.CancelIcon(), func() {
			diag.Dismiss()
		})

		diag.SetButtons([]fyne.CanvasObject{
			confirmBtn,
			cancelBtn,
		})

		diag.SetOnClosed(func() {
			if confirmed {
				if onConfirm != nil {
					onConfirm()
				}
			} else {
				if onCancel != nil {
					onCancel()
				}
			}

		})

		diag.Show()
	})
}

func (a *App) ShowInputDialog(title, message string, onConfirm func(string), onCancel func()) {
	a.DoWhenNoOverlay(func() {
		entry := widget.NewEntry()

		diag := a.CreateCustomDialog(title, lang.L("OK"),
			container.NewVBox(
				widget.NewForm(
					widget.NewFormItem(message, entry),
				),
				Wrap(NewFiller(""), func(o fyne.CanvasObject) {
					w := o.(*Filler)
					w.SetMinSize(fyne.NewSize(0, theme.CurrentForWidget(w).Size(theme.SizeNameInnerPadding)))
				}),
			),
			nil,
		)

		confirmed := false

		confirmBtn := widget.NewButtonWithIcon(lang.L("Confirm"), theme.ConfirmIcon(), func() {
			confirmed = true
			diag.Hide()
		})
		confirmBtn.Importance = widget.HighImportance

		cancelBtn := widget.NewButtonWithIcon(lang.L("Cancel"), theme.CancelIcon(), func() {
			diag.Dismiss()
		})

		diag.SetButtons([]fyne.CanvasObject{
			confirmBtn,
			cancelBtn,
		})

		entry.OnSubmitted = func(string) {
			confirmBtn.OnTapped()
		}

		diag.SetOnClosed(func() {
			if confirmed {
				if onConfirm != nil {
					onConfirm(entry.Text)
				}
			} else {
				if onCancel != nil {
					onCancel()
				}
			}

		})

		diag.Show()
	})
}

func (a *App) getDialogMinSize() fyne.Size {
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
