package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

type TranscoderSpec struct {
	Input   string
	Output  string
	Width   int // pixels; 0 = keep aspect ration
	Height  int // pixels; 0 = keep aspect ration
	Bitrate int // bits per second; 0 = encoder default

}

func (s TranscoderSpec) Validate() error {
	var errs []error

	if s.Input == "" {
		fmt.Printf("s.Input: %v\n", s.Input)
		errs = append(errs, errors.New("input file is required"))
	}
	if s.Output == "" {
		errs = append(errs, errors.New("input file is required"))
	}
	if s.Width < 0 {
		errs = append(errs, fmt.Errorf("width must me >0, got %d", s.Width))
	}
	if s.Height < 0 {
		errs = append(errs, fmt.Errorf("height must me >0, got %d", s.Height))
	}
	if s.Bitrate < 0 {
		errs = append(errs, fmt.Errorf("bitrate must me >0, got %d", s.Bitrate))
	}

	return errors.Join(errs...)

}

func startTranscode(ctx context.Context, spec TranscoderSpec) (err error) {
	a, err := ffmpeg_go.Probe(spec.Input)
	if err != nil {
		err = fmt.Errorf("transcoder: probing input file failed: %w", err)
		return
	}

	totalDuration, err := probeDuration(a)
	if err != nil {
		err = fmt.Errorf("transcoder: probing for duration of input file failed: %w", err)
		return
	}

	args := ffmpeg_go.KwArgs{}
	if spec.Bitrate > 0 {
		args["b:v"] = spec.Bitrate
	}

	stream := ffmpeg_go.Input(spec.Input)
	if spec.Width > 0 || spec.Height > 0 {
		scale := fmt.Sprintf("%d:%d", spec.Width, spec.Height)
		stream = stream.Filter("scale", ffmpeg_go.Args{scale})
	}
	err = stream.
		Output(spec.Output, args).
		GlobalArgs("-progress", "unix://"+TempSock(totalDuration)).
		OverWriteOutput().Run()

	if err != nil {

		err = fmt.Errorf("transcoder: transcoding with parameters %v failed: %w", args, err)
		return
	}

	return
}
func TempSock(totalDuration float64) string {
	// serve

	sockFileName := path.Join(os.TempDir(), fmt.Sprintf("%d_sock", rand.Int()))
	l, err := net.Listen("unix", sockFileName)
	if err != nil {
		panic(err)
	}

	go func() {
		re := regexp.MustCompile(`out_time_ms=(\d+)`)
		fd, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}
		buf := make([]byte, 16)
		data := ""
		progress := ""
		for {
			_, err := fd.Read(buf)
			if err != nil {
				return
			}
			data += string(buf)
			a := re.FindAllStringSubmatch(data, -1)
			cp := ""
			if len(a) > 0 && len(a[len(a)-1]) > 0 {
				c, _ := strconv.Atoi(a[len(a)-1][len(a[len(a)-1])-1])
				cp = fmt.Sprintf("%.2f", float64(c)/totalDuration/1000000)
			}
			if strings.Contains(data, "progress=end") {
				cp = "done"
			}
			if cp == "" {
				cp = ".0"
			}
			if cp != progress {
				progress = cp
				fmt.Println("progress: ", progress)
			}
		}
	}()

	return sockFileName
}

type probeFormat struct {
	Duration string `json:"duration"`
}

type probeData struct {
	Format probeFormat `json:"format"`
}

func probeDuration(a string) (float64, error) {
	pd := probeData{}
	err := json.Unmarshal([]byte(a), &pd)
	if err != nil {
		return 0, err
	}
	f, err := strconv.ParseFloat(pd.Format.Duration, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}
