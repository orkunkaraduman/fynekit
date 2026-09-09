package fynekit

import (
	"sync"

	"fyne.io/fyne/v2/data/binding"
)

var (
	R  = Walk
	W  = Wrap
	WM = WrapWithMinSize
)

type attachToItemBase struct {
	attached []struct {
		binding.DataItem
		binding.DataListener
	}
	attachedMu sync.Mutex
}

func (b *attachToItemBase) AttachListenerToItem(item binding.DataItem, listener binding.DataListener) {
	item.AddListener(listener)
	b.attachedMu.Lock()
	defer b.attachedMu.Unlock()
	b.attached = append(b.attached, struct {
		binding.DataItem
		binding.DataListener
	}{item, listener})
}

func (b *attachToItemBase) AttachFunctionToItem(item binding.DataItem, fn func()) {
	b.AttachListenerToItem(item, binding.NewDataListener(fn))
}

func (b *attachToItemBase) Release() {
	b.attachedMu.Lock()
	defer b.attachedMu.Unlock()
	for _, a := range b.attached {
		a.DataItem.RemoveListener(a.DataListener)
	}
	b.attached = nil
}
