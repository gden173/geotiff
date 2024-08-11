package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/gden173/geotiff/geotiff"
)

var geoHelp = flag.Bool("help", false, "Print the help message.")

// usage
//
// Prints the command line arguments usage
func usage() {
	fmt.Fprint(os.Stderr, "Usage:  [options] <geotiff> <x> <y> \n\n")
	fmt.Fprint(os.Stderr, "Prints the value at the given coordinates\n\n")
	fmt.Fprint(os.Stderr, "Options:\n\n")
	fmt.Fprint(os.Stderr, "  <x> <y>  Coordinates to get the value from\n")
	fmt.Fprint(os.Stderr, "  <geotiff>  Geotiff file to read\n")
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

	if flag.NArg() != 3 {
		fmt.Fprintf(os.Stderr, "[ERROR] Missing arguments\n")
		usage()
	}

	geoFile := flag.Arg(0)
	xs := flag.Arg(1)
	ys := flag.Arg(2)

	// Check if file exists
	if _, err := os.Stat(geoFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "[ERROR] File %s does not exist\n", geoFile)
		usage()
		os.Exit(1)
	}

	// Convert to ints
	x, err := strconv.Atoi(xs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] x value %s is not a number\n", xs)
		return
	}
	y, err := strconv.Atoi(ys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] y value %s is not a number\n", xs)
		return
	}

	f, err := os.Open(geoFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Could not open file %s\n", geoFile)
		return
	}
	defer f.Close()

	g, err := geotiff.Read(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Could not read geotiff file %s\n", geoFile)
		return
	}

	val, err := g.AtPoint(x, y)
	if err != nil {
		fmt.Println(g.ImageHeight())
		fmt.Println(g.ImageWidth())
		panic(fmt.Errorf("%w", err))
	}
	fmt.Printf("Value at coordinates %d, %d: %f\n", x, y, val)
}
