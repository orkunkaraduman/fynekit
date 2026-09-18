package fynekit

import "math"

func hav(x float64) float64 {
	return math.Pow(math.Sin(x/2), 2)
}

func ahav(x float64) float64 {
	return 2 * math.Asin(math.Sqrt(x))
}
