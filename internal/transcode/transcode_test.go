package transcode_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/antoniguss/transcoder/main/internal/transcode"
)

func TestProgress_Fraction(t *testing.T) {
	tests := []struct {
		name    string             // description of this test case
		p       transcode.Progress // receiver
		want    float64
		wantHas bool
	}{
		{"unknown total", transcode.Progress{Processed: time.Second}, 0, false},
		{"negative total", transcode.Progress{Processed: time.Second, Total: -time.Second * 5}, 0, false},
		{"halfway done", transcode.Progress{Processed: time.Second, Total: time.Second * 2}, 0.5, true},
		{"no progress", transcode.Progress{Processed: 0, Total: time.Second * 5}, 0, true},
		{"past completion", transcode.Progress{Processed: time.Second * 6, Total: time.Second * 5}, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotHas := tt.p.Fraction()
			if got != tt.want || gotHas != tt.wantHas {
				t.Errorf("Fraction() = (%v, %v), want (%v, %v)", got, gotHas, tt.want, tt.wantHas)
			}
		})
	}
}

// makeTestClip generates a small synthetic video and returns its path.
// It skips the calling test with -short or when ffmpeg is unavailable.
func makeTestClip(t *testing.T, seconds int) string {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}

	clip := filepath.Join(t.TempDir(), "in.mp4")
	src := fmt.Sprintf("testsrc=duration=%d:size=320x240:rate=10", seconds)
	cmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", src, "-pix_fmt", "yuv420p", clip)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generating clip: %v\n%s", err, out)
	}
	return clip
}

// probeSize returns the video dimensions of path as "WxH".
func probeSize(t *testing.T, path string) string {
	t.Helper()
	out, err := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0",
		"-show_entries", "stream=width,height", "-of", "csv=p=0:s=x", path).Output()
	if err != nil {
		t.Fatalf("probing %s: %v", path, err)
	}
	return strings.TrimSpace(string(out))
}

func TestRun_InvalidSpec(t *testing.T) {
	called := false
	err := transcode.Run(context.Background(), transcode.TranscoderSpec{}, func(transcode.Progress) {
		called = true
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if called {
		t.Error("onProgress was called; validation should stop Run before ffmpeg starts")
	}
}

func TestRun_Transcodes(t *testing.T) {
	in := makeTestClip(t, 2)
	out := filepath.Join(t.TempDir(), "out.mp4")

	var updates []transcode.Progress
	err := transcode.Run(context.Background(), transcode.TranscoderSpec{Input: in, Output: out},
		func(p transcode.Progress) { updates = append(updates, p) })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if info, err := os.Stat(out); err != nil || info.Size() == 0 {
		t.Fatalf("output missing or empty: %v", err)
	}
	if len(updates) == 0 {
		t.Fatal("onProgress was never called")
	}
	for i := 1; i < len(updates); i++ {
		if updates[i].Processed < updates[i-1].Processed {
			t.Errorf("progress went backwards: %v then %v", updates[i-1].Processed, updates[i].Processed)
		}
	}
	last := updates[len(updates)-1]
	if last.Total < 1900*time.Millisecond || last.Total > 2100*time.Millisecond {
		t.Errorf("Total = %v, want about 2s", last.Total)
	}
}

func TestRun_Scales(t *testing.T) {
	in := makeTestClip(t, 1) // 320x240

	tests := []struct {
		name          string
		width, height int
		want          string
	}{
		{"width only keeps aspect ratio", 160, 0, "160x120"},
		{"height only keeps aspect ratio", 0, 120, "160x120"},
		{"both set", 200, 100, "200x100"},
		{"neither set keeps size", 0, 0, "320x240"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "out.mp4")
			spec := transcode.TranscoderSpec{Input: in, Output: out, Width: tt.width, Height: tt.height}
			if err := transcode.Run(context.Background(), spec, func(transcode.Progress) {}); err != nil {
				t.Fatalf("Run: %v", err)
			}
			if got := probeSize(t, out); got != tt.want {
				t.Errorf("output size = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRun_MissingInput(t *testing.T) {
	makeTestClip(t, 1) // only used for its skip checks
	spec := transcode.TranscoderSpec{
		Input:  filepath.Join(t.TempDir(), "does-not-exist.mp4"),
		Output: filepath.Join(t.TempDir(), "out.mp4"),
	}
	err := transcode.Run(context.Background(), spec, func(transcode.Progress) {})
	if err == nil {
		t.Fatal("expected error for missing input, got nil")
	}
	if !strings.Contains(err.Error(), "No such file") {
		t.Errorf("error should include ffmpeg's stderr, got: %v", err)
	}
}

func TestRun_Cancel(t *testing.T) {
	in := makeTestClip(t, 60)
	out := filepath.Join(t.TempDir(), "out.mp4")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	time.AfterFunc(300*time.Millisecond, cancel)

	start := time.Now()
	spec := transcode.TranscoderSpec{Input: in, Output: out, Preset: transcode.PresetVerySlow}
	err := transcode.Run(ctx, spec, func(transcode.Progress) {})

	if err == nil {
		t.Fatal("expected error after cancellation, got nil")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("Run took %v after cancellation; ffmpeg was not stopped promptly", elapsed)
	}
	// The behaviour the runner needs: tell "aborted" apart from "failed".
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v; want it to wrap context.Canceled", err)
	}
}

func TestRunHLS(t *testing.T) {
	in := makeTestClip(t, 15)
	testDir := t.TempDir()

	ctx := t.Context()

	spec := transcode.HLSSpec{InputFile: in, OutputDir: testDir}
	if err := transcode.RunHLS(ctx, spec, func(p transcode.Progress) {}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	filepath.Walk(testDir, func(path string, info fs.FileInfo, err error) error {
		fmt.Println(path)
		return nil
	})
	fmt.Printf("testDir: %v\n", testDir)

}
