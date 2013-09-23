package haversine

import "testing"

func TestHaverine(t *testing.T) {
	//distance b/w lyon and paris
	if got, want := 391.970466; Haversine(45.7597, 4.8422, 48.8567, 2.3508) != want {
		t.Errorf("Havesine = %f; want %f", got, want)
	}

	if got, want := 2884.632202; Haversine(36.12, -86.67, 33.94, -118.40) != want {
		t.Errorf("Havesine = %f; want %f", got, want)
	}

}
