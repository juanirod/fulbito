package espn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var baseURL = "https://site.api.espn.com/apis/site/v2/sports/soccer"

var httpClient = &http.Client{Timeout: 10 * time.Second}

type ScoreboardResponse struct {
	Events []Event `json:"events"`
}

type Event struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	ShortName    string        `json:"shortName"`
	Date         string        `json:"date"`
	Competitions []Competition `json:"competitions"`
}

type Competition struct {
	Status      Status       `json:"status"`
	Competitors []Competitor `json:"competitors"`
	Details     []Detail     `json:"details,omitempty"`
}

type Status struct {
	Clock        float64    `json:"clock"`
	DisplayClock string     `json:"displayClock"`
	Type         StatusType `json:"type"`
}

type StatusType struct {
	State       string `json:"state"`
	Completed   bool   `json:"completed"`
	Description string `json:"description"`
	Detail      string `json:"detail"`
	ShortDetail string `json:"shortDetail"`
}

type Competitor struct {
	HomeAway string `json:"homeAway"`
	Score    string `json:"score"`
	Team     Team   `json:"team"`
}

type Team struct {
	ID           string `json:"id"`
	DisplayName  string `json:"displayName"`
	Abbreviation string `json:"abbreviation"`
}

type Detail struct {
	Type             DetailType        `json:"type"`
	Clock            DetailClock       `json:"clock"`
	ScoringPlay      bool              `json:"scoringPlay"`
	RedCard          bool              `json:"redCard"`
	YellowCard       bool              `json:"yellowCard"`
	PenaltyKick      bool              `json:"penaltyKick"`
	OwnGoal          bool              `json:"ownGoal"`
	Team             DetailTeam        `json:"team"`
	AthletesInvolved []AthleteInvolved `json:"athletesInvolved,omitempty"`
}

type DetailTeam struct {
	ID string `json:"id"`
}

type DetailType struct {
	Text string `json:"text"`
}

type DetailClock struct {
	DisplayValue string `json:"displayValue"`
}

type AthleteInvolved struct {
	ShortName string `json:"shortName"`
}

func Fetch(leagueSlug, date string) (*ScoreboardResponse, error) {
	url := fmt.Sprintf("%s/%s/scoreboard?dates=%s", baseURL, leagueSlug, date)

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("espn: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("espn: unexpected status %d", resp.StatusCode)
	}

	var out ScoreboardResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("espn: decode failed: %w", err)
	}

	return &out, nil
}
