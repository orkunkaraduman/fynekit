package fynekit

import (
	"sync"

	"fyne.io/fyne/v2/data/binding"
)

type attachToItem struct {
	attached []struct {
		binding.DataItem
		binding.DataListener
	}
	attachedMu sync.Mutex
}

func (a *attachToItem) AttachListenerToItem(item binding.DataItem, listener binding.DataListener) {
	a.attachedMu.Lock()
	defer a.attachedMu.Unlock()
	a.attached = append(a.attached, struct {
		binding.DataItem
		binding.DataListener
	}{item, listener})
	item.AddListener(listener)
}

func (a *attachToItem) AttachFunctionToItem(item binding.DataItem, fn func()) {
	a.AttachListenerToItem(item, binding.NewDataListener(fn))
}

func (a *attachToItem) Release() {
	a.attachedMu.Lock()
	defer a.attachedMu.Unlock()
	for _, val := range a.attached {
		val.DataItem.RemoveListener(val.DataListener)
	}
	a.attached = nil
}
