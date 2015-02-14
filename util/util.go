package util

import (
	"fmt"
	"math"
	"time"
)

// PrettyDuration pretty-prints a duration in the form minutes:seconds.
// The seconds part is zero-padded; the minutes part is not.
func PrettyDuration(dur time.Duration) string {
	return fmt.Sprintf("%d:%02d", int(dur.Minutes()), int(math.Mod(dur.Seconds(), 60)))
}
