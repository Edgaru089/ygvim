package util

import "time"

// Clock is a clock to measure elapsed time.
// It is a shorthand for the time.Since() and time.Now() combo.
type Clock struct {
	t time.Time
}

// NewClock returns a new clock.
// It also starts it.
func NewClock() (c *Clock) {
	return &Clock{t: time.Now()}
}

// Restart resets the start time.
// It also returns the elapsed time.
func (c *Clock) Restart() (t time.Duration) {
	t = time.Since(c.t)
	c.t = time.Now()
	return
}

// Elapsed returns the elapsed time.
func (c *Clock) Elapsed() time.Duration {
	return time.Since(c.t)
}
