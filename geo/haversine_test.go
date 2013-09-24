package geo

import (
	"fmt"
	"testing"
)

func check(t* testing.T, want string, lat1, lon1, lat2, lon2 float64) {
	if got := fmt.Sprintf("%.2f", Haversine(lat1, lon1, lat2, lon2)); got != want {
		t.Errorf("Haversine = got %f; want %f", got, want)
	}

}

func TestHaverine(t *testing.T) {
	//distance b/w lyon and paris
	check(t, "391.97", 45.7597, 4.8422, 48.8567, 2.3508)
	check(t, "2884.63", 36.12, -86.67, 33.94, -118.40)
}
