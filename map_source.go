package fynekit

import (
	"context"
	"image"
)

type MapSource interface {
	TileSize() int
	GetTile(ctx context.Context, x, y, zoom int) (image.Image, error)
}
