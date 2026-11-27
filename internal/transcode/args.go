package transcode

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// Builds args for an ffmpeg command with:
// Input: HLSSpec.InputFile
// At HLSSpec.Output produces a master playlist master.m3u8
// subdirectories for each Rendition.Name with a media playlist media.m3u8 and segments segment<i>.ts
func buildHLSArgs(spec HLSSpec, hasAudio bool) (args []string) {

	args = append(args, "-y", "-nostdin", "-loglevel", "warning", "-nostats", "-progress", "pipe:1")

	args = append(args, "-i", spec.InputFile)

	var bArgs []string
	var mArgs []string
	var streamMap []string
	for i, r := range Ladder {
		bArgs = append(bArgs, fmt.Sprintf("-b:v:%d", i), strconv.Itoa(r.VideoBitrate))
		mArgs = append(mArgs, "-map", "0:v")

		if hasAudio {
			bArgs = append(bArgs, fmt.Sprintf("-b:a:%d", i), strconv.Itoa(r.AudioBitrate))
			mArgs = append(mArgs, "-map", "0:a")
		}

		if hasAudio {
			streamMap = append(streamMap, fmt.Sprintf("v:%d,a:%d,name:%s", i, i, r.Name))
		} else {
			streamMap = append(streamMap, fmt.Sprintf("v:%d,name:%s", i, r.Name))
		}
	}
	args = append(args, bArgs...)
	args = append(args, mArgs...)

	args = append(args, "-f", "hls")
	args = append(args, "-var_stream_map", strings.Join(streamMap, " "))

	args = append(args, "-hls_segment_filename", filepath.Join(spec.OutputDir, "%v", "segment_%06d.ts"))
	args = append(args, "-master_pl_name", "master.m3u8")

	args = append(args, filepath.Join(spec.OutputDir, "%v", "media.m3u8"))

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
