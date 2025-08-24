# CRUSH.md - Ora-Ora-Ora Project Commands and Guidelines

## Build/Run Commands
```bash
go run .               # Run CLI
go run . start-tui     # Run TUI
go build -o bin/ora-ora-ora .  # Build executable
```

## Test Commands
```bash
go test ./... -cover          # Run all tests with coverage
go test -run TestName ./path  # Run specific test
go test -v ./path             # Run tests with verbose output
```

## Lint/Format Commands
```bash
go fmt ./...       # Format code
go vet ./...       # Vet code
goimports ./...    # Auto-format imports (if available)
```

## Code Style Guidelines

### General
- Use `go fmt` for formatting (Tab indent, visual 2 spaces)
- Package names: short, lowercase, no underscores (e.g., `tui`, `agent`)
- File names: lowercase, underscores if needed (e.g., `add_note.go`)

### Naming Conventions
- Public identifiers: UpperCamelCase
- Private identifiers: lowerCamelCase

### Imports
- Standard library imports first, separated by blank line
- Third-party imports next
- Local imports last
- Use grouped imports with blank lines between groups

### Types
- Prefer structs with clear field names
- Use appropriate Go types (time.Time for timestamps)
- Add JSON tags for serialization
- Follow documented interface contracts (see AGENTS.md)

### Error Handling
- Always handle errors explicitly
- Wrap errors with context when propagating
- Use fmt.Errorf("action failed: %w", err) for wrapping

### TUI Components
- Follow Bubble Tea architecture (model/update/view)
- Keep state immutable
- Update functions return new models

### Agent Layer
- Keep Ollama interactions loosely coupled
- Support non-streaming responses first
- Maintain modularity for future model replacements

### Testing
- Use table-driven tests when possible
- Place test files alongside implementation (`xxx_test.go`)
- Cover key branches and error conditions