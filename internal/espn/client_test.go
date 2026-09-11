package espn

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const fixtureJSON = `{
  "events": [
    {
      "id": "123",
      "name": "Boca Juniors at River Plate",
      "shortName": "BOC @ RIV",
      "date": "2026-09-10T20:00Z",
      "competitions": [
        {
          "status": {
            "clock": 2730,
            "displayClock": "45:30",
            "type": {
              "state": "in",
              "completed": false,
              "shortDetail": "45:30 - 1st Half"
            }
          },
          "competitors": [
            {
              "homeAway": "home",
              "score": "2",
              "team": { "displayName": "River Plate", "abbreviation": "RIV" }
            },
            {
              "homeAway": "away",
              "score": "1",
              "team": { "displayName": "Boca Juniors", "abbreviation": "BOC" }
            }
          ],
          "details": [
            {
              "type": { "text": "Goal" },
              "clock": { "displayValue": "23'" },
              "scoringPlay": true,
              "redCard": false,
              "athletesInvolved": [{ "shortName": "J. Alvarez" }]
            }
          ]
        }
      ]
    }
  ]
}`

func TestFetch(t *testing.T) {
	t.Run("parses events, scores and status from a valid response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("dates") != "20260910" {
				t.Errorf("expected dates query param 20260910, got %q", r.URL.Query().Get("dates"))
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fixtureJSON))
		}))
		defer server.Close()

		original := baseURL
		baseURL = server.URL
		defer func() { baseURL = original }()

		resp, err := Fetch("arg.1", "20260910")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(resp.Events))
		}

		event := resp.Events[0]
		if event.ShortName != "BOC @ RIV" {
			t.Errorf("expected shortName %q, got %q", "BOC @ RIV", event.ShortName)
		}
		if len(event.Competitions) != 1 {
			t.Fatalf("expected 1 competition, got %d", len(event.Competitions))
		}

		comp := event.Competitions[0]
		if comp.Status.Type.State != "in" {
			t.Errorf("expected state %q, got %q", "in", comp.Status.Type.State)
		}
		if comp.Status.Type.ShortDetail != "45:30 - 1st Half" {
			t.Errorf("expected shortDetail %q, got %q", "45:30 - 1st Half", comp.Status.Type.ShortDetail)
		}

		if len(comp.Competitors) != 2 {
			t.Fatalf("expected 2 competitors, got %d", len(comp.Competitors))
		}
		home := comp.Competitors[0]
		if home.HomeAway != "home" || home.Score != "2" || home.Team.Abbreviation != "RIV" {
			t.Errorf("unexpected home competitor: %+v", home)
		}
		away := comp.Competitors[1]
		if away.HomeAway != "away" || away.Score != "1" || away.Team.Abbreviation != "BOC" {
			t.Errorf("unexpected away competitor: %+v", away)
		}

		if len(comp.Details) != 1 {
			t.Fatalf("expected 1 detail, got %d", len(comp.Details))
		}
		detail := comp.Details[0]
		if !detail.ScoringPlay {
			t.Error("expected detail to be a scoring play")
		}
		if detail.Type.Text != "Goal" {
			t.Errorf("expected detail type %q, got %q", "Goal", detail.Type.Text)
		}
		if len(detail.AthletesInvolved) != 1 || detail.AthletesInvolved[0].ShortName != "J. Alvarez" {
			t.Errorf("unexpected athletes involved: %+v", detail.AthletesInvolved)
		}
	})

	t.Run("returns an error on non-2xx status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		original := baseURL
		baseURL = server.URL
		defer func() { baseURL = original }()

		if _, err := Fetch("eng.1", "20260910"); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("returns an error on malformed JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		original := baseURL
		baseURL = server.URL
		defer func() { baseURL = original }()

		if _, err := Fetch("eng.1", "20260910"); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}
