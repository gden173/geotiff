package geotiff

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestFileServer starts a local HTTP file server rooted at ./testdata.
// The returned server supports HTTP range requests via http.FileServer.
func newTestFileServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.FileServer(http.Dir("./testdata")))
}

func TestReadFromURL_Stats(t *testing.T) {
	server := newTestFileServer(t)
	defer server.Close()

	gtiff, err := ReadFromURL(server.URL+"/test.tif", server.Client())
	if err != nil {
		t.Fatalf("ReadFromURL failed: %s", err)
	}

	// Values extracted via: gdalinfo -stats testdata/test.tif
	wantMinimum := 46.557
	wantMaximum := 942.159
	wantMean := 234.399
	wantStdev := 106.601
	tolerance := 0.001

	got := gtiff.Stats()

	if !checkToTolerance(float64(got.Min), wantMinimum, tolerance) {
		t.Errorf("minimum value incorrect: want %f got %f", wantMinimum, got.Min)
	}
	if !checkToTolerance(float64(got.Max), wantMaximum, tolerance) {
		t.Errorf("maximum value incorrect: want %f got %f", wantMaximum, got.Max)
	}
	if !checkToTolerance(float64(got.Mean), wantMean, 0.01) {
		t.Errorf("mean value incorrect: want %f got %f", wantMean, got.Mean)
	}
	if !checkToTolerance(float64(got.StdDev), wantStdev, 0.01) {
		t.Errorf("standard deviation incorrect: want %f got %f", wantStdev, got.StdDev)
	}
}

func TestReadFromURL_AtCoord(t *testing.T) {
	server := newTestFileServer(t)
	defer server.Close()

	gtiff, err := ReadFromURL(server.URL+"/test.tif", server.Client())
	if err != nil {
		t.Fatalf("ReadFromURL failed: %s", err)
	}

	// Commands used to extract values:
	// gdallocationinfo -wgs84 testdata/test.tif lon lat
	testLocations := []struct {
		lon           float64
		lat           float64
		expectedValue float64
	}{
		{lon: 138, lat: -23, expectedValue: 261.246856689453},
		{lon: 139.5, lat: -24, expectedValue: 116.877799987793},
	}

	for _, tl := range testLocations {
		val, err := gtiff.AtCoord(tl.lon, tl.lat, false)
		if err != nil {
			t.Errorf("AtCoord(%f, %f) error: %s", tl.lon, tl.lat, err)
			continue
		}
		if !checkToTolerance(float64(val), tl.expectedValue, 0.0001) {
			t.Errorf("AtCoord(%f, %f) = %f, want %f", tl.lon, tl.lat, val, tl.expectedValue)
		}
	}
}

func TestReadFromURL_NotFound(t *testing.T) {
	server := newTestFileServer(t)
	defer server.Close()

	_, err := ReadFromURL(server.URL+"/nonexistent.tif", server.Client())
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestReadFromURL_InvalidURL(t *testing.T) {
	_, err := ReadFromURL("http://localhost:0/nonexistent.tif", nil)
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}
