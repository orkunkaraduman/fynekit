package fynekit

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	mapHttpClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout: 3 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:    3 * time.Second,
			MaxIdleConns:           10,
			IdleConnTimeout:        65 * time.Second,
			ResponseHeaderTimeout:  5 * time.Second,
			ExpectContinueTimeout:  1 * time.Second,
			MaxResponseHeaderBytes: 1 << 20,
		},
	}
)

var _ fyne.Widget = (*Map)(nil)

type Map struct {
	widget.BaseWidget
	Lat, Lon float64
	Zoom     int

	source MapSource
	cache  *mapCache
	runner *Runner

	dragging bool
	draggedX float32
	draggedY float32
}

type MapOption func(*Map)

type MapSource interface {
	TileSize() int
	GetTile(ctx context.Context, x, y, zoom int) (image.Image, error)
	AttributionHidden() bool
	AttributionLabel() string
	AttributionURL() string
}

type MapSourceOption func(MapSource)

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

func (m *Map) InvalidateCache() {
	m.cache.Invalidate()
	m.Refresh()
}

func (m *Map) PanToLatLon(lat, lon float64) {
	m.Lat, m.Lon = lat, lon
	m.Refresh()
}

func (m *Map) PanWest(pix int) {
	pos := m.getPosFromLatLon(m.Lat, m.Lon, m.Zoom)
	pos.X -= float32(pix)
	m.Lat, m.Lon = m.getLatLonFromPos(pos, m.Zoom)
	m.Refresh()
}

func (m *Map) PanNorth(pix int) {
	pos := m.getPosFromLatLon(m.Lat, m.Lon, m.Zoom)
	pos.Y -= float32(pix)
	m.Lat, m.Lon = m.getLatLonFromPos(pos, m.Zoom)
	m.Refresh()
}

func (m *Map) PanEast(pix int) {
	pos := m.getPosFromLatLon(m.Lat, m.Lon, m.Zoom)
	pos.X += float32(pix)
	m.Lat, m.Lon = m.getLatLonFromPos(pos, m.Zoom)
	m.Refresh()
}

func (m *Map) PanSouth(pix int) {
	pos := m.getPosFromLatLon(m.Lat, m.Lon, m.Zoom)
	pos.Y += float32(pix)
	m.Lat, m.Lon = m.getLatLonFromPos(pos, m.Zoom)
	m.Refresh()
}

func (m *Map) SetZoom(zoom int) {
	m.Zoom = zoom
	m.Refresh()
}

func (m *Map) ZoomIn() {
	m.Zoom++
	m.Refresh()
}

func (m *Map) ZoomOut() {
	if m.Zoom <= 0 {
		return
	}
	m.Zoom--
	m.Refresh()
}

func (m *Map) Dragged(ev *fyne.DragEvent) {
	pos := m.getPosFromLatLon(m.Lat, m.Lon, m.Zoom)
	pos.X -= ev.Dragged.DX
	pos.Y -= ev.Dragged.DY
	m.Lat, m.Lon = m.getLatLonFromPos(pos, m.Zoom)
	m.dragging = true
	m.draggedX -= ev.Dragged.DX
	m.draggedY -= ev.Dragged.DY
	m.Refresh()
}

func (m *Map) DragEnd() {
	m.dragging = false
	m.draggedX = 0
	m.draggedY = 0
	m.Refresh()
}

func (m *Map) getEmptyRGBAImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, theme.ColorForWidget(theme.ColorNameDisabled, m))
	return img
}

func (m *Map) getEmptyUniformImage() *image.Uniform {
	img := image.NewUniform(theme.ColorForWidget(theme.ColorNameDisabled, m))
	return img
}

func (m *Map) worldSize(zoom int) int {
	return m.source.TileSize() * (1 << zoom)
}

func (m *Map) getPosFromLatLon(lat, lon float64, zoom int) fyne.Position {
	n := float64(m.worldSize(zoom))
	x := (lon + 180.0) / 360.0 * n
	latRad := lat * math.Pi / 180.0
	y := (1.0 - math.Log(math.Tan(latRad)+1.0/math.Cos(latRad))/math.Pi) / 2.0 * n
	return fyne.Position{
		X: float32(x),
		Y: float32(y),
	}
}

func (m *Map) getLatLonFromPos(pos fyne.Position, zoom int) (lat float64, lon float64) {
	n := float64(m.worldSize(zoom))
	lon = float64(pos.X)/n*360.0 - 180.0
	latRad := math.Atan(math.Sinh(math.Pi * (1.0 - 2.0*float64(pos.Y)/n)))
	lat = latRad * 180.0 / math.Pi
	return
}

var _ fyne.WidgetRenderer = (*mapRenderer)(nil)

type mapRenderer struct {
	m         *Map
	canvImg   *canvas.Image
	drawImg   *image.RGBA
	copyright *fyne.Container
}

func newMapRenderer(m *Map) *mapRenderer {
	r := &mapRenderer{
		m:       m,
		canvImg: canvas.NewImageFromImage(m.getEmptyRGBAImage()),
		drawImg: m.getEmptyRGBAImage(),
	}

	u, _ := url.Parse(m.source.AttributionURL())
	link := widget.NewHyperlink(m.source.AttributionLabel(), u)
	link.Alignment = fyne.TextAlignTrailing
	link.SizeName = theme.SizeNameCaptionText
	link.TextStyle.Bold = true
	r.copyright = container.NewHBox(layout.NewSpacer(), link)

	r.Refresh()
	return r
}

func (r *mapRenderer) Destroy() {
}

func (r *mapRenderer) Layout(s fyne.Size) {
	r.canvImg.Resize(s)
	ms := r.copyright.MinSize()
	r.copyright.Resize(fyne.NewSize(s.Width, ms.Height))
	r.copyright.Move(fyne.NewPos(0, s.Height-ms.Height-theme.Padding()))
	r.Refresh()
}

func (r *mapRenderer) MinSize() fyne.Size {
	tileSize := float32(r.m.source.TileSize())
	return fyne.NewSize(tileSize, tileSize)
}

func (r *mapRenderer) Objects() []fyne.CanvasObject {
	objs := []fyne.CanvasObject{r.canvImg}
	if !r.m.source.AttributionHidden() {
		objs = append(objs, r.copyright)
	}
	return objs
}

func (r *mapRenderer) Refresh() {
	tileSize := r.m.source.TileSize()
	size := r.canvImg.Size()
	bounds := image.Rect(0, 0,
		int(math.Round(float64(size.Width))), int(math.Round(float64(size.Height))))
	if r.m.dragging {
		img := image.NewRGBA(bounds)
		draw.Draw(img, bounds,
			r.m.getEmptyUniformImage(), image.Point{}, draw.Over)
		draw.Draw(img, bounds,
			r.drawImg, image.Point{X: int(r.m.draggedX), Y: int(r.m.draggedY)}, draw.Over)
		r.canvImg.Image = img
		r.canvImg.Refresh()
		return
	}
	r.drawImg = image.NewRGBA(bounds)
	draw.Draw(r.drawImg, bounds,
		r.m.getEmptyUniformImage(), image.Point{}, draw.Over)
	r.canvImg.Image = r.drawImg
	r.canvImg.Refresh()
	center := r.m.getPosFromLatLon(r.m.Lat, r.m.Lon, r.m.Zoom)
	zoom := r.m.Zoom
	for y := float32(0); y < size.Height+float32(tileSize); y += float32(tileSize) {
		for x := float32(0); x < size.Width+float32(tileSize); x += float32(tileSize) {
			r.m.runner.RunAsync(func(ctx context.Context) {
				r.fill(ctx, tileSize, size, r.drawImg, center, zoom, x, y)
			})
		}
	}
}

func (r *mapRenderer) fill(ctx context.Context, tileSize int, size fyne.Size, drawImg draw.Image, center fyne.Position, zoom int, x, y float32) {
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
	bounds := image.Rectangle{
		Min: image.Point{X: int(trunc.X - start.X), Y: int(trunc.Y - start.Y)},
		Max: image.Point{X: int(trunc.X-start.X) + tileSize, Y: int(trunc.Y-start.Y) + tileSize},
	}
	var sp image.Point
	if bounds.Min.X < 0 {
		sp.X = -bounds.Min.X
		bounds.Min.X = 0
	}
	if bounds.Min.Y < 0 {
		sp.Y = -bounds.Min.Y
		bounds.Min.Y = 0
	}
	if v := int(math.Round(float64(size.Width))); bounds.Max.X > v {
		bounds.Max.X = v
	}
	if v := int(math.Round(float64(size.Height))); bounds.Max.Y > v {
		bounds.Max.Y = v
	}
	if s := bounds.Size(); s.X <= 0 || s.Y <= 0 {
		return
	}
	if v := float32(r.m.worldSize(zoom) / tileSize); !(0 <= floor.X && floor.X < v) || !(0 <= floor.Y && floor.Y < v) {
		return
	}
	tile, err := r.m.cache.GetTile(ctx, int(floor.X), int(floor.Y), zoom)
	if err != nil {
		// TODO: log
		log.Println(err)
		return
	}
	fyne.DoAndWait(func() {
		draw.Draw(drawImg, bounds, tile, sp, draw.Over)
		r.canvImg.Refresh()
	})
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
