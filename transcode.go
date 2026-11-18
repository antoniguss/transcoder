package main

import (
	"encoding/json"
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

func startJob(input string, output string, bitrate int, width int, height int) (err error) {
	a, err := ffmpeg_go.Probe(input)
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
	if bitrate > 0 {
		args["b:v"] = bitrate
	}

	stream := ffmpeg_go.Input(input)
	if width > 0 || height > 0 {
		scale := fmt.Sprintf("%d:%d", width, height)
		stream = stream.Filter("scale", ffmpeg_go.Args{scale})
	}
	err = stream.
		Output(output, args).
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
