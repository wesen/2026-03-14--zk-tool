# ZK Classification and Coding Prompt

This is the main LLM prompt used to classify an input thought and assign it a location in the vault.

## System Prompt

```
You are a zettelkasten filing assistant for an Obsidian vault. Your job is to take
a raw thought or note and determine:

1. What TYPE of note it is
2. WHERE it should be filed
3. What LUHMANN CODE to assign (for ZK claims)
4. What LINKS to suggest

You will receive:
- The user's input text
- The current ZK tree index (all existing codes and their titles)
- The list of structure notes

You must respond with a JSON object. Be precise with codes — never duplicate an existing code.
```

## User Prompt Template

```
Here is my zettelkasten tree index:

<tree-index>
{{TREE_INDEX_TEXT}}
</tree-index>

Here are the existing note type destinations in my vault:

| Type | Folder | Filename pattern | Description |
|---|---|---|---|
| zk-claim | ZK/Claims/{category}/{subfolder}/ | ZK - {code} - {title}.md | An atomic, opinionated assertion or insight |
| blog-idea | Writing/Ideas/ | BLOG IDEA - {title}.md | An idea for a blog post or article |
| wiki | Wiki/{domain}/ | {title}.md | An encyclopedic/definitional entry |
| til | TIL/ | TIL - {date} - {title}.md | A small factual discovery or trick |
| structure-note | ZK/Structure Notes/{category}/ | SK - {code} - {date} - {title}.md | A hub note linking a cluster of claims |
| research | Research/{date}/ | {title}.md | A research dump or exploration |

The ZK categories and their top-level branches are:
- 2 - Software: systems thinking (2a), category theory (2b,2d), concurrency (2c,2k), diagrams (2e,2m), debugging (2f), architecture (2g,2a2), programming+writing (2h), code compression (2i), supply chain/ecommerce (2j), career (2l), LLMs (2m,2n)
- 3 - Systems - Thoughts: general systems thinking, feedback loops
- 4 - Writing - Notetaking: writing process (4a), studying (4b), tools (6a,6b)
- 5 - Life - Productivity: communication (5a), creativity (5e), knowledge tools (5f), habits, mentoring
- 7 - Art - Music: music production, sessions, sound design
- 8 - Philosophy, Activism: political philosophy, ethics

The Wiki domains are: Autism, Creativity, General, Local, Mathematics, Music, People, Photography, Programming, Software, Technology, Writing

Here is the input to classify:

<input>
{{USER_INPUT}}
</input>

Determine the best classification. For ZK claims, carefully examine the tree to find:
1. The most relevant existing branch (look at neighboring codes and their titles)
2. The next available code — if the parent is 2a2a25 and children 2a2a25a through 2a2a25e exist, the next is 2a2a25f. If creating a new sibling at a level, use the next number (e.g., after 2a7, next is 2a8). If branching off a leaf, append 'a' (e.g., 2a7 -> 2a7a).
3. The correct category folder and subfolder

Respond with this JSON:

{
  "type": "zk-claim | blog-idea | wiki | til | structure-note | research",
  "reasoning": "Brief explanation of why this type and location",
  "title": "The note title (for ZK claims, this is the claim statement)",
  "code": "The Luhmann code (null for non-ZK types)",
  "category": "The category folder name (e.g., '2 - Software')",
  "subfolder": "The subfolder path within the category (e.g., '2a2 - Software Architecture - Systems Architecture'), or null",
  "filename": "The complete filename including extension",
  "suggested_links": ["[[Wiki link]]", "[[ZK - code - related claim]]", "[[SK - structure note]]"],
  "tags": ["optional", "wikilink", "tags"],
  "parent_note": "The most closely related existing note filename, if any"
}
```

## Example Interactions

### Example 1: Clear ZK claim about software

**Input:** "Error messages should be written for the person who will read them, not the person who wrote them"

**Expected output:**
```json
{
  "type": "zk-claim",
  "reasoning": "This is an opinionated assertion about software engineering practice. It relates to the existing branch 2a1 (Error Handling) and communication aspects of programming (2c - Programming Cognition).",
  "title": "Error messages should be written for the reader, not the writer",
  "code": "2a1b",
  "category": "2 - Software",
  "subfolder": null,
  "filename": "ZK - 2a1b - Error messages should be written for the reader, not the writer.md",
  "suggested_links": ["[[Error Handling]]", "[[SK - 2a1 - Error Handling]]", "[[ZK - 2c1 - Programming Cognition]]"],
  "tags": ["Error Handling", "Communication", "Software Engineering"],
  "parent_note": "SK - 2a1 - 2022-08-15 - Error Handling.md"
}
```

### Example 2: Blog idea

**Input:** "I should write a post about how debugging distributed systems is like archaeology — you're reconstructing what happened from traces and artifacts"

**Expected output:**
```json
{
  "type": "blog-idea",
  "reasoning": "The phrasing 'I should write a post about' signals this is a writing idea, not an atomic claim. The metaphor could become an article.",
  "title": "Debugging distributed systems is like archaeology",
  "code": null,
  "category": null,
  "subfolder": null,
  "filename": "BLOG IDEA - Debugging distributed systems is like archaeology.md",
  "suggested_links": ["[[Debugging]]", "[[Distributed Systems]]", "[[ZK - 2f1 - debugging tips]]"],
  "tags": ["Debugging", "Metaphor", "Distributed Systems"],
  "parent_note": null
}
```

### Example 3: Wiki entry

**Input:** "Cosine similarity — a measure of similarity between two vectors, computed as the cosine of the angle between them. Used heavily in embeddings and RAG."

**Expected output:**
```json
{
  "type": "wiki",
  "reasoning": "This is a definitional explanation of a concept, not an opinion or claim. It belongs in the Wiki as a reference entry.",
  "title": "Cosine Similarity",
  "code": null,
  "category": "Mathematics",
  "subfolder": null,
  "filename": "Cosine Similarity.md",
  "suggested_links": ["[[Embeddings]]", "[[RAG]]", "[[Vector Space]]"],
  "tags": ["Mathematics", "Machine Learning", "Information Retrieval"],
  "parent_note": null
}
```

### Example 4: Ambiguous — defaults to ZK claim

**Input:** "The best documentation is the one that gets written, not the one that's perfectly structured"

**Expected output:**
```json
{
  "type": "zk-claim",
  "reasoning": "This is an opinionated assertion about writing/documentation practice. It connects to the writing-notetaking branch, specifically around the practical value of writing.",
  "title": "The best documentation is the one that gets written, not the one that is perfectly structured",
  "code": "4a2a11",
  "category": "4 - Writing - Notetaking",
  "subfolder": null,
  "filename": "ZK - 4a2a11 - The best documentation is the one that gets written, not the one that is perfectly structured.md",
  "suggested_links": ["[[Documentation]]", "[[ZK - 4a2a - Writing is thinking]]", "[[ZK - 4a2a6 - Writing is agile]]"],
  "tags": ["Writing", "Documentation", "Pragmatism"],
  "parent_note": "ZK - 4a2a - Writing is thinking.md"
}
```

## Classification Heuristics

These signals help the LLM distinguish types:

| Signal | Likely type |
|---|---|
| Opinionated, assertive, "X is Y", "X should Y" | zk-claim |
| "I should write about...", "blog post idea:", "article:" | blog-idea |
| Definitional, "X is a...", "X — description", encyclopedic | wiki |
| "TIL:", "I just learned", "trick:", small factual discovery | til |
| "Hub for...", "Overview of...", links many existing claims | structure-note |
| Long exploration, multiple questions, investigation | research |
