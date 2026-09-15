package fynekit

import (
	"fmt"
	"image"
	"image/png"
	"net/http"
)

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

func (c *OsmMapSource) GetTile(x, y, zoom int) (tile image.Image, err error) {
	u := fmt.Sprintf(c.address, zoom, x, y)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}
	return png.Decode(resp.Body)
}
