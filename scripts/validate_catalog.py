"""Validate the HERMOS AI Marketplace catalog.

Checks the catalog manifest, every plugin manifest, the bilingual documentation
pairs and all JSON files in the repository. Exits non-zero on errors; warnings
(such as the deliberate non-kebab-case plugin name) do not fail the build.

Usage: python scripts/validate_catalog.py [repo-root]
"""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(sys.argv[1] if len(sys.argv) > 1 else ".").resolve()
DOC_PAIRS = ("README", "CHANGELOG", "RELEASE_NOTES")
KEBAB = re.compile(r"^[a-z0-9]+(-[a-z0-9]+)*$")

# Single source of truth for catalog categories. Keep in sync with the
# "Categories" table in README.md / README_de.md.
BUSINESS_UNITS = ("tnt", "fis", "rfid", "sales", "marketing", "ai-dev")
OPERATING_MODES = ("operations", "networking", "desktop")
ALLOWED_CATEGORIES = frozenset(BUSINESS_UNITS + OPERATING_MODES)

errors: list[str] = []
warnings: list[str] = []


def load_json(path: Path):
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        errors.append(f"missing file: {path.relative_to(ROOT)}")
    except json.JSONDecodeError as exc:
        errors.append(f"invalid JSON in {path.relative_to(ROOT)}: {exc}")
    return None


def check_doc_pairs(folder: Path, label: str) -> None:
    for stem in DOC_PAIRS:
        en, de = folder / f"{stem}.md", folder / f"{stem}_de.md"
        if en.exists() and not de.exists():
            errors.append(f"{label}: {stem}.md has no German counterpart {stem}_de.md")
        if de.exists() and not en.exists():
            errors.append(f"{label}: {stem}_de.md has no English counterpart {stem}.md")


def check_binaries(folder: Path, version: str, label: str) -> None:
    """Shipped binaries must carry the plugin version — catches stale release copies."""
    binaries = [
        path
        for path in sorted(folder.iterdir())
        if path.is_file() and (path.suffix == ".exe" or path.name.endswith("-linux"))
    ]
    if not binaries:
        return
    needle = version.encode("ascii")
    for binary in binaries:
        blob = binary.read_bytes()
        if needle in blob:
            print(f"  ok: {label}/{binary.name} carries version {version}")
        else:
            errors.append(
                f"{label}: {binary.name} does not contain version string {version} — stale binary?"
            )


def main() -> int:
    catalog_path = ROOT / ".claude-plugin" / "marketplace.json"
    catalog = load_json(catalog_path)
    if catalog is None:
        print_report()
        return 1

    for field in ("name", "plugins"):
        if field not in catalog:
            errors.append(f"marketplace.json: missing field '{field}'")
    if not isinstance(catalog.get("plugins"), list) or not catalog.get("plugins"):
        errors.append("marketplace.json: 'plugins' must be a non-empty list")
        print_report()
        return 1

    print(f"catalog '{catalog.get('name')}' version {catalog.get('version', 'n/a')}")

    listed_sources = set()
    for entry in catalog["plugins"]:
        name = entry.get("name", "<unnamed>")
        source = entry.get("source")
        if not source:
            errors.append(f"{name}: entry has no 'source'")
            continue
        listed_sources.add(source.strip("./").replace("\\", "/"))
        folder = (ROOT / source).resolve()
        if not folder.is_dir():
            errors.append(f"{name}: source folder {source} does not exist")
            continue

        manifest = load_json(folder / ".claude-plugin" / "plugin.json")
        if manifest is None:
            continue
        if manifest.get("name") != name:
            errors.append(
                f"{name}: plugin.json name '{manifest.get('name')}' differs from marketplace entry"
            )
        if entry.get("version") and manifest.get("version") != entry["version"]:
            errors.append(
                f"{name}: version mismatch — marketplace.json {entry['version']}, "
                f"plugin.json {manifest.get('version')}"
            )
        category = entry.get("category")
        if not category:
            warnings.append(f"{name}: no 'category' — business unit cannot be derived")
        elif category not in ALLOWED_CATEGORIES:
            errors.append(
                f"{name}: unknown category '{category}' — allowed: "
                + ", ".join(sorted(ALLOWED_CATEGORIES))
            )
        if not KEBAB.match(name):
            warnings.append(
                f"{name}: not kebab-case; Claude Code accepts it, the Claude.ai org sync requires kebab-case"
            )
        mcp = folder / ".mcp.json"
        if mcp.exists():
            load_json(mcp)
        check_doc_pairs(folder, f"plugins/{folder.name}")
        if manifest.get("version"):
            check_binaries(folder, manifest["version"], f"plugins/{folder.name}")
        print(f"  ok: {name} {manifest.get('version')} ({entry.get('category', 'no category')})")

    plugins_dir = ROOT / "plugins"
    if plugins_dir.is_dir():
        for folder in sorted(p for p in plugins_dir.iterdir() if p.is_dir()):
            rel = f"plugins/{folder.name}"
            if rel not in listed_sources:
                errors.append(f"{rel} exists but is not listed in marketplace.json")

    used = {e.get("category") for e in catalog["plugins"] if e.get("category")}
    unused = sorted(ALLOWED_CATEGORIES - used)
    if unused:
        print(f"  categories defined, no entries yet: {', '.join(unused)}")

    check_doc_pairs(ROOT, "repository root")
    docs = ROOT / "docs"
    if docs.is_dir():
        for en in sorted(docs.glob("*.md")):
            if en.stem.endswith("_de"):
                continue
            if not (docs / f"{en.stem}_de.md").exists():
                errors.append(f"docs: {en.name} has no German counterpart {en.stem}_de.md")

    for path in sorted(ROOT.rglob("*.json")):
        if ".git" in path.parts or "node_modules" in path.parts:
            continue
        load_json(path)

    print_report()
    return 1 if errors else 0


def print_report() -> None:
    for warning in warnings:
        print(f"::warning::{warning}")
    for error in errors:
        print(f"::error::{error}")
    print(f"\n{len(errors)} error(s), {len(warnings)} warning(s)")


if __name__ == "__main__":
    raise SystemExit(main())