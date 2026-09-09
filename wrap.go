package fynekit

import (
	"fyne.io/fyne/v2"
)

func Wrap(object fyne.CanvasObject, fn func(o fyne.CanvasObject)) fyne.Widget {
	if fn != nil {
		fn(object)
	}
	v := new(View)
	v.ExtendBaseWidget(v)
	v.SetContent(object)
	return v
}

func WrapWithMinSize(object fyne.CanvasObject, minSize fyne.Size, fn func(o fyne.CanvasObject)) fyne.Widget {
	if fn != nil {
		fn(object)
	}
	v := new(View)
	v.ExtendBaseWidget(v)
	v.SetContent(object)
	v.SetMinSize(minSize)
	return v
}
