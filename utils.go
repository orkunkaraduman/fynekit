package fynekit

import (
	"image"
	"image/color"
	"image/draw"
)

func drawCircle(
	img draw.Image,
	cx, cy, radius, thickness int,
	c color.Color,
) {
	bounds := img.Bounds()

	var filled bool
	if thickness <= 0 {
		thickness = 1
		filled = true
	}

	outer2 := radius * radius

	inner := radius - thickness
	if inner < 0 {
		inner = 0
	}
	inner2 := inner * inner

	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if !image.Pt(x, y).In(bounds) {
				continue
			}

			dx := x - cx
			dy := y - cy
			d2 := dx*dx + dy*dy

			if filled {
				if d2 <= outer2 {
					img.Set(x, y, c)
				}
			} else {
				if d2 <= outer2 && d2 >= inner2 {
					img.Set(x, y, c)
				}
			}
		}
	}
}

func drawLine(
	img draw.Image,
	x1, y1, x2, y2 int,
	thickness int,
	c color.Color,
) {
	bounds := img.Bounds()

	dx := absInt(x2 - x1)
	dy := absInt(y2 - y1)

	sx := -1
	if x1 < x2 {
		sx = 1
	}

	sy := -1
	if y1 < y2 {
		sy = 1
	}

	e := dx - dy
	r := thickness / 2

	for {
		for oy := -r; oy <= r; oy++ {
			for ox := -r; ox <= r; ox++ {
				x := x1 + ox
				y := y1 + oy

				if image.Pt(x, y).In(bounds) {
					img.Set(x, y, c)
				}
			}
		}

		if x1 == x2 && y1 == y2 {
			break
		}

		e2 := e * 2

		if e2 >= -dy {
			e -= dy
			x1 += sx
		}

		if e2 <= dx {
			e += dx
			y1 += sy
		}
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
