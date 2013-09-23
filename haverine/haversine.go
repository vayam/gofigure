//Haversine to find distance b/w to locations
package haverine

import "math"

func Radians(deg float64) float64 {
	return deg * math.Pi / 180
}

func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	lat1 = Radians(lat1)
	lon1 = Radians(lon1)
	lat2 = Radians(lat2)
	lon2 = Radians(lon2)

	dLon := lon2 - lon1
	dLat := lat2 - lat1

	a := math.Pow(math.Sin(dLat/2), 2) + math.Pow(math.Sin(dLon/2), 2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Asin(math.Sqrt(a))
	km := 6367 * c
	return km
}
