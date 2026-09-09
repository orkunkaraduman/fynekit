package fynekit

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/mobile"
)

type touchableObject struct {
	View

	deliverEvents bool
	touchDown     func(*mobile.TouchEvent)
	touchUp       func(*mobile.TouchEvent)
	touchCancel   func(*mobile.TouchEvent)
}

var _ fyne.Widget = (*tappableObject)(nil)
var _ mobile.Touchable = (*touchableObject)(nil)

func MakeTouchable(object fyne.CanvasObject, deliverEvents bool,
	touchDown func(*mobile.TouchEvent),
	touchUp func(*mobile.TouchEvent),
	touchCancel func(*mobile.TouchEvent),
) fyne.CanvasObject {
	if object == nil {
		panic("object is nil")
	}
	o := &touchableObject{
		deliverEvents: deliverEvents,
		touchDown:     touchDown,
		touchUp:       touchUp,
		touchCancel:   touchCancel,
	}
	o.ExtendBaseWidget(o)
	o.SetContent(object)
	return o
}

func (o *touchableObject) TouchDown(ev *mobile.TouchEvent) {
	if o.touchDown != nil {
		o.touchDown(ev)
	}
	if obj, _ := o.content.(mobile.Touchable); obj != nil && o.deliverEvents {
		obj.TouchDown(ev)
	}
}

func (o *touchableObject) TouchUp(ev *mobile.TouchEvent) {
	if o.touchUp != nil {
		o.touchUp(ev)
	}
	if obj, _ := o.content.(mobile.Touchable); obj != nil && o.deliverEvents {
		obj.TouchUp(ev)
	}
}

func (o *touchableObject) TouchCancel(ev *mobile.TouchEvent) {
	if o.touchCancel != nil {
		o.touchCancel(ev)
	}
	if obj, _ := o.content.(mobile.Touchable); obj != nil && o.deliverEvents {
		obj.TouchCancel(ev)
	}
}
