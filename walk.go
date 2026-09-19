package fynekit

import (
	_ "unsafe"

	"fyne.io/fyne/v2"
)

// Walk walks the object tree rooted at object, calling fn for [fyne.CanvasObject] in the tree, including object.
// If the function fn returns false, Walk skips all remaining objects.
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
