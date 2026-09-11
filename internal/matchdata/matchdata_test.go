package matchdata

import (
	"regexp"
	"testing"

	"fulbito/internal/espn"
)

func TestBuildMatchRow(t *testing.T) {
	t.Run("scheduled match shows dash scores and short detail status", func(t *testing.T) {
		event := espn.Event{
			ShortName: "BOC @ RIV",
			Competitions: []espn.Competition{
				{
					Status: espn.Status{
						Type: espn.StatusType{State: "pre", ShortDetail: "8:00 PM"},
					},
					Competitors: []espn.Competitor{
						{HomeAway: "home", Score: "0", Team: espn.Team{DisplayName: "River Plate"}},
						{HomeAway: "away", Score: "0", Team: espn.Team{DisplayName: "Boca Juniors"}},
					},
				},
			},
		}

		row := BuildMatchRow(event)

		if row.Home != "River Plate" || row.Away != "Boca Juniors" {
			t.Errorf("unexpected team names: %+v", row)
		}
		if row.Score != "- : -" {
			t.Errorf("expected dash score for scheduled match, got %q", row.Score)
		}
		if row.Status != "8:00 PM" {
			t.Errorf("expected status %q, got %q", "8:00 PM", row.Status)
		}
		if row.Events != "" {
			t.Errorf("expected no events, got %q", row.Events)
		}
	})

	t.Run("in-progress match shows live score and key events", func(t *testing.T) {
		event := espn.Event{
			ShortName: "BOC @ RIV",
			Competitions: []espn.Competition{
				{
					Status: espn.Status{
						Type: espn.StatusType{State: "in", ShortDetail: "45:30 - 1st Half"},
					},
					Competitors: []espn.Competitor{
						{HomeAway: "home", Score: "2", Team: espn.Team{DisplayName: "River Plate"}},
						{HomeAway: "away", Score: "1", Team: espn.Team{DisplayName: "Boca Juniors"}},
					},
					Details: []espn.Detail{
						{
							Type:             espn.DetailType{Text: "Goal"},
							Clock:            espn.DetailClock{DisplayValue: "23'"},
							ScoringPlay:      true,
							AthletesInvolved: []espn.AthleteInvolved{{ShortName: "J. Alvarez"}},
						},
						{
							Type:             espn.DetailType{Text: "Red Card"},
							Clock:            espn.DetailClock{DisplayValue: "55'"},
							RedCard:          true,
							AthletesInvolved: []espn.AthleteInvolved{{ShortName: "E. Diaz"}},
						},
					},
				},
			},
		}

		row := BuildMatchRow(event)

		if row.Score != "2 : 1" {
			t.Errorf("expected score %q, got %q", "2 : 1", row.Score)
		}
		if row.Status != "45:30 - 1st Half" {
			t.Errorf("expected status %q, got %q", "45:30 - 1st Half", row.Status)
		}
		want := "23' Goal - J. Alvarez; 55' Red Card - E. Diaz"
		if row.Events != want {
			t.Errorf("expected events %q, got %q", want, row.Events)
		}
	})

	t.Run("non-scoring, non-card details are ignored", func(t *testing.T) {
		event := espn.Event{
			Competitions: []espn.Competition{
				{
					Status: espn.Status{Type: espn.StatusType{State: "in", ShortDetail: "10:00"}},
					Competitors: []espn.Competitor{
						{HomeAway: "home", Score: "0", Team: espn.Team{DisplayName: "Home"}},
						{HomeAway: "away", Score: "0", Team: espn.Team{DisplayName: "Away"}},
					},
					Details: []espn.Detail{
						{Type: espn.DetailType{Text: "Substitution"}},
					},
				},
			},
		}

		row := BuildMatchRow(event)

		if row.Events != "" {
			t.Errorf("expected no events, got %q", row.Events)
		}
	})

	t.Run("missing status detail falls back to TBD", func(t *testing.T) {
		event := espn.Event{
			Competitions: []espn.Competition{
				{
					Competitors: []espn.Competitor{
						{HomeAway: "home", Score: "0", Team: espn.Team{DisplayName: "Home"}},
						{HomeAway: "away", Score: "0", Team: espn.Team{DisplayName: "Away"}},
					},
				},
			},
		}

		row := BuildMatchRow(event)

		if row.Status != "TBD" {
			t.Errorf("expected status %q, got %q", "TBD", row.Status)
		}
	})

	t.Run("event without competitions returns an empty row", func(t *testing.T) {
		row := BuildMatchRow(espn.Event{ShortName: "N/A"})

		if row != (MatchRow{}) {
			t.Errorf("expected zero-value row, got %+v", row)
		}
	})
}

func TestBuildTimeline(t *testing.T) {
	const homeID, awayID = "1", "2"

	t.Run("orders most recent event first", func(t *testing.T) {
		details := []espn.Detail{
			{
				Type: espn.DetailType{Text: "Goal"}, Clock: espn.DetailClock{DisplayValue: "10'"},
				ScoringPlay: true, Team: espn.DetailTeam{ID: homeID},
				AthletesInvolved: []espn.AthleteInvolved{{ShortName: "A"}},
			},
			{
				Type: espn.DetailType{Text: "Goal"}, Clock: espn.DetailClock{DisplayValue: "75'"},
				ScoringPlay: true, Team: espn.DetailTeam{ID: awayID},
				AthletesInvolved: []espn.AthleteInvolved{{ShortName: "B"}},
			},
		}

		rows := BuildTimeline(details, homeID, awayID)

		if len(rows) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(rows))
		}
		if rows[0].Minute != "75'" {
			t.Errorf("expected most recent event (75') first, got %q", rows[0].Minute)
		}
		if rows[1].Minute != "10'" {
			t.Errorf("expected oldest event (10') last, got %q", rows[1].Minute)
		}
	})

	t.Run("orders stoppage time correctly relative to the next half's minutes", func(t *testing.T) {
		details := []espn.Detail{
			{Type: espn.DetailType{Text: "Yellow Card"}, Clock: espn.DetailClock{DisplayValue: "46'"}, YellowCard: true, Team: espn.DetailTeam{ID: homeID}},
			{Type: espn.DetailType{Text: "Yellow Card"}, Clock: espn.DetailClock{DisplayValue: "45+2'"}, YellowCard: true, Team: espn.DetailTeam{ID: homeID}},
		}

		rows := BuildTimeline(details, homeID, awayID)

		if rows[0].Minute != "46'" || rows[1].Minute != "45+2'" {
			t.Errorf("expected 46' before 45+2', got order %q, %q", rows[0].Minute, rows[1].Minute)
		}
	})

	t.Run("goal with assist renders scorer and assist on the same side", func(t *testing.T) {
		details := []espn.Detail{
			{
				Type: espn.DetailType{Text: "Goal"}, Clock: espn.DetailClock{DisplayValue: "23'"},
				ScoringPlay: true, Team: espn.DetailTeam{ID: homeID},
				AthletesInvolved: []espn.AthleteInvolved{{ShortName: "J. Alvarez"}, {ShortName: "L. Messi"}},
			},
		}

		rows := BuildTimeline(details, homeID, awayID)

		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}
		row := rows[0]
		if len(row.Home) != 2 || row.Home[0] != "⚽ J. Alvarez" || row.Home[1] != "🅰️ L. Messi" {
			t.Errorf("unexpected home lines: %+v", row.Home)
		}
		if len(row.Away) != 0 {
			t.Errorf("expected no away lines, got %+v", row.Away)
		}
	})

	t.Run("own goal has no assist even with a second athlete listed", func(t *testing.T) {
		details := []espn.Detail{
			{
				Type: espn.DetailType{Text: "Goal"}, Clock: espn.DetailClock{DisplayValue: "40'"},
				ScoringPlay: true, OwnGoal: true, Team: espn.DetailTeam{ID: awayID},
				AthletesInvolved: []espn.AthleteInvolved{{ShortName: "C. Romero"}, {ShortName: "Should Not Appear"}},
			},
		}

		rows := BuildTimeline(details, homeID, awayID)

		if len(rows[0].Away) != 1 || rows[0].Away[0] != "⚽ (OG) C. Romero" {
			t.Errorf("unexpected own goal row: %+v", rows[0].Away)
		}
	})

	t.Run("penalty scored has no assist even with a second athlete listed", func(t *testing.T) {
		details := []espn.Detail{
			{
				Type: espn.DetailType{Text: "Penalty - Scored"}, Clock: espn.DetailClock{DisplayValue: "60'"},
				ScoringPlay: true, PenaltyKick: true, Team: espn.DetailTeam{ID: homeID},
				AthletesInvolved: []espn.AthleteInvolved{{ShortName: "L. Messi"}, {ShortName: "Should Not Appear"}},
			},
		}

		rows := BuildTimeline(details, homeID, awayID)

		if len(rows[0].Home) != 1 || rows[0].Home[0] != "⚽ (P) L. Messi" {
			t.Errorf("unexpected penalty row: %+v", rows[0].Home)
		}
	})

	t.Run("cards render with their icon on the scoring team's side", func(t *testing.T) {
		details := []espn.Detail{
			{Type: espn.DetailType{Text: "Yellow Card"}, Clock: espn.DetailClock{DisplayValue: "30'"}, YellowCard: true, Team: espn.DetailTeam{ID: homeID}, AthletesInvolved: []espn.AthleteInvolved{{ShortName: "N. Fernandez"}}},
			{Type: espn.DetailType{Text: "Red Card"}, Clock: espn.DetailClock{DisplayValue: "80'"}, RedCard: true, Team: espn.DetailTeam{ID: awayID}, AthletesInvolved: []espn.AthleteInvolved{{ShortName: "E. Diaz"}}},
		}

		rows := BuildTimeline(details, homeID, awayID)

		if len(rows) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(rows))
		}
		if rows[0].Away[0] != "🟥 E. Diaz" {
			t.Errorf("expected red card row first (most recent), got %+v", rows[0])
		}
		if rows[1].Home[0] != "🟨 N. Fernandez" {
			t.Errorf("expected yellow card row second, got %+v", rows[1])
		}
	})

	t.Run("unrecognized event type falls back to its raw text instead of disappearing", func(t *testing.T) {
		details := []espn.Detail{
			{Type: espn.DetailType{Text: "Substitution"}, Clock: espn.DetailClock{DisplayValue: "65'"}, Team: espn.DetailTeam{ID: homeID}},
		}

		rows := BuildTimeline(details, homeID, awayID)

		if len(rows) != 1 || len(rows[0].Home) != 1 || rows[0].Home[0] != "Substitution" {
			t.Errorf("expected raw fallback text, got %+v", rows)
		}
	})

	t.Run("no details produces no rows", func(t *testing.T) {
		rows := BuildTimeline(nil, homeID, awayID)
		if len(rows) != 0 {
			t.Errorf("expected 0 rows, got %d", len(rows))
		}
	})
}

func TestFormatKickoff(t *testing.T) {
	t.Run("valid ISO date formats as HH:MM", func(t *testing.T) {
		got := FormatKickoff("2026-09-11T19:00Z")
		if !regexp.MustCompile(`^\d{2}:\d{2}$`).MatchString(got) {
			t.Errorf("expected HH:MM format, got %q", got)
		}
	})

	t.Run("unparseable date falls back to placeholder", func(t *testing.T) {
		got := FormatKickoff("not-a-date")
		if got != "--:--" {
			t.Errorf("expected placeholder, got %q", got)
		}
	})
}

func TestBuildMatchRowKickoff(t *testing.T) {
	event := espn.Event{
		Date: "2026-09-11T19:00Z",
		Competitions: []espn.Competition{
			{
				Competitors: []espn.Competitor{
					{HomeAway: "home", Team: espn.Team{DisplayName: "Home"}},
					{HomeAway: "away", Team: espn.Team{DisplayName: "Away"}},
				},
			},
		},
	}

	row := BuildMatchRow(event)

	if !regexp.MustCompile(`^\d{2}:\d{2}$`).MatchString(row.Kickoff) {
		t.Errorf("expected row.Kickoff in HH:MM format, got %q", row.Kickoff)
	}
}

func TestSortEventsByKickoff(t *testing.T) {
	t.Run("sorts ascending by kickoff time", func(t *testing.T) {
		events := []espn.Event{
			{Name: "20:00", Date: "2026-09-11T20:00Z"},
			{Name: "14:00", Date: "2026-09-11T14:00Z"},
			{Name: "17:00", Date: "2026-09-11T17:00Z"},
		}

		SortEventsByKickoff(events)

		want := []string{"14:00", "17:00", "20:00"}
		for i, w := range want {
			if events[i].Name != w {
				t.Errorf("position %d: expected %q, got %q", i, w, events[i].Name)
			}
		}
	})

	t.Run("unparseable dates don't panic and keep a stable order", func(t *testing.T) {
		events := []espn.Event{
			{Name: "b", Date: "not-a-date"},
			{Name: "a", Date: "also-not-a-date"},
		}

		SortEventsByKickoff(events)

		if len(events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(events))
		}
	})
}

func TestHomeAway(t *testing.T) {
	comp := espn.Competition{
		Competitors: []espn.Competitor{
			{HomeAway: "away", Team: espn.Team{DisplayName: "Away Team"}},
			{HomeAway: "home", Team: espn.Team{DisplayName: "Home Team"}},
		},
	}

	home, away := HomeAway(comp)

	if home.Team.DisplayName != "Home Team" {
		t.Errorf("expected home team %q, got %q", "Home Team", home.Team.DisplayName)
	}
	if away.Team.DisplayName != "Away Team" {
		t.Errorf("expected away team %q, got %q", "Away Team", away.Team.DisplayName)
	}
}
