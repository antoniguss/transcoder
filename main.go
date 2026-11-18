package main

import (
	"flag"
	"fmt"
	"log"
)

var (
	input   = flag.String("i", "", "input path to video file, relative or absolute")
	output  = flag.String("o", "", "output path to video file, relative or absolute")
	bitrate = flag.Int("b", -1, "bitrate of destination file")
	width   = flag.Int("w", -1, "width of destination file")
	height  = flag.Int("l", -1, "height of destination file")
)

func main() {

	flag.Parse()
	if *input == "" || *output == "" {
		log.Println("Usage: -i <input path> -o <output path>")
		return
	}

	if *width < 0 {
		*width = -1
	}

	if *height < 0 {
		*height = -1
	}

	// Build args
	if err := startJob(*input, *output, *bitrate, *width, *height); err != nil {
		log.Println(fmt.Errorf("main: running job failed: %w"), err)
		return
	}

}
