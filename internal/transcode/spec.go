package transcode

import (
	"errors"
	"fmt"
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
