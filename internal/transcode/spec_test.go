package transcode_test

import (
	"testing"

	"github.com/antoniguss/transcoder/main/internal/transcode"
)

func TestPreset_Validate(t *testing.T) {
	tests := []struct {
		input string // description of this test case
		want  bool
	}{
		{"ultrafast", true},
		{"superfast", true},
		{"veryfast", true},
		{"faster", true},
		{"fast", true},
		{"medium", true},
		{"slow", true},
		{"slower", true},
		{"veryslow", true},
		{"", true},

		{"quick", false},
		{"fastest", false},
		{"speed", false},
		{"low", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := transcode.Preset(tt.input)
			got := p.Validate()
			if got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
