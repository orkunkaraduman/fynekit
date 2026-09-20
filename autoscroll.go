package fynekit

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type AutoScrollEntry struct {
	widget.Entry
	Scroll *container.Scroll
}

func NewAutoScrollEntry(scroll *container.Scroll) *AutoScrollEntry {
	s := &AutoScrollEntry{
		Entry:  widget.Entry{Wrapping: fyne.TextWrap(fyne.TextTruncateClip)},
		Scroll: scroll,
	}
	s.ExtendBaseWidget(s)
	return s
}

func NewAutoScrollEntryWithData(scroll *container.Scroll, data binding.String) *AutoScrollEntry {
	entry := NewAutoScrollEntry(scroll)
	entry.Bind(data)

	return entry
}

func (s *AutoScrollEntry) FocusGained() {
	s.Entry.FocusGained()
	go func() {
		<-time.After(time.Second / 2)
		fyne.Do(func() {
			setAutoScroll(s, s.Scroll)
		})
	}()
}

func setAutoScroll(wid fyne.Widget, scroll *container.Scroll) {
	pos := wid.Position()
	pos.X = scroll.Offset.X
	pos.Y -= theme.CurrentForWidget(wid).Size(theme.SizeNameInnerPadding)
	if pos.Y < 0 {
		pos.Y = 0
	}
	scroll.ScrollToBottom()
	if scroll.Offset.Y > pos.Y {
		scroll.ScrollToOffset(pos)
	}
}

/*type AutoScrollBase interface {
	fyne.Disableable
	fyne.Draggable
	fyne.Focusable
	fyne.Tappable
	fyne.Widget
	desktop.Mouseable
	desktop.Keyable
	mobile.Keyboardable
	mobile.Touchable
	fyne.Tabbable
	ExtendBaseWidget(fyne.Widget)
}

type AutoScroll struct {
	widget.BaseWidget
	Scroll *container.Scroll

	wid AutoScrollBase
}

func NewAutoScroll(wid AutoScrollBase, scroll *container.Scroll) *AutoScroll {
	s := &AutoScroll{
		Scroll: scroll,
		wid:    wid,
	}
	s.ExtendBaseWidget(s)
	return s
}

func (s *AutoScroll) Enable() {
	s.wid.Enable()
}

func (s *AutoScroll) Disable() {
	s.wid.Disable()
}

func (s *AutoScroll) Disabled() bool {
	return s.wid.Disabled()
}

func (s *AutoScroll) Dragged(ev *fyne.DragEvent) {
	s.wid.Dragged(ev)
}

func (s *AutoScroll) DragEnd() {
	s.wid.DragEnd()
}

func (s *AutoScroll) FocusGained() {
	s.wid.FocusGained()
	go func() {
		<-time.After(time.Second / 4)
		fyne.Do(func() {
			pos := s.Position()
			s.Scroll.ScrollToTop()
			s.Scroll.ScrollToBottom()
			if s.Scroll.Offset.X > pos.X || s.Scroll.Offset.Y > pos.Y {
				s.Scroll.ScrollToOffset(pos)
			}
		})
	}()
}

func (s *AutoScroll) FocusLost() {
	s.wid.FocusLost()
}

func (s *AutoScroll) TypedRune(r rune) {
	s.wid.TypedRune(r)
}

func (s *AutoScroll) TypedKey(ev *fyne.KeyEvent) {
	s.wid.TypedKey(ev)
}

func (s *AutoScroll) Tapped(ev *fyne.PointEvent) {
	c := fyne.CurrentApp().Driver().CanvasForObject(s)
	if c != nil {
		c.Focus(s)
	}
	s.wid.Tapped(ev)
}

func (s *AutoScroll) CreateRenderer() fyne.WidgetRenderer {
	return s.wid.CreateRenderer()
	//return widget.NewSimpleRenderer(s.wid)
}

func (s *AutoScroll) MouseDown(ev *desktop.MouseEvent) {
	s.wid.MouseDown(ev)
}

func (s *AutoScroll) MouseUp(ev *desktop.MouseEvent) {
	s.wid.MouseUp(ev)
}

func (s *AutoScroll) KeyDown(ev *fyne.KeyEvent) {
	s.wid.KeyDown(ev)
}

func (s *AutoScroll) KeyUp(ev *fyne.KeyEvent) {
	s.wid.KeyUp(ev)
}

func (s *AutoScroll) Keyboard() mobile.KeyboardType {
	return s.wid.Keyboard()
}

func (s *AutoScroll) TouchDown(ev *mobile.TouchEvent) {
	s.wid.TouchDown(ev)
}

func (s *AutoScroll) TouchUp(ev *mobile.TouchEvent) {
	s.wid.TouchUp(ev)
}

func (s *AutoScroll) TouchCancel(ev *mobile.TouchEvent) {
	s.wid.TouchCancel(ev)
}

func (s *AutoScroll) AcceptsTab() bool {
	return s.wid.AcceptsTab()
}*/
