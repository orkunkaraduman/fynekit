package fynekit

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*Map)(nil)

type Map struct {
	widget.BaseWidget

	source MapSource
	runner *Runner
}

func NewMap(source MapSource) *Map {
	m := &Map{
		source: source,
		runner: NewRunner(),
	}
	m.ExtendBaseWidget(m)
	return m
}

func (m *Map) CreateRenderer() fyne.WidgetRenderer {
	return newMapRenderer(m)
}

func (m *Map) Stop() {
	m.runner.Stop()
}

func (m *Map) worldSize(zoom int) int {
	return m.source.TileSize() * (1 << zoom)
}

var _ fyne.WidgetRenderer = (*mapRenderer)(nil)

type mapRenderer struct {
	m   *Map
	obj *canvas.Rectangle
}

func newMapRenderer(m *Map) *mapRenderer {
	r := &mapRenderer{
		m:   m,
		obj: canvas.NewRectangle(color.Black),
	}
	return r
}

func (r *mapRenderer) Destroy() {
}

func (r *mapRenderer) Layout(s fyne.Size) {
	/*r.obj.Move(fyne.NewPos(-10, -10))
	s.Width += 10
	s.Height += 10*/
	r.obj.Resize(s)
}

func (r *mapRenderer) MinSize() fyne.Size {
	return fyne.NewSize(100, 100)
}

func (r *mapRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.obj}
}

func (r *mapRenderer) Refresh() {
	r.obj.Refresh()
}
