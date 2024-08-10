# Geotiff

[![Go Reference](https://pkg.go.dev/badge/github.com/gden173/geotiff.svg)](https://pkg.go.dev/github.com/gden173/geotiff)
`main` ![main ci status](https://github.com/gden173/geotiff/actions/workflows/go.yml/badge.svg?branch=main)

<!--toc:start-->
- [Geotiff](#geotiff)
  - [Installation](#installation)
  - [Command Line Example](#command-line-example)
  - [Library Example](#library-example)
  - [Licence](#licence)
<!--toc:end-->


A golang geotiff parsing library.  This is a pure golang implementation with no
dependencies on gdal or C compilation. This is meant to implement a relatively
small subset of gdal. However, more features may be implemented in the future.

## Installation

```bash
go get github.com/gden173/geotiff@latest
```

## Command Line Example 

The command line tool `geotiff` can be used to read in a geotiff file and print 
details about the geotiff file. 

```bash
go build -o geotiff cmd/geotiff/main.go
./geotiff -info geotiff/testdata/WCSServer.tif

# GeoTiff Info
# Bounds:
# Upper Left   ( 135.0000000,  -20.0000000 )
# Lower Left   ( 135.0000000,  -25.3000000 )
# Upper Right  ( 140.0000000,  -20.0000000 )
# Lower Right  ( 140.0000000,  -25.3000000 )
#
# Stats:
# Minimum=46.557, Maximum=942.159, Mean=234.397, StdDev=106.603

```

## Library Example 

An example of reading in a tiled geotiff is located in the `main.go` file.

```go
package main

import (
	"fmt"
	"os"

	"github.com/gden173/geotiff/geotiff"
)

func main() {
	f, err := os.Open("geotiff/testdata/WCSServer.tif")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// read the geotiff
	gtiff, err := geotiff.Read(f)
    if err != nil {
        panic(err)
    }

	// get the geotiff bounds
	bounds, err := gtiff.Bounds()
	if err != nil {
		panic(err)
	}
	fmt.Println(bounds)
}

// Upper Left   ( 114.0000000,  -11.0000000 )
// Lower Left   ( 114.0000000,  -44.0000000 )
// Upper Right  ( 153.9000000,  -11.0000000 )
// Lower Right  ( 153.9000000,  -44.0000000 )

```

## Licence 

 - [MIT](LICENCE)
