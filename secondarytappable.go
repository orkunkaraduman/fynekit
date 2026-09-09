package fynekit

import (
	"fyne.io/fyne/v2"
)

type secondaryTappableObject struct {
	View

	deliverEvents   bool
	tappedSecondary func(*fyne.PointEvent)
}

var _ fyne.Widget = (*secondaryTappableObject)(nil)
var _ fyne.SecondaryTappable = (*secondaryTappableObject)(nil)

func MakeSecondaryTappable(object fyne.CanvasObject, deliverEvents bool, tappedSecondary func(*fyne.PointEvent)) fyne.CanvasObject {
	if object == nil {
		panic("object is nil")
	}
	o := &secondaryTappableObject{
		deliverEvents:   deliverEvents,
		tappedSecondary: tappedSecondary,
	}
	o.ExtendBaseWidget(o)
	o.SetContent(object)
	return o
}

func (o *secondaryTappableObject) TappedSecondary(ev *fyne.PointEvent) {
	if o.tappedSecondary != nil {
		o.tappedSecondary(ev)
	}
	if obj, _ := o.content.(fyne.SecondaryTappable); obj != nil && o.deliverEvents {
		obj.TappedSecondary(ev)
	}
}
