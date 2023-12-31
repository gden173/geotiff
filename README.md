[![Go Reference](https://pkg.go.dev/badge/github.com/gden173/geotiff.svg)](https://pkg.go.dev/github.com/gden173/geotiff)

<!--toc:start-->
- [Geotiff](#geotiff)
  - [Installation](#installation)
  - [Example](#example)
  - [Licence](#licence)
<!--toc:end-->

# Geotiff

A golang geotiff parsing library.  This is a pure golang implementation with no
dependencies on gdal or C compilation. This is meant to implement a relatively
small subset of gdal. However, more features may be implemented in the future.

## Installation

```bash
go get github.com/gden173/geotiff@latest
```


## Example 

An example of reading in a tiled geotiff is 

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

## Licence 

 - [MIT](LICENCE)
