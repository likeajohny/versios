# Versios (vrs) - Design Spec

## Context

There is no existing CLI tool that combines zero-config project type detection, multi-ecosystem version bumping, lock file handling, and safety warnings. Existing tools either require configuration files (Knope, tbump), only support one ecosystem (npm version, cargo-release), or are GitHub bots rather than local CLIs (release-please).

Versios (`vrs`) fills this gap: a single binary you run in any project directory to bump the version with safety checks and lock file awareness.

## Scope

**v1 ecosystems:** JS/TS, PHP, Rust, Go

**Out of scope for v1:** Python, Java, Dart, monorepo workspace support, changelog generation, CI/CD integration.

## CLI Interface

### Commands

```
vrs 2.1.12          # bump to explicit version
vrs major           # 1.2.3 -> 2.0.0
vrs minor           # 1.2.3 -> 1.3.0
vrs patch           # 1.2.3 -> 1.2.4
vrs                 # no args: show current version + ecosystem info
```

### Interactive Flow (example)

```
$ vrs 2.0.0

  Detected: Node.js project (pnpm)
  Current version: 1.5.3
  Target version:  2.0.0

  Warning: This is a MAJOR version bump (1.x -> 2.x). Continue? [Y/n]

  Updated package.json

  Lock file is out of sync. How should we update it?
  > Run `pnpm install` (recommended)
    Edit pnpm-lock.yaml directly
    Skip

  Running pnpm install... done

  Create git commit and tag? [Y/n]
  Created commit: "bump version to 2.0.0"
  Created tag: v2.0.0

  Done! Version bumped from 1.5.3 -> 2.0.0
```

### Safety Warnings

Triggered when:
- **Major version bump** (x.0.0 change)
- **Downgrade** (new version < current version)
- **Pre-release** version (e.g., `2.0.0-rc.1`)

## Architecture

### Plugin-based Ecosystem Detection

Each ecosystem implements a common Go interface:

```go
type Ecosystem interface {
    Name() string                              // "nodejs", "php", "go", "rust"
    Detect(dir string) bool                    // checks for manifest files
    ReadVersion(dir string) (string, error)    // reads current version
    WriteVersion(dir string, version string) error
    LockFileStrategy() LockStrategy            // "run_install", "direct_edit", "none"
    PackageManager(dir string) string          // "npm", "pnpm", "yarn", "composer", "cargo"
}
```

Detection walks all registered ecosystems. If multiple match (monorepo), list them and let the user pick or bump all.

### Ecosystem Details

| Ecosystem | Manifest | Lock files | PM detection | Version location |
|-----------|----------|------------|--------------|-----------------|
| JS/TS | `package.json` | `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock` | Which lock file exists | `version` field in package.json |
| PHP | `composer.json` | `composer.lock` | Always `composer` | `version` field in composer.json |
| Rust | `Cargo.toml` | `Cargo.lock` | Always `cargo` | `version` in `[package]` section |
| Go | `go.mod` | N/A | Always `go` | Git tag only (no file edit) |

### Lock File Strategy

| Ecosystem | Recommended | Direct edit? |
|-----------|-------------|-------------|
| JS (npm) | `npm install` | Yes |
| JS (pnpm) | `pnpm install` | Risky (YAML + checksums) |
| JS (yarn) | `yarn install` | No (binary in yarn 2+) |
| PHP | `composer update --lock` | Not recommended |
| Rust | `cargo update` | No (checksums) |
| Go | N/A | N/A (git tag only) |

### Go Special Case

Go has no standard version file. `go.mod` contains the module path, not the project version. Go projects version via git tags. For Go projects:
- `Detect()` checks for `go.mod`
- `ReadVersion()` reads the latest git tag matching `v*` pattern (e.g., `v1.2.3` → `1.2.3`)
- `WriteVersion()` creates a new git tag (file edits are skipped)
- Lock file strategy is `none`

## Project Structure

```
versios/
├── main.go
├── cmd/
│   └── root.go              # Cobra root command, arg parsing, flow orchestration
├── ecosystem/
│   ├── ecosystem.go          # Interface definition + registry
│   ├── nodejs.go
│   ├── php.go
│   ├── golang.go
│   └── rust.go
├── version/
│   └── semver.go             # Semver parsing, comparison, bump logic
├── go.mod
└── go.sum
```

## Error Handling

- **No manifest found:** "No supported project detected in current directory" + list of what was scanned
- **Invalid version string:** Reject non-semver input (must match `x.y.z` with optional pre-release)
- **Read-only files:** Fail with clear error
- **Multiple ecosystems detected:** List all, let user choose (or `--all`)
- **Version field missing:** Warn and offer to add it
- **Git operations fail:** File changes are preserved, warn that git ops were skipped

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- Standard library for JSON/TOML/YAML parsing where possible
- `github.com/BurntSushi/toml` - for Cargo.toml parsing
- `gopkg.in/yaml.v3` - for pnpm-lock.yaml (if direct edit is implemented)

## Verification

1. Create test projects for each ecosystem (JS with npm, JS with pnpm, PHP, Rust, Go)
2. Run `vrs` with no args in each - should display current version
3. Run `vrs patch` - should bump patch version correctly
4. Run `vrs 2.0.0` from `1.x.x` - should trigger major version warning
5. Run `vrs 0.1.0` from `1.x.x` - should trigger downgrade warning
6. Test lock file update flow for each ecosystem
7. Test git commit + tag creation
8. Test in directory with no manifest - should show helpful error
9. Test in directory with multiple manifests - should list and prompt
