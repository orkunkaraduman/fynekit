package fynekit

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*Filler)(nil)

type Filler struct {
	widget.BaseWidget
	ColorName fyne.ThemeColorName

	minSize fyne.Size
}

func NewFiller(colorName fyne.ThemeColorName) *Filler {
	f := &Filler{
		ColorName: colorName,
	}
	f.ExtendBaseWidget(f)
	return f
}

func (f *Filler) CreateRenderer() fyne.WidgetRenderer {
	return newFillerRenderer(f)
}

func (f *Filler) SetMinSize(size fyne.Size) {
	f.minSize = size
}

var _ fyne.WidgetRenderer = (*fillerRenderer)(nil)

type fillerRenderer struct {
	filler *Filler
	rect   *canvas.Rectangle
}

func newFillerRenderer(filler *Filler) *fillerRenderer {
	r := &fillerRenderer{
		filler: filler,
		rect:   canvas.NewRectangle(color.Transparent),
	}
	r.Refresh()
	return r
}

func (r *fillerRenderer) Destroy() {
}

func (r *fillerRenderer) Layout(s fyne.Size) {
	r.rect.Resize(s)
}

func (r *fillerRenderer) MinSize() fyne.Size {
	return r.filler.minSize
}

func (r *fillerRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.rect}
}

func (r *fillerRenderer) Refresh() {
	r.rect.FillColor = color.Transparent
	if r.filler.ColorName != "" {
		r.rect.FillColor = theme.ColorForWidget(r.filler.ColorName, r.filler)
	}
	r.rect.Refresh()
}
