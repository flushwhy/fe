#!/usr/bin/env python3
"""Smoke-test the shipped Blender and Godot project templates."""

import ast
import json
import re
import sys
import tomllib
from pathlib import Path


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(f"template validation failed: {message}")


def validate_blender(root: Path) -> None:
    addon = root / "blender-addon"
    manifest_path = addon / "blender_manifest.toml"
    manifest = tomllib.loads(manifest_path.read_text(encoding="utf-8"))

    for field in ("schema_version", "id", "version", "name", "tagline",
                  "maintainer", "blender_version_min", "license"):
        require(field in manifest and manifest[field], f"Blender manifest missing {field}")
    require(manifest.get("type") == ["add-on"],
            'Blender manifest type must be ["add-on"]')
    require(isinstance(manifest.get("license"), list) and manifest["license"],
            "Blender manifest license must be a non-empty array")

    source = addon / "__init__.py"
    ast.parse(source.read_text(encoding="utf-8"), filename=str(source))
    require((addon / "README.md").is_file(), "Blender README is missing")


def validate_godot(root: Path) -> None:
    project = root / "godot-2d"
    project_file = project / "project.godot"
    content = project_file.read_text(encoding="utf-8")

    match = re.search(r'^run/main_scene\s*=\s*"res://([^"]+)"', content, re.MULTILINE)
    require(match is not None, "Godot project has no res:// main scene")
    scene_path = project / match.group(1)
    require(scene_path.is_file(), f"Godot main scene does not exist: {match.group(1)}")

    scene = scene_path.read_text(encoding="utf-8")
    references = re.findall(r'path\s*=\s*"res://([^"]+)"', scene)
    require(references, "Godot main scene has no external resource references")
    for reference in references:
        require((project / reference).is_file(),
                f"Godot scene resource does not exist: {reference}")

    script = project / "scripts" / "main.gd"
    require(script.is_file(), "Godot movement script is missing")
    require((project / "README.md").is_file(), "Godot README is missing")


def main() -> None:
    root = Path(sys.argv[1] if len(sys.argv) > 1 else "fe-templates").resolve()
    catalog = json.loads((root / "catalog.json").read_text(encoding="utf-8"))
    require(catalog.get("version") == 1, "unsupported template catalog version")

    ids = [item["id"] for item in catalog.get("templates", [])]
    require(ids, "template catalog is empty")
    require(len(ids) == len(set(ids)), "template catalog contains duplicate IDs")
    for template_id in ids:
        require((root / template_id).is_dir(),
                f"catalog entry has no template directory: {template_id}")

    validate_blender(root)
    validate_godot(root)
    print(f"Validated {len(ids)} catalog entries, Blender add-on, and Godot 2D project.")


if __name__ == "__main__":
    main()
