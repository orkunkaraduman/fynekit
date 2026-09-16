package fynekit

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
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
	tileSize := float32(r.m.source.TileSize())
	return fyne.NewSize(tileSize, tileSize)
}

func (r *mapRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.img}
}

func (r *mapRenderer) Refresh() {
	tileSize := r.m.source.TileSize()
	size := r.img.Size()
	img := image.NewRGBA(image.Rect(0, 0,
		int(math.Round(float64(size.Width))), int(math.Round(float64(size.Height)))))
	r.img.Image = img
	r.img.Refresh()
	center := r.m.getCenterPos()
	zoom := r.m.Zoom
	for y := float32(0); y < size.Height; y += float32(tileSize) {
		for x := float32(0); y < size.Width; x += float32(tileSize) {
			r.m.runner.RunAsync(func(ctx context.Context) {
				r.fill(ctx, tileSize, size, img, center, zoom, x, y)
			})
		}
	}
}

func (r *mapRenderer) getEmptyImage() image.Image {
	return image.NewUniform(color.Gray{Y: 0xc0})
}

func (r *mapRenderer) fill(ctx context.Context, tileSize int, size fyne.Size, img *image.RGBA, center fyne.Position, zoom int, x, y float32) {
	start := fyne.Position{
		X: center.X - size.Width/2,
		Y: center.Y - size.Height/2,
	}
	current := fyne.Position{
		X: start.X + x,
		Y: start.Y + y,
	}
	floor := fyne.Position{
		X: float32(math.Floor(float64(current.X) / float64(tileSize))),
		Y: float32(math.Floor(float64(current.Y) / float64(tileSize))),
	}
	trunc := fyne.Position{
		X: floor.X * float32(tileSize),
		Y: floor.Y * float32(tileSize),
	}
	rem := fyne.Position{
		X: current.X - trunc.X,
		Y: current.Y - trunc.Y,
	}
	bounds := image.Rectangle{
		Min: image.Point{X: int(trunc.X - start.X), Y: int(trunc.Y - start.Y)},
		Max: image.Point{X: int(trunc.X-start.X) + tileSize, Y: int(trunc.Y-start.Y) + tileSize},
	}
	if bounds.Min.X < 0 {
		bounds.Min.X = 0
	}
	if bounds.Min.Y < 0 {
		bounds.Min.Y = 0
	}
	if bounds.Max.X > tileSize {
		bounds.Max.X = tileSize
	}
	if bounds.Max.Y > tileSize {
		bounds.Max.Y = tileSize
	}
	if s := bounds.Size(); s.X <= 0 || s.Y <= 0 {
		return
	}
	tile, err := r.m.cache.GetTile(ctx, int(floor.X), int(floor.Y), zoom)
	if err != nil {
		// TODO: log
		log.Println(err)
		return
	}
	draw.Draw(img, img.Bounds(), tile, image.Point{X: int(rem.X), Y: int(rem.Y)}, draw.Src)
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
