package tui

import (
	"testing"
	"time"
)

func TestShiftAndFormatDate(t *testing.T) {
	cases := []struct {
		name string
		base time.Time
		days int
		want string
	}{
		{"same day", time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC), 0, "20260911"},
		{"next day", time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC), 1, "20260912"},
		{"previous day", time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC), -1, "20260910"},
		{"month rollover forward", time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC), 1, "20261001"},
		{"month rollover backward", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), -1, "20260930"},
		{"year rollover forward", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), 1, "20270101"},
		{"year rollover backward", time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), -1, "20261231"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := shiftAndFormatDate(c.base, c.days)
			if got != c.want {
				t.Errorf("shiftAndFormatDate(%v, %d) = %q, want %q", c.base, c.days, got, c.want)
			}
		})
	}
}

func TestFormatDateEs(t *testing.T) {
	cases := []struct {
		name string
		date time.Time
		want string
	}{
		{"friday", time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC), "Vie 11 Sep 2026"},
		{"december", time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC), "Vie 25 Dic 2026"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := formatDateEs(c.date)
			if got != c.want {
				t.Errorf("formatDateEs(%v) = %q, want %q", c.date, got, c.want)
			}
		})
	}
}
