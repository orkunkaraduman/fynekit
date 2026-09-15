package fynekit

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*Map)(nil)

type Map struct {
	widget.BaseWidget

	source MapSource
	runner *Runner

	// center lat, lon
	lat, lon float64
	// zoom
	zoom int
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

func (m *Map) PanToLatLon(lat, lon float64) {

}

func (m *Map) worldSize(zoom int) int {
	return m.source.TileSize() * (1 << zoom)
}

func (m *Map) getPixFromLatLon(lat, lon float64) (x, y float64) {
	n := float64(m.worldSize(m.zoom))
	x = (lon + 180.0) / 360.0 * n
	latRad := lat * math.Pi / 180.0
	y = (1.0 - math.Log(math.Tan(latRad)+1.0/math.Cos(latRad))/math.Pi) / 2.0 * n
	return
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
