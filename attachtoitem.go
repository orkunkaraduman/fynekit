package fynekit

import (
	"sync"

	"fyne.io/fyne/v2/data/binding"
)

type AttachToItem struct {
	attached []struct {
		binding.DataItem
		binding.DataListener
	}
	attachedMu sync.Mutex
}

func (a *AttachToItem) AttachListenerToItem(item binding.DataItem, listener binding.DataListener) {
	a.attachedMu.Lock()
	defer a.attachedMu.Unlock()
	a.attached = append(a.attached, struct {
		binding.DataItem
		binding.DataListener
	}{item, listener})
	item.AddListener(listener)
}

func (a *AttachToItem) AttachFunctionToItem(item binding.DataItem, fn func()) {
	a.AttachListenerToItem(item, binding.NewDataListener(fn))
}

func (a *AttachToItem) Release() {
	a.attachedMu.Lock()
	defer a.attachedMu.Unlock()
	for _, val := range a.attached {
		val.DataItem.RemoveListener(val.DataListener)
	}
	a.attached = nil
}
