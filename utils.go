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
	filled bool,
) {
	bounds := img.Bounds()

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
