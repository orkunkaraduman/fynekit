package fynekit

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*Map)(nil)

type Map struct {
	widget.BaseWidget
	Lat, Lon float64
	Zoom     int

	source MapSource
	cache  *mapCache
	runner *Runner
}

type MapSource interface {
	TileSize() int
	GetTile(ctx context.Context, x, y, zoom int) (image.Image, error)
}

func NewMap(source MapSource) *Map {
	m := &Map{
		source: source,
		cache:  newMapCache(source),
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
	m.Lat, m.Lon = lat, lon
	m.Refresh()
}

func (m *Map) SetZoom(zoom int) {
	m.Zoom = zoom
	m.Refresh()
}

func (m *Map) worldSize() int {
	return m.source.TileSize() * (1 << m.Zoom)
}

func (m *Map) getPosFromLatLon(lat, lon float64) fyne.Position {
	n := float64(m.worldSize())
	x := (lon + 180.0) / 360.0 * n
	latRad := lat * math.Pi / 180.0
	y := (1.0 - math.Log(math.Tan(latRad)+1.0/math.Cos(latRad))/math.Pi) / 2.0 * n
	return fyne.Position{
		X: float32(x),
		Y: float32(y),
	}
}

func (m *Map) getCenterPos() fyne.Position {
	return m.getPosFromLatLon(m.Lat, m.Lon)
}

var _ fyne.WidgetRenderer = (*mapRenderer)(nil)

type mapRenderer struct {
	m   *Map
	img *canvas.Image
}

func newMapRenderer(m *Map) *mapRenderer {
	r := &mapRenderer{
		m: m,
	}
	r.img = canvas.NewImageFromImage(r.getEmptyImage())
	r.Refresh()
	return r
}

func (r *mapRenderer) Destroy() {
}

func (r *mapRenderer) Layout(s fyne.Size) {
	r.img.Resize(s)
}

func (r *mapRenderer) MinSize() fyne.Size {
	n := float32(r.m.source.TileSize())
	return fyne.NewSize(n, n)
}

func (r *mapRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.img}
}

func (r *mapRenderer) Refresh() {
	sz := r.img.Size()
	img := image.NewRGBA(image.Rect(0, 0,
		int(math.Round(float64(sz.Width))), int(math.Round(float64(sz.Height)))))
	r.img.Image = img
	r.img.Refresh()
	centerPos := r.m.getCenterPos()
	zoom := r.m.Zoom
	n := float32(r.m.worldSize())
	for y := float32(0); y < sz.Height; y += n {
		for x := float32(0); y < sz.Width; x += n {
			pos := fyne.Position{
				X: centerPos.X + x - sz.Width/2,
				Y: centerPos.Y + y - sz.Height/2,
			}
			r.m.runner.RunAsync(func(ctx context.Context) {
				r.fill(ctx, img, pos, centerPos, zoom)
			})
		}
	}
}

func (r *mapRenderer) getEmptyImage() image.Image {
	return image.NewUniform(color.Gray{Y: 0xc0})
}

func (r *mapRenderer) fill(ctx context.Context, img *image.RGBA, pos, centerPos fyne.Position, zoom int) {
	ts := r.m.source.TileSize()
	tileX, tileY := int(pos.X/float32(ts)), int(pos.Y/float32(ts))
	if tileX < 0 || tileY < 0 {
		return
	}
	tile, err := r.m.cache.GetTile(ctx, tileX, tileY, zoom)
	if err != nil {
		// TODO: log
		_ = err
		return
	}
	draw.Draw(img, img.Bounds(), tile, image.Point{}, draw.Src)
}

type mapCache struct {
	source  MapSource
	cache   map[string]image.Image
	cacheMu sync.RWMutex
	nl      *Namedlock
}

func newMapCache(source MapSource) *mapCache {
	c := &mapCache{
		source: source,
		cache:  make(map[string]image.Image),
		nl:     NewNamedlock(),
	}
	return c
}

func (c *mapCache) GetTile(ctx context.Context, x, y, zoom int) (tile image.Image, err error) {
	key := fmt.Sprintf("%d/%d/%d", zoom, x, y)
	locker := c.nl.Locker(key)
	locker.Lock()
	defer locker.Unlock()
	c.cacheMu.RLock()
	tile = c.cache[key]
	c.cacheMu.RUnlock()
	if tile == nil {
		tile, err = c.source.GetTile(ctx, x, y, zoom)
		if err != nil {
			return
		}
	}
	c.cacheMu.Lock()
	c.cache[key] = tile
	c.cacheMu.Unlock()
	return tile, nil
}

func (c *mapCache) Invalidate() {
	c.cacheMu.Lock()
	c.cache = make(map[string]image.Image)
	c.cacheMu.Unlock()
}
