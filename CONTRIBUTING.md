# Contributing

Contributions are welcome — this is a small, focused tool, so keep changes proportional to that.

- **Adding a league**: append an entry to `internal/leagues/All` in `internal/leagues/leagues.go` with the league's ESPN slug (find it by swapping the league segment in `https://site.api.espn.com/apis/site/v2/sports/soccer/{slug}/scoreboard`), a display name, and a short `Cmd` (used for `--league` in quick mode).
- **Tests**: this project follows strict TDD for pure/testable logic (parsing, formatting, sorting, argument parsing). Write the test first, watch it fail, then implement. Bubbletea view rendering itself doesn't need unit tests, but any pure helper it depends on does.
- **Before opening a PR**, make sure these all pass:
  ```sh
  go build ./...
  go vet ./...
  go test ./...
  gofmt -l .   # should print nothing
  ```
  CI runs the same checks on every PR and blocks merging until they're green.
- Keep PRs focused — no unrelated refactors or abstractions beyond what the change needs.
