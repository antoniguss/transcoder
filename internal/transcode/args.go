package transcode

import (
	"fmt"
	"strconv"
)

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
