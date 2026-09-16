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
	address string
}

func NewOsmMapSource() *OsmMapSource {
	return NewOsmMapSourceWithAddress("")
}

func NewOsmMapSourceWithAddress(address string) *OsmMapSource {
	if address == "" {
		address = "https://tile.openstreetmap.org/%d/%d/%d.png"
	}
	s := &OsmMapSource{
		address: address,
	}
	return s
}

func (s *OsmMapSource) TileSize() int {
	return 256
}

func (s *OsmMapSource) GetTile(ctx context.Context, x, y, zoom int) (tile image.Image, err error) {
	u := fmt.Sprintf(s.address, zoom, x, y)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "github.com/orkunkaraduman/fynekit.Map/1.0")
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
