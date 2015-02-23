package baps3

import (
	"testing"
	"time"
)

func TestUtil(t *testing.T) {
	cases := []struct {
		dur  time.Duration
		want string
	}{
		{
			time.Duration(60 * time.Second),
			"1:00",
		},
		// Test negative durations because what the hell
		{
			time.Duration(-60 * time.Second),
			"-1:00",
		},
		// Make sure no rounding shennanigans are happening
		{
			time.Duration(750 * time.Millisecond),
			"0:00",
		},
		{
			time.Duration(31 * time.Second),
			"0:31",
		},
	}

	for _, c := range cases {
		got := PrettyDuration(c.dur)
		if got != c.want {
			t.Errorf("PrettyDuration(%q) == %q, want %q", c.dur, got, c.want)
		}
	}
}
