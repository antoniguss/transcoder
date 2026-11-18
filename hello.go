package main

import (
	"flag"
	"fmt"
)

func main() {

	pathPtr := flag.String("path", "", "A path to the video file. Relative or absolute.")
	bitratePtr := flag.Int("bitrate", -1, "Bitrate of the destination video.")
	destPathPtr := flag.String("destPath", "", "A path to the destination video file. Relative or absolute.")

	flag.Parse()
	fmt.Printf("path: %v\n", *pathPtr)
	fmt.Printf("bitrate: %v\n", *bitratePtr)
	fmt.Printf("destination path: %v\n", *destPathPtr)

	fmt.Println("Hello, world!")
}
