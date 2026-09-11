# fulbito

A terminal app for live soccer scores, powered by ESPN's public scoreboard API. Browse matches from several leagues in a TUI, or query them straight from a script.

## Install

Requires Go 1.25+.

```sh
git clone <this-repo-url>
cd fulbito
go install .
```

This puts a `fulbito` binary in `$(go env GOBIN)` (usually `~/go/bin` — make sure that's on your `PATH`).

## Usage

### Interactive TUI

```sh
fulbito
```

| Key | Action |
|---|---|
| `Tab` | Switch focus between the league sidebar and the main panel |
| `j`/`k` or `↑`/`↓` (sidebar focused) | Change league — the main panel updates instantly |
| `j`/`k` or `↑`/`↓` (main panel focused) | Move the match cursor |
| `←`/`→` or `h`/`l` (main panel focused) | Change the displayed date |
| `Enter` (sidebar focused) | Move focus to the main panel |
| `Enter` (main panel focused) | Open the selected match's event timeline |
| `Esc`/`b` | Close the match detail popup |
| `q` / `Ctrl+C` | Quit |

The sidebar's first entry, "Partidos del día", merges every tracked league's matches for the day, sorted by kickoff time.

### Quick mode (scriptable, no TUI)

Passing any flag skips the TUI and prints plain text instead:

```sh
fulbito --all                          # today's matches across every tracked league
fulbito --league premier               # today's matches for one league
fulbito --league premier 2 --details   # chronological event timeline for match #2
```

The match index is a bare argument (no `--index` flag needed) and can appear anywhere among the other flags.

## Architecture

```
main.go            entry point: routes to the TUI or quick mode
cli_args.go         quick-mode argument parsing
cli_quick.go         quick-mode plain-text output
internal/espn/       ESPN scoreboard API client and response types
internal/matchdata/    shared match/timeline formatting logic (used by both the TUI and quick mode)
internal/leagues/     the list of tracked leagues
internal/tui/         the Bubbletea interactive app (sidebar, scoreboard, match detail popup)
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).
