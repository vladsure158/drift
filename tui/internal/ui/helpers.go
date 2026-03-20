package ui

import (
	"fmt"
	"strings"
	"time"
)

func TimeSince(iso string) string {
	if iso == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		// Try without fractional seconds
		t, err = time.Parse("2006-01-02T15:04:05Z", iso)
		if err != nil {
			return ""
		}
	}
	diff := time.Since(t)
	m := int(diff.Minutes())
	if m < 1 {
		return "now"
	}
	if m < 60 {
		return fmt.Sprintf("%dm", m)
	}
	h := m / 60
	if h < 24 {
		return fmt.Sprintf("%dh", h)
	}
	d := h / 24
	if d < 7 {
		return fmt.Sprintf("%dd", d)
	}
	w := d / 7
	if w < 5 {
		return fmt.Sprintf("%dw", w)
	}
	return fmt.Sprintf("%dmo", d/30)
}

func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	return s[:max-1] + "…"
}

func PadRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func Clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
