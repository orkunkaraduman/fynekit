package fynekit

import (
	"fmt"
	"image"
	"image/color"
	"sync"

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
	r.obj.Move(fyne.NewPos(-10, -10))
	s.Width += 10
	s.Height += 10
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

type MapSource interface {
	TileSize() int
	GetTile(x, y, zoom int) (image.Image, error)
}

type MapCache struct {
	source  MapSource
	cache   map[string]image.Image
	cacheMu sync.RWMutex
	nl      *Namedlock
}

func NewMapCache(source MapSource) *MapCache {
	c := &MapCache{
		source: source,
		cache:  make(map[string]image.Image),
		nl:     NewNamedlock(),
	}
	return c
}

func (c *MapCache) TileSize() int {
	return c.source.TileSize()
}

func (c *MapCache) GetTile(x, y, zoom int) (tile image.Image, err error) {
	key := fmt.Sprintf("%d/%d/%d", zoom, x, y)
	locker := c.nl.Locker(key)
	locker.Lock()
	defer locker.Unlock()
	c.cacheMu.RLock()
	tile = c.cache[key]
	c.cacheMu.RUnlock()
	if tile == nil {
		tile, err = c.source.GetTile(x, y, zoom)
		if err != nil {
			return
		}
	}
	c.cacheMu.Lock()
	c.cache[key] = tile
	c.cacheMu.Unlock()
	return tile, nil
}

func (c *MapCache) Invalidate() {
	c.cacheMu.Lock()
	c.cache = make(map[string]image.Image)
	c.cacheMu.Unlock()
}
