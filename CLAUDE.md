# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```sh
go build -o vrs .                    # build binary
go build -ldflags "-X github.com/likeajohny/versios/cmd.appVersion=0.1.0" -o vrs .  # with version
./vrs                                # run (detects ecosystem in cwd)
```

## Testing

```sh
go test ./...                        # all tests
go test ./ecosystem/                 # single package
go test -run TestNodeJSPatch ./...   # single test by name
go vet ./...                         # static analysis
```

Integration tests (`integration_test.go`) build the binary and run it as a subprocess against temp directories with git repos. They cover multi-ecosystem detection, tag prefix conventions, error cases, and CI mode (`--yes`).

## Architecture

**Entry point:** `main.go` -> `cmd.Execute()` -> single Cobra root command in `cmd/root.go`.

**Core flow in `cmd/root.go:run()`:**
1. `ecosystem.DetectAll()` scans cwd for manifest files
2. User selects which ecosystem(s) to bump (or `--yes` bumps all)
3. First selected ecosystem provides the reference version
4. `version.Resolve()` turns "major"/"minor"/"patch"/explicit string into a `Version`
5. Safety checks (major bump, downgrade, pre-release) with confirmation prompts
6. `eco.WriteVersion()` updates each selected ecosystem's manifest
7. Lock file update offered per ecosystem (runs detected package manager)
8. Git commit + tag (annotated or lightweight, auto-detected prefix convention)

**Packages:**

| Package | Role |
|---------|------|
| `ecosystem` | `Ecosystem` interface + registry pattern. Each ecosystem registers via `init()`. Shared JSON helpers (`readJSONVersion`, `writeJSONVersion`) use regex replacement to preserve file formatting. |
| `version` | Immutable `Version` struct, semver parsing/comparison, `Resolve()` for bump keywords. |
| `git` | Wrappers around git subprocess calls: tag creation, commit, prefix detection (`v` vs plain). |
| `prompt` | Thin wrappers around `huh/v2` TUI: `Confirm`, `Select`, `Input`, `SelectOrAll`. |

**Adding a new ecosystem:** Implement the `Ecosystem` interface in `ecosystem/`, call `Register()` in an `init()` function. The registry auto-discovers it.

**Key conventions:**
- JSON manifest files are modified via regex to preserve formatting (not marshal/unmarshal)
- Go ecosystem has no manifest version field; version comes from git tags only (`WritesFiles()` returns false)
- `--yes` / `-y` flag skips all interactive prompts for CI usage
- All user-facing output goes to stderr; only the version string goes to stdout
