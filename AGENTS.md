# AGENTS.md

Instructions for coding agents that are asked to create, configure, build, or package a project with Fe.

This guide describes Fe's current behavior. Treat the installed `fe --help`, each command's `--help`, the project's own README, and its checked-in build files as the final authority if they differ from this document. Do not assume that a capability is available just because a project type is detected.

## 1. Start by understanding the project

Before changing files or running build commands:

1. Confirm that `fe` is installed and available:
   ```sh
   fe --version
   fe --help
   ```
2. Inspect the project:
   ```sh
   fe inspect .
   fe setup .
   fe doctor
   ```
   - `fe inspect` detects project markers and may suggest a conventional build command.
   - `fe setup` is read-only. It reports detected tool requirements; it does not install tools or edit files.
   - `fe doctor` checks Fe and project configuration. Use the relevant ecosystem-specific doctor for Godot or O3DE projects.
3. Read the project's README, build scripts, manifests, and existing CI configuration.
4. Read `.fe.yaml` if present. Do not replace an existing build system or overwrite project configuration just to make Fe fit.
5. Identify the actual build/test command and required SDKs before installing anything.

Inspection is not permission to execute project code. Review commands from project files before running them, especially build scripts, package lifecycle scripts, custom preflight checks, and container commands.

## 2. Create a project with Fe

Fe can scaffold one of its built-in language starters or a named template from the remote `flushwhy/fe-templates` catalog.

### Built-in language starters

The currently supported `fe init` language names are:

- `odin` — default template: `odin-basic`
- `c` — default template: `c-make`
- `cpp` — default template: `cpp-cmake`
- `zig` — default template: `zig-build`
- `go` — default template: `go-basic`

Example:

```sh
fe init go --name my-project
cd my-project
```

For Go, set the module path when the project needs a real import path:

```sh
fe init go --name my-project --module github.com/owner/my-project
```

The destination directory must not already exist. Use a new directory or choose a different project name; do not delete an existing directory to make scaffolding succeed.

### Named templates

List the live catalog rather than guessing template IDs:

```sh
fe template list
fe init --template <template-id> --name my-project
```

Templates are downloaded and cached locally. Refresh a cached template with:

```sh
fe template update <template-id>
```

After scaffolding, enter the new directory and read its README. A named template may have its own dependencies, build steps, and conventions. Do not assume that every template uses the same build command or produces the same output.

### What scaffolding does

Fe writes the selected template into a new project directory and renders its template files. C and C++ starters also receive an initial `compile_commands.json` stub; that stub is not a real compilation database. Generate a real database using the project's build system before relying on clangd.

Outside CI, `fe init` also runs Fe's local project-sync step. Check the resulting files and report any tool or editor setup that still needs attention.

## 3. Project types Fe recognizes

Detection is based on files and markers. It does not mean Fe can automatically build every detected ecosystem.

| Detected project type | Typical markers | Fe assistance |
|---|---|---|
| Go | `go.mod` | Inspection, Go tool checks, conventional build suggestion for a root module |
| Rust | `Cargo.toml` | Inspection, Cargo tool checks, conventional release-build suggestion for a root crate |
| Python | `pyproject.toml`, `requirements.txt`, `Pipfile`, `setup.py` | Inspection, Python tool checks, configured preflight/static checks |
| Node.js | `package.json` | Inspection and Node/package-manager checks |
| C/C++ | `CMakeLists.txt`, `Makefile` | Inspection, compiler/build-tool checks, compilation database repair and editor setup |
| Godot | `project.godot` | Engine detection/management, export presets, native extension workflow |
| Blender add-on | `blender_manifest.toml` or legacy `bl_info` metadata | Static validation and ZIP packaging |
| O3DE | `engine.json` | O3DE discovery, diagnostics, project/Gem helpers, existing CMake build directory support |
| Lua | `.lua` or `.rockspec` files | Inspection and Lua runtime/preflight checks |

Important boundaries:

- Project inspection is read-only and does not execute project code.
- Automatic build suggestions are deliberately limited to root-level Go and Rust manifests. For other project types, use the project's own documented build system or explicit `.fe.yaml` targets.
- Fe does not infer a safe build command from every detected marker.
- Nested projects may be detected, but conventional build suggestions are only made when the relevant manifest is at the inspected project root.

## 4. Configure a rebuildable project

Fe reads `.fe.yaml` from the working directory by default. You can select another file with `--config`.

Configuration precedence is:

1. Environment-variable overrides supported by Fe.
2. Values in `.fe.yaml`.
3. Fe conventions and defaults.

A minimal example for a custom C project:

```yaml
itchio:
  username: ""
  game: ""

butler:
  directory: "builds"

build:
  targets:
    - name: linux-release
      platform: linux-x64
      cmd: "gcc -O2 -o builds/linux-x64/my-game src/main.c"
      output: "builds/linux-x64/my-game"
```

Build targets have these fields:

- `name`: optional unique target name, used by `fe build --target`.
- `platform`: optional platform label, used by `fe build --platform`.
- `cmd`: required executable and arguments as a command string. Fe splits this into arguments; it does not run it through a shell. Do not put shell pipelines, redirections, or `&&` chains in this field. Use a checked-in build script as the command when a multi-step shell workflow is needed.
- `output`: optional declared output path.

Example commands:

```sh
fe build --list-targets
fe build
fe build --target linux-release
fe build --platform linux-x64
```

Fe runs every configured target by default. Validate that each target's compiler, SDK, and platform toolchain is actually available on the current host. A target being listed does not mean the host can cross-compile it.

### Useful configuration sections

Fe's typed configuration currently includes:

- `build.targets` and optional `build.container`
- `itchio.username` and `itchio.game`
- `butler.directory` and `butler.userversion`
- `godot.executable` and `godot.output_dir`
- `pack.input`, `pack.output`, and named `pack.sheets`
- `transcode` input/output and codec options
- `preflight.enabled`, `preflight.detect`, and configured `preflight.checks`
- `gumroad.stable` version-to-product-ID mappings

Do not add optional integrations to a project unless the task needs them. Keep credentials out of committed YAML.

Supported environment overrides include:

| Environment variable | Configuration value |
|---|---|
| `ITCHIO_USERNAME` | `itchio.username` |
| `ITCHIO_GAME` | `itchio.game` |
| `FE_USERVERSION` | `butler.userversion` |
| `FE_BUILD_DIR` | `butler.directory` |

## 5. Set up editor and language tools

Use `fe setup .` to see detected requirements. It is a report, not an installer.

Fe can add project editor/LSP configuration for:

- `clangd` — C/C++
- `ols` — Odin
- `zls` — Zig
- `gopls` — Go

Examples:

```sh
fe add clangd
fe add ols
fe add zls
fe add gopls
```

For CMake projects, generate a real compilation database before configuring clangd:

```sh
cmake -S . -B build -DCMAKE_EXPORT_COMPILE_COMMANDS=ON
fe fix compile-commands
```

Use the project's actual configure/build options; the example is not a universal CMake recipe.

## 6. Build and package by ecosystem

### Configured build targets

Use `fe build` when the project has valid `build.targets` in `.fe.yaml`. Review every configured command before execution. Fe validates the configuration and runs configured preflight checks before building.

### Godot

For a Godot project:

1. Read the project's export presets and required assets.
2. Run `fe godot doctor` and `fe godot engines`.
3. Ensure the matching engine version and export templates are installed.
4. Inspect presets and output paths:
   ```sh
   fe export --list-presets
   ```
5. Export all presets or one named preset:
   ```sh
   fe export
   fe export --preset "Windows Desktop"
   ```

Fe also provides `fe godot install`, `fe godot open`, and Godot 4 C++ GDExtension helpers under `fe godot extension`. Godot 3 GDNative is not the same workflow as Godot 4 GDExtension; inspect the project's existing native-extension setup before changing it.

### Blender add-ons

Fe can validate and package Blender extensions and legacy add-ons:

```sh
fe blender validate .
fe blender package . --output dist/my-addon.zip
```

Validation reads metadata and parses Python syntax when Python is available; it does not import or execute add-on code. Blender itself is not required for static validation or ZIP creation, but use Blender for the final in-application installation test.

### O3DE

Fe supplements an existing O3DE installation; it does not replace O3DE's project manager or configure an engine from scratch.

Start with:

```sh
fe o3de doctor
fe o3de find engines
fe o3de find projects
fe o3de find gems
fe o3de find templates
```

Use `fe o3de project` and `fe o3de gem` only when the task requires project or Gem registration/management. To build, pass an already configured CMake build directory:

```sh
fe o3de build --build-dir build
```

Use the project's generator and configuration requirements. Do not assume that Fe can create or configure the O3DE build directory for you.

### Odin packages

Fe can install Odin source packages by alias or HTTPS repository URL:

```sh
fe install <package>
fe get https://github.com/owner/package
```

The default destination is `mylibs/` (or `packages.directory` if configured). Existing destinations are not overwritten. Review the repository and its provenance before adding a dependency.

### Asset and media helpers

- `fe pack` packs sprite assets using `pack.input`, `pack.output`, or configured named sheets.
- `fe transcode` uses FFmpeg. Install FFmpeg separately and review the input/output and codec settings before converting files.
- `fe package` creates ZIP or `tar.gz` archives from build artifacts.
- `fe manifest` writes an artifact inventory with file sizes and SHA-256 checksums.
- `fe changelog` generates Markdown release notes from Git history and tags.

These helpers do not replace the project's own asset pipeline unless the project is deliberately configured to use them.

## 7. Preflight, checks, and troubleshooting

Useful commands:

```sh
fe inspect .
fe setup .
fe validate
fe preflight
fe check .
fe doctor
```

- `fe validate` checks Fe configuration and relevant paths.
- `fe preflight` checks toolchains and runs configured pre-build checks. Review configured commands before allowing them to run.
- `fe check` runs supported local static analyzers when they are installed. Missing tools are reported; Fe does not silently install them.
- `fe doctor` diagnoses the local environment and project configuration.

Fix the underlying issue, rerun the narrowest relevant check, and then rerun the build. Do not hide failures by deleting tests, disabling checks, or changing build flags without understanding the cause.

## 8. Publishing is a separate, explicit action

Building and publishing are different operations. Do not publish, upload, create a release, or change a live product unless the user explicitly requested that action.

- `fe bmp` publishes build output to itch.io through Butler.
- `fe release` orchestrates the configured release workflow and may publish to itch.io.
- `fe gumroad release` uploads a stable version to its explicitly configured Gumroad product.
- `fe release --gumroad` includes the configured stable Gumroad upload.
- `fe release --dry-run` previews the release plan; use it before any authorized release.

Gumroad stable uploads require explicit version-to-product mappings. Do not invent product IDs, reuse a product for another version, or retry an ambiguous upload without checking the product state.

For normal agent build tasks, stop after a successful build and report the artifact paths. Ask before running any publishing command.

## 9. Completion checklist

Before reporting that a project is rebuildable:

- [ ] The project README and template-specific instructions were read.
- [ ] Fe's detected project type and required tools were checked.
- [ ] The exact build command and required toolchain are known.
- [ ] Required dependencies are installed or clearly reported as missing.
- [ ] `.fe.yaml` is valid and does not contain invented or unnecessary settings.
- [ ] The build was run from a clean, understood state.
- [ ] Tests or project-specific validation were run when available.
- [ ] Build outputs and their paths were verified.
- [ ] No publishing action was taken without explicit authorization.

If a build cannot be reproduced on the current host, state which tool, SDK, platform, or external service is missing. Do not claim success based only on a successful Fe configuration check.
