package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"fulbito/internal/espn"
	"fulbito/internal/leagues"
	"fulbito/internal/matchdata"
)

func runQuick(args []string) {
	q, err := parseQuickArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fulbito:", err)
		os.Exit(1)
	}

	date := time.Now().Format("20060102")

	switch {
	case q.all:
		printAllLeagues(date)
	case q.details:
		printMatchDetails(q.league, date, *q.index)
	default:
		printLeagueMatches(q.league, date)
	}
}

func findLeague(cmd string) (leagues.League, bool) {
	for _, l := range leagues.All {
		if l.Cmd == cmd {
			return l, true
		}
	}
	return leagues.League{}, false
}

func validLeagueCmds() string {
	cmds := make([]string, len(leagues.All))
	for i, l := range leagues.All {
		cmds[i] = l.Cmd
	}
	return strings.Join(cmds, ", ")
}

func requireLeague(cmd string) leagues.League {
	league, ok := findLeague(cmd)
	if !ok {
		fmt.Fprintf(os.Stderr, "fulbito: unknown league %q; valid leagues: %s\n", cmd, validLeagueCmds())
		os.Exit(1)
	}
	return league
}

func fetchOrExit(slug, date string) *espn.ScoreboardResponse {
	resp, err := espn.Fetch(slug, date)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fulbito:", err)
		os.Exit(1)
	}
	return resp
}

func printMatchLines(events []espn.Event) {
	if len(events) == 0 {
		fmt.Println("  (no matches today)")
		return
	}
	for i, event := range events {
		row := matchdata.BuildMatchRow(event)
		fmt.Printf("%d) %s  %s %s %s (%s)\n", i+1, row.Kickoff, row.Home, row.Score, row.Away, row.Status)
	}
}

func printLeagueMatches(cmd, date string) {
	league := requireLeague(cmd)
	resp := fetchOrExit(league.Slug, date)
	matchdata.SortEventsByKickoff(resp.Events)
	printMatchLines(resp.Events)
}

func printAllLeagues(date string) {
	for _, league := range leagues.All {
		fmt.Printf("%s (%s)\n", league.Name, league.Cmd)
		resp, err := espn.Fetch(league.Slug, date)
		if err != nil {
			fmt.Printf("  error: %v\n", err)
		} else {
			matchdata.SortEventsByKickoff(resp.Events)
			printMatchLines(resp.Events)
		}
		fmt.Println()
	}
}

func printMatchDetails(cmd, date string, index int) {
	league := requireLeague(cmd)
	resp := fetchOrExit(league.Slug, date)
	matchdata.SortEventsByKickoff(resp.Events)

	if index < 1 || index > len(resp.Events) {
		fmt.Fprintf(os.Stderr, "fulbito: match index %d not found (only %d matches today)\n", index, len(resp.Events))
		os.Exit(1)
	}

	event := resp.Events[index-1]
	if len(event.Competitions) == 0 {
		fmt.Println("No hay eventos detallados disponibles.")
		return
	}
	comp := event.Competitions[0]
	home, away := matchdata.HomeAway(comp)

	homeScore, awayScore := home.Score, away.Score
	if comp.Status.Type.State == "pre" {
		homeScore, awayScore = "-", "-"
	}

	fmt.Printf("%s vs %s — %s - %s\n", home.Team.DisplayName, away.Team.DisplayName, homeScore, awayScore)

	rows := matchdata.BuildTimeline(comp.Details, home.Team.ID, away.Team.ID)
	if len(rows) == 0 {
		fmt.Println("No hay eventos detallados disponibles.")
		return
	}

	for _, row := range rows {
		lines := row.Home
		if len(row.Away) > 0 {
			lines = row.Away
		}
		for i, line := range lines {
			minute := row.Minute
			if i > 0 {
				minute = ""
			}
			fmt.Printf("%-5s %s\n", minute, line)
		}
	}
}
