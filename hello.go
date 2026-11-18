package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	ffmpeg_go "github.com/u2takey/ffmpeg-go"
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

	if err := openInputFile(); err != nil {
		log.Println(fmt.Errorf("main: opening input file failed: %w", err))
		return
	}

	if *width < 0 {
		*width = -1
	}

	if *height < 0 {
		*height = -1
	}

	// Build args
	args := ffmpeg_go.KwArgs{}
	if *bitrate > 0 {
		args["b:v"] = *bitrate
	}

	if *width > 0 || *height > 0 {
		scale := fmt.Sprintf("%d:%d", *width, *height)
		err := ffmpeg_go.Input(*input).
			Filter("scale", ffmpeg_go.Args{scale}).
			Output(*output, args).
			OverWriteOutput().Run()

		if err != nil {
			log.Println(fmt.Errorf("main: transcoding with parameters %v failed: %w"), args, err)
		}
	}

}

func openInputFile() (err error) {
	if _, err = os.Stat(*input); errors.Is(err, os.ErrNotExist) {
		err = fmt.Errorf("main: opening input failed: %w", err)
		return
	}

	return
}
