package fynekit

import (
	"fyne.io/fyne/v2"
)

type tappableObject struct {
	View

	deliverEvents bool
	tapped        func(*fyne.PointEvent)
}

var _ fyne.Widget = (*tappableObject)(nil)
var _ fyne.Tappable = (*tappableObject)(nil)

func MakeTappable(object fyne.CanvasObject, deliverEvents bool, tapped func(*fyne.PointEvent)) fyne.CanvasObject {
	if object == nil {
		panic("object is nil")
	}
	o := &tappableObject{
		deliverEvents: deliverEvents,
		tapped:        tapped,
	}
	o.ExtendBaseWidget(o)
	o.SetContent(object)
	return o
}

func (o *tappableObject) Tapped(ev *fyne.PointEvent) {
	if o.tapped != nil {
		o.tapped(ev)
	}
	if obj, _ := o.content.(fyne.Tappable); obj != nil && o.deliverEvents {
		obj.Tapped(ev)
	}
}
