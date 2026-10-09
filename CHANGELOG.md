# Changelog

All notable changes to this project will be documented in this file. (after 0.0.1)

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/en/2.0.0/).

## [0.1.0](https://github.com/flushwhy/fe/compare/fe-v0.0.2...fe-v0.1.0) (2026-10-09)


### Features

* add Chocolatey package metadata ([365850f](https://github.com/flushwhy/fe/commit/365850fdd3b35c8bb52c54b538319078d4712fbd))
* add Chocolatey package metadata ([80cfb11](https://github.com/flushwhy/fe/commit/80cfb1113fce5c12e162172bd7560b46e2a8bacb))
* add Chocolatey package metadata ([f8e6f45](https://github.com/flushwhy/fe/commit/f8e6f45c552c45fec9e244aaf04826d367b9cec6))
* add safe Odin repository installer ([103c12d](https://github.com/flushwhy/fe/commit/103c12d66ecf785488a714eb1a3f161b30ced62b))
* automate Chocolatey packaging on releases ([f42c748](https://github.com/flushwhy/fe/commit/f42c7484b8026d925fb9ed947e721bc90e8e69a0))
* expose install and get commands ([61a805a](https://github.com/flushwhy/fe/commit/61a805a00c34f8ef8ca2c40b6f23a4c6bf668dce))


### Bug Fixes

* build binaries and attach them to GitHub releases ([25c1b68](https://github.com/flushwhy/fe/commit/25c1b68fe95eccb94f971f240b99a79f90e0be6b))
* conditionally publish Chocolatey package using secret ([b1ed148](https://github.com/flushwhy/fe/commit/b1ed1484ada18dd5fab1011a4e537140d5207c0d))

## [Unreleased]

---

## [0.0.4]

### Added
- Full architectural rewrite — all business logic moved to `internal/` packages (`butler`, `transcode`, `pack`, `scaffold`, `build`, `config`, `tty`, `ui`, `tooling`, `doctor`). Cobra commands are now thin wiring only.
- New `fe init <lang> --name <project>` command replacing the old `init`. Supports `odin`, `c`, `cpp`, `zig`, and `go` with per-language templates embedded directly in the binary via `go:embed`.
- New `fe add <tool>` command for wiring up LSP and editor tooling without manual config. Supports `clangd`, `ols`, `zls`, and `gopls`.
  - `fe add clangd` detects your build system (cmake vs bear/make), writes `.clangd` with the correct `CompilationDatabase` path, and creates a `compile_commands.json` stub so clangd starts without errors.
  - `fe add ols` writes or patches `ols.json` for the Odin Language Server.
  - `fe add zls` writes `zls.json` for the Zig Language Server.
  - `fe add gopls` writes `.gopls.yaml` workspace settings for Go.
- New `fe build` command — config-driven compiler wrapper. Define per-platform build commands in `.fe.yaml` under `build.targets`.
- New `fe release` command — orchestrates the full pipeline: `build → pack → bmp` in one shot, with `--skip-build`, `--skip-pack`, `--skip-bmp` flags.
- New `fe validate` command — pre-flight checker for config and paths, CI-safe (non-zero exit on failure).
- New `fe doctor` command — auto-detects project language and checks binaries, LSP configs, `compile_commands.json` validity, and cmake flags. Prints fix hints for every failed check.
- Shared `internal/ui` package providing spinners, step runners, styled output helpers (`OK`, `Fail`, `Warn`, `Dim`, `Title`, `Step`, `Hint`, `Elapsed`), and a `Confirm` prompt. Spinners degrade to plain log lines automatically in CI.
- Spinners on every command. All long-running operations show a Bubble Tea spinner in interactive terminals.
- Full Bubble Tea TUI (`fe tui`) with four tabs: Project, Ship, Doctor, Config. Supports keyboard navigation, inline project name prompts, per-command result views, and active config display. TTY-guarded — refuses to run in CI.
- Typed config struct in `internal/config` loaded via `viper.Unmarshal`. Config priority: env vars > `.fe.yaml` > conventions > defaults.
- CI-friendly env var overrides: `ITCHIO_USERNAME`, `ITCHIO_GAME`, `FE_USERVERSION`, `FE_BUILD_DIR`.
- Convention-based defaults — `fe` works with zero config if your project follows the standard layout (`builds/`, `assets/sprites/`, etc.).
- Remote template support — set `templates.source` in `.fe.yaml` to fetch `<lang>.tar.gz` from a custom URL, with automatic fallback to built-ins on failure.
- `fe-templates` submodule linked at [github.com/flushwhy/fe-templates](https://github.com/flushwhy/fe-templates).
- GitHub Actions workflows for building `fe` itself as a release binary across 5 platforms (linux-x64, linux-arm64, windows-x64, macos-x64, macos-arm64).
- Butler push now supports both flat convention layout (`builds/windows-x64/`) and legacy two-level layout (`builds/windows/x64/`), with `CombinedOutput()` so errors from butler are visible.

### Changed
- `fe init` now takes a language argument (`fe init odin --name mygame`) instead of just a project name.
- `fe bmp` now resolves platform channels from both flat and two-level directory structures.
- All commands use `RunE` (returning errors) instead of `Run` + `log.Fatal` for proper error propagation.
- `internal/transcode` ffmpeg-go arg keys corrected — no leading dash (e.g. `"c:v"` not `"-c:v"`).

### Fixed
- `pack.go` mutex-in-loop bug removed. File removal before write is now a clean single operation.
- `fs.Sub` embed path now uses `path.Join` (forward slashes) instead of `filepath.Join`, fixing template resolution on Windows.
- `compile_commands.json` stub written immediately on `fe init c/cpp` and `fe add clangd` so clangd never errors on startup before the first build.

---

## [0.0.3]

### Added
- Started work on a TUI.
- Added [fe-templates](https://github.com/flushwhy/fe-templates) submodule, to be integrated into the `init` command.

### Fixed
- Fixed file access permissions and added thread safety to the `pack` command.

---

## [0.0.2]

### Added
- Viper support for `.fe.yaml` config files.
- New `pack` command that packs multiple PNGs into a single PNG.
- New `init` command.

### Changed
- How `transcode` works; updated flags and options.

### Removed
- `resizetexture` and `PNGJoiner` commands.

### Fixed
- `transcode` and `bmp` commands.

---
