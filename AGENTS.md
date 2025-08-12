# Repository Guidelines

## Project Structure & Module Organization
- Root entrypoint: `main.go` initializes Cobra and CLI.
- Commands: `cmd/` (root in `cmd.go`, subpackages under `cmd/core/`).
- TUI components: `tui/` (e.g., `add_note.go`).
- Go module: `go.mod` with Bubble Tea + Cobra dependencies.

## Build, Test, and Development Commands
- Run locally: `go run .` (starts CLI; add `start-tui` to launch TUI).
- Build binary: `go build -o ./bin/ora-ora-ora .` (create local binary).
- Format: `go fmt ./...` and imports via `goimports` if installed.
- Static checks: `go vet ./...`.
- Tests: `go test ./...` (use `-cover` for coverage), even if none exist yet.

## Coding Style & Naming Conventions
- Use standard Go formatting (`go fmt`); tabs/2-space visual indent.
- Package names: short, all-lowercase, no underscores (e.g., `tui`, `core`).
- Exported identifiers: UpperCamelCase; unexported: lowerCamelCase.
- Files: lowercase with underscores if needed (e.g., `add_note.go`).
- Keep commands cohesive; avoid cross-package circular deps.

## Testing Guidelines
- Framework: standard `testing` package.
- Place tests next to code: `pkg/file_test.go` or `tui/add_note_test.go`.
- Name tests `TestXxx`; table-driven tests preferred for logic.
- Aim to cover parsing and command behavior; run `go test ./... -cover`.

## Commit & Pull Request Guidelines
- Commit style: Conventional Commits (e.g., `feat: ...`, `fix: ...`); project history includes `feat:`.
- Keep PRs focused and small; describe intent, approach, and impact.
- Link related issues; include CLI output or TUI screenshots when UI changes.
- Checklist: passes `go fmt`, `go vet`, and `go test`.

## Architecture & Tips
- Flow: `main.go` → `cmd.NewOraCmd()` (Cobra) → optional `StartTui()`.
- TUI uses Bubble Tea’s model/update/view; keep state immutable where possible.
- Dependencies are managed via Go modules; run `go mod tidy` after changes.
- Future config (e.g., search/AI backends) should use env vars and a small `config` package.

