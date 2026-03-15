# ZK Tool — Quick Zettelkasten Note Filing

A CLI tool that takes a quick thought and uses an LLM to classify, code, and file it into the correct location in this Obsidian vault.

## Vault Analysis

### ZK System (551 notes)

**Claims** (501 notes) — atomic assertions using Luhmann-style branching codes.
Filed into category folders under `ZK/Claims/`:

| Category | Notes | Example sub-branches |
|---|---|---|
| `2 - Software` | 252 | 2a0 Systems Thinking, 2a2 Architecture, 2e Diagrams, 2k State, 2m LLMs |
| `5 - Life - Productivity` | 51 | 5a Communication, 5e Creativity, 5f Knowledge Tools |
| `4 - Writing - Notetaking` | 47 | 4a Writing process, 4b Studying |
| `3 - Systems - Thoughts` | 34 | 3b-3g various |
| `7 - Art - Music` | 17 | 7 prefix codes |
| `8 - Philosophy, Activism` | 5 | 8 prefix codes |
| `Inbox` (unfiled) | 95 | dated or uncoded claims |

**Structure Notes** (50 notes) — hub notes (`SK - {code} - {topic}`) that link clusters of claims.

### Naming Conventions

- ZK Claims: `ZK - {Luhmann code} - {claim}.md` (395 coded), `ZK - {date} - {claim}.md` (124 dated), `ZK - {claim}.md` (89 uncoded)
- Structure Notes: `SK - {code} - {date} - {topic}.md`
- Blog Ideas: `BLOG IDEA - {topic}.md` (in `Writing/Ideas/`)
- Notes use prefix patterns: `BOOK -`, `DRAFT -`, `TALK -`, `COURSE -`, `DR -`

### Luhmann Code System

Alternating number-letter hierarchical IDs. Top-level number = category, then alternating letter (branch) and number (sequence):

```
2       = Software (category)
2a      = first branch (Systems)
2a0     = first sub-branch
2a0a    = first leaf
2a0a1   = deeper leaf
```

Deepest observed: 7 characters (e.g. `2a2a9b1`).

Sub-folders within `ZK/Claims/2 - Software/` group by branch code:
- `2a0 - Systems Programming - Systems Thinking/`
- `2a2 - Software Architecture - Systems Architecture/`
- `2d - Category Theory for Engineering/`
- `2e - Diagrams/`
- `2m - LLMs/`
- `Inbox/` (unfiled within category)

### Parallel Knowledge Areas

| Area | Location | Count | Purpose |
|---|---|---|---|
| Wiki | `Wiki/{domain}/` | 709 | Encyclopedic entries |
| Notes | `Notes/{type}/` | 585 | Reference material (Books: 393) |
| Writing Ideas | `Writing/Ideas/` | ~20 | Blog post ideas |
| Inbox | `Inbox/` | 188 | Unsorted capture |
| TIL | `TIL/` | small | Today I Learned |
| Snippets | `Snippets/` | small | Code snippets |

### ZK Claim Content Structure

A typical claim file:

```markdown
# {Claim as title}
[[Tag1]] [[Tag2]]

{1-3 paragraphs explaining the claim, with [[wikilinks]] to other ZK notes, books, concepts}

## Brainstorm
...

## Links
- [[related notes]]

## Logs
[[YYYY-MM-DD]]
- Created
```

No YAML frontmatter is used in ZK claims.

## Tool Design

### Usage

```bash
# Quick capture
zk "Software architecture is about managing rate of change across boundaries"

# Explicit type
zk --type blog-idea "Illustrated guide to debugging distributed systems"
zk --type wiki "Cosine similarity"
zk --type til "You can use git worktrees to run agents in isolation"

# Interactive — opens $EDITOR after scaffolding
zk -i "Prototyping with LLMs is like pair programming with an alien"

# Triage inbox
zk triage                  # walks ZK/Claims/Inbox/
zk triage --all Inbox/     # triage the main Inbox too
```

### Pipeline

1. **Build tree index** (`scripts/build_tree_index.py`) — scan filenames, produce `ZK/.tree-index.json`
2. **Classify + code** (`scripts/prompts/classify_and_code.md`) — LLM determines type, branch, code
3. **Generate note** — apply template, seed links
4. **Confirm** — show proposed filing, let user accept/edit/reject
5. **Write** — create the file

### File Map

- `PROJ - ZK Tool.md` — this file
- `scripts/build_tree_index.py` — scans vault, builds tree index JSON
- `scripts/prompts/classify_and_code.md` — the main LLM prompt
- `scripts/prompts/generate_note.md` — prompt for expanding a claim into a full note
- `scripts/zk_create.py` — the main CLI entry point (TODO)

## Open Questions

- Should the tool also update structure notes when adding a new claim to their branch?
- Should we auto-detect wikilinks from the existing vault vocabulary (all note titles)?
- Should triage mode suggest merging near-duplicate claims?
- MCP server vs standalone CLI vs Obsidian plugin?
