package geotiff

import (
	"bytes"
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

func Test_Stats_Happy(t *testing.T) {
	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}

	gtiff, err := Read(r)
	if err != nil {
		t.Fatal(err)
	}

	// Check the statistics for the file
	// Extracted via gdalinfo -stats go/runoffarea/internal/geotiff/testdata/test.tif
	wantMinimum := 46.557
	wantMaximum := 942.159
	wantMean := 234.399
	wantStdev := 106.601

	got := gtiff.Stats()
	tolerance := 0.001

	if !checkToTolerance(float64(got.Min), wantMinimum, tolerance) {
		t.Errorf("minimum value incorrect: want %f got %f", float32(wantMinimum), got.Min)
	}

	if !checkToTolerance(float64(got.Max), wantMaximum, tolerance) {
		t.Errorf("maximum value incorrect: want %f got %f", float32(wantMaximum), got.Max)
	}

	// need to be a bit nicer here, as the floating point error will be larger
	if !checkToTolerance(float64(got.Mean), wantMean, 0.01) {
		t.Errorf("mean value incorrect: want %f got %f", float32(wantMean), got.Mean)
	}

	if !checkToTolerance(float64(got.StdDev), wantStdev, 0.01) {
		t.Errorf("standard deviation value incorrect: want %f got %f", float32(wantStdev), got.StdDev)
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

// makeTestStripTIFF builds a minimal valid little-endian striped GeoTIFF binary
// entirely in memory.
//
// Image properties:
//   - 4 × 6 pixels (width × height), float32, no compression
//   - 2 rows per strip → 3 strips
//   - Pixel values 1 … 24 in row-major order
//   - PixelScale (0.1, 0.1), upper-left tiepoint at (135.0, -20.0)
func makeTestStripTIFF(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	w := func(v interface{}) {
		t.Helper()
		if err := binary.Write(&buf, binary.LittleEndian, v); err != nil {
			t.Fatalf("binary.Write: %v", err)
		}
	}

	const (
		imgWidth      = 4
		imgHeight     = 6
		rps           = 2 // RowsPerStrip
		numStrips     = (imgHeight + rps - 1) / rps // = 3 (ceiling division)
		bytesPerStrip = imgWidth * rps * 4           // float32 = 4 bytes each
	)

	// Byte offsets for each section of the file.
	const (
		ifdStart            = 8
		numIFDEntries       = 13
		ifdSize             = 2 + numIFDEntries*12 + 4 // 162
		stripOffsetsOff     = ifdStart + ifdSize        // 170
		stripByteCountsOff  = stripOffsetsOff + numStrips*4  // 182
		pixelScaleOff       = stripByteCountsOff + numStrips*4 // 194
		tiepointOff         = pixelScaleOff + 3*8             // 218
		imageDataOff        = tiepointOff + 6*8               // 266
	)

	// Header
	w(uint16(0x4949))    // II = little-endian
	w(uint16(42))        // TIFF magic
	w(uint32(ifdStart))  // offset to first IFD

	// IFD entry count
	w(uint16(numIFDEntries))

	// Helpers to write individual IFD entries.
	writeShort := func(tag, val uint16) {
		w(tag); w(uint16(3)); w(uint32(1)); w(val); w(uint16(0))
	}
	writeLongOff := func(tag uint16, count, offset uint32) {
		w(tag); w(uint16(4)); w(count); w(offset)
	}
	writeDoubleOff := func(tag uint16, count, offset uint32) {
		w(tag); w(uint16(12)); w(count); w(offset)
	}

	// IFD entries – must be sorted in ascending tag order.
	writeShort(256, imgWidth)                                       // ImageWidth
	writeShort(257, imgHeight)                                      // ImageLength
	writeShort(258, 32)                                             // BitsPerSample
	writeShort(259, 1)                                              // Compression = none
	writeShort(262, 1)                                              // PhotometricInterpretation
	writeLongOff(273, numStrips, stripOffsetsOff)                   // StripOffsets
	writeShort(277, 1)                                              // SamplesPerPixel
	writeShort(278, rps)                                            // RowsPerStrip
	writeLongOff(279, numStrips, stripByteCountsOff)                // StripByteCounts
	writeShort(284, 1)                                              // PlanarConfiguration
	writeShort(339, 3)                                              // SampleFormat = IEEE float
	writeDoubleOff(33550, 3, pixelScaleOff)                         // ModelPixelScale
	writeDoubleOff(33922, 6, tiepointOff)                           // ModelTiepoint

	w(uint32(0)) // next IFD = none

	// StripOffsets values
	for i := 0; i < numStrips; i++ {
		w(uint32(imageDataOff + i*bytesPerStrip))
	}
	// StripByteCounts values
	for i := 0; i < numStrips; i++ {
		w(uint32(bytesPerStrip))
	}
	// ModelPixelScale: (scaleX, scaleY, scaleZ)
	w(float64(0.1)); w(float64(0.1)); w(float64(0.0))
	// ModelTiepoint: (I, J, K, X, Y, Z)  – upper-left pixel at (135.0, -20.0)
	w(float64(0)); w(float64(0)); w(float64(0))
	w(float64(135.0)); w(float64(-20.0)); w(float64(0))

	// Image data: float32 values 1 … 24 in row-major order
	var v float32 = 1.0
	for i := 0; i < numStrips*imgWidth*rps; i++ {
		w(v)
		v++
	}

	return buf.Bytes()
}

func Test_ReadStripData_Happy(t *testing.T) {
	data := makeTestStripTIFF(t)
	r := bytes.NewReader(data)

	tags, h, err := readTags(r)
	if err != nil {
		t.Fatal(err)
	}

	strips, err := readStripData(r, tags, h)
	if err != nil {
		t.Fatal(err)
	}

	// 3 strips × 8 pixels (4 wide, 2 rows) each
	if len(strips) != 3 {
		t.Fatalf("expected 3 strips, got %d", len(strips))
	}
	for i, strip := range strips {
		if len(strip) != 8 {
			t.Errorf("strip %d: expected 8 pixels, got %d", i, len(strip))
		}
	}

	// Strip 0 holds rows 0-1: values 1..8
	if strips[0][0] != 1.0 {
		t.Errorf("strips[0][0]: got %f, want 1.0", strips[0][0])
	}
	if strips[0][7] != 8.0 {
		t.Errorf("strips[0][7]: got %f, want 8.0", strips[0][7])
	}
	// Strip 2 holds rows 4-5: values 17..24
	if strips[2][0] != 17.0 {
		t.Errorf("strips[2][0]: got %f, want 17.0", strips[2][0])
	}
	if strips[2][7] != 24.0 {
		t.Errorf("strips[2][7]: got %f, want 24.0", strips[2][7])
	}
}

func Test_Read_StripFormat_Happy(t *testing.T) {
	data := makeTestStripTIFF(t)
	r := bytes.NewReader(data)

	gtiff, err := Read(r)
	if err != nil {
		t.Fatal(err)
	}

	// Row-major pixel layout (0-indexed):
	//   Row 0: [ 1,  2,  3,  4]
	//   Row 1: [ 5,  6,  7,  8]
	//   Row 2: [ 9, 10, 11, 12]
	//   Row 3: [13, 14, 15, 16]
	//   Row 4: [17, 18, 19, 20]
	//   Row 5: [21, 22, 23, 24]
	tests := []struct {
		x, y int
		want float32
	}{
		{0, 0, 1.0},   // row 0, col 0
		{3, 1, 8.0},   // row 1, col 3
		{2, 2, 11.0},  // row 2, col 2
		{1, 4, 18.0},  // row 4, col 1
		{3, 5, 24.0},  // row 5, col 3
	}

	for _, tt := range tests {
		got, err := gtiff.loc(tt.x, tt.y)
		if err != nil {
			t.Errorf("loc(%d, %d) error: %v", tt.x, tt.y, err)
			continue
		}
		if got != tt.want {
			t.Errorf("loc(%d, %d) = %f, want %f", tt.x, tt.y, got, tt.want)
		}
	}
}

func Test_Read_StripFormat_Bounds(t *testing.T) {
	data := makeTestStripTIFF(t)
	r := bytes.NewReader(data)

	gtiff, err := Read(r)
	if err != nil {
		t.Fatal(err)
	}

	bounds, err := gtiff.Bounds()
	if err != nil {
		t.Fatal(err)
	}

	// Tiepoint (0,0) → upper-left at (135.0, -20.0)
	// Width=4, Height=6, PixelScale=(0.1, 0.1)
	// LowerRight = (135.0 + 4*0.1, -20.0 - 6*0.1) = (135.4, -20.6)
	tolerance := 1e-9
	if !checkToTolerance(bounds.UpperLeft.Lon, 135.0, tolerance) {
		t.Errorf("UpperLeft.Lon: got %f, want 135.0", bounds.UpperLeft.Lon)
	}
	if !checkToTolerance(bounds.UpperLeft.Lat, -20.0, tolerance) {
		t.Errorf("UpperLeft.Lat: got %f, want -20.0", bounds.UpperLeft.Lat)
	}
	if !checkToTolerance(bounds.LowerRight.Lon, 135.4, tolerance) {
		t.Errorf("LowerRight.Lon: got %f, want 135.4", bounds.LowerRight.Lon)
	}
	if !checkToTolerance(bounds.LowerRight.Lat, -20.6, tolerance) {
		t.Errorf("LowerRight.Lat: got %f, want -20.6", bounds.LowerRight.Lat)
	}
}

// makeTestStripTIFFLongDims builds a minimal valid little-endian striped
// GeoTIFF binary where ImageWidth, ImageLength, BitsPerSample, and
// RowsPerStrip are stored as LONG (uint32) fields rather than SHORT (uint16).
// This mirrors the layout used by real-world elevation files such as the NRW
// DGM1 dataset (https://www.opengeodata.nrw.de/produkte/geobasis/hm/dgm1_tiff/).
//
// The image properties are intentionally identical to makeTestStripTIFF so
// that the same expected values can be reused.
func makeTestStripTIFFLongDims(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	w := func(v interface{}) {
		t.Helper()
		if err := binary.Write(&buf, binary.LittleEndian, v); err != nil {
			t.Fatalf("binary.Write: %v", err)
		}
	}

	const (
		imgWidth      = 4
		imgHeight     = 6
		rps           = 2 // RowsPerStrip
		numStrips     = (imgHeight + rps - 1) / rps // = 3
		bytesPerStrip = imgWidth * rps * 4           // float32 = 4 bytes
	)

	// Byte offsets for each section of the file.
	// All dimension IFD entries are written as LONG (8 bytes: type+count+value)
	// rather than SHORT (which packs the value directly into the 4-byte offset
	// field), so the offsets differ from makeTestStripTIFF.
	const (
		ifdStart           = 8
		numIFDEntries      = 13
		ifdSize            = 2 + numIFDEntries*12 + 4 // 162
		stripOffsetsOff    = ifdStart + ifdSize        // 170
		stripByteCountsOff = stripOffsetsOff + numStrips*4
		pixelScaleOff      = stripByteCountsOff + numStrips*4
		tiepointOff        = pixelScaleOff + 3*8
		imageDataOff       = tiepointOff + 6*8
	)

	// Header
	w(uint16(0x4949))   // II = little-endian
	w(uint16(42))       // TIFF magic
	w(uint32(ifdStart)) // offset to first IFD

	// IFD entry count
	w(uint16(numIFDEntries))

	// writeLong writes a single-value LONG IFD entry whose value fits in the
	// 4-byte value-offset field (TIFF spec: values ≤ 4 bytes are stored inline).
	writeLong := func(tag uint16, val uint32) {
		w(tag); w(uint16(4)); w(uint32(1)); w(val)
	}
	writeLongOff := func(tag uint16, count, offset uint32) {
		w(tag); w(uint16(4)); w(count); w(offset)
	}
	writeDoubleOff := func(tag uint16, count, offset uint32) {
		w(tag); w(uint16(12)); w(count); w(offset)
	}

	// IFD entries – must be in ascending tag order.
	// Dimension/sample fields use LONG instead of SHORT to exercise the
	// SHORT/LONG-agnostic reading path introduced for NRW DGM1 support.
	writeLong(256, imgWidth)  // ImageWidth      – LONG
	writeLong(257, imgHeight) // ImageLength     – LONG
	writeLong(258, 32)        // BitsPerSample   – LONG
	writeLong(259, 1)         // Compression     – LONG (none)
	writeLong(262, 1)         // PhotometricInterpretation – LONG
	writeLongOff(273, numStrips, stripOffsetsOff)    // StripOffsets
	writeLong(277, 1)                                // SamplesPerPixel – LONG
	writeLong(278, rps)                              // RowsPerStrip    – LONG
	writeLongOff(279, numStrips, stripByteCountsOff) // StripByteCounts
	writeLong(284, 1)                                // PlanarConfiguration – LONG
	writeLong(339, 3)                                // SampleFormat – LONG (IEEE float)
	writeDoubleOff(33550, 3, pixelScaleOff)          // ModelPixelScale
	writeDoubleOff(33922, 6, tiepointOff)            // ModelTiepoint

	w(uint32(0)) // next IFD = none

	// StripOffsets values
	for i := 0; i < numStrips; i++ {
		w(uint32(imageDataOff + i*bytesPerStrip))
	}
	// StripByteCounts values
	for i := 0; i < numStrips; i++ {
		w(uint32(bytesPerStrip))
	}
	// ModelPixelScale: (scaleX, scaleY, scaleZ)
	w(float64(0.1)); w(float64(0.1)); w(float64(0.0))
	// ModelTiepoint: (I, J, K, X, Y, Z) – upper-left pixel at (135.0, -20.0)
	w(float64(0)); w(float64(0)); w(float64(0))
	w(float64(135.0)); w(float64(-20.0)); w(float64(0))

	// Image data: float32 values 1 … 24 in row-major order
	var v float32 = 1.0
	for i := 0; i < numStrips*imgWidth*rps; i++ {
		w(v)
		v++
	}

	return buf.Bytes()
}

// Test_Read_StripFormat_LongDims verifies that a striped GeoTIFF whose
// dimension tags (ImageWidth, ImageLength, BitsPerSample, RowsPerStrip) are
// encoded as LONG (uint32) rather than SHORT (uint16) can be read correctly.
// This exercises the fix for NRW DGM1 files which use LONG types.
func Test_Read_StripFormat_LongDims(t *testing.T) {
	data := makeTestStripTIFFLongDims(t)
	r := bytes.NewReader(data)

	gtiff, err := Read(r)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Same pixel layout as Test_Read_StripFormat_Happy
	tests := []struct {
		x, y int
		want float32
	}{
		{0, 0, 1.0},
		{3, 1, 8.0},
		{2, 2, 11.0},
		{1, 4, 18.0},
		{3, 5, 24.0},
	}

	for _, tt := range tests {
		got, err := gtiff.loc(tt.x, tt.y)
		if err != nil {
			t.Errorf("loc(%d, %d) error: %v", tt.x, tt.y, err)
			continue
		}
		if got != tt.want {
			t.Errorf("loc(%d, %d) = %f, want %f", tt.x, tt.y, got, tt.want)
		}
	}
}

// Test_ReadStripData_LongDims verifies that readStripData accepts dimension
// tags stored as LONG instead of SHORT.
func Test_ReadStripData_LongDims(t *testing.T) {
	data := makeTestStripTIFFLongDims(t)
	r := bytes.NewReader(data)

	tags, h, err := readTags(r)
	if err != nil {
		t.Fatal(err)
	}

	strips, err := readStripData(r, tags, h)
	if err != nil {
		t.Fatalf("readStripData failed on LONG-dimension TIFF: %v", err)
	}

	if len(strips) != 3 {
		t.Fatalf("expected 3 strips, got %d", len(strips))
	}
	for i, strip := range strips {
		if len(strip) != 8 {
			t.Errorf("strip %d: expected 8 pixels, got %d", i, len(strip))
		}
	}

	if strips[0][0] != 1.0 {
		t.Errorf("strips[0][0]: got %f, want 1.0", strips[0][0])
	}
	if strips[2][7] != 24.0 {
		t.Errorf("strips[2][7]: got %f, want 24.0", strips[2][7])
	}
}
