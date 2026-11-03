# Fe 0.9.0 roadmap

Fe 0.9.0 is a capability milestone: make the existing project, build, integration, and release workflows work together as one predictable developer tool.

This roadmap starts from the current `master` implementation. Status describes what is present in the repository, not a promise that every edge case is production-ready. Each feature PR must verify behavior and add focused tests before calling an area complete.

## Release goals

1. Make configured builds easier to select, validate, and understand without breaking existing `.fe.yaml` files.
2. Turn project detection into useful, safe workflows for more than conventional compiled applications.
3. Finish the highest-value ecosystem workflows, beginning with Blender and targeted Godot/O3DE gaps.
4. Make templates and container builds reliable, documented parts of the normal developer loop.
5. Verify the complete release path from project outputs to published artifacts.

## Current capability audit

| Area | Current state on `master` | 0.9.0 direction |
|---|---|---|
| Project inspection | Implemented. Read-only detection covers Go, Rust, Python, Node.js, C/C++, Godot, Blender add-ons, O3DE, and Lua. Safe build suggestions are intentionally limited to root Go and Rust manifests. | Keep detection conservative; connect only well-understood workflows to execution. |
| Configuration | Implemented around `.fe.yaml`, Viper, environment overrides, conventions, and command-specific validation. | Improve validation and target selection additively; preserve existing configuration. |
| Build targets | Implemented with `platform`, `cmd`, and `output`; `fe build --platform` filters targets. Commands are split into argv and do not interpret shell syntax. | Add named target selection and clear listing/diagnostics without silently changing command execution semantics. |
| Preflight and diagnostics | `fe preflight`, `fe validate`, `fe doctor`, and `fe check` exist. | Close concrete gaps discovered through integration tests; avoid duplicating these commands. |
| Templates | Remote catalog listing, download, local caching, and refresh exist through `fe template list/update` and `fe init`. | Improve metadata, discoverability, and tests only where the current template repository demonstrates a need. |
| Godot | Project/engine detection, engine selection/opening, native extension scaffolding, preset export, and export listing exist. | Audit version/platform edge cases and unify overlapping export entry points only where it reduces confusion. |
| O3DE | Discovery, doctor, project/engine registration helpers, Gem operations, and configured build-directory support exist. | Identify the remaining high-value project/watch workflow gaps; do not duplicate O3DE's own project manager. |
| Blender | Inspection recognizes extension manifests and legacy `bl_info` metadata. No complete Blender add-on validate/package command exists. | Deliver a safe, deterministic installable add-on ZIP workflow. |
| Container builds | Docker/Podman execution exists as an opt-in `fe build --container` path with a configured image and command. | Test failure/output behavior and document host/container boundaries; no container manager or speculative cache layer. |
| Odin packages | `fe install` / `fe get` resolve an alias or GitHub repository and safely install Odin source packages. | Validate usability and configuration/documentation gaps before adding features. |
| Release pipeline | Git changelog, ZIP/tar.gz archive packaging, artifact manifest, and optional release orchestration are implemented. | Exercise complete release paths and verify checksums, assets, skip flags, and stable/rolling separation. |
| Distribution | Stable/rolling workflows and existing package channels are maintained in CI. | Keep Gumroad automation out of the critical path until documented upload/replacement behavior is available. |

## Delivery sequence

Work in small, independently reviewable PRs. Do not stack dependent feature branches on unmerged work unless the dependency is explicit and the stack is kept current.

### 0.9.0-1 — Build target selection

- Extend the existing target model rather than replace it.
- Add an optional stable target name while retaining `platform` as a backward-compatible field.
- Add `fe build --target <name>` and a target-listing mode.
- Define duplicate-name, missing-name, and unknown-target diagnostics.
- Keep the current all-target default and `--platform` behavior.
- Add config, CLI, and regression tests using old configuration examples.

### 0.9.0-2 — Build configuration validation

- Validate target definitions before starting any build.
- Report the target name/platform with actionable errors.
- Ensure dry-run/inspection output uses the same resolved target selection as execution.
- Do not introduce a general plugin or build-graph framework.

### 0.9.0-3 — Blender add-on workflow

- Detect an explicit add-on root, including monorepo paths.
- Read extension manifest or legacy metadata without importing project Python.
- Validate required metadata and Python syntax using available local tooling.
- Package a deterministic ZIP with the installable directory layout.
- Exclude VCS metadata, caches, virtual environments, and generated output.
- Add tests for detection, validation, archive contents, and ignored paths.
- Reuse the existing archive/release pipeline; do not create a Blender-specific publisher.

### 0.9.0-4 — Godot workflow audit and polish

- Test engine discovery and project/version selection.
- Test preset parsing and output mapping across supported desktop platforms.
- Review the overlap between `fe export` and `fe godot export`; simplify only if behavior can be made clearer without removing useful options.
- Make unsupported platform and missing export-template errors actionable.

### 0.9.0-5 — Template workflow improvements

- Check the current `fe-templates` catalog and real template layouts before designing metadata.
- Improve template descriptions/category display if the catalog supports them.
- Test offline cache reuse, refresh, invalid IDs, and interrupted downloads.
- Keep templates in the external repository; avoid hard-coding template definitions in Fe.

### 0.9.0-6 — O3DE workflow gap closure

- Audit the current project/engine/Gem commands and general `fe watch` behavior.
- Prioritize one concrete missing workflow, likely project-aware watch/rebuild behavior.
- Use O3DE's existing CLI/build tools rather than recreating engine registration or build orchestration.
- Keep the first increment narrowly scoped and testable.

### 0.9.0-7 — Container and Odin workflow verification

- Add integration-level tests for container argument construction, mounts, exit status, and output persistence where practical.
- Verify Docker and Podman behavior and clearly report missing runtimes/images.
- Exercise Odin package installation with invalid archives and existing destinations.
- Add features only when these tests reveal a real user-facing gap.

### 0.9.0-8 — Release hardening

- Test changelog, manifest, archive, and release orchestration together.
- Verify checksums and archive contents against the manifest.
- Verify optional stages and skip flags independently.
- Verify stable and rolling release paths remain separate.
- Confirm clean-environment install smoke tests and published documentation.

## Release acceptance criteria

- Existing `.fe.yaml` projects continue to work without migration.
- Named targets can be selected and listed, with old platform selection preserved.
- Build configuration errors are found before commands execute.
- Blender add-ons can be validated and packaged without importing or executing add-on code.
- Godot's documented export paths match actual output behavior.
- Template cache and refresh behavior have regression coverage.
- The selected O3DE workflow is useful without duplicating O3DE itself.
- Container and Odin workflows have meaningful failure-path coverage.
- Release artifacts, checksums, and manifests agree.
- Stable and rolling distribution are verified independently.
- CLI help, README, and docs describe implemented behavior.
- CI passes on the final release candidate.

## Explicitly out of scope

- Automated Gumroad uploads until the required file replacement API is documented and verified.
- A plugin marketplace or general-purpose integration framework.
- AI-generated configuration.
- Replacing CMake, Godot, O3DE, Blender, or other ecosystem tools.
- A built-in container/image manager.
- A GUI project manager.

## Working rules

- Prefer additive changes and preserve existing CLI/config behavior.
- Keep ecosystem-specific behavior at the edge; share code only after multiple real workflows need it.
- Never execute detected project scripts during read-only inspection or validation.
- Do not install tools, rewrite project files, or publish artifacts without explicit user configuration/action.
- Each PR should include tests and update documentation for the behavior it actually implements.
