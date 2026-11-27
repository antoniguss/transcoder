package transcode

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func probeDuration(ctx context.Context, input string) (time.Duration, error) {

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

func probeStreams(ctx context.Context, input string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-loglevel", "error",
		"-show_entries", "stream=codec_type",
		"-of", "csv=p=0",
		input)
	log.Printf("cmd: %v\n", cmd)

	var streams []string
	data, err := cmd.Output()
	if err != nil {
		return streams, fmt.Errorf("probing for streams: %w", err)
	}

	return strings.Fields(string(data)), nil

}
