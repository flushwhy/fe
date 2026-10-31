# Project workflows and release architecture

Fe's long-term goal is to make software and creative-project delivery repeatable without requiring every user to learn each ecosystem's build, packaging, and publishing tools.

The user describes the desired outcome in `.fe.yaml`; Fe detects what it can, explains what it cannot do, and runs a transparent workflow.

## Product principles

- **One configuration, many project types.** A project can be a CLI, desktop application, game, library, Blender add-on, or a custom project.
- **Automatic by default, explicit when needed.** Detect project markers and sensible defaults, but let users override every important decision.
- **YAML describes intent.** Fe owns the implementation details for supported workflows and tools.
- **Explain every step.** Show what is running, why it is needed, what succeeded, and actionable recovery guidance when it fails.
- **No silent side effects.** Installing tools, modifying project metadata, and publishing artifacts require explicit configuration or confirmation.
- **Graceful capability limits.** A target that cannot be built on the current host should be reported with the reason and a practical alternative.
- **Extensible without bloating the core.** Ecosystem-specific behavior belongs in focused workflows and integrations.

## Conceptual architecture

### Core engine

Shared capabilities used by all workflows:

- project inspection and configuration loading;
- tool and dependency checks;
- build target execution;
- artifact collection and packaging;
- version and changelog generation;
- metadata synchronization;
- release planning and orchestration;
- publishing integrations;
- consistent progress output and diagnostics.

### Project workflows

A workflow understands the conventions of one ecosystem and translates them into core operations. Examples include:

- Go, Rust, Python, Node, and C/C++ projects;
- Godot and other game projects;
- Blender add-ons;
- libraries, editor extensions, and plugins;
- custom projects described entirely in YAML.

A workflow should detect and validate ecosystem-specific metadata, identify required tools, define build/package steps, and produce artifacts that the core release pipeline can publish.

### Integrations

Publishing destinations and external services are separate from project workflows. Initial candidates include itch.io (Butler) and GitHub Releases. Additional destinations can be added without changing project detection or packaging.

## Intended user experience

The commands should form a discoverable progression:

- `fe inspect`: explain what Fe recognizes about the current project and show safe build suggestions where conventions are reliable.
- `fe setup`: show required tools and offer safe, explicit installation/configuration actions.
- `fe doctor`: diagnose project and environment problems with actionable fixes.
- `fe build`: build configured or explicitly selected detected targets.
- `fe changelog`: generate release notes from Git history and tags.
- `fe package`: create configured distribution artifacts.
- `fe release`: coordinate checks, build, metadata, packaging, and publishing.

Inspection is read-only. A detected marker is evidence about a project, not permission to execute its build scripts. Suggestions should only be offered when the command is conventional and its assumptions can be explained. In the first increment, that means root-level Go and Rust manifests; other detected ecosystems are reported without guessing a build command. Explicit `.fe.yaml` targets remain authoritative and are never silently replaced by detection.

The exact command surface can evolve; the shared behavior and clear explanations matter more than preserving these names.

## Release pipeline

The intended high-level order is:

1. Inspect project and load configuration.
2. Validate configuration and run preflight checks.
3. Resolve required tools and build capabilities.
4. Build configured targets that are supported on this host or through a configured remote/container environment.
5. Generate or review changelog and release metadata.
6. Synchronize metadata into other project files only when explicitly configured.
7. Package configured formats.
8. Publish selected artifacts to configured destinations.
9. Report artifacts, skipped targets, published destinations, and actionable failures.

A dry-run should show the plan without building, modifying metadata, or publishing. Publishing should be independently skippable, and existing artifacts should be publishable without rebuilding.

## Configuration direction

The following is illustrative, not a finalized schema:

```yaml
project:
  name: my-project
  version: "1.2.0"
  type: auto

build:
  auto_detect: true
  targets:
    - platform: linux-x64
      cmd: go
      args: [build, ./...]

changelog:
  source: git
  since: last-release
  output: CHANGELOG.md

packages:
  - zip
  - deb
  - rpm

publish:
  github:
    enabled: true
  itch:
    enabled: false
```

Build target `cmd` is the executable when `args` is present; each `args` entry is passed as one argument, so paths and values containing spaces stay intact. Existing `cmd`-only targets remain supported with the legacy whitespace split. Configuration should remain approachable. Users should only need to specify values that differ from useful defaults. Credentials must come from environment variables or secure credential storage, never committed YAML.

## Incremental delivery plan

1. Make project inspection explain detected types and offer conservative, read-only build suggestions.
2. Connect those suggestions to the shared build planner, while keeping explicit `.fe.yaml` targets authoritative.
3. Add tool detection and capability-aware preflight so Fe can explain what is needed on the current host or elsewhere.
4. Generate changelogs from Git history and tags.
5. Add native archive packaging, followed by DEB/RPM support.
6. Add metadata synchronization with preview/dry-run safeguards.
7. Build a release orchestrator that composes inspection, setup, build, changelog, and packaging.
8. Extend publishing, starting with the existing itch.io/Butler integration and GitHub Releases.
9. Expand ecosystem workflows based on real use cases, not a speculative plugin framework.

Each phase should ship independently and preserve existing commands. Keep the core small: add a shared abstraction only when multiple real workflows need it.
