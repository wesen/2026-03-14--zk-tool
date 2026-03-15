#!/usr/bin/env python3
"""
Build a tree index of all ZK claims and structure notes from filenames.

Scans the ZK/ directory, extracts Luhmann codes from filenames,
and produces a JSON index used by the LLM classification prompt.

Usage:
    python build_tree_index.py [--vault-path /path/to/vault] [--output json|text]

Output: ZK/.tree-index.json (or stdout in text mode)
"""

import argparse
import json
import os
import re
import sys
from collections import defaultdict
from pathlib import Path


def extract_code_and_title(filename: str, prefix: str = "ZK") -> tuple[str | None, str | None]:
    """Extract Luhmann code and title from a ZK filename.

    Handles three patterns:
      ZK - 2a0a - Some claim title.md          -> ("2a0a", "Some claim title")
      ZK - 2022-08-13 - Some dated claim.md    -> (None, "2022-08-13 - Some dated claim")
      ZK - Some uncoded claim.md                -> (None, "Some uncoded claim")
    """
    stem = filename.removesuffix(".md")

    # Pattern 1: coded — "ZK - {code} - {title}"
    m = re.match(rf"^{prefix} - ([0-9][a-z0-9]+) - (.+)$", stem)
    if m:
        return m.group(1), m.group(2).strip()

    # Pattern 2: dated — "ZK - 2022-08-13 - {title}"
    m = re.match(rf"^{prefix} - (\d{{4}}-\d{{2}}-\d{{2}}) - (.+)$", stem)
    if m:
        return None, f"{m.group(1)} - {m.group(2).strip()}"

    # Pattern 3: uncoded — "ZK - {title}"
    m = re.match(rf"^{prefix} - (.+)$", stem)
    if m:
        return None, m.group(1).strip()

    return None, None


def scan_claims(vault_path: Path) -> dict:
    """Scan ZK/Claims/ and return structured tree data."""
    claims_dir = vault_path / "ZK" / "Claims"
    if not claims_dir.exists():
        print(f"Error: {claims_dir} does not exist", file=sys.stderr)
        sys.exit(1)

    categories = defaultdict(lambda: {"coded": [], "uncoded": [], "subfolders": set()})

    for root, dirs, files in os.walk(claims_dir):
        rel = Path(root).relative_to(claims_dir)
        parts = rel.parts

        # Determine category (first directory level)
        if len(parts) == 0:
            continue
        category = parts[0]

        # Track subfolder names
        if len(parts) > 1:
            categories[category]["subfolders"].add(str(rel))

        for f in files:
            if not f.endswith(".md"):
                continue
            code, title = extract_code_and_title(f, prefix="ZK")
            if title is None:
                continue

            entry = {
                "code": code,
                "title": title,
                "path": str(Path(root) / f),
                "subfolder": str(rel) if len(parts) > 1 else None,
            }

            if code:
                categories[category]["coded"].append(entry)
            else:
                categories[category]["uncoded"].append(entry)

    # Sort coded entries by code
    for cat in categories:
        categories[cat]["coded"].sort(key=lambda x: x["code"])
        categories[cat]["subfolders"] = sorted(categories[cat]["subfolders"])

    return dict(categories)


def scan_structure_notes(vault_path: Path) -> list:
    """Scan ZK/Structure Notes/ and return list of structure notes."""
    sn_dir = vault_path / "ZK" / "Structure Notes"
    if not sn_dir.exists():
        return []

    notes = []
    for root, dirs, files in os.walk(sn_dir):
        for f in sorted(files):
            if not f.endswith(".md"):
                continue
            code, title = extract_code_and_title(f, prefix="SK")
            if title is None:
                continue
            rel = Path(root).relative_to(sn_dir)
            notes.append({
                "code": code,
                "title": title,
                "category": str(rel) if str(rel) != "." else None,
            })

    return notes


def build_index(vault_path: Path) -> dict:
    """Build the full tree index."""
    return {
        "vault_path": str(vault_path),
        "claims": scan_claims(vault_path),
        "structure_notes": scan_structure_notes(vault_path),
    }


def render_text(index: dict) -> str:
    """Render the tree index as human-readable text (for LLM prompts)."""
    lines = []
    lines.append("# ZK Tree Index\n")

    lines.append("## Structure Notes\n")
    for sn in index["structure_notes"]:
        code_str = sn["code"] or "?"
        cat = sn["category"] or "root"
        lines.append(f"  SK {code_str}: {sn['title']} [{cat}]")

    lines.append("\n## Claims by Category\n")
    for cat_name in sorted(index["claims"].keys()):
        cat = index["claims"][cat_name]
        lines.append(f"### {cat_name}")

        if cat["subfolders"]:
            lines.append(f"  Subfolders: {', '.join(cat['subfolders'])}")

        for entry in cat["coded"]:
            lines.append(f"  {entry['code']}: {entry['title']}")

        if cat["uncoded"]:
            lines.append(f"  ({len(cat['uncoded'])} uncoded notes)")

        lines.append("")

    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(description="Build ZK tree index")
    parser.add_argument(
        "--vault-path",
        type=Path,
        default=Path(__file__).resolve().parent.parent.parent,
        help="Path to the Obsidian vault root",
    )
    parser.add_argument(
        "--output",
        choices=["json", "text"],
        default="json",
        help="Output format",
    )
    parser.add_argument(
        "--stdout",
        action="store_true",
        help="Print to stdout instead of writing to file",
    )
    args = parser.parse_args()

    index = build_index(args.vault_path)

    if args.output == "text":
        text = render_text(index)
        if args.stdout:
            print(text)
        else:
            out_path = args.vault_path / "ZK" / ".tree-index.txt"
            out_path.write_text(text)
            print(f"Wrote {out_path}", file=sys.stderr)
    else:
        if args.stdout:
            json.dump(index, sys.stdout, indent=2)
        else:
            out_path = args.vault_path / "ZK" / ".tree-index.json"
            out_path.write_text(json.dumps(index, indent=2))
            print(f"Wrote {out_path}", file=sys.stderr)


if __name__ == "__main__":
    main()
