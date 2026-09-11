package main

import "testing"

func intPtr(n int) *int { return &n }

func TestParseQuickArgs(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want quickArgs
	}{
		{"all flag", []string{"--all"}, quickArgs{all: true}},
		{"league only", []string{"--league", "eng.1"}, quickArgs{league: "eng.1"}},
		{
			"league, index, then details",
			[]string{"--league", "eng.1", "2", "--details"},
			quickArgs{league: "eng.1", index: intPtr(2), details: true},
		},
		{
			"league, details, then index",
			[]string{"--league", "eng.1", "--details", "2"},
			quickArgs{league: "eng.1", index: intPtr(2), details: true},
		},
		{
			"index before any flag",
			[]string{"2", "--league", "eng.1", "--details"},
			quickArgs{league: "eng.1", index: intPtr(2), details: true},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := parseQuickArgs(c.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.all != c.want.all || got.league != c.want.league || got.details != c.want.details {
				t.Fatalf("got %+v, want %+v", got, c.want)
			}
			if (got.index == nil) != (c.want.index == nil) {
				t.Fatalf("index nilness mismatch: got %+v, want %+v", got, c.want)
			}
			if got.index != nil && *got.index != *c.want.index {
				t.Fatalf("index value mismatch: got %d, want %d", *got.index, *c.want.index)
			}
		})
	}
}

func TestParseQuickArgsErrors(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"bare index without league", []string{"2"}},
		{"details without index", []string{"--league", "eng.1", "--details"}},
		{"all combined with league", []string{"--all", "--league", "eng.1"}},
		{"all combined with index", []string{"--all", "2"}},
		{"league missing value", []string{"--league"}},
		{"non-numeric extra argument", []string{"--league", "eng.1", "foo"}},
		{"two bare index-like arguments", []string{"--league", "eng.1", "1", "2"}},
		{"unknown flag", []string{"--unknown"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := parseQuickArgs(c.args); err == nil {
				t.Fatalf("expected an error for args %v, got none", c.args)
			}
		})
	}
}
