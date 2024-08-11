// Package main
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gden173/geotiff/geotiff"
)

var geoInfo = flag.Bool("info", false, "Print the geotiff file information.")
var geoHelp = flag.Bool("help", false, "Print the help message.")

// usage
//
// Prints the command line arguments usage
func usage() {
	fmt.Fprint(os.Stderr, "Usage: geotiff  [options] <geotiff> \n\n")
	fmt.Fprint(os.Stderr, "Options:\n\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {

	if len(os.Args) == 1 {
		usage()
	}

	flag.Parse()

	if *geoHelp {
		usage()
	}

	geoFile := flag.Arg(0)

	// Check if file exists
	if _, err := os.Stat(geoFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "[ERROR] File %s does not exist\n", geoFile)
		usage()
		os.Exit(1)
	}

	if *geoInfo {
		f, err := os.Open(geoFile)
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
		fmt.Println("GeoTiff Info:")
		fmt.Println("Bounds:")
		fmt.Println(bounds)
		fmt.Println("Stats:")
		fmt.Println(gtiff.Stats())
	}

}
