package main

import "fmt"
import "time"
import "math"

func PrettyDuration(dur time.Duration) string {
	return fmt.Sprintf("%02d:%02d", int(dur.Minutes()), int(math.Mod(dur.Seconds(), 60)))
}
