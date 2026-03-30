# versios (vrs)

Zero-config CLI tool that detects your project type and bumps the version in the right place.

## Usage

```
vrs                 # show current version and detected ecosystem
vrs patch           # 1.2.3 -> 1.2.4
vrs minor           # 1.2.3 -> 1.3.0
vrs major           # 1.2.3 -> 2.0.0
vrs 2.1.0           # set explicit version
vrs 2.1.0 --yes     # skip all prompts (CI mode)
```

## Supported Ecosystems

| Ecosystem | Manifest | Lock file handling | Version source |
|-----------|----------|-------------------|---------------|
| JS/TS | `package.json` | npm/pnpm/yarn install | `version` field |
| PHP | `composer.json` | `composer update --lock` | `version` field |
| Rust | `Cargo.toml` | `cargo update` | `[package].version` |
| Go | `go.mod` | N/A | Git tags |

Package manager is auto-detected (e.g., `pnpm-lock.yaml` present = pnpm).

## Features

- **Zero config** -- just run `vrs` in any supported project directory
- **Safety warnings** -- alerts on major version bumps, downgrades, and pre-release versions
- **Lock file awareness** -- offers to run the appropriate install command after bumping
- **Git integration** -- optional commit + tag (annotated or lightweight) after bumping
- **Multi-ecosystem** -- if multiple manifests are found, choose which to update or pick "All"
- **CI-friendly** -- `--yes` flag skips all prompts, bumps all detected ecosystems

## Install

```sh
go install github.com/likeajohny/versios@latest
```

Or build from source:

```sh
git clone git@github.com:likeajohny/versios.git
cd versios
go build -o vrs .
```

To include a version number in the binary:

```sh
go build -ldflags "-X github.com/likeajohny/versios/cmd.appVersion=0.1.0" -o vrs .
```

## Example

```
$ vrs 2.0.0

  Detected: nodejs project (pnpm)
  Current version: 1.5.3
  Target version:  2.0.0

  This is a MAJOR version bump (1.x -> 2.x). Continue? Yes / No

  Updated package.json

  Lock file may be out of sync. How should we update it?
  > Run `pnpm install` (recommended)
    Skip

  Running pnpm install... done

  Create git commit and tag? Yes / No

  Tag type?
  > Annotated (recommended)
    Lightweight

  Created commit: "bump version to 2.0.0"
  Created annotated tag: v2.0.0

  Done! Version bumped from 1.5.3 -> 2.0.0
```

## License

MIT
