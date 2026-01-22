package retry

import (
	"math"
	"math/rand"
	"time"
)

type BackoffStrategy interface {
	NextDelay(retryCount int) time.Duration
	Name() string
}

type ExponentialBackoff struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Factor       float64
	Jitter       bool
}

func (e *ExponentialBackoff) Name() string { return "exponential" }

func (e *ExponentialBackoff) NextDelay(retryCount int) time.Duration {
	if retryCount == 0 {
		return e.InitialDelay
	}
	
	delay := float64(e.InitialDelay) * math.Pow(e.Factor, float64(retryCount))
	
	if delay > float64(e.MaxDelay) {
		delay = float64(e.MaxDelay)
	}
	
	if e.Jitter {
		jitter := rand.Float64() * 0.2 * delay // ±10% jitter
		delay = delay + jitter - (jitter / 2)
	}
	
	return time.Duration(delay)
}

type LinearBackoff struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Increment    time.Duration
	Jitter       bool
}

func (l *LinearBackoff) Name() string { return "linear" }

func (l *LinearBackoff) NextDelay(retryCount int) time.Duration {
	if retryCount == 0 {
		return l.InitialDelay
	}
	
	delay := l.InitialDelay + (l.Increment * time.Duration(retryCount))
	if delay > l.MaxDelay {
		delay = l.MaxDelay
	}
	
	if l.Jitter {
		jitter := rand.Float64() * 0.1 * float64(delay)
		delay = delay + time.Duration(jitter) - time.Duration(jitter/2)
	}
	
	return delay
}

type FixedBackoff struct {
	Delay  time.Duration
	Jitter bool
}

func (f *FixedBackoff) Name() string { return "fixed" }

func (f *FixedBackoff) NextDelay(retryCount int) time.Duration {
	delay := f.Delay
	
	if f.Jitter {
		jitter := rand.Float64() * 0.2 * float64(delay)
		delay = delay + time.Duration(jitter) - time.Duration(jitter/2)
	}
	
	return delay
}

type NoBackoff struct{}

func (n *NoBackoff) Name() string { return "none" }
func (n *NoBackoff) NextDelay(retryCount int) time.Duration { return 0 }