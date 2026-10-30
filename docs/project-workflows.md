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

- `fe inspect`: explain what Fe recognizes about the current project.
- `fe setup`: show required tools and offer safe, explicit installation/configuration actions.
- `fe doctor`: diagnose project and environment problems with actionable fixes.
- `fe build`: build configured or detected targets.
- `fe changelog`: generate release notes from Git history and tags.
- `fe package`: create configured distribution artifacts.
- `fe release`: coordinate checks, build, metadata, packaging, and publishing.

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
    - name: linux-x64
      command: go
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

Configuration should remain approachable. Users should only need to specify values that differ from useful defaults. Credentials must come from environment variables or secure credential storage, never committed YAML.

## Incremental delivery plan

1. Project inspection and explicit project identity/type metadata.
2. Unified build target model and capability-aware preflight.
3. Git-based changelog generation.
4. Native archive packaging, followed by DEB/RPM support.
5. Metadata synchronization with preview/dry-run safeguards.
6. A generalized release orchestrator that composes the above capabilities.
7. Publishing adapters, starting with the existing itch.io/Butler integration and GitHub Releases.
8. Ecosystem workflows, starting with Blender add-ons and expanding based on real use cases.

Each phase should ship independently and preserve existing commands. Avoid a speculative plugin framework until multiple real workflows demonstrate the need for one.
