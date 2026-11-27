package transcode

import (
	"context"
	"slices"
	"time"
)

type Progress struct {
	Processed time.Duration
	Total     time.Duration
}

func (p Progress) Fraction() (f float64, has bool) {
	if p.Total <= 0 {
		return 0, false
	}

	return min(float64(p.Processed)/float64(p.Total), 1), true
}

func RunHLS(ctx context.Context, spec HLSSpec, onProgress func(Progress)) error {

	s, err := probeStreams(ctx, spec.InputFile)
	if err != nil {
		return err
	}

	hasAudio := slices.Contains(s, "audio")
	args := buildHLSArgs(spec, hasAudio)

	return runFFmpeg(ctx, args, func(d time.Duration) {})

}

func Run(ctx context.Context, spec TranscoderSpec, onProgress func(Progress)) error {

	if err := spec.Validate(); err != nil {
		return err
	}

	total, err := probeDuration(ctx, spec.Input)
	if err != nil {
		total = 0
	}

	return runFFmpeg(ctx, buildArgs(spec), func(processed time.Duration) {
		onProgress(Progress{Processed: processed, Total: total})
	})
}
