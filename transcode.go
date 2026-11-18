package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type TranscoderSpec struct {
	Input   string
	Output  string
	Width   int    // pixels; 0 = keep aspect ration
	Height  int    // pixels; 0 = keep aspect ration
	Bitrate int    // bits per second; 0 = encoder default
	Preset  Preset //

}

type Preset string

const (
	PresetUltraFast Preset = "ultrafast"
	PresetSuperFast Preset = "superfast"
	PresetVeryFast  Preset = "veryfast"
	PresetFaster    Preset = "faster"
	PresetFast      Preset = "fast"
	PresetMedium    Preset = "medium"
	PresetSlow      Preset = "slow"
	PresetSlower    Preset = "slower"
	PresetVerySlow  Preset = "veryslow"
)

func (p Preset) Validate() bool {
	switch p {
	case "", PresetUltraFast, PresetSuperFast, PresetVeryFast, PresetFaster, PresetFast, PresetMedium, PresetSlow, PresetSlower, PresetVerySlow:
		return true
	}
	return false
}

func (s TranscoderSpec) Validate() error {
	var errs []error

	if s.Input == "" {
		errs = append(errs, errors.New("input file is required"))
	}
	if s.Output == "" {
		errs = append(errs, errors.New("output file is required"))
	}
	if s.Width < 0 {
		errs = append(errs, fmt.Errorf("width must me >= 0, got %d", s.Width))
	}
	if s.Height < 0 {
		errs = append(errs, fmt.Errorf("height must me >= 0, got %d", s.Height))
	}
	if s.Bitrate < 0 {
		errs = append(errs, fmt.Errorf("bitrate must me >= 0, got %d", s.Bitrate))
	}

	if !s.Preset.Validate() {
		errs = append(errs, fmt.Errorf("invalid preset, got %s", s.Preset))
	}

	return errors.Join(errs...)

}

type State string

const (
	StateSubmitted   State = "submitted"
	StateTranscoding State = "transcoding"
	StateDone        State = "done"
)

type Progress struct {
	Processed time.Duration
	Total     time.Duration
}

func (p Progress) Fraction() (float64, bool) {
	if p.Total <= 0 {
		return 0, false
	}

	return min(float64(p.Processed)/float64(p.Total), 1), true
}

func startTranscode(ctx context.Context, spec TranscoderSpec) (err error) {

	args := buildArgs(spec)

	total, err := getDuration(ctx, spec.Input)
	if err != nil {
		log.Printf("probe to get duration: %w", err)
	}

	onProgress := func(processed time.Duration) {
		p := Progress{Processed: processed, Total: total}
		if f, ok := p.Fraction(); ok {
			fmt.Printf("%.0f%%\n", f*100)
		} else {
			fmt.Printf("processed %s\n", processed.Round(time.Second))
		}
	}

	err = runFFmpeg(ctx, args, onProgress)
	if err != nil {
		return
	}

	return
}

func buildArgs(spec TranscoderSpec) (args []string) {

	args = append(args, "-y", "-nostdin", "-loglevel", "error", "-nostats", "-progress", "pipe:1")

	args = append(args, "-i", spec.Input)
	if spec.Bitrate > 0 {
		args = append(args, "-b:v", strconv.Itoa(spec.Bitrate))
	}

	if spec.Width == 0 {
		spec.Width = -2
	}
	if spec.Height == 0 {
		spec.Height = -2
	}

	if spec.Width > 0 || spec.Height > 0 {
		args = append(args, "-vf", fmt.Sprintf("scale=%d:%d", spec.Width, spec.Height))
	}

	if spec.Preset != "" {
		args = append(args, "-preset", string(spec.Preset))
	}

	args = append(args, spec.Output)
	return args
}

func runFFmpeg(ctx context.Context, args []string, onProgress func(time.Duration)) (err error) {

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	log.Printf("cmd: %v\n", cmd)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting ffmpeg: %w", err)
	}

	if err = parseProgress(stdout, onProgress); err != nil {
		log.Printf("progress tracking stopped: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return
}

func parseProgress(stdout io.ReadCloser, onProgress func(time.Duration)) error {
	// If an error happens empty out stdout
	defer io.Copy(io.Discard, stdout)

	scanner := bufio.NewScanner(stdout)
	var latest time.Duration
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if !ok {
			continue
		}

		switch key {
		case "out_time_us":
			if us, err := strconv.ParseInt(value, 10, 64); err == nil {
				latest = time.Duration(us) * time.Microsecond
			}
		case "progress":
			onProgress(latest)
		}
	}

	return scanner.Err()
}

func getDuration(ctx context.Context, input string) (time.Duration, error) {

	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration", "-of", "json",
		input)
	log.Printf("cmd: %v\n", cmd)

	data, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("reading ffprobe: %w", err)
	}

	var out struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return 0, fmt.Errorf("parsing ffprobe output: %w", err)
	}

	secs, err := strconv.ParseFloat(out.Format.Duration, 64)
	if err != nil {
		return 0, err
	}

	return time.Duration(secs * float64(time.Second)), nil

}
