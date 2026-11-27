package transcode

import (
	"errors"
	"fmt"
)

type TranscoderSpec struct {
	Input   string `json:"input"`   // input file path
	Output  string `json:"output"`  // output file path
	Width   int    `json:"width"`   // pixels; 0 = keep aspect ration
	Height  int    `json:"height"`  // pixels; 0 = keep aspect ration
	Bitrate int    `json:"bitrate"` // bits per second; 0 = encoder default
	Preset  Preset `json:"preset"`
}

type HLSSpec struct {
	InputFile  string      `json:"input"`      // Path to input file
	OutputDir  string      `json:"output"`     // Path to output directory
	Renditions []Rendition `json:"renditions"` // Optional: required renditions, default Ladder used if none provided
}

type Rendition struct {
	Name         string
	Width        int
	Height       int
	VideoBitrate int
	AudioBitrate int
}

var Ladder = []Rendition{
	{
		Name:         "top1080p",
		Width:        1920,
		Height:       1080,
		VideoBitrate: 6000_000,
		AudioBitrate: 128_000,
	},
	{
		Name:         "high720p",
		Width:        1280,
		Height:       720,
		VideoBitrate: 4500_000,
		AudioBitrate: 128_000,
	},
	{
		Name:         "low720p",
		Width:        1280,
		Height:       720,
		VideoBitrate: 3000_000,
		AudioBitrate: 128_000,
	},
	{
		Name:         "540p",
		Width:        960,
		Height:       540,
		VideoBitrate: 2000_000,
		AudioBitrate: 128_000,
	},
	{
		Name:         "high432p",
		Width:        768,
		Height:       432,
		VideoBitrate: 1100_000,
		AudioBitrate: 128_000,
	},
	{
		Name:         "low432p",
		Width:        768,
		Height:       432,
		VideoBitrate: 730_000,
		AudioBitrate: 128_000,
	},
	{
		Name:         "234p",
		Width:        416,
		Height:       234,
		VideoBitrate: 730_000,
		AudioBitrate: 128_000,
	},
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
