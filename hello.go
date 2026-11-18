package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	input   = flag.String("i", "", "input path to video file, relative or absolute")
	output  = flag.String("o", "", "output path to video file, relative or absolute")
	bitrate = flag.Int("b", -1, "bitrate of destination file")
)

func main() {

	flag.Parse()
	if *input == "" || *output == "" {
		log.Println("Usage: -i <input path> -o <output path>")
		return
	}

	if err := openInputFile(); err != nil {
		log.Println(fmt.Errorf("main: opening input file failed: %w", err))
		return
	}

}

func openInputFile() (err error) {
	if _, err = os.Stat(*input); errors.Is(err, os.ErrNotExist) {
		err = fmt.Errorf("main: opening input failed: %w", err)
		return
	}

	return
}
