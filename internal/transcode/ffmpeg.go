package transcode

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func runFFmpeg(ctx context.Context, args []string, onProgress func(time.Duration)) (err error) {

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

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
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("ffmpeg stopped: %w", ctxErr)
		}

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
