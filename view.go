package fynekit

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*View)(nil)

type View struct {
	widget.BaseWidget

	attachToItem
	content fyne.CanvasObject
	minSize fyne.Size
}

func (v *View) CreateRenderer() fyne.WidgetRenderer {
	return &viewRenderer{
		view: v,
	}
}

func (v *View) SetContent(content fyne.CanvasObject) {
	v.content = content
	v.Refresh()
}

func (v *View) SetMinSize(size fyne.Size) {
	v.minSize = size
	v.Refresh()
}

func (v *View) Release() {
	defer v.attachToItem.Release()
	Walk(v.content, func(o fyne.CanvasObject) bool {
		if obj, ok := o.(interface {
			fyne.Widget
			Release()
		}); ok {
			obj.Release()
			return false
		}
		if obj, ok := o.(interface {
			fyne.Widget
			Unbind()
		}); ok {
			obj.Unbind()
		}
		return true
	})
}

var _ fyne.WidgetRenderer = (*viewRenderer)(nil)

type viewRenderer struct {
	view *View
}

func (r *viewRenderer) Destroy() {
}

func (r *viewRenderer) Layout(s fyne.Size) {
	if r.view.content == nil {
		return
	}
	r.view.content.Resize(s)
}

func (r *viewRenderer) MinSize() fyne.Size {
	if r.view.content == nil {
		return r.view.minSize
	}
	return r.view.content.MinSize().Max(r.view.minSize)
}

func (r *viewRenderer) Objects() []fyne.CanvasObject {
	if r.view.content == nil {
		return nil
	}
	return []fyne.CanvasObject{r.view.content}
}

func (r *viewRenderer) Refresh() {
	if r.view.content == nil {
		return
	}
	r.view.content.Refresh()
}
