package geotiff

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
)

// head contains the tiff file header
//
// Per the TIFF 6.0 Specification (p.13)
//
// A TIFF file begins with an 8-byte image file header that points to an image
// file directory (IFD). An image file directory contains information about the
// image, as well as pointers to the actual image data.
type head struct {
	// Bytes 0-1
	//
	// The byte order used within the file
	byteOrder binary.ByteOrder

	// Bytes 2-3
	//
	// An arbitrary but carefully chosen number (42) that further identifies the file as a TIFF file.
	// The byte order depends on the value of Bytes 0-1.
	tIFIdentifier uint16

	// Bytes 4-7
	//
	// The offset (in bytes) of the first IFD
	iFDByteOffset uint32
}

// iFDEntry contains the image file directory (IFD) entries.
//
// Per the TIFF 6.0 Specification (p.14)
//
// An Image File Directory (IFD) consists of a 2-byte count of the number of
// directory entries (i.e., the number of fields), followed by a sequence of
// 12-byte field entries, followed by a 4-byte offset of the next IFD (or 0 if
// none). (Do not forget to write the 4 bytes of 0 after the last IFD.) There
// must be at least 1 IFD in a TIFF file and each IFD must have at least one
// entry.
type iFDEntry struct {
	// Bytes 0-1
	//
	// The Tag that identifies the field.
	Tag Tag

	// Bytes 2-3
	//
	//The field FType.
	FType fieldType

	//  Bytes 4-7
	//
	// The number of values, Count of the indicated Type.
	Count uint32

	// Bytes 8-11
	//
	// The Value Offset, the file offset (in bytes) of the Value for
	// the field. The Value is expected to begin on a word boundary; the
	// corresponding Value Offset will thus be an even number. This file offset
	// may point anywhere in the file, even after the image data.
	ValueOffset uint32
}

func (ifd *iFDEntry) totalBytes() uint32 {
	return ifd.Count * ifd.FType.bytes()
}

// value reads the directory value
func (ifd *iFDEntry) value(r io.ReadSeeker, byteOrder binary.ByteOrder) (*tagData, error) {
	t := tagData{
		fType:  ifd.FType,
		length: ifd.Count,
	}
	offset := int64(ifd.ValueOffset)
	if _, err := r.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}
	switch ifd.FType {
	case BYTE:
		t.byteData = make([]uint8, ifd.Count)
		if err := binary.Read(r, byteOrder, &t.byteData); err != nil {
			return nil, err
		}
	case ASCII:
		p := make([]uint8, ifd.Count)
		if err := binary.Read(r, byteOrder, p); err != nil {
			return nil, err
		}
		t.asciiData = string(p)
	case SHORT:
		t.shortData = make([]uint16, ifd.Count)
		if err := binary.Read(r, byteOrder, &t.shortData); err != nil {
			return nil, err
		}
	case LONG:
		t.longData = make([]uint32, ifd.Count)
		if err := binary.Read(r, byteOrder, &t.longData); err != nil {
			return nil, err
		}
	case FLOAT:
		t.floatData = make([]float32, ifd.Count)
		if err := binary.Read(r, byteOrder, t.floatData); err != nil {
			return nil, err
		}
	case DOUBLE:
		t.doubleData = make([]float64, ifd.Count)
		if err := binary.Read(r, byteOrder, t.doubleData); err != nil {
			return nil, err
		}
	}
	return &t, nil
}

var errByteOrder = errors.New("could not parse tiff header byte order")

// readHeader reads the header of a GeoTIFF file
func readHeader(r io.Reader) (head, error) {
	var fileHeader head

	// Byte Order
	var byteOrder uint16
	err := binary.Read(r, binary.BigEndian, &byteOrder)
	if err != nil {
		return fileHeader, fmt.Errorf("%w: %s", errByteOrder, err)
	}

	switch byteOrder {
	case littleEndian:
		fileHeader.byteOrder = binary.LittleEndian
	case bigEndian:
		fileHeader.byteOrder = binary.BigEndian
	default:
		return fileHeader, errByteOrder
	}

	// Identifier
	err = binary.Read(r, fileHeader.byteOrder, &fileHeader.tIFIdentifier)
	if err != nil {
		return fileHeader, fmt.Errorf("err: failed to read tiff header: %w", err)
	}

	if fileHeader.tIFIdentifier != tiffIdentifier {
		return fileHeader, fmt.Errorf("invalid tiff file identifier: expected %d got %d",
			tiffIdentifier, fileHeader.tIFIdentifier)
	}

	// Offset
	err = binary.Read(r, fileHeader.byteOrder, &fileHeader.iFDByteOffset)
	if err != nil {
		return fileHeader, fmt.Errorf("failed to read IFD byte offset: %w", err)
	}
	return fileHeader, nil
}

// tagData holds the tag data for each tag
// is supposed to act similar to a union
// where only one data field is used at any one time
type tagData struct {
	fType      fieldType
	length     uint32
	byteData   []uint8
	asciiData  string
	shortData  []uint16
	longData   []uint32
	floatData  []float32
	doubleData []float64
}

// Tags holds the tag files
type Tags map[Tag]tagData

// String converts types to strings
func (t tagData) String() string {
	var dataStr string
	switch t.fType {
	case SHORT:
		dataStr = fmt.Sprintf("%v", t.shortData)
	case LONG:
		dataStr = fmt.Sprintf("%v", t.longData)
	case FLOAT:
		dataStr = fmt.Sprintf("%v", t.floatData)
	case DOUBLE:
		dataStr = fmt.Sprintf("%v", t.doubleData)
	case BYTE:
		dataStr = fmt.Sprintf("%v", t.byteData)
	case ASCII:
		dataStr = fmt.Sprintf("%v", t.asciiData)
	}
	return t.fType.String() + " " + fmt.Sprintf("%d", t.length) + " " + dataStr
}

func (t tagData) value() (fieldType, []interface{}) {
	switch t.fType {
	case SHORT:
		return t.fType, []interface{}{t.shortData}
	case LONG:
		return t.fType, []interface{}{t.longData}
	case FLOAT:
		return t.fType, []interface{}{t.floatData}
	case DOUBLE:
		return t.fType, []interface{}{t.doubleData}
	case BYTE:
		return t.fType, []interface{}{t.byteData}
	case ASCII:
		return t.fType, []interface{}{t.asciiData}
	}
	return NONE, nil
}

// readTags reads the tags of a GeoTIFF file
func readTags(r io.ReadSeeker) (Tags, head, error) {
	tags := make(Tags)

	// read the header tag to extract the IFD Byte offset
	h, err := readHeader(r)
	if err != nil {
		return tags, h, fmt.Errorf("failed to read tiff header: %w", err)
	}

	// Get the first IFD entry via the IFD Byte offset recorded in the header
	// via SeekStart
	//
	// From the TIFF 6.0 Specification (p.13)
	//
	// The offset (in bytes) of the first IFD. The directory may be at any location in the
	// file after the header but must begin on a word boundary. In particular, an
	// Image File Directory may follow the image data it describes. Readers
	// must follow the pointers wherever they may lead. The term byte offset is
	// always used in this document to refer to a location with respect to the
	// beginning of the TIFF file. The first byte of the file has an offset of
	// 0.
	iFDOffset := h.iFDByteOffset

	// Jump to the first IFD Byte Offset
	if _, err := r.Seek(int64(iFDOffset), io.SeekStart); err != nil {
		return tags, h, errors.New("error: unable to read IFD Start")
	}

	for iFDOffset != 0 {
		// Per the TIFF 6.0 Specification (p.14)
		//
		// The number of Directory Entries is contained in the
		// first two bytes of each IFD
		var numDirectoryEntries uint16
		if err := binary.Read(r, h.byteOrder, &numDirectoryEntries); err != nil {
			return tags, h, errors.New("error: unable to read directory entry")
		}
		var nextDirOffset int64
		for i := uint16(0); i < numDirectoryEntries; i++ {

			var iFDEntry iFDEntry
			if err := binary.Read(r, h.byteOrder, &iFDEntry); err != nil {
				return tags, h, err
			}

			if iFDEntry.FType.bytes() == 0 {
				return tags, h, fmt.Errorf("error: unrecognized tag %d", iFDEntry.Tag)
			}

			// Per  the TIFF 6.0 Specification
			//
			// If the Value is shorter than 4 bytes, it is left-justified
			// within the 4-byte Value Offset, i.e., stored in the
			// lower numbered bytes.
			//
			// Whether the Value fits within 4 bytes is determined by the Type
			// and Count of the field
			if iFDEntry.totalBytes() <= fourByte {
				currentOffset, _ := r.Seek(0, io.SeekCurrent)
				iFDEntry.ValueOffset = uint32(currentOffset) - fourByte
			}

			nextDirOffset, _ = r.Seek(0, io.SeekCurrent)

			// Read the tags
			tagName := iFDEntry.Tag
			tagvalue, err := iFDEntry.value(r, h.byteOrder)
			if err != nil {
				return tags, h, err
			}
			tags[tagName] = *tagvalue

			// Jump to the next directory entry
			if _, err := r.Seek(nextDirOffset, io.SeekStart); err != nil {
				return tags, h, fmt.Errorf("err: could not jump to next directory: %w", err)
			}
		}

		if err = binary.Read(r, h.byteOrder, &iFDOffset); err != nil {
			return tags, h, fmt.Errorf("err: could not jump to next file: %w", err)
		}
	}
	return tags, h, nil
}

var errGeoTIFFData = errors.New("could not read GeoTIFF data")

// readData reads the data from a  tiled GeoTIFF file
// into a 1D 32bit float array
func readData(r io.ReadSeeker, tags Tags, header head) ([][]float32, error) {
	fields := [...]Tag{BitsPerSample, ImageLength, ImageWidth, TileWidth, TileLength, TileByteCounts, TileOffsets}
	shortData := make(map[Tag][]uint16, 0)
	longData := make(map[Tag][]uint32, 0)
	for _, f := range fields {
		v, ok := tags[f]
		if !ok {
			return nil, fmt.Errorf("%w, could not retrieve %s", errGeoTIFFData, f)
		}
		ftype, elem := v.value()
		switch ftype {
		case SHORT:
			shortData[f] = elem[0].([]uint16)
		case LONG:
			longData[f] = elem[0].([]uint32)
		default:
			return nil, fmt.Errorf("%w, incorrect data type for %s - %s", errGeoTIFFData, f, ftype)

		}
	}

	imageWidth := shortData[ImageWidth][0]
	imageLength := shortData[ImageLength][0]
	tileWidth := shortData[TileWidth][0]
	tileLength := shortData[TileLength][0]
	bitsPerSample := shortData[BitsPerSample][0]
	offsets := longData[TileOffsets]
	byteCounts := longData[TileByteCounts]

	// From the Tiff 6.0 Specification (p. 67)
	tilesAcross := (imageWidth + tileWidth - 1) / tileWidth
	tilesDown := (imageLength + tileLength - 1) / tileLength
	tilesPerImage := tilesAcross * tilesDown

	if int(tilesPerImage) != len(offsets) {
		return nil, errors.New("invalid number of offsets for tiles")
	}

	data := make([][]float32, 0, tilesPerImage)
	for i, offset := range offsets {
		numPixels := uint32(byteCounts[i]) / (uint32(bitsPerSample) / eightByte)
		tileData := make([]float32, numPixels)
		if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("%w: could not find offset: got %s", errGeoTIFFData, err)
		}

		if err := binary.Read(r, header.byteOrder, &tileData); err != nil {
			return nil, fmt.Errorf("%w: could not read bytes: got %s", errGeoTIFFData, err)
		}

		data = append(data, tileData)
	}
	return data, nil
}

// GeoTIFF a geotiff object
type GeoTIFF struct {
	tags        Tags
	data        [][]float32
	imageWidth  uint16
	imageLength uint16
	tileWidth   uint16
	tileLength  uint16
	PixelScaleX float64
	PixelScaleY float64
}

// AtCoord returns the value closest to the requested latitude and longitude value
//
// If the value does not exist (i.e., the location requested) falls in between
// multiple grid points.
//
// Interp indicates if bilinear interpolation should be done along
// either direction
func (g *GeoTIFF) AtCoord(x float64, y float64, interp bool) (float32, error) {
	rect, err := g.Bounds()
	if err != nil {
		return 0, err
	}
	p := Point{Lon: x, Lat: y}
	if !rect.Contains(p) {
		return 0, fmt.Errorf("requested point %s does not fall inside the image bounds", p)
	}

	if interp {
		return g.interp(p)
	}

	xIDx := int(math.Abs(rect.UpperLeft.Lon-p.Lon) / g.PixelScaleX)
	yIDx := int(math.Abs(p.Lat-rect.UpperLeft.Lat) / g.PixelScaleY)
	val, err := g.loc(xIDx, yIDx)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// interp computes bilinear interpolation between
// four point values
//
// This isn't perfect as it doesn't completely solve for the face that data is only
// available at grid points
func (g *GeoTIFF) interp(p Point) (float32, error) {
	points := []Point{
		{
			Lon: p.Lon - g.PixelScaleX,
			Lat: p.Lat,
		},
		{
			Lon: p.Lon + g.PixelScaleX,
			Lat: p.Lat,
		},
		{
			Lon: p.Lon,
			Lat: p.Lat - g.PixelScaleY,
		},
		{
			Lon: p.Lon,
			Lat: p.Lat + g.PixelScaleY,
		},
	}

	pointValues, err := g.AtPoints(points, false)
	if err != nil {
		return 0, err
	}
	// See: https://en.wikipedia.org/wiki/Bilinear_interpolation
	// All the coefficients are equivalent here as it is assumed the grid is equal (which is not correct)
	// therefore the result is just the mean value
	var pv float32 = 0.0
	for _, pvals := range pointValues {
		pv += pvals
	}
	return pv / 4.0, nil
}

// AtPoints returns image values at
// a specified slice of points
func (g *GeoTIFF) AtPoints(points []Point, interp bool) ([]float32, error) {
	data := make([]float32, 0, len(points))
	for i, p := range points {
		v, err := g.AtCoord(p.Lon, p.Lat, interp)
		if err != nil {
			return nil, err
		}
		data[i] = v
	}
	return data, nil
}

// loc returns data by location (i.e. an X, and Y point on the image)
func (g *GeoTIFF) loc(x int, y int) (float32, error) {
	if x < 0 || x >= int(g.imageWidth) || y < 0 || y >= int(g.imageLength) {
		return 0.0, errors.New("point lies outside image")
	}

	// The data is read into the array from the top left corner in tiles
	// it starts at (0, 0) and reads in to (0, tileWidth) before repeating this to
	// (tileWidth, tileLength) before beginning to read in at (tileWidth, 0) and
	// repeating this process. The buffer is placed on the bottom and the right hand side
	//
	//  This means the data is read in in the order shown below
	//   ------------------------------    ...
	//   |         |         |      | 0 0 ..  |   ...
	//   |    0    |    1    |   2  | 0 0 ..  |   ...
	//   |         |         |      | 0 0 ..  |   ...
	//   |---------|---------|------|---------|   ...
	//   |         |         |      | 0 0 ..  |   ...
	//   |    3    |    4    |   5  | 0 0 ..  |   ...
	//   |         |         |      | 0 0 ..  |   ...
	//   |---------|---------|------|---------|   ...
	//   |         |         |      | 0 0 ..  |   ...
	//   |    6    |    7    |   8  | 0 0 ..  |   ...
	//   |    .    |    .    |   .  | 0 0 ..  |   ...
	//   |    .    |    .    |   .  | 0 0 ..  |   ...
	//
	tilesAcross := int(g.imageWidth+g.tileWidth-1) / int(g.tileWidth)
	idAcross := x / int(g.tileWidth)
	idDown := y / int(g.tileLength)
	tileNum := tilesAcross*idDown + idAcross
	idI := x % int(g.tileWidth)
	idJ := (y % int(g.tileLength)) * int(g.tileWidth)
	return g.data[tileNum][idJ+idI], nil
}

// Bounds returns the bounding rectangle of the image
func (g *GeoTIFF) Bounds() (*CornerCoordinates, error) {
	// Check for the model tie point
	tiePoint, ok := g.tags[ModelTiepoint]

	// TODO: add transformation
	if !ok {
		return nil, errors.New("unable to retrieve model tiepoint:  ModelTransformationTag not implemented")
	}

	// https://freeimage.sourceforge.io/fnet/html/38F9430A.htm
	//
	// ModelTiePoints = (...,I,J,K, X,Y,Z...), where (I,J,K) is the point at
	// location (I,J) in raster space with pixel-value K, and (X,Y,Z) is a
	// vector in model space. In most cases the model space is only
	// two-dimensional, in which case both K and Z should be set to zero; this
	// third dimension is provided in anticipation of future support for 3D
	// digital elevation models and vertical coordinate systems. A raster image
	// may be georeferenced simply by specifying its location, size and
	// orientation in the model coordinate space M. This may be done by
	// specifying the location of three of the four bounding corner points.
	// However, tiepoints are only to be considered exact at the points
	// specified; thus defining such a set of bounding tiepoints does not imply
	// that the model space locations of the interior of the image may be
	// exactly computed by a linear interpolation of these tiepoints.
	tiePointLen := 6
	if int(tiePoint.length) != tiePointLen {
		return nil, fmt.Errorf("%s has invalid length %d", ModelTiepoint, tiePoint.length)
	}
	var tiePointValues []float64
	ftype, e := tiePoint.value()
	switch ftype {
	case DOUBLE:
		tiePointValues = e[0].([]float64)
	default:
		return nil, errors.New("unrecognized value for tiepoint")
	}
	cc := CornerCoordinates{}
	if tiePointValues[0] == 0 && tiePointValues[1] == 0 {
		cc.UpperLeft = Point{Lon: tiePointValues[3], Lat: tiePointValues[4]}
		cc.LowerLeft = Point{Lon: cc.UpperLeft.Lon, Lat: cc.UpperLeft.Lat - float64(g.imageLength)*g.PixelScaleY}
		cc.LowerRight = Point{Lon: cc.UpperLeft.Lon + float64(g.imageWidth)*g.PixelScaleX, Lat: cc.LowerLeft.Lat}
		cc.UpperRight = Point{Lon: cc.LowerRight.Lon, Lat: cc.UpperLeft.Lat}
	} else {
		cc.UpperLeft = Point{
			Lon: tiePointValues[3] - float64(tiePointValues[0])*g.PixelScaleX,
			Lat: tiePointValues[4] + float64(tiePointValues[1])*g.PixelScaleY,
		}
		cc.LowerRight = Point{
			Lon: tiePointValues[3] + float64(tiePointValues[0])*g.PixelScaleX,
			Lat: tiePointValues[4] - float64(tiePointValues[1])*g.PixelScaleY,
		}
		cc.LowerLeft = Point{Lon: cc.LowerLeft.Lon, Lat: cc.LowerRight.Lat}
		cc.UpperRight = Point{Lon: cc.LowerRight.Lon, Lat: cc.UpperLeft.Lat}
	}
	return &cc, nil
}

// Point contains X, Y longitude and latitude points
type Point struct {
	Lon float64
	Lat float64
}

// String matches the formatting from gdalinfo
func (p Point) String() string {
	return fmt.Sprintf("%011.7f,  %011.7f", p.Lon, p.Lat)
}

// Equals checks if two points are equal
func (p *Point) Equals(px Point) bool {
	return p.Lon == px.Lon && p.Lat == px.Lat
}

// Distance returns the haversine distance between any two points
// in metres
func (p *Point) Distance(px Point) float64 {
	// https://en.wikipedia.org/wiki/Earth_radius#Mean_radius
	const earthRadiusInMetres = 6_371_008.8
	d2R := func(deg float64) float64 {
		return math.Pi * deg / 180.0
	}
	latSin := math.Pow(math.Sin(d2R(px.Lat-p.Lat)/2.0), 2)
	lonSin := math.Pow(math.Sin(d2R(px.Lon-p.Lon)/2.0), 2)
	cosLat := math.Cos(d2R(px.Lat)) * math.Cos(d2R(p.Lat))
	return float64(earthRadiusInMetres*2) * math.Asin(math.Sqrt(latSin+lonSin*cosLat))
}

// CornerCoordinates contains the GeoTiffs corners
type CornerCoordinates struct {
	UpperLeft  Point
	LowerLeft  Point
	UpperRight Point
	LowerRight Point
}

// Contains checks if a point falls inside the corner coordinates
func (cc CornerCoordinates) Contains(p Point) bool {
	return cc.UpperLeft.Lon <= p.Lon && p.Lon <= cc.UpperRight.Lon &&
		cc.LowerLeft.Lat <= p.Lat && p.Lat <= cc.UpperLeft.Lat
}

func (cc CornerCoordinates) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Upper Left   ( %s )\n", cc.UpperLeft))
	sb.WriteString(fmt.Sprintf("Lower Left   ( %s )\n", cc.LowerLeft))
	sb.WriteString(fmt.Sprintf("Upper Right  ( %s )\n", cc.UpperRight))
	sb.WriteString(fmt.Sprintf("Lower Right  ( %s )\n", cc.LowerRight))
	return sb.String()
}

// Read reads the GeoTIFF file
func Read(r io.ReadSeeker) (*GeoTIFF, error) {
	gTags, header, err := readTags(r)
	if err != nil {
		return nil, err
	}

	gData, err := readData(r, gTags, header)
	if err != nil {
		return nil, err
	}

	fields := [...]Tag{ImageLength, ImageWidth, TileWidth, TileLength}
	shortData := make(map[Tag][]uint16, 0)
	for _, f := range fields {
		v, ok := gTags[f]
		if !ok {
			return nil, fmt.Errorf("%w, could not retrieve %s", errGeoTIFFData, f)
		}
		ftype, elem := v.value()
		switch ftype {
		case SHORT:
			shortData[f] = elem[0].([]uint16)
		default:
			return nil, fmt.Errorf("%w, incorrect data type for %s - %s", errGeoTIFFData, f, ftype)

		}
	}
	imageWidth := shortData[ImageWidth][0]
	imageLength := shortData[ImageLength][0]
	tileWidth := shortData[TileWidth][0]
	tileLength := shortData[TileLength][0]

	pixelScale := gTags[ModelPixelScale]
	pixelScaleLen := 3
	if int(pixelScale.length) != pixelScaleLen {
		return nil, fmt.Errorf("%s has invalid length %d", ModelPixelScale, pixelScale.length)
	}
	var pixelScaleValues []float64
	f, e := pixelScale.value()
	switch f {
	case DOUBLE:
		pixelScaleValues = e[0].([]float64)
	default:
		return nil, fmt.Errorf("unrecognized value for %s", ModelPixelScale)
	}

	return &GeoTIFF{
		tags:        gTags,
		data:        gData,
		imageWidth:  imageWidth,
		imageLength: imageLength,
		tileWidth:   tileWidth,
		tileLength:  tileLength,
		PixelScaleX: pixelScaleValues[0],
		PixelScaleY: pixelScaleValues[1],
	}, nil
}

// New creates a new instance of a geotiff object
//
// # This constructor is primarily used for testing
//
// This constructor does not verify what tags have been included
func New(data [][]float32, iWidth uint16, iLength uint16, tWidth uint16, tLength uint16, pX float64, pY float64, tags Tags) (*GeoTIFF, error) {
	if pX < 0 || pY < 0 {
		return nil, errors.New("pixel scale tags should be > 0")
	}

	tilesAcross := (iWidth + tWidth - 1) / tWidth
	tilesDown := (iLength + tLength - 1) / tLength
	tilesPerImage := tilesAcross * tilesDown

	if int(tilesPerImage) != len(data) {
		return nil, errors.New("invalid number of tiles for data")
	}

	for _, d := range data {
		if len(d) != int(tWidth*tLength) {
			return nil, errors.New("invalid amount of tile data passed")
		}
	}

	g := &GeoTIFF{}
	g.data = data
	g.imageWidth = iWidth
	g.imageLength = iLength
	g.tileWidth = tWidth
	g.tileLength = tLength
	g.PixelScaleX = pX
	g.PixelScaleY = pY
	g.tags = tags
	return g, nil
}
