package fynekit

import (
	"context"
	"fmt"
	"image"
	"image/color"
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
var _ fyne.Tappable = (*Map)(nil)
var _ fyne.Draggable = (*Map)(nil)

type Map struct {
	widget.BaseWidget
	Lat, Lon float64
	Zoom     int
	Scale    float32
	Overlays []MapOverlay
	Circles  []MapCircle
	Lines    []MapLine
	OnTapped func(lat, lon float64)

	source MapSource
	cache  *mapCache
	runner *runner

	dragging bool
	draggedX float32
	draggedY float32
}

func NewMap(source MapSource, opts ...MapOption) *Map {
	m := &Map{
		Scale:  1.0,
		source: source,
		cache:  newMapCache(source),
		runner: newRunner(),
	}
	m.ExtendBaseWidget(m)
	for _, opt := range opts {
		opt(m)
	}
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
	pos.X -= float32(pix) / m.Scale
	m.Lat, m.Lon = m.getLatLonFromPos(pos, m.Zoom)
	m.Refresh()
}

func (m *Map) PanNorth(pix int) {
	pos := m.getPosFromLatLon(m.Lat, m.Lon, m.Zoom)
	pos.Y -= float32(pix) / m.Scale
	m.Lat, m.Lon = m.getLatLonFromPos(pos, m.Zoom)
	m.Refresh()
}

func (m *Map) PanEast(pix int) {
	pos := m.getPosFromLatLon(m.Lat, m.Lon, m.Zoom)
	pos.X += float32(pix) / m.Scale
	m.Lat, m.Lon = m.getLatLonFromPos(pos, m.Zoom)
	m.Refresh()
}

func (m *Map) PanSouth(pix int) {
	pos := m.getPosFromLatLon(m.Lat, m.Lon, m.Zoom)
	pos.Y += float32(pix) / m.Scale
	m.Lat, m.Lon = m.getLatLonFromPos(pos, m.Zoom)
	m.Refresh()
}

func (m *Map) SetZoom(zoom int) {
	m.Zoom = zoom
	m.Refresh()
}

func (m *Map) SetScale(scale float32) {
	if scale <= 0 {
		scale = 1.0
	}
	m.Scale = scale
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

func (m *Map) Tapped(ev *fyne.PointEvent) {
	if m.OnTapped == nil {
		return
	}
	sz := m.Size()
	pos := m.getPosFromLatLon(m.Lat, m.Lon, m.Zoom)
	pos.X += ev.Position.X*m.Scale - sz.Width/2
	pos.Y += ev.Position.Y*m.Scale - sz.Height/2
	lat, lon := m.getLatLonFromPos(pos, m.Zoom)
	m.OnTapped(lat, lon)
}

func (m *Map) Dragged(ev *fyne.DragEvent) {
	ev.Dragged.DX /= m.Scale
	ev.Dragged.DY /= m.Scale
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

type MapOption func(*Map)

type MapSource interface {
	TileSize() int
	GetTile(ctx context.Context, x, y, zoom int) (image.Image, error)
	AttributionHidden() bool
	AttributionLabel() string
	AttributionURL() string
}

type MapSourceOption func(MapSource)

type MapOverlay struct {
	Lat, Lon float64
	Image    image.Image
}

type MapCircle struct {
	Lat, Lon  float64
	Color     color.Color
	Radius    int
	Thickness int
}

type MapLine struct {
	Lat1, Lon1 float64
	Lat2, Lon2 float64
	Color      color.Color
	Thickness  int
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
	size.Width /= r.m.Scale
	size.Height /= r.m.Scale
	bounds := image.Rectangle{
		Min: image.Point{},
		Max: image.Point{
			X: int(math.Round(float64(size.Width))),
			Y: int(math.Round(float64(size.Height))),
		},
	}
	drawImg := image.NewRGBA(bounds)
	if r.m.dragging {
		draw.Draw(drawImg, bounds,
			r.m.getEmptyUniformImage(), image.Point{}, draw.Over)
		draw.Draw(drawImg, bounds,
			r.drawImg, image.Point{X: int(r.m.draggedX), Y: int(r.m.draggedY)}, draw.Over)
		r.canvImg.Image = drawImg
		r.canvImg.Refresh()
		return
	}
	r.drawImg = drawImg
	draw.Draw(drawImg, bounds,
		r.m.getEmptyUniformImage(), image.Point{}, draw.Over)
	r.canvImg.Image = drawImg
	r.canvImg.Refresh()
	center := r.m.getPosFromLatLon(r.m.Lat, r.m.Lon, r.m.Zoom)
	zoom := r.m.Zoom
	var wg sync.WaitGroup
	for y := float32(0); y < size.Height+float32(tileSize); y += float32(tileSize) {
		for x := float32(0); x < size.Width+float32(tileSize); x += float32(tileSize) {
			wg.Add(1)
			r.m.runner.RunAsync(func(ctx context.Context) {
				defer wg.Done()
				r.drawTile(ctx, tileSize, size, drawImg, center, zoom, x, y)
			})
		}
	}
	overlays := make([]MapOverlay, len(r.m.Overlays))
	for i, overlay := range r.m.Overlays {
		overlays[i] = overlay
	}
	circles := make([]MapCircle, len(r.m.Circles))
	for i, circle := range r.m.Circles {
		circles[i] = circle
	}
	lines := make([]MapLine, len(r.m.Lines))
	for i, line := range r.m.Lines {
		lines[i] = line
	}
	r.m.runner.RunAsync(func(ctx context.Context) {
		wg.Wait()
		overlaysImg := r.drawOverlays(ctx, size, drawImg, center, zoom, overlays)
		circlesImg := r.drawCircles(ctx, size, drawImg, center, zoom, circles)
		linesImg := r.drawLines(ctx, size, drawImg, center, zoom, lines)
		fyne.DoAndWait(func() {
			if r.canvImg.Image == drawImg {
				draw.Draw(drawImg, drawImg.Bounds(), overlaysImg, image.Point{}, draw.Over)
				draw.Draw(drawImg, drawImg.Bounds(), circlesImg, image.Point{}, draw.Over)
				draw.Draw(drawImg, drawImg.Bounds(), linesImg, image.Point{}, draw.Over)
				r.canvImg.Refresh()
			}
		})
	})
}

func (r *mapRenderer) drawTile(ctx context.Context, tileSize int, size fyne.Size, drawImg draw.Image, center fyne.Position, zoom int, x, y float32) {
	bMin := drawImg.Bounds().Min
	bMax := drawImg.Bounds().Max
	start := fyne.Position{
		X: center.X - size.Width/2,
		Y: center.Y - size.Height/2,
	}
	pos := fyne.Position{
		X: start.X + x,
		Y: start.Y + y,
	}
	floor := fyne.Position{
		X: float32(math.Floor(float64(pos.X) / float64(tileSize))),
		Y: float32(math.Floor(float64(pos.Y) / float64(tileSize))),
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
	if bounds.Min.X < bMin.X {
		sp.X += bMin.X - bounds.Min.X
		bounds.Min.X = bMin.X
	}
	if bounds.Min.Y < bMin.Y {
		sp.Y += bMin.Y - bounds.Min.Y
		bounds.Min.Y = bMin.Y
	}
	if bounds.Max.X > bMax.X {
		bounds.Max.X = bMax.X
	}
	if bounds.Max.Y > bMax.Y {
		bounds.Max.Y = bMax.Y
	}
	if s := bounds.Size(); s.X <= 0 || s.Y <= 0 {
		return
	}
	if v := float32(r.m.worldSize(zoom) / tileSize); !(0 <= floor.X && floor.X < v) || !(0 <= floor.Y && floor.Y < v) {
		return
	}
	tile, err := r.m.cache.GetTile(ctx, int(floor.X), int(floor.Y), zoom)
	if err != nil {
		log.Printf("unable to get tile: %v", err)
		return
	}
	if ctx.Err() != nil {
		return
	}
	fyne.DoAndWait(func() {
		draw.Draw(drawImg, bounds, tile, sp, draw.Over)
	})
}

func (r *mapRenderer) drawOverlays(ctx context.Context, size fyne.Size, drawImg draw.Image, center fyne.Position, zoom int, overlays []MapOverlay) image.Image {
	img := image.NewRGBA(drawImg.Bounds())
	bMin := drawImg.Bounds().Min
	bMax := drawImg.Bounds().Max
	start := fyne.Position{
		X: center.X - size.Width/2,
		Y: center.Y - size.Height/2,
	}
	for _, overlay := range overlays {
		if ctx.Err() != nil {
			return img
		}
		overlaySz := overlay.Image.Bounds().Size()
		pos := r.m.getPosFromLatLon(overlay.Lat, overlay.Lon, zoom)
		bounds := image.Rectangle{
			Min: image.Point{
				X: int(pos.X - start.X - float32(overlaySz.X/2)),
				Y: int(pos.Y - start.Y - float32(overlaySz.Y/2)),
			},
			Max: image.Point{
				X: int(pos.X-start.X-float32(overlaySz.X/2)) + overlaySz.X,
				Y: int(pos.Y-start.Y-float32(overlaySz.Y/2)) + overlaySz.Y,
			},
		}
		sp := overlay.Image.Bounds().Min
		if bounds.Min.X < bMin.X {
			sp.X += bMin.X - bounds.Min.X
			bounds.Min.X = bMin.X
		}
		if bounds.Min.Y < bMin.Y {
			sp.Y += bMin.Y - bounds.Min.Y
			bounds.Min.Y = bMin.Y
		}
		if bounds.Max.X > bMax.X {
			bounds.Max.X = bMax.X
		}
		if bounds.Max.Y > bMax.Y {
			bounds.Max.Y = bMax.Y
		}
		if s := bounds.Size(); s.X <= 0 || s.Y <= 0 {
			continue
		}
		draw.Draw(img, bounds, overlay.Image, sp, draw.Over)
	}
	return img
}

func (r *mapRenderer) drawCircles(ctx context.Context, size fyne.Size, drawImg draw.Image, center fyne.Position, zoom int, circles []MapCircle) image.Image {
	img := image.NewRGBA(drawImg.Bounds())
	start := fyne.Position{
		X: center.X - size.Width/2,
		Y: center.Y - size.Height/2,
	}
	for _, circle := range circles {
		if ctx.Err() != nil {
			return img
		}
		pos := r.m.getPosFromLatLon(circle.Lat, circle.Lon, zoom)
		drawCircle(img,
			int(pos.X-start.X), int(pos.Y-start.Y),
			circle.Radius, circle.Thickness,
			circle.Color,
		)
	}
	return img
}

func (r *mapRenderer) drawLines(ctx context.Context, size fyne.Size, drawImg draw.Image, center fyne.Position, zoom int, lines []MapLine) image.Image {
	img := image.NewRGBA(drawImg.Bounds())
	start := fyne.Position{
		X: center.X - size.Width/2,
		Y: center.Y - size.Height/2,
	}
	for _, line := range lines {
		if ctx.Err() != nil {
			return img
		}
		pos1 := r.m.getPosFromLatLon(line.Lat1, line.Lon1, zoom)
		pos2 := r.m.getPosFromLatLon(line.Lat2, line.Lon2, zoom)
		drawLine(img,
			int(pos1.X-start.X), int(pos1.Y-start.Y),
			int(pos2.X-start.X), int(pos2.Y-start.Y),
			line.Thickness,
			line.Color,
		)
	}
	return img
}

type mapCache struct {
	source  MapSource
	cache   map[string]image.Image
	cacheMu sync.RWMutex
	nl      *namedLock
}

func newMapCache(source MapSource) *mapCache {
	c := &mapCache{
		source: source,
		nl:     newNamedLock(),
	}
	c.Invalidate()
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
	c.cache = make(map[string]image.Image, 64)
	c.cacheMu.Unlock()
}
