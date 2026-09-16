package fynekit

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"net/http"
)

var _ MapSource = (*OsmMapSource)(nil)

type OsmMapSource struct {
	httpClient        *http.Client
	tileSource        string
	userAgent         string
	attributionHidden bool
	attributionLabel  string
	attributionURL    string
}

func OsmMapSourceOptionWithHttpClient(httpClient *http.Client) MapSourceOption {
	return func(s MapSource) {
		s.(*OsmMapSource).httpClient = httpClient
	}
}

func OsmMapSourceOptionWithTileSource(tileSource string) MapSourceOption {
	return func(s MapSource) {
		s.(*OsmMapSource).tileSource = tileSource
	}
}

func OsmMapSourceOptionWithUserAgent(userAgent string) MapSourceOption {
	return func(s MapSource) {
		s.(*OsmMapSource).userAgent = userAgent
	}
}

func OsmMapSourceOptionWithAttribution(enable bool, label, _url string) MapSourceOption {
	return func(s MapSource) {
		ms := s.(*OsmMapSource)
		ms.attributionHidden = !enable
		ms.attributionLabel = label
		ms.attributionURL = _url
	}
}

func NewOsmMapSource(opts ...MapSourceOption) *OsmMapSource {
	s := &OsmMapSource{
		httpClient:        mapHttpClient,
		tileSource:        "https://tile.openstreetmap.org/%d/%d/%d.png",
		userAgent:         "github.com/orkunkaraduman/fynekit.Map/1.0",
		attributionHidden: false,
		attributionLabel:  "© OpenStreetMap",
		attributionURL:    "https://openstreetmap.org",
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *OsmMapSource) TileSize() int {
	return 256
}

func (s *OsmMapSource) GetTile(ctx context.Context, x, y, zoom int) (tile image.Image, err error) {
	u := fmt.Sprintf(s.tileSource, zoom, x, y)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code %d", resp.StatusCode)
	}
	tile, err = png.Decode(resp.Body)
	if err != nil {
		return
	}
	ts := s.TileSize()
	sz := tile.Bounds().Size()
	if sz.X != ts || sz.Y != ts {
		err = errors.New("tile size mismatch")
		return
	}
	return
}

func (s *OsmMapSource) AttributionHidden() bool {
	return s.attributionHidden
}

func (s *OsmMapSource) AttributionLabel() string {
	return s.attributionLabel
}

func (s *OsmMapSource) AttributionURL() string {
	return s.attributionURL
}
