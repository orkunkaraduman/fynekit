package fynekit

import (
	"math"
)

const earthRadius = 6371009

type MapCoordinate struct {
	Lat float64
	Lon float64
}

func (c MapCoordinate) LatRad() float64 {
	return c.Lat * 2 * math.Pi / 360
}

func (c MapCoordinate) LonRad() float64 {
	return c.Lon * 2 * math.Pi / 360
}

func (c MapCoordinate) Distance(d MapCoordinate) float64 {
	angle := hav(d.LatRad()-c.LatRad()) +
		math.Cos(c.LatRad())*math.Cos(d.LatRad())*hav(d.LonRad()-c.LonRad())
	return earthRadius * ahav(angle)
}

func (c MapCoordinate) Bearing(d MapCoordinate) float64 {
	radLonDiff := d.LonRad() - c.LonRad()
	radBearing := math.Atan2(math.Sin(radLonDiff)*math.Cos(d.LatRad()),
		math.Cos(c.LatRad())*math.Sin(d.LatRad())-
			math.Sin(c.LatRad())*math.Cos(d.LatRad())*math.Cos(radLonDiff))
	return math.Mod(360*radBearing/(2*math.Pi)+360, 360)
}

func (c MapCoordinate) Offset(distance, bearing float64) MapCoordinate {
	radBearing := bearing * 2 * math.Pi / 360
	radLat := math.Asin(math.Sin(c.LatRad())*math.Cos(distance/earthRadius) +
		math.Cos(c.LatRad())*math.Sin(distance/earthRadius)*math.Cos(radBearing))
	radLon := c.LonRad() +
		math.Atan2(math.Sin(radBearing)*math.Sin(distance/earthRadius)*math.Cos(c.LatRad()),
			math.Cos(distance/earthRadius)-math.Sin(c.LatRad())*math.Sin(radLat))
	return MapCoordinate{Lat: 360 * radLat / (2 * math.Pi), Lon: 360 * radLon / (2 * math.Pi)}
}
