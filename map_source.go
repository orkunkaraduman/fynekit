package fynekit

import (
	"image"
)

type MapSource interface {
	TileSize() int
	GetTile(x, y, zoom int) (image.Image, error)
}
