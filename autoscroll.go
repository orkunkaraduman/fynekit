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
