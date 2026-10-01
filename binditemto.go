package fynekit

import (
	"sync"

	"fyne.io/fyne/v2/data/binding"
)

type bindItemTo struct {
	binded []struct {
		binding.DataItem
		binding.DataListener
	}
	bindedMu sync.Mutex
}

func (a *bindItemTo) BindItemToListener(item binding.DataItem, listener binding.DataListener) {
	a.bindedMu.Lock()
	defer a.bindedMu.Unlock()
	a.binded = append(a.binded, struct {
		binding.DataItem
		binding.DataListener
	}{item, listener})
	item.AddListener(listener)
}

func (a *bindItemTo) BindItemToFunction(item binding.DataItem, fn func()) {
	a.BindItemToListener(item, binding.NewDataListener(fn))
}

func (a *bindItemTo) UnbindAll() {
	a.bindedMu.Lock()
	defer a.bindedMu.Unlock()
	for _, val := range a.binded {
		val.DataItem.RemoveListener(val.DataListener)
	}
	a.binded = nil
}
