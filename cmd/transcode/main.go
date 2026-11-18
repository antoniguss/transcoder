package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/antoniguss/transcoder/main/internal/transcode"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer, args []string) error {
	fs := flag.NewFlagSet("transcode", flag.ContinueOnError)

	var spec transcode.TranscoderSpec
	fs.StringVar(&spec.Input, "i", "", "input file path")
	fs.StringVar(&spec.Output, "o", "", "output file path")
	fs.IntVar(&spec.Width, "w", 0, "output width in pixels (0 = auto)")
	fs.IntVar(&spec.Height, "h", 0, "output height in pixels (0 = auto)")
	fs.IntVar(&spec.Bitrate, "b", 0, "output bitrate (0 = auto)")
	fs.StringVar((*string)(&spec.Preset), "p", "", "preset (e.g fast, slow)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if err := spec.Validate(); err != nil {
		fs.Usage()
		return fmt.Errorf("invalid arguments: %w", err)
	}

	return transcode.Run(ctx, spec, func(p transcode.Progress) {
		if f, ok := p.Fraction(); ok {
			fmt.Fprintf(w, "%.0f%%\n", f*100)
		} else {
			fmt.Fprintf(w, "processed %s\n", p.Processed.Round(time.Second))
		}
	})
}
