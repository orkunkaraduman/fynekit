package fynekit

import (
	_ "unsafe"

	"fyne.io/fyne/v2"
)

func Walk(fn func(o fyne.CanvasObject) bool, objects ...fyne.CanvasObject) {
	if fn == nil {
		return
	}
	for _, object := range objects {
		if object == nil {
			continue
		}
		if !fn(object) {
			continue
		}
		switch obj := object.(type) {
		case *fyne.Container:
			for _, o := range obj.Objects {
				Walk(fn, o)
			}
		case fyne.Widget:
			r := GetRenderer(obj)
			if r == nil {
				break
			}
			for _, o := range r.Objects() {
				Walk(fn, o)
			}
		}
	}
}
