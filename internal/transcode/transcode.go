package transcode

import (
	"context"
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
