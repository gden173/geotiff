package geotiff

import (
	"encoding/binary"
	"math"
	"os"
	"testing"
)

var testfile = "./testdata/test.tif"

func Test_ReadHeader_Happy(t *testing.T) {
	wantHeader := head{
		byteOrder:     binary.LittleEndian,
		tIFIdentifier: 42,
		iFDByteOffset: 8, // TODO: find way to not hardcode this
	}

	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	got, err := readHeader(r)
	if err != nil {
		t.Fail()
	}

	if got.byteOrder != wantHeader.byteOrder {
		t.Error("incorrect byte order:")
	}

	if got.tIFIdentifier != wantHeader.tIFIdentifier {
		t.Error("incorrect TIFID")
	}

	if got.iFDByteOffset != wantHeader.iFDByteOffset {
		t.Error("incorrect Byte Offset")
	}

	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
}

func Test_ReadTags_Happy(t *testing.T) {
	tagsTests := []struct {
		name         string
		testfile     string
		expectedTags map[Tag][]uint64
	}{
		{
			name:     "test",
			testfile: "./testdata/test.tif",
			expectedTags: map[Tag][]uint64{
				Compression:               {1},
				SampleFormat:              {3},
				ImageWidth:                {180},
				ImageLength:               {191},
				PhotometricInterpretation: {1},
				TileWidth:                 {128},
				TileLength:                {128},
				TileByteCounts:            {65536, 65536, 65536, 65536},
				TileOffsets:               {416, 65952, 131488, 197024},
				BitsPerSample:             {32},
				SamplesPerPixel:           {1},
				PlanarConfiguration:       {1},
			},
		},
		{
			name:     "australia",
			testfile: "./testdata/WCSServer.tif",
			expectedTags: map[Tag][]uint64{
				Compression:               {1},
				TileWidth:                 {128},
				SampleFormat:              {3},
				ImageWidth:                {1437},
				ImageLength:               {1188},
				PhotometricInterpretation: {1},
				TileLength:                {128},
				TileByteCounts: {
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
					65536, 65536, 65536, 65536, 65536, 65536, 65536, 65536,
				},
				TileOffsets: {
					1344, 66880, 132416, 197952, 263488, 329024, 394560,
					460096, 525632, 591168, 656704, 722240, 787776, 853312,
					918848, 984384, 1049920, 1115456, 1180992, 1246528,
					1312064, 1377600, 1443136, 1508672, 1574208, 1639744,
					1705280, 1770816, 1836352, 1901888, 1967424, 2032960,
					2098496, 2164032, 2229568, 2295104, 2360640, 2426176,
					2491712, 2557248, 2622784, 2688320, 2753856, 2819392,
					2884928, 2950464, 3016000, 3081536, 3147072, 3212608,
					3278144, 3343680, 3409216, 3474752, 3540288, 3605824,
					3671360, 3736896, 3802432, 3867968, 3933504, 3999040,
					4064576, 4130112, 4195648, 4261184, 4326720, 4392256,
					4457792, 4523328, 4588864, 4654400, 4719936, 4785472,
					4851008, 4916544, 4982080, 5047616, 5113152, 5178688,
					5244224, 5309760, 5375296, 5440832, 5506368, 5571904,
					5637440, 5702976, 5768512, 5834048, 5899584, 5965120,
					6030656, 6096192, 6161728, 6227264, 6292800, 6358336,
					6423872, 6489408, 6554944, 6620480, 6686016, 6751552,
					6817088, 6882624, 6948160, 7013696, 7079232, 7144768,
					7210304, 7275840, 7341376, 7406912, 7472448, 7537984,
					7603520, 7669056, 7734592, 7800128,
				},
				BitsPerSample:       {32},
				SamplesPerPixel:     {1},
				PlanarConfiguration: {1},
			},
		},
	}

	for _, tt := range tagsTests {

		r, err := os.Open(tt.testfile)
		if err != nil {
			t.Fatal(err)
		}

		gotTags, _, err := readTags(r)
		if err != nil {
			t.Fatalf("error %s", err)
		}

		for k, v := range tt.expectedTags {
			gotV, ok := gotTags[k]
			if !ok {
				t.Errorf("unrecognized key %s", k)
			}

			ftype, val := gotV.value()
			ev := make([]uint64, 0)
			switch ftype {
			case SHORT:
				e := val[0].([]uint16)
				for _, evv := range e {
					ev = append(ev, uint64(evv))
				}
			case LONG:
				e := val[0].([]uint32)
				for _, evv := range e {
					ev = append(ev, uint64(evv))
				}
			}

			if len(ev) != len(v) {
				t.Errorf("incorrect number of tags returned: got %v - want %v", ev, v)
			}
			for i, vv := range v {
				if vv != ev[i] {
					t.Errorf("invalid tag value for tag %s: want %d got %d", k, vv, ev[i])
				}
			}
		}

		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func checkToTolerance(x float64, y float64, tolerance float64) bool {
	return math.Abs(x-y) <= tolerance
}

func Test_ReadData_Happy(t *testing.T) {
	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}

	tt, h, err := readTags(r)
	if err != nil {
		t.Fatal(err)
	}
	data, err := readData(r, tt, h)
	if err != nil {
		t.Fatal(err)
	}

	// Check the statistics for the file
	// Extracted via gdalinfo -stats go/runoffarea/internal/geotiff/testdata/test.tif
	wantImageSize := 180 * 191
	wantMinimum := 46.557
	wantMaximum := 942.159
	wantMean := 234.399

	var min float32 = math.MaxFloat32
	var max float32 = math.SmallestNonzeroFloat32
	var mean float32
	var nonzero float32

	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data[i]); j++ {
			d := data[i][j]
			if d != 0 {
				nonzero++
				mean += d
				if d < min {
					min = d
				}
				if d > max {
					max = d
				}
			}
		}
	}
	mean /= nonzero
	tolerance := 0.001

	if !checkToTolerance(float64(nonzero), float64(wantImageSize), tolerance) {
		t.Errorf("number of nonzero pixels incorrect: want %f got %f", float32(wantImageSize), nonzero)
	}

	if !checkToTolerance(float64(min), wantMinimum, tolerance) {
		t.Errorf("minimum value incorrect: want %f got %f", float32(wantMinimum), min)
	}

	if !checkToTolerance(float64(max), wantMaximum, tolerance) {
		t.Errorf("maximum value incorrect: want %f got %f", float32(wantMaximum), max)
	}

	// need to be a bit nicer here, as the floating point error will be larger
	if !checkToTolerance(float64(mean), wantMean, 0.01) {
		t.Errorf("mean value incorrect: want %f got %f", float32(wantMean), mean)
	}

	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
}

func Test_LocationData_Happy(t *testing.T) {
	t.Run("small testfile", func(t *testing.T) {
		r, err := os.Open(testfile)
		if err != nil {
			t.Fatal(err)
		}
		geo, err := Read(r)
		if err != nil {
			t.Fatal(err)
		}

		// Commands used to extract values
		// gdallocationinfo testdata/test.tif x y
		testLocations := []struct {
			x             int
			y             int
			expectedValue float64
		}{
			{
				x:             0,
				y:             0,
				expectedValue: 284.911102294922,
			},
			{
				x:             130,
				y:             130,
				expectedValue: 114.861305236816,
			},
			{
				x:             0,
				y:             130,
				expectedValue: 492.600250244141,
			},
			{
				x:             130,
				y:             0,
				expectedValue: 328.951202392578,
			},
			{
				x:             150,
				y:             150,
				expectedValue: 83.6148529052734,
			},
		}

		for _, tl := range testLocations {
			val, err := geo.loc(tl.x, tl.y)
			if err != nil {
				t.Errorf("got err %s", err)
			}

			if !checkToTolerance(float64(val), tl.expectedValue, 0.0001) {
				t.Errorf("got incorrect value %f want %f for %d, %d", val, tl.expectedValue, tl.x, tl.y)
			}
		}
	})

	t.Run("australia test file", func(t *testing.T) {
		testFile := "./testdata/WCSServer.tif"
		r, err := os.Open(testFile)
		if err != nil {
			t.Fatal(err)
		}
		geo, err := Read(r)
		if err != nil {
			t.Fatal(err)
		}

		// Commands used to extract values
		// gdallocationinfo testdata/WCSServer.tif x y
		testLocations := []struct {
			x             int
			y             int
			expectedValue float64
		}{
			{
				x:             0,
				y:             0,
				expectedValue: 0,
			},
			{
				x:             1000,
				y:             100,
				expectedValue: 38.1251182556152,
			},
			{
				x:             1000,
				y:             666,
				expectedValue: 199.612075805664,
			},
		}

		for _, tl := range testLocations {
			val, err := geo.loc(tl.x, tl.y)
			if err != nil {
				t.Errorf("got err %s", err)
			}

			if !checkToTolerance(float64(val), tl.expectedValue, 0.0001) {
				t.Errorf("got incorrect value %f want %f for %d, %d", val, tl.expectedValue, tl.x, tl.y)
			}
		}
	})
}

func Test_AtCoord_Happy(t *testing.T) {
	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	geo, err := Read(r)
	if err != nil {
		t.Fatal(err)
	}

	// Commands used to extract values
	// gdallocationinfo -wgs84 testdata/test.tif  lon lat
	testLocations := []struct {
		lon           float64
		lat           float64
		expectedValue float64
	}{
		{
			lon:           138,
			lat:           -23,
			expectedValue: 261.246856689453,
		},
		{
			lon:           139.5,
			lat:           -24,
			expectedValue: 116.877799987793,
		},
	}

	for _, tl := range testLocations {
		val, err := geo.AtCoord(tl.lon, tl.lat, false)
		if err != nil {
			t.Errorf("got err %s", err)
		}
		if !checkToTolerance(float64(val), tl.expectedValue, 0.0001) {
			t.Errorf("got incorrect value %f want %f for %f, %f", val, tl.expectedValue, tl.lon, tl.lat)
		}
	}
}

func Test_New_Happy(t *testing.T) {
	g, err := New(
		[][]float32{
			{1, 2, 3, 4, 6, 7, 0, 0},
			{1, 2, 3, 4, 6, 7, 0, 0},
			{1, 2, 3, 4, 6, 7, 0, 0},
			{1, 2, 3, 4, 6, 7, 0, 0},
		},
		6, 4, 4, 2, 1, 1, nil)
	if err != nil {
		t.Fatalf("failed with %s", err)
	}

	t.Run("point (0, 0)", func(t *testing.T) {
		gt1, err := g.loc(0, 0)
		if err != nil {
			t.Fail()
		}
		if gt1 != 1 {
			t.Fail()
		}
	})

	t.Run("point (5, 1)", func(t *testing.T) {
		var want float32 = 7.0
		gt1, err := g.loc(5, 1)
		if err != nil {
			t.Fail()
		}
		if gt1 != want {
			t.Errorf("got %f wanted %f", gt1, want)
		}
	})

	t.Run("outside bounds", func(t *testing.T) {
		_, err := g.loc(7, 4)
		if err == nil {
			t.Fail()
		}
	})
}

func Test_New_Sad(t *testing.T) {
	t.Run("negative resolution", func(t *testing.T) {
		_, err := New(
			[][]float32{
				{1, 2, 3, 4, 6, 7, 0, 0},
				{1, 2, 3, 4, 6, 7, 0, 0},
				{1, 2, 3, 4, 6, 7, 0, 0},
				{1, 2, 3, 4, 6, 7, 0, 0},
			},
			6, 4, 4, 2, -1, -0.001, nil)

		if err == nil {
			t.Fail()
		}
	})

	t.Run("incorrect image size", func(t *testing.T) {
		_, err := New(
			[][]float32{
				{1, 2, 3, 4, 6, 7, 0, 0},
				{1, 2, 3, 4, 6, 7, 0, 0},
				{1, 2, 3, 4, 6, 7, 0, 0},
				{1, 2, 3, 4, 6, 7, 0, 0},
			},
			7, 4, 4, 2, -1, -0.001, nil)

		if err == nil {
			t.Fail()
		}
	})

	t.Run("incorrect tile size", func(t *testing.T) {
		_, err := New(
			[][]float32{
				{1, 2, 3, 4, 6, 7, 0, 0},
				{1, 2, 3, 4, 6, 7, 0, 0},
				{1, 2, 3, 4, 6, 7, 0, 0},
				{1, 2, 3, 4, 6, 7, 0, 0},
			},
			6, 4, 4, 3, -1, -0.001, nil)

		if err == nil {
			t.Fail()
		}
	})
}

func Test_HaversineDistance(t *testing.T) {
	// Examples used from https://pypi.org/project/haversine/
	t.Run("lyon to paris", func(t *testing.T) {
		lyon := Point{Lat: 45.7597, Lon: 4.8422}
		paris := Point{Lat: 48.8567, Lon: 2.3508}
		wantDistanceInMetres := 392217.2595594006
		if !checkToTolerance(lyon.Distance(paris), wantDistanceInMetres, 0.0001) {
			t.Errorf("got %f want %f", lyon.Distance(paris), wantDistanceInMetres)
		}
	})
}
