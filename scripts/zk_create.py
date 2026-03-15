#!/usr/bin/env python3
"""
zk_create.py — Quick ZK note creation with LLM-assisted classification and filing.

Usage:
    python zk_create.py "Some thought or claim"
    python zk_create.py --type blog-idea "Article about debugging"
    python zk_create.py --type wiki "Cosine similarity"
    python zk_create.py --dry-run "Test without writing"

Requires: ANTHROPIC_API_KEY environment variable (or uses claude CLI).
"""

import argparse
import json
import os
import subprocess
import sys
from datetime import date
from pathlib import Path

# Resolve vault path relative to this script
SCRIPT_DIR = Path(__file__).resolve().parent
PROJECT_DIR = SCRIPT_DIR.parent
VAULT_PATH = PROJECT_DIR.parent.parent  # Projects/ZK Tool/scripts -> vault root

PROMPTS_DIR = SCRIPT_DIR / "prompts"
CLASSIFY_PROMPT = PROMPTS_DIR / "classify_and_code.md"
GENERATE_PROMPT = PROMPTS_DIR / "generate_note.md"


def build_tree_index_text() -> str:
    """Run build_tree_index.py and capture the text output."""
    result = subprocess.run(
        [sys.executable, str(SCRIPT_DIR / "build_tree_index.py"),
         "--vault-path", str(VAULT_PATH),
         "--output", "text", "--stdout"],
        capture_output=True, text=True
    )
    if result.returncode != 0:
        print(f"Error building tree index: {result.stderr}", file=sys.stderr)
        sys.exit(1)
    return result.stdout


def build_classify_prompt(user_input: str, tree_index: str, type_override: str | None = None) -> str:
    """Build the full classification prompt from the template."""
    # Read the prompt template and extract the user prompt section
    # For now, construct it inline (the .md file documents the design)

    type_hint = ""
    if type_override:
        type_hint = f"\nThe user has explicitly requested this be filed as: {type_override}\nRespect this override.\n"

    return f"""You are a zettelkasten filing assistant for an Obsidian vault. Your job is to take
a raw thought or note and determine:
1. What TYPE of note it is
2. WHERE it should be filed
3. What LUHMANN CODE to assign (for ZK claims)
4. What LINKS to suggest

Here is the zettelkasten tree index:

<tree-index>
{tree_index}
</tree-index>

Here are the note type destinations:

| Type | Folder | Filename pattern | Description |
|---|---|---|---|
| zk-claim | ZK/Claims/{{category}}/{{subfolder}}/ | ZK - {{code}} - {{title}}.md | An atomic, opinionated assertion or insight |
| blog-idea | Writing/Ideas/ | BLOG IDEA - {{title}}.md | An idea for a blog post or article |
| wiki | Wiki/{{domain}}/ | {{title}}.md | An encyclopedic/definitional entry |
| til | TIL/ | TIL - {date.today().isoformat()} - {{title}}.md | A small factual discovery or trick |
| structure-note | ZK/Structure Notes/{{category}}/ | SK - {{code}} - {date.today().isoformat()} - {{title}}.md | A hub note linking a cluster of claims |
| research | Research/{date.today().isoformat()}/ | {{title}}.md | A research dump or exploration |

The ZK categories and their top-level branches are:
- 2 - Software: systems thinking (2a), category theory (2b,2d), concurrency (2c,2k), diagrams (2e,2m), debugging (2f), architecture (2g,2a2), programming+writing (2h), code compression (2i), supply chain/ecommerce (2j), career (2l), LLMs (2m,2n)
- 3 - Systems - Thoughts: general systems thinking, feedback loops
- 4 - Writing - Notetaking: writing process (4a), studying (4b), tools (6a,6b)
- 5 - Life - Productivity: communication (5a), creativity (5e), knowledge tools (5f), habits, mentoring
- 7 - Art - Music: music production, sessions, sound design
- 8 - Philosophy, Activism: political philosophy, ethics

Wiki domains: Autism, Creativity, General, Local, Mathematics, Music, People, Photography, Programming, Software, Technology, Writing
{type_hint}
Here is the input to classify:

<input>
{user_input}
</input>

Determine the best classification. For ZK claims, carefully examine the tree to find
the most relevant branch and the next available code.

Respond with ONLY this JSON (no markdown fences, no explanation outside the JSON):

{{
  "type": "zk-claim | blog-idea | wiki | til | structure-note | research",
  "reasoning": "Brief explanation of why this type and location",
  "title": "The note title",
  "code": "The Luhmann code or null",
  "category": "The category folder name or null",
  "subfolder": "The subfolder path or null",
  "filename": "The complete filename.md",
  "suggested_links": ["[[link1]]", "[[link2]]"],
  "tags": ["tag1", "tag2"],
  "parent_note": "Most closely related existing note filename or null"
}}"""


def build_generate_prompt(classification: dict, user_input: str) -> str:
    """Build the note content generation prompt."""
    note_type = classification["type"]
    title = classification["title"]
    code = classification.get("code", "")
    links = ", ".join(classification.get("suggested_links", []))
    tags = ", ".join(classification.get("tags", []))
    parent = classification.get("parent_note", "")
    today = date.today().isoformat()

    type_instructions = {
        "zk-claim": f"""Generate a ZK claim note with this structure:

# {title}
{{wikilink tags from the tags list, as [[Tag1]] [[Tag2]] on the second line}}

{{1-3 paragraphs expanding on the claim. Use [[wikilinks]] to connect to concepts.
Write in first person, conversational. Explain WHY this is true, give an example
if possible, connect to related ideas.}}

## Brainstorm
{{2-3 bullet points of directions to develop this}}

## Links
{{Links to related ZK notes, structure notes, wiki entries}}

## Logs
[[{today}]]
- Created""",

        "blog-idea": f"""Generate a blog idea note:

# {title}

## Core idea
{{1-2 paragraphs capturing the essence}}

## Outline
{{5-8 bullet points sketching a structure}}

## Related
{{Links to ZK claims, wiki entries that could feed this article}}""",

        "wiki": f"""Generate a wiki entry:

# {title}

{{2-4 paragraphs defining and explaining the concept. Include [[wikilinks]].
Encyclopedic tone — explain what it IS.}}

## See Also
{{Related wiki entries and ZK claims}}""",

        "til": f"""Generate a TIL note:

# {title}

{{1-2 paragraphs explaining what was learned and why it is useful.
Include code examples if relevant.}}

## Links
{{Related notes}}""",
    }

    instructions = type_instructions.get(note_type, type_instructions["zk-claim"])

    return f"""You are a zettelkasten writing assistant. Generate the note content
matching the vault's voice: conversational, first-person, with liberal [[wikilinks]].

Type: {note_type}
Title: {title}
Code: {code}
Suggested links: {links}
Tags: {tags}
Parent note: {parent}

Original input:
<input>
{user_input}
</input>

{instructions}

Output ONLY the markdown content. No code fences wrapping it."""


def call_llm(prompt: str) -> str:
    """Call the LLM. Tries anthropic SDK first, falls back to claude CLI."""
    try:
        import anthropic
        client = anthropic.Anthropic()
        response = client.messages.create(
            model="claude-sonnet-4-20250514",
            max_tokens=2000,
            messages=[{"role": "user", "content": prompt}],
        )
        return response.content[0].text
    except ImportError:
        pass

    # Fallback: use claude CLI or curl
    api_key = os.environ.get("ANTHROPIC_API_KEY")
    if not api_key:
        print("Error: No anthropic SDK and no ANTHROPIC_API_KEY set.", file=sys.stderr)
        print("Install: pip install anthropic", file=sys.stderr)
        sys.exit(1)

    import urllib.request

    data = json.dumps({
        "model": "claude-sonnet-4-20250514",
        "max_tokens": 2000,
        "messages": [{"role": "user", "content": prompt}],
    }).encode()

    req = urllib.request.Request(
        "https://api.anthropic.com/v1/messages",
        data=data,
        headers={
            "Content-Type": "application/json",
            "x-api-key": api_key,
            "anthropic-version": "2023-06-01",
        },
    )

    with urllib.request.urlopen(req) as resp:
        result = json.loads(resp.read())
        return result["content"][0]["text"]


def resolve_output_path(classification: dict) -> Path:
    """Determine the full filesystem path for the new note."""
    note_type = classification["type"]
    filename = classification["filename"]

    if note_type == "zk-claim":
        category = classification.get("category", "Inbox")
        subfolder = classification.get("subfolder")
        if subfolder:
            return VAULT_PATH / "ZK" / "Claims" / subfolder / filename
        else:
            return VAULT_PATH / "ZK" / "Claims" / category / filename

    elif note_type == "blog-idea":
        return VAULT_PATH / "Writing" / "Ideas" / filename

    elif note_type == "wiki":
        domain = classification.get("category", "General")
        return VAULT_PATH / "Wiki" / domain / filename

    elif note_type == "til":
        return VAULT_PATH / "TIL" / filename

    elif note_type == "structure-note":
        category = classification.get("category", "")
        return VAULT_PATH / "ZK" / "Structure Notes" / category / filename

    elif note_type == "research":
        return VAULT_PATH / "Research" / date.today().isoformat() / filename

    else:
        return VAULT_PATH / "Inbox" / filename


def display_classification(classification: dict, output_path: Path):
    """Show the proposed classification for user confirmation."""
    print()
    print(f"  Type:     {classification['type']}")
    if classification.get("code"):
        print(f"  Code:     {classification['code']}")
    print(f"  Title:    {classification['title']}")
    print(f"  Path:     {output_path.relative_to(VAULT_PATH)}")
    print(f"  File:     {classification['filename']}")
    if classification.get("suggested_links"):
        print(f"  Links:    {', '.join(classification['suggested_links'][:5])}")
    if classification.get("reasoning"):
        print(f"  Reason:   {classification['reasoning']}")
    print()


def confirm(prompt: str = "  [Enter] accept  [e] edit title  [t] change type  [s] skip  > ") -> str:
    """Get user confirmation."""
    try:
        return input(prompt).strip().lower()
    except (EOFError, KeyboardInterrupt):
        print()
        return "s"


def main():
    parser = argparse.ArgumentParser(
        description="Create a ZK note with LLM-assisted classification and filing"
    )
    parser.add_argument("input", nargs="*", help="The thought/claim/idea to file")
    parser.add_argument("--type", "-t", choices=[
        "zk-claim", "blog-idea", "wiki", "til", "structure-note", "research"
    ], help="Override the auto-detected type")
    parser.add_argument("--dry-run", "-n", action="store_true",
                        help="Show what would be created without writing")
    parser.add_argument("--no-generate", action="store_true",
                        help="Only classify, don't generate note content")
    parser.add_argument("--vault-path", type=Path, default=VAULT_PATH,
                        help="Path to vault root")
    args = parser.parse_args()

    if not args.input:
        parser.print_help()
        sys.exit(1)

    user_input = " ".join(args.input)
    print(f"Input: {user_input}")
    print("Building tree index...")

    tree_index = build_tree_index_text()

    print("Classifying...")
    classify_prompt = build_classify_prompt(user_input, tree_index, args.type)
    raw_response = call_llm(classify_prompt)

    # Parse JSON from response (handle possible markdown fences)
    json_text = raw_response.strip()
    if json_text.startswith("```"):
        json_text = "\n".join(json_text.split("\n")[1:-1])

    try:
        classification = json.loads(json_text)
    except json.JSONDecodeError:
        print(f"Error: Could not parse LLM response as JSON:", file=sys.stderr)
        print(raw_response, file=sys.stderr)
        sys.exit(1)

    output_path = resolve_output_path(classification)
    display_classification(classification, output_path)

    if args.dry_run:
        print("  (dry run — not writing)")
        return

    choice = confirm()
    if choice == "s":
        print("Skipped.")
        return
    elif choice == "e":
        new_title = input("  New title: ").strip()
        if new_title:
            classification["title"] = new_title
            # Rebuild filename
            if classification.get("code"):
                classification["filename"] = f"ZK - {classification['code']} - {new_title}.md"
            output_path = resolve_output_path(classification)
            display_classification(classification, output_path)
            if confirm() == "s":
                return

    if args.no_generate:
        # Write a minimal stub
        content = f"# {classification['title']}\n\n{user_input}\n\n## Links\n\n## Logs\n[[{date.today().isoformat()}]]\n- Created\n"
    else:
        print("Generating note content...")
        gen_prompt = build_generate_prompt(classification, user_input)
        content = call_llm(gen_prompt)

    # Ensure parent directory exists
    output_path.parent.mkdir(parents=True, exist_ok=True)

    if output_path.exists():
        print(f"  Warning: {output_path.relative_to(VAULT_PATH)} already exists!")
        if confirm("  Overwrite? [y/N] > ") != "y":
            print("Skipped.")
            return

    output_path.write_text(content)
    print(f"  Created: {output_path.relative_to(VAULT_PATH)}")


if __name__ == "__main__":
    main()
