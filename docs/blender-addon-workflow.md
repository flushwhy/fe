# Blender add-on workflow

Fe can inspect, validate, and package a Blender add-on without importing or executing project code.

## Commands

Run from the add-on directory or pass its path:

```sh
fe blender validate
fe blender validate ./addons/my-addon
fe blender package ./addons/my-addon
fe blender package ./addons/my-addon --output dist/my-addon.zip
``

Packaging defaults to `dist/<id>-<version>.zip` for extensions and `dist/<directory>-dev.zip` for legacy add-ons.

## Supported layouts

### Blender extensions

An extension must contain `blender_manifest.toml` and `__init__.py` at its root. Fe validates the 1.0.0 manifest schema fields used by add-on extensions:

- `schema_version`, `id`, `version`, `name`, `tagline`, `maintainer`, and `type`
- `blender_version_min`
- At least one SPDX license identifier

The initial workflow supports `type = "add-on"`. IDs must use lowercase letters, digits, and underscores; versions must use semantic versioning.

### Legacy add-ons

A legacy add-on must contain a root `__init__.py` with `bl_info` metadata. Its package ZIP keeps the add-on directory as the top-level folder, which is the layout expected by the legacy add-on installer.

## Validation and safety

- Metadata and manifest are read as data; add-on Python is never imported or executed.
- If Python is available on PATH, Fe parses every Python source file with Python's AST parser to catch syntax errors.
- If Python is unavailable, metadata validation still succeeds and Fe reports that syntax checking was skipped.
- A local Blender installation is not required for static validation or packaging.
- Validation does not install dependencies or rewrite project files.

## Packaging

- Extension files are placed at the ZIP root, including `blender_manifest.toml`.
- Legacy add-ons retain their top-level package directory.
- Git metadata, Fe metadata, virtual environments, caches, build output, distribution output, bytecode, and nested ZIP files are excluded.
- ZIP entries are written in stable path order with a fixed timestamp for reproducible output.
- Symlinks and non-regular files are excluded.

The ZIP can be installed from disk in Blender for a final in-application test. Fe does not publish to the Blender Extensions Platform.

## References

- [Blender extension manifest schema](https://developer.blender.org/docs/features/extensions/schema/1.0.0/)
- [Blender extension packaging guide](https://docs.blender.org/manual/en/latest/advanced/extensions/getting_started.html)
