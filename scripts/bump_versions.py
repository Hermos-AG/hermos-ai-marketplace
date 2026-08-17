"""Sync the gpu-mcp plugin version from its Go source and bump the catalog version.

Line-based edits keep the JSON formatting (and diffs) untouched.

Usage: python scripts/bump_versions.py [new-catalog-version]
"""
from __future__ import annotations

import json
import os
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
PLUGIN_DIR = ROOT / "plugins" / "gpu-mcp"
CATALOG = ROOT / ".claude-plugin" / "marketplace.json"
MANIFEST = PLUGIN_DIR / ".claude-plugin" / "plugin.json"
VERSION_LINE = re.compile(r'^(\s*"version"\s*:\s*")([^"]+)(".*)$')


def read_lines(path: Path) -> list[str]:
    return path.read_text(encoding="utf-8").split("\n")


def write_lines(path: Path, lines: list[str]) -> None:
    with open(path, "w", encoding="utf-8", newline="\n") as fh:
        fh.write("\n".join(lines))


def set_version_at(lines: list[str], index: int, new: str, label: str) -> None:
    match = VERSION_LINE.match(lines[index])
    if not match:
        sys.exit(f"{label}: line {index + 1} is not a version line: {lines[index]!r}")
    if match.group(2) != new:
        print(f"{label}: {match.group(2)} -> {new}")
    lines[index] = f"{match.group(1)}{new}{match.group(3)}"


def source_version() -> str:
    text = (PLUGIN_DIR / "main.go").read_text(encoding="utf-8")
    match = re.search(r'serverVersion\s*=\s*"([0-9]+\.[0-9]+\.[0-9]+)"', text)
    if not match:
        sys.exit("serverVersion not found in plugins/gpu-mcp/main.go")
    return match.group(1)


def bump_patch(version: str) -> str:
    major, minor, patch = (int(part) for part in version.split("."))
    return f"{major}.{minor}.{patch + 1}"


def find_entry_version(lines: list[str], entry_index: int, catalog_index: int) -> int | None:
    """Version line belonging to the object that holds line `entry_index`."""
    for i in range(entry_index + 1, len(lines)):
        if VERSION_LINE.match(lines[i]):
            return i
        if lines[i].strip() in ("},", "}"):
            break
    for i in range(entry_index - 1, -1, -1):
        if i != catalog_index and VERSION_LINE.match(lines[i]):
            return i
        if lines[i].strip() == "{":
            break
    return None


def main() -> int:
    wanted = sys.argv[1].strip() if len(sys.argv) > 1 and sys.argv[1].strip() else None
    plugin_version = source_version()

    manifest_lines = read_lines(MANIFEST)
    for index, line in enumerate(manifest_lines):
        if VERSION_LINE.match(line):
            set_version_at(manifest_lines, index, plugin_version, "plugin.json")
            break
    write_lines(MANIFEST, manifest_lines)

    catalog_lines = read_lines(CATALOG)
    catalog_version_index = next(
        (i for i, line in enumerate(catalog_lines) if VERSION_LINE.match(line)), None
    )
    if catalog_version_index is None:
        sys.exit("marketplace.json: no catalog version line found")
    current = VERSION_LINE.match(catalog_lines[catalog_version_index]).group(2)
    new_catalog = wanted or bump_patch(current)
    set_version_at(catalog_lines, catalog_version_index, new_catalog, "marketplace.json (catalog)")

    entry_index = next(
        (i for i, line in enumerate(catalog_lines) if "plugins/gpu-mcp" in line), None
    )
    if entry_index is None:
        sys.exit("marketplace.json: no entry with source plugins/gpu-mcp")
    entry_version_index = find_entry_version(catalog_lines, entry_index, catalog_version_index)
    if entry_version_index is None:
        sys.exit("marketplace.json: gpu-mcp entry has no version line")
    set_version_at(catalog_lines, entry_version_index, plugin_version, "marketplace.json (gpu-mcp)")
    write_lines(CATALOG, catalog_lines)

    # sanity: both files still parse
    json.loads(MANIFEST.read_text(encoding="utf-8"))
    json.loads(CATALOG.read_text(encoding="utf-8"))

    print(f"plugin {plugin_version}, catalog {new_catalog}")
    output = os.environ.get("GITHUB_OUTPUT")
    if output:
        with open(output, "a", encoding="utf-8") as fh:
            fh.write(f"plugin_version={plugin_version}\n")
            fh.write(f"catalog_version={new_catalog}\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())