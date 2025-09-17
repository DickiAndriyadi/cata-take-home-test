package util

import "time"

type BackoffStrategy interface {
	Duration(attempt int) time.Duration
}

type ExponentialBackoff struct {
	Initial time.Duration
	Max     time.Duration
}

func NewExponentialBackoff(initial, max time.Duration) *ExponentialBackoff {
	return &ExponentialBackoff{
		Initial: initial,
		Max:     max,
	}
}

func (b *ExponentialBackoff) Duration(attempt int) time.Duration {
	dur := b.Initial << attempt // initial * 2^attempt
	if dur > b.Max {
		return b.Max
	}
	return dur
}
