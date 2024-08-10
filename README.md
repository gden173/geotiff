# Geotiff

[![Go Reference](https://pkg.go.dev/badge/github.com/gden173/geotiff.svg)](https://pkg.go.dev/github.com/gden173/geotiff)
`main` ![main ci status](https://github.com/gden173/geotiff/actions/workflows/go.yml/badge.svg?branch=main)

<!--toc:start-->
- [Geotiff](#geotiff)
  - [Installation](#installation)
  - [Example](#example)
  - [Licence](#licence)
<!--toc:end-->


A golang geotiff parsing library.  This is a pure golang implementation with no
dependencies on gdal or C compilation. This is meant to implement a relatively
small subset of gdal. However, more features may be implemented in the future.

## Installation

```bash
go get github.com/gden173/geotiff@latest
```


## Example 

An example of reading in a tiled geotiff is located in the `main.go` file.

```go
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
```

```
Upper Left   ( 114.0000000,  -11.0000000 )
Lower Left   ( 114.0000000,  -44.0000000 )
Upper Right  ( 153.9000000,  -11.0000000 )
Lower Right  ( 153.9000000,  -44.0000000 )

```

## Licence 

 - [MIT](LICENCE)
