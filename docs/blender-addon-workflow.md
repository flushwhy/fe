# Blender add-on workflow

Blender add-ons are a useful first example of Fe supporting projects that are not conventional compiled applications. The workflow should turn an add-on directory into a validated, versioned, installable release archive without making its author learn a separate release toolchain.

This document describes intended behavior. It does not claim that Blender-specific commands are implemented yet.

## Project recognition

Fe should recognize both common Blender add-on layouts:

- Legacy add-ons that expose Blender add-on metadata in Python source (commonly `bl_info`).
- Extension packages that include a `blender_manifest.toml` manifest.

Detection should report what was found and avoid guessing when a repository contains multiple unrelated add-ons. Users must be able to select the intended add-on explicitly in `.fe.yaml`.

## Intended configuration

Illustrative configuration; field names are not final:

```yaml
project:
  name: my-blender-addon
  version: "1.2.0"
  type: blender-addon

blender:
  path: addon/
  compatibility:
    minimum: "4.2"
    maximum: "5.0"
  validate: true
  package:
    format: zip
    output: dist/

changelog:
  source: git
  since: last-release

publish:
  github:
    enabled: true
```

The project version and compatibility range should have one clear source of truth. If Blender metadata also contains a version, synchronization should be explicit and previewable rather than silently rewriting source files.

## Workflow responsibilities

### Inspect

- Locate the add-on or extension root.
- Identify legacy add-on metadata or an extension manifest.
- Read its declared name, version, and Blender compatibility where available.
- Explain missing or ambiguous metadata with a suggested fix.

### Validate

- Check required metadata and expected package structure.
- Check Python syntax without importing or executing the add-on.
- Validate the extension manifest when present.
- Report compatibility declarations that are missing or malformed.
- Keep validation read-only.

### Package

- Build a ZIP with the correct installable directory layout.
- Exclude Git metadata, caches, local virtual environments, test artifacts, and Fe's own output directory.
- Include required runtime files and manifest/metadata.
- List archive contents in dry-run or verbose output so users can verify what will ship.

### Release

- Use the shared Fe release pipeline for checks, changelog generation, artifact naming, and publishing.
- Allow packaging without publishing.
- Allow publishing an already-created ZIP without rebuilding.
- Keep credentials outside `.fe.yaml`.

## User-facing diagnostics

Errors should explain the consequence and the next action. For example:

```text
Blender add-on validation failed

Problem:
  No add-on metadata or blender_manifest.toml was found in addon/.

What this means:
  Fe cannot determine the add-on name or compatibility information.

How to fix:
  Point blender.path at the add-on root, or add the required Blender metadata.
```

A missing local Blender installation should not automatically prevent static validation and ZIP packaging. Only checks that genuinely require Blender should be skipped or marked unavailable, with the reason shown.

## Acceptance criteria for an initial implementation

- Detect legacy add-ons and extension packages without executing project Python code.
- Support an explicit add-on path for monorepos and ambiguous repositories.
- Validate required metadata and produce actionable errors.
- Create a deterministic ZIP with a correct installable layout.
- Exclude common generated and private development files.
- Add tests for detection, validation, archive contents, and ignored paths.
- Integrate with the shared release pipeline without making Blender publishing a special case.
