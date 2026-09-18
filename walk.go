package fynekit

import (
	_ "unsafe"

	"fyne.io/fyne/v2"
)

func Walk(object fyne.CanvasObject, fn func(fyne.CanvasObject) bool) {
	if !fn(object) {
		return
	}
	switch obj := object.(type) {
	case *fyne.Container:
		for _, o := range obj.Objects {
			Walk(o, fn)
		}
	case fyne.Widget:
		r := GetRenderer(obj)
		if r == nil {
			break
		}
		for _, o := range r.Objects() {
			Walk(o, fn)
		}
	}
}
