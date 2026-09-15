package fynekit

import (
	"context"
	"fmt"
	"image"
	"sync"
)

var _ MapSource = (*MapCache)(nil)

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

func (c *MapCache) GetTile(ctx context.Context, x, y, zoom int) (tile image.Image, err error) {
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

func (c *MapCache) Invalidate() {
	c.cacheMu.Lock()
	c.cache = make(map[string]image.Image)
	c.cacheMu.Unlock()
}
