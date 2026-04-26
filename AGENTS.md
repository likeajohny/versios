# AGENTS.md

This file provides guidance to coding agents working in this repository.

## Project

- Name: `versios` (`vrs`)
- Type: Go CLI
- Purpose: detect the current project ecosystem and bump its version in the right place

## Build And Run

```sh
go build -o vrs .
go build -ldflags "-X github.com/likeajohny/versios/cmd.appVersion=0.1.0" -o vrs .
./vrs
```

## Test Commands

```sh
go test ./...
go test ./ecosystem/
go test -run TestNodeJSPatch ./...
go vet ./...
```

Integration tests in `integration_test.go` build the binary and exercise it against temporary directories and git repositories.

## Architecture

- Entry point: `main.go` -> `cmd.Execute()`
- CLI definition and main workflow: `cmd/root.go`
- Ecosystem detection and manifest updates: `ecosystem/`
- Semantic version parsing and bump resolution: `version/`
- Git commit, tag, and tag-prefix handling: `git/`
- Interactive prompts: `prompt/`

## Main Flow

1. Detect ecosystems from the current working directory with `ecosystem.DetectAll()`.
2. Select one, many, or all ecosystems to update.
3. Use the first selected ecosystem as the reference version.
4. Resolve the requested bump with `version.Resolve()`.
5. Confirm risky operations like major bumps, downgrades, or pre-release targets.
6. Write the new version through each ecosystem's `WriteVersion()`.
7. Offer lock file updates where relevant.
8. Optionally create a git commit and tag, preserving the existing tag prefix convention.

## Conventions

- Preserve file formatting when updating JSON manifests. Shared helpers intentionally use regex replacement instead of reformatting via marshal/unmarshal.
- Go projects do not write a manifest version. Their version source is git tags.
- `--yes` and `-y` must skip prompts for CI-style usage.
- User-facing status output belongs on stderr. Stdout should stay suitable for scripting.
- New ecosystems should implement the `Ecosystem` interface and register themselves via `init()`.

## Agent Notes

- Prefer small, targeted changes. This repository is compact and test coverage is already structured by package.
- When changing ecosystem behavior, check both unit tests in `ecosystem/*_test.go` and integration coverage in `integration_test.go`.
- Avoid introducing formatting churn in manifest-writing code.
