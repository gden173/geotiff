// Package main
package main

import (
	"flag"
	"fmt"
	"os"
)

var geoFile = flag.String("file", "", "geotiff file path.")
var geoInfo = flag.String("info", "", "Print the geotiff files information.") 
// var geoHelp = flag.String("help", "", "Print the usage  information and exit.") 

// usage
//
// Prints the command line arguments usage
func usage() {
	fmt.Fprint(os.Stderr, "Usage: geotiff  [options] <geotiff> \n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {

	if len(os.Args) == 1 {
		usage()
	}

	flag.Parse()

	if *geoInfo != "" {
		fmt.Println("info")
	}


}

