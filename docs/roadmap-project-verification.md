# Project verification roadmap

This document captures a project-health and post-build verification initiative for Fe, informed by integration work on GeSystemCore/O3DE. The goal is to help developers catch problems that a successful compile alone cannot detect, while keeping Fe a general-purpose project and release tool.

This is a roadmap, not a claim that the proposed behavior is implemented. Existing commands are the starting point: extend them where they already own the responsibility instead of introducing overlapping commands.

## Why this matters

The GeSystemCore integration exposed several failure classes that were easy to miss when looking only at build status:

- Two copies of the same Gem were present in the engine's Gems locations, making it unclear which copy the project was loading.
- Player, ground, and target entities were marked editor-only. They appeared in the editor but were excluded from the game runtime.
- A camera component existed and compiled, but the game had not selected it as the active render view.
- Runtime interaction needed to be exercised and confirmed from application logs; compilation and launch alone did not prove the feature worked.

These are examples of a broader distinction:

- A healthy development environment does not guarantee a valid project.
- A valid project and successful build do not guarantee a correct runtime.
- A running application does not guarantee that expected behavior or release artifacts are present.

Fe should make those stages explicit and verifiable.

## Command responsibilities

Fe already has `fe doctor`, `fe validate`, and `fe preflight`. Preserve and clarify those responsibilities:

| Command | Responsibility | Examples |
|---|---|---|
| `fe doctor` | Diagnose the host development environment. | Required tools missing, compiler/SDK unavailable, invalid environment variables, inaccessible configured paths. |
| `fe validate` | Inspect project and Fe configuration for structural or semantic problems without running project code. | Invalid configuration, missing referenced files, duplicate or conflicting project components, undeclared local dependencies. |
| `fe preflight` | Check whether the selected build or release operation can begin. | Build tool/configuration availability, unresolved build prerequisites, invalid target selection, missing required inputs. |
| `fe verify` | Check the outputs of a completed build and, when explicitly configured, exercise the application. | Required artifacts, executable launch, smoke-test conditions, expected or forbidden log messages, timeout and exit status. |

The exact boundary between current command behavior should be confirmed against the implementation before changing it. Do not create a second command for a check that already has a clear home.

## Proposed capabilities

### 1. Strengthen environment diagnostics

Extend `fe doctor` with actionable checks where the current implementation has gaps:

- Detect required toolchains and SDKs, and report discovered versions and paths.
- Distinguish missing tools from tools that exist but cannot be executed.
- Check configured directories and permissions without silently modifying the environment.
- Explain which project/build feature requires a missing tool.
- Keep diagnostics useful in local terminals and CI logs.

Doctor must remain diagnostic: it should not install tools, rewrite environment variables, or make system changes automatically.

### 2. Add project integrity checks to validation

Extend `fe validate` with safe, read-only checks such as:

- Duplicate modules, plugins, Gems, or dependency copies in known project search locations.
- Conflicting dependency declarations or ambiguous local dependency paths.
- Missing files and directories referenced by project configuration.
- References that resolve outside the expected project/engine/workspace roots.
- Required assets or configuration entries that are absent.
- Project-type-specific checks supplied by an appropriate validator.

Checks should report the affected path/component, why it is suspicious, and a concrete next step. A duplicate should be reported as a possible conflict rather than automatically deleted or rewritten.

Validation must not import or execute detected project scripts, build steps, editor extensions, or other project-controlled code. Avoid false certainty when an ecosystem's project format cannot be interpreted safely.

### 3. Improve build preflight

Use the existing `fe preflight` workflow to catch build-time problems before invoking a long build:

- Confirm the selected target and configuration resolve consistently with the actual build command.
- Check required tools, SDKs, inputs, and generated configuration.
- Report stale or missing generated files when the project type provides a reliable way to identify them.
- Make it clear which checks are informational, warnings, or blockers.
- Keep dry-run output and actual execution based on the same resolved settings.

Preflight should orchestrate existing ecosystem tools rather than recreate their build systems.

### 4. Introduce post-build verification

Add `fe verify` as a distinct stage after build. It should support progressively stronger levels of verification:

1. **Artifact checks:** confirm required executables, libraries, packages, assets, and manifests exist and are non-empty.
2. **Artifact consistency:** where metadata is available, compare expected outputs with the actual output set and release manifest.
3. **Launch checks:** optionally launch a configured executable, capture its exit status and bounded output, and enforce a timeout.
4. **Smoke-test checks:** optionally evaluate configured success/failure conditions, such as expected log messages or a project-provided test command.

A minimal illustrative configuration could look like:

```yaml
verify:
  artifacts:
    required:
      - path: build/bin/MyGame.exe
        type: file

  runtime:
    command: [build/bin/MyGame.exe, "--smoke-test"]
    timeout: 30s
    expected_exit_code: 0
    logs:
      required:
        - "Startup complete"
      forbidden:
        - "Failed to initialize"
```

This is an example of the desired capability, not a finalized configuration schema. Any schema must fit Fe's existing configuration conventions and validation rules.

Runtime execution must be opt-in and explicit. Fe must not launch arbitrary project binaries as part of read-only `validate` or `preflight`. Verification should capture useful diagnostics, enforce timeouts, and clearly distinguish a failed application from a failed verification configuration.

### 5. Use lightweight ecosystem validators

Engine- and framework-specific rules should live at the edge of Fe. For example, an O3DE validator could eventually inspect known Gem references and prefab/entity metadata for runtime hazards such as editor-only entities in a runtime level.

The shared Fe layer should own common reporting, severity, selection, and exit-status behavior. Ecosystem validators should own interpretation of their formats and should reuse the ecosystem's own tools where practical.

Start with a small, built-in provider/validator interface only when at least one real integration needs it. Do not build a plugin marketplace, dynamic extension host, general-purpose rule language, or a second implementation of CMake/O3DE/Unity/Godot/Blender tooling.

## Suggested delivery sequence

Keep each increment small and independently testable:

1. **Audit existing commands.** Map current `doctor`, `validate`, and `preflight` checks, their output, and their tests. Document overlaps and gaps before changing behavior.
2. **Close high-confidence validation gaps.** Add safe project-integrity checks with focused fixtures, including duplicate dependency/module paths and missing references.
3. **Harden preflight.** Ensure preflight resolves the same targets and prerequisites as build execution and reports actionable blockers.
4. **Define verification configuration.** Design the smallest additive schema that fits existing Fe configuration; preserve existing projects and defaults.
5. **Implement artifact verification.** Add required-file checks and clear exit statuses before adding process execution.
6. **Add opt-in runtime smoke checks.** Support command arguments, bounded output capture, timeouts, exit codes, and log conditions.
7. **Add the first ecosystem validator.** Choose a concrete, testable integration gap rather than creating speculative provider infrastructure.

## Acceptance criteria

- Existing `.fe.yaml` files and command behavior remain backward-compatible unless a change is explicitly documented.
- `doctor`, `validate`, and `preflight` have distinct, documented responsibilities and do not duplicate one another.
- Read-only inspection and validation never execute project-controlled code or mutate the project/environment.
- Duplicate or ambiguous dependencies are reported with paths and actionable context; Fe does not automatically delete or replace them.
- Build preflight checks the same resolved target/configuration that the build will use.
- `fe verify` can validate required artifacts and return a meaningful non-zero status when required outputs are missing.
- Runtime execution is explicitly configured, time-bounded, and reports exit status and captured diagnostics.
- Ecosystem-specific checks remain isolated from Fe's general project/release core.
- Tests cover success, failure, malformed configuration, missing tools/files, timeouts, and platform-specific path behavior where applicable.
- CLI help and documentation describe only behavior that has actually shipped.

## Non-goals

- Replacing ecosystem build systems, editors, project managers, or test runners.
- Automatically repairing, deleting, moving, or rewriting project files during validation.
- Running arbitrary project scripts during read-only inspection.
- A general plugin marketplace or dynamic third-party validator platform.
- Making every project type support runtime smoke tests.
- Treating log substring matching as a substitute for proper application tests; it is an optional integration-level signal.

## Design principles

- **Extend before duplicating:** build on Fe's existing diagnostics and validation commands.
- **Read-only by default:** inspection must not alter a user's project or machine.
- **Explicit execution:** launching an application belongs to an opt-in verification step.
- **Actionable output:** identify the failing item, its location, its severity, and the next useful action.
- **Conservative ecosystem support:** delegate to existing tools and add specialized checks only for demonstrated gaps.
- **Backward compatibility:** add capabilities without silently changing existing project configuration or release behavior.
