// Package main
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
	gtiff, _ := geotiff.Read(f)

	// get the geotiff bounds
	bounds, err := gtiff.Bounds()
	if err != nil {
		panic(err)
	}
	fmt.Println(bounds)
}
