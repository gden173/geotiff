// Package geotiff contains an implementation of a geotiff reader
package geotiff

import "fmt"

// From the Tiff 6.0 Specification (p.13)
//
// An Image File Directory (IFD) consists of a 2-byte count of the number of
// directory entries (i.e., the number of fields), followed by a sequence of
// 12-byte field entries, followed by a 4-byte offset of the next IFD (or 0 if
// none). (Do not forget to write the 4 bytes of 0 after the last IFD.) There
// must be at least 1 IFD in a TIFF file and each IFD must have at least one
// entry.
//
// IFD Entry
//
// Each 12-byte IFD entry has the following format:
// Bytes 0-1 The Tag that identifies the field.
// Bytes 2-3 The field Type.
// Bytes 4-7 The number of values, Count of the indicated Type
// Bytes 8-11 The Value Offset, the file offset (in bytes) of the Value for the field.
// The Value is expected to begin on a word boundary; the corresponding Value
// Offset will thus be an even number. This file offset may point anywhere in
// the file, even after the image data.

// Tiff byte codes for byte ordering
const (
	littleEndian = 0x4949
	bigEndian    = 0x4D4D
)

// TIFF File Identifier
//
// From the Tiff 6.0 Specification (p.13)
//
// An arbitrary but carefully chosen number (42) that further identifies the
// file as a TIFF file.
const (
	tiffIdentifier uint16 = 42
)

// From the Tiff 6.0 Specification (p.16)
type fieldType uint16

const (
	NONE      fieldType = 0  // None      =  Default for no field type
	BYTE      fieldType = 1  // BYTE      =  8-bit unsigned integer.
	ASCII     fieldType = 2  // ASCII     =  8-bit byte that contains a 7-bit ASCII code; the last byte must be NULL (binary zero).
	SHORT     fieldType = 3  // SHORT     =  16-bit (2-byte) unsigned integer.
	LONG      fieldType = 4  // LONG      =  32-bit (4-byte) unsigned integer.
	RATIONAL  fieldType = 5  // RATIONAL  =  Two LONGS: the first represents the numerator of a
	SBYTE     fieldType = 6  // SBYTE     =  An 8-bit signed (twos-complement) integer.
	UNDEFINED fieldType = 7  // UNDEFINED =  An 8-bit byte that may contain anything, depending on the definition of the field.
	SSHORT    fieldType = 8  // SSHORT    =  A 16-bit (2-byte) signed (twos-complement) integer.
	SLONG     fieldType = 9  // SLONG     =  A 32-bit (4-byte) signed (twos-complement) integer.
	SRATIONAL fieldType = 10 // SRATIONAL =  Two SLONG’s: the first represents the numerator of a fraction, the second the denominator.
	FLOAT     fieldType = 11 // FLOAT     =  Single precision (4-byte) IEEE format.
	DOUBLE    fieldType = 12 // DOUBLE    =  Double precision (8-byte) IEEE format
)

const (
	zeroByte  = 0
	oneByte   = 1
	twoByte   = 2
	fourByte  = 4
	eightByte = 8
)

// fieldTypeLen is the length of every field type in bytes
var fieldTypeLen = [...]uint32{
	zeroByte, oneByte, oneByte, twoByte,
	fourByte, eightByte, oneByte, oneByte,
	twoByte, fourByte, eightByte, fourByte, eightByte,
}

// bytes returns the number of bytes in each data type
//
// returns 0 if unrecognized
func (f fieldType) bytes() uint32 {
	if f == 0 || int(f) > len(fieldTypeLen) {
		return fieldTypeLen[0]
	}
	return fieldTypeLen[int(f)]
}

var fieldTypeToLabel = map[fieldType]string{
	BYTE:      "BYTE",
	ASCII:     "ASCII",
	SHORT:     "SHORT",
	LONG:      "LONG",
	RATIONAL:  "RATIONAL",
	SBYTE:     "SBYTE",
	UNDEFINED: "UNDEFINED",
	SSHORT:    "SSHORT",
	SLONG:     "SLONG",
	SRATIONAL: "SRATIONAL",
	FLOAT:     "FLOAT",
	DOUBLE:    "DOUBLE",
}

func (f fieldType) String() string {
	v, ok := fieldTypeToLabel[f]
	if !ok {
		return fmt.Sprintf("unrecognized field type %d", f)
	}
	return v
}

// Tag contains TIFF/GeoTIFF image tags
//
// Currently only the necessary tags are included for parsing tile formatted GeoTIFF,
type Tag uint16

//nolint:unused
const (
	ImageWidth                Tag = 256 // ImageWidth
	ImageLength               Tag = 257 // ImageLength
	BitsPerSample             Tag = 258 // BitsPerSample
	Compression               Tag = 259 // Compression
	PhotometricInterpretation Tag = 262 // PhotometricInterpretation
	FillOrder                 Tag = 266 // FillOrder
	StripOffsets              Tag = 273 // StripOffsets
	SamplesPerPixel           Tag = 277 // SamplesPerPixel
	RowsPerStrip              Tag = 278 // RowsPerStrip
	StripByteCounts           Tag = 279 // StripByteCounts
	PlanarConfiguration       Tag = 284 // PlanarConfiguration
	T4Options                 Tag = 292 // CCITT Group 3 options, a set of 32 flag bits.
	T6Options                 Tag = 293 // CCITT Group 4 options, a set of 32 flag bits.
	TileWidth                 Tag = 322 // TileWidth
	TileLength                Tag = 323 // TileLength
	TileOffsets               Tag = 324 // TileOffsets
	TileByteCounts            Tag = 325 // TileByteCounts
	XResolution               Tag = 282 // XResolution
	YResolution               Tag = 283 // YResolution
	ResolutionUnit            Tag = 296 // ResolutionUnit
	Predictor                 Tag = 317 // Predictor
	ColorMap                  Tag = 320 // ColorMap
	ExtraSamples              Tag = 338 // ExtraSamples
	SampleFormat              Tag = 339 // SampleFormat

	// GeoTIFF Specific Tags
	GeoKeyDirectory     Tag = 34735
	GeoDoubleParams     Tag = 34736
	GeoASCIIParams      Tag = 34737
	ModelPixelScale     Tag = 33550
	ModelTiepoint       Tag = 33922
	ModelTransformation Tag = 34264
)

var tagToLabel = map[Tag]string{
	ImageWidth:                "ImageWidth",
	ImageLength:               "ImageLength",
	BitsPerSample:             "BitsPerSample",
	Compression:               "Compression",
	PhotometricInterpretation: "PhotometricInterpretation",
	FillOrder:                 "FillOrder",
	StripOffsets:              "StripOffsets",
	SamplesPerPixel:           "SamplesPerPixel",
	RowsPerStrip:              "RowsPerStrip",
	StripByteCounts:           "StripByteCounts",
	T4Options:                 "T4Options",
	T6Options:                 "T6Options",
	TileWidth:                 "TileWidth",
	TileLength:                "TileLength",
	TileOffsets:               "TileOffsets",
	TileByteCounts:            "TileByteCounts",
	XResolution:               "XResolution",
	YResolution:               "YResolution",
	ResolutionUnit:            "ResolutionUnit",
	Predictor:                 "Predictor",
	ColorMap:                  "ColorMap",
	ExtraSamples:              "ExtraSamples",
	SampleFormat:              "SampleFormat",
	PlanarConfiguration:       "PlanarConfiguration",
	GeoKeyDirectory:           "GeoKeyDirectory",
	GeoDoubleParams:           "GeoDoubleParams",
	GeoASCIIParams:            "GeoAsciiParams",
	ModelPixelScale:           "ModelPixelScale",
	ModelTiepoint:             "ModelTiepoint",
	ModelTransformation:       "ModelTransformation",
}

func (t Tag) String() string {
	v, ok := tagToLabel[t]
	if !ok {
		return fmt.Sprintf("%d", t)
	}
	return v
}

//nolint:unused
var tagToLen = map[Tag]uint32{
	ImageWidth:                1,
	ImageLength:               1,
	BitsPerSample:             1,
	Compression:               1,
	PhotometricInterpretation: 1,
	FillOrder:                 1,
	StripOffsets:              0,
	SamplesPerPixel:           1,
	RowsPerStrip:              1,
	StripByteCounts:           0,
	T4Options:                 0,
	T6Options:                 0,
	TileWidth:                 1,
	TileLength:                1,
	TileOffsets:               0,
	TileByteCounts:            0,
	XResolution:               1,
	YResolution:               1,
	ResolutionUnit:            0,
	Predictor:                 0,
	ColorMap:                  0,
	ExtraSamples:              0,
	SampleFormat:              0,
	PlanarConfiguration:       0,
	GeoKeyDirectory:           0,
	GeoDoubleParams:           0,
	GeoASCIIParams:            0,
	ModelPixelScale:           0,
	ModelTiepoint:             0,
	ModelTransformation:       0,
}

//nolint:unused
type photometricInterpretation uint32

//nolint:unused
const (
	whiteIsZero photometricInterpretation = 0
	blackIsZero photometricInterpretation = 1
	rGB         photometricInterpretation = 2
	paletted    photometricInterpretation = 3
	transMask   photometricInterpretation = 4 // transparency mask
	cMYK        photometricInterpretation = 5
	yCbCr       photometricInterpretation = 6
	cIELab      photometricInterpretation = 8
)
