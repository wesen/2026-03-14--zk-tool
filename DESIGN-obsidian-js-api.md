# obsidian.js — A Clean JS API for Obsidian CLI Scripting

## Context & Goals

The [official Obsidian CLI](https://help.obsidian.md/cli) (shipped in v1.12, Feb 2026) provides
70+ commands for vault manipulation. The existing wrapper [`obsidian-ts`](https://github.com/kitschpatrol/obsidian-ts)
is a faithful 1:1 mirror of the CLI's command structure — but it's just a thin shell-out layer.

This library aims to be **a scripting toolkit**, not just a CLI wrapper. It should make it easy
to write concise, composable scripts — vault health reports, batch refactors, AI-powered filing,
daily automation — in plain JavaScript.

### Design Principles

1. **Concise by default** — common operations should be one-liners
2. **Composable** — return arrays/objects that chain with standard JS (`.filter`, `.map`, `for await`)
3. **Batteries included** — markdown parsing, frontmatter, wikilink extraction built in
4. **Lazy where possible** — don't fetch what you don't need; cache what you already have
5. **Plain JS** — works in Node scripts, not just TypeScript projects; no build step required
6. **Escape hatch** — always possible to drop down to raw CLI execution

---

## Architecture

```
obsidian.js
├── core/            # CLI execution, config, parsing
│   ├── exec.js      # spawn obsidian binary, parse output
│   ├── config.js    # vault selection, binary path
│   └── parse.js     # JSON, key-value, line list parsers
├── commands/        # 1:1 CLI wrappers (thin layer)
│   ├── file.js
│   ├── search.js
│   ├── link.js
│   ├── property.js
│   ├── tag.js
│   ├── task.js
│   ├── daily.js
│   ├── template.js
│   ├── plugin.js
│   └── ...
├── note.js          # High-level Note object with markdown ops
├── query.js         # Fluent query builder
├── batch.js         # Batch/bulk operations with progress
├── markdown.js      # Frontmatter, wikilinks, headings parser
└── index.js         # Main entry point
```

---

## API Design

### 1. Initialization & Configuration

```js
import obs from 'obsidian.js'

// Configure once
obs.configure({ vault: 'My Vault' })

// Or use per-call
const files = await obs.files({ vault: 'Other Vault' })

// Check connection
const v = await obs.version()  // '1.12.4'
```

### 2. Files — The Core Primitives

```js
// List files (returns string[] of paths)
const all = await obs.files()
const mdOnly = await obs.files({ folder: 'ZK/Claims', ext: 'md' })

// CRUD
const content = await obs.read('ZK - 2a0 - Systems thinking.md')
const content2 = await obs.read({ path: 'ZK/Claims/2 - Software/note.md' })

await obs.create('New Note', { content: '# Hello', folder: 'Inbox' })
await obs.create('From Template', { template: 'ZK Claim' })

await obs.append('Daily Log', '- 3:45 PM finished review')
await obs.prepend('Meeting Notes', '## Action Items\n')

// Move with automatic wikilink rewriting (the killer feature)
await obs.move('Old Location/note.md', 'New Location/')
await obs.rename('old-name.md', 'better-name.md')

// Delete (trash by default, permanent opt-in)
await obs.delete('scratch.md')
await obs.delete('scratch.md', { permanent: true })
```

**Design decision**: `obs.read(name)` accepts a string for wikilink-style resolution.
Pass `{ path: '...' }` for exact paths. This mirrors how Obsidian itself resolves links.

### 3. Note — Rich Object for a Single Note

```js
const note = await obs.note('ZK - 2a0 - Systems thinking')

// Metadata
note.path       // 'ZK/Claims/2 - Software/2a0 - .../ZK - 2a0 - Systems thinking.md'
note.name       // 'ZK - 2a0 - Systems thinking'
note.created    // Date
note.modified   // Date
note.size       // number (bytes)

// Content
note.content    // raw markdown string (lazy-loaded on first access)
note.frontmatter // parsed YAML frontmatter as object (or null)
note.body       // markdown without frontmatter
note.headings   // ['Systems thinking', 'Brainstorm', 'Links', 'Logs']
note.wikilinks  // ['[[Category Theory]]', '[[2a1 - Feedback loops]]']
note.tags       // ['#software', '#systems']
note.tasks      // [{ text: 'Follow up', done: false, line: 14 }]

// Links
note.backlinks  // [{ file: 'SK - 2a - Systems.md', count: 3 }, ...]
note.outgoing   // ['Category Theory.md', '2a1 - Feedback loops.md']

// Mutations (write through CLI so wikilinks update)
await note.setProperty('status', 'reviewed')
await note.removeProperty('draft')
await note.append('\n## New Section\n')
await note.moveTo('ZK/Claims/2 - Software/2a0/')
await note.rename('ZK - 2a0 - Better title')
```

**Why a Note object?** Scripts that touch individual notes often need content + metadata + links
together. Fetching them separately is tedious and slow (multiple CLI calls). The Note object
batches these and caches results.

### 4. Query Builder — Find Notes Fluently

```js
// Simple search
const results = await obs.search('distributed systems')

// Fluent queries
const orphans = await obs.query()
  .orphans()
  .inFolder('ZK/Claims')
  .run()

const recentZK = await obs.query()
  .inFolder('ZK/Claims')
  .withTag('software')
  .modifiedAfter('2026-01-01')
  .sortBy('modified', 'desc')
  .limit(20)
  .run()

// Full-text search with context
const hits = await obs.query()
  .search('rate of change')
  .withContext()          // include surrounding lines
  .run()
// => [{ file: 'ZK - 2g1 - ...', matches: [{ line: 5, text: '...' }] }]

// Property-based queries
const drafts = await obs.query()
  .whereProperty('status', 'draft')
  .run()

// Unresolved links (broken wikilinks)
const broken = await obs.query().unresolved().run()

// Dead ends (notes with no outgoing links)
const deadEnds = await obs.query().deadEnds().run()
```

**Implementation**: The query builder composes CLI calls. `.inFolder()` maps to `folder=`,
`.search()` uses `search` or `search:context`, `.orphans()` uses `orphans`, etc.
When filters can't be expressed as CLI args, the builder post-filters in JS.

### 5. Batch Operations — Bulk With Progress

```js
// Iterate over notes matching a query
for await (const note of obs.each({ folder: 'ZK/Claims' })) {
  if (!note.content.includes('## Links')) {
    await note.append('\n## Links\n')
  }
}

// Batch with progress callback
await obs.batch(
  obs.query().inFolder('Inbox'),
  async (note) => {
    await note.setProperty('reviewed', 'false')
  },
  {
    concurrency: 1,     // sequential (safe for Obsidian)
    dryRun: false,
    onProgress: (done, total) => console.log(`${done}/${total}`)
  }
)

// Bulk tag rename (uses CLI's native bulk rename)
await obs.tags.rename('old-tag', 'new-tag')

// Collect stats
const stats = await obs.batch(
  obs.query().inFolder('ZK/Claims'),
  async (note) => ({
    name: note.name,
    links: note.outgoing.length,
    words: (await obs.wordcount(note.path)).words,
  })
)
// stats is an array of results
```

### 6. Markdown Utilities

```js
import { md } from 'obsidian.js'

// Parse frontmatter
const { frontmatter, body } = md.parse(rawMarkdown)
// frontmatter: { tags: ['foo'], status: 'draft' }
// body: '# Title\n...'

// Extract wikilinks from text
const links = md.wikilinks('See [[Category Theory]] and [[2a0 - Systems]]')
// ['Category Theory', '2a0 - Systems']

// Extract tags
const tags = md.tags('Text with #software and #systems-thinking')
// ['software', 'systems-thinking']

// Extract headings
const headings = md.headings(body)
// [{ level: 1, text: 'Title' }, { level: 2, text: 'Links' }]

// Build note content
const content = md.note({
  title: 'New Claim',
  wikiTags: ['Software', 'Architecture'],
  body: 'Architecture is about boundaries...',
  sections: {
    'Brainstorm': '- Explore boundary patterns\n- Rate of change',
    'Links': '- [[2g - Architecture]]\n- [[2a0 - Systems]]',
    'Logs': `[[${new Date().toISOString().slice(0,10)}]]\n- Created`,
  },
})
// Returns formatted markdown string
```

### 7. Daily Notes

```js
// Today's daily note
const today = await obs.daily.read()
await obs.daily.append('- 15:30 Finished code review')
await obs.daily.prepend('## Morning Goals\n- Ship feature X')

// Open in Obsidian
await obs.daily.open()

// Get path without opening
const path = await obs.daily.path()
```

### 8. Tags

```js
const tags = await obs.tags.list()                  // [{ tag: 'software', count: 42 }, ...]
const tags = await obs.tags.list({ sort: 'count' }) // sorted by frequency
const info = await obs.tags.info('software')         // detailed usage info
await obs.tags.rename('old-name', 'new-name')       // vault-wide rename
```

### 9. Properties (Frontmatter)

```js
// Vault-wide property index
const props = await obs.properties.list()
// [{ name: 'status', type: 'text', count: 234 }, ...]

// Per-file
const status = await obs.properties.read('note.md', 'status')
await obs.properties.set('note.md', 'status', 'reviewed')
await obs.properties.set('note.md', 'priority', 5, { type: 'number' })
await obs.properties.remove('note.md', 'draft')

// Aliases
const aliases = await obs.properties.aliases('note.md')
```

### 10. Tasks

```js
const todos = await obs.tasks.list({ todo: true })
// [{ file: 'Daily/2026-03-15.md', line: 12, status: ' ', text: 'Ship feature' }]

const dailyTasks = await obs.tasks.list({ daily: true })

// Toggle/complete
await obs.tasks.toggle({ path: 'Daily/2026-03-15.md', line: 12 })
await obs.tasks.done({ path: 'Daily/2026-03-15.md', line: 12 })
```

### 11. Links & Graph

```js
// Per-note
const back = await obs.links.backlinks('ZK - 2a0 - Systems')
const out  = await obs.links.outgoing('ZK - 2a0 - Systems')

// Vault health
const orphans    = await obs.links.orphans()     // no incoming links
const deadEnds   = await obs.links.deadEnds()    // no outgoing links
const unresolved = await obs.links.unresolved()  // broken [[links]]
```

### 12. Plugins & Themes

```js
const plugins = await obs.plugins.list()
await obs.plugins.enable('dataview')
await obs.plugins.disable('obsidian-git')
await obs.plugins.install('yanki', { enable: true })
await obs.plugins.reload('my-plugin')

await obs.themes.set('Minimal')
```

### 13. Templates

```js
const templates = await obs.templates.list()
const content = await obs.templates.read('ZK Claim', { resolve: true, title: 'My Title' })
await obs.templates.insert('ZK Claim')  // into active file
```

### 14. Dev / Eval — Escape Hatch

```js
// Execute arbitrary JS in Obsidian's runtime
const result = await obs.eval('app.vault.getMarkdownFiles().length')

// Raw CLI execution (ultimate escape hatch)
const raw = await obs.exec('some:command', { key: 'value' }, ['flag'])
```

### 15. Vault

```js
const vaults = await obs.vaults.list()        // ['My Vault', 'Work']
const info = await obs.vaults.info()           // { name, path, files, folders, size }
await obs.vaults.open('Work')
```

---

## Higher-Level Patterns (The Scripting Power)

These patterns show why a proper API layer matters beyond raw CLI wrapping.

### Vault Health Report

```js
import obs from 'obsidian.js'
obs.configure({ vault: 'My Vault' })

const orphans    = await obs.links.orphans()
const deadEnds   = await obs.links.deadEnds()
const unresolved = await obs.links.unresolved()
const totalFiles = await obs.files.total()
const totalTags  = await obs.tags.total()

const report = `# Vault Health — ${new Date().toLocaleDateString()}
- **Files**: ${totalFiles}
- **Tags**: ${totalTags}
- **Orphans**: ${orphans.length} notes with no incoming links
- **Dead ends**: ${deadEnds.length} notes with no outgoing links
- **Broken links**: ${unresolved.length} unresolved wikilinks

## Top broken links
${unresolved.slice(0, 10).map(u => `- [[${u.link}]] (${u.count} references)`).join('\n')}

## Orphaned notes
${orphans.slice(0, 20).map(f => `- [[${f}]]`).join('\n')}
`

await obs.create('Vault Health Report', {
  content: report,
  folder: 'Meta',
  overwrite: true,
})
```

### ZK Filing Pipeline (Replaces zk_create.py)

```js
import obs from 'obsidian.js'
import { md } from 'obsidian.js'
import Anthropic from '@anthropic-ai/sdk'

obs.configure({ vault: 'My Vault' })
const llm = new Anthropic()

export async function zkFile(thought) {
  // 1. Build context from existing vault
  const claims = await obs.files({ folder: 'ZK/Claims' })
  const tags = await obs.tags.list({ sort: 'count' })

  // 2. Classify with LLM
  const classification = await classify(thought, claims, tags)

  // 3. Generate content
  const content = md.note({
    title: classification.title,
    wikiTags: classification.tags,
    body: classification.body,
    sections: {
      'Brainstorm': classification.brainstorm,
      'Links': classification.links.map(l => `- ${l}`).join('\n'),
      'Logs': `[[${new Date().toISOString().slice(0,10)}]]\n- Created`,
    },
  })

  // 4. Create via CLI (wikilinks auto-index)
  await obs.create(classification.filename, {
    content,
    folder: classification.folder,
  })

  return classification
}
```

### Batch Inbox Triage

```js
import obs from 'obsidian.js'

for await (const note of obs.each({ folder: 'Inbox' })) {
  const links = note.wikilinks
  const tags = note.tags

  if (links.length === 0 && tags.length === 0) {
    console.log(`ORPHAN: ${note.name} — no links, no tags`)
  }

  // Auto-tag based on content keywords
  if (note.content.includes('debug') && !tags.includes('#debugging')) {
    await note.append('\n#debugging')
  }
}
```

### Morning Automation

```js
import obs from 'obsidian.js'

// Ensure daily note exists and add morning template
await obs.daily.open()

const yesterday = await obs.tasks.list({ daily: true, todo: true })
if (yesterday.length > 0) {
  const carried = yesterday.map(t => `- [ ] ${t.text} (carried over)`).join('\n')
  await obs.daily.append(`\n## Carried Over\n${carried}`)
}

// Vault stats in daily note
const total = await obs.files.total()
await obs.daily.append(`\n> Vault: ${total} files`)
```

### Find Similar Notes (for dedup)

```js
import obs from 'obsidian.js'

const note = await obs.note('ZK - 2a0 - Systems thinking')
const keywords = note.wikilinks.slice(0, 3).map(l => l.replace(/\[\[|\]\]/g, ''))

for (const keyword of keywords) {
  const related = await obs.search(keyword)
  console.log(`\n"${keyword}" appears in:`)
  for (const path of related.slice(0, 5)) {
    if (path !== note.path) console.log(`  - ${path}`)
  }
}
```

---

## Implementation Notes

### CLI Execution Model

The Obsidian CLI is a **remote control** — it talks to a running Obsidian instance via IPC.
This means:

- Obsidian must be running (auto-launches if not)
- Commands are **sequential** — no parallel execution against the same vault
- Every command is a subprocess spawn (`obsidian [vault=X] <command> key=val flag`)
- Output is text (lines, key-value, or JSON depending on `format=` flag)

The `exec` layer handles:
- Argument serialization (`key=value` pairs + bare flags)
- Vault prefix injection from config
- Output parsing (JSON, key-value, line lists)
- Error detection (both non-zero exit and `Error:` prefix in stdout)
- Version compatibility checking

### What We Add Over obsidian-ts

| Feature | obsidian-ts | obsidian.js |
|---|---|---|
| CLI wrapping | Yes (1:1) | Yes (simplified surface) |
| String-based file resolution | Yes | Yes (`obs.read('name')`) |
| Note object with content + links | No | Yes (`obs.note()`) |
| Markdown parsing (frontmatter, links) | No | Yes (`md.parse()`, `md.wikilinks()`) |
| Query builder | No | Yes (`obs.query().inFolder()...`) |
| Batch operations with progress | No | Yes (`obs.batch()`, `obs.each()`) |
| Markdown content builder | No | Yes (`md.note()`) |
| Caching | No | Yes (Note object caches metadata) |
| Pure JS (no build step) | No (TS only) | Yes |
| Node version | 24.1+ | 18+ |

### Dependencies (Minimal)

- **No build step** — pure ESM, works with `node --experimental-modules` or modern Node
- `yaml` — for frontmatter parsing (battle-tested, small)
- That's it. The CLI wrapper itself is just `child_process.execFile`.

### Error Handling

```js
import { ObsidianError, NotFoundError, VersionError } from 'obsidian.js'

try {
  await obs.read('nonexistent.md')
} catch (err) {
  if (err instanceof NotFoundError) {
    // Obsidian CLI binary not on PATH
  } else if (err instanceof VersionError) {
    // CLI version incompatible
  } else if (err instanceof ObsidianError) {
    // Command failed (file not found, invalid args, etc.)
    console.error(err.message)  // human-readable error from CLI
    console.error(err.stdout)   // raw stdout
    console.error(err.stderr)   // raw stderr
  }
}
```

---

## File Layout

```
obsidian.js/
├── package.json
├── index.js              # Main export: obs default + named { md, ObsidianError, ... }
├── lib/
│   ├── exec.js           # CLI subprocess execution
│   ├── config.js         # Global config (vault, binary path)
│   ├── parse.js          # Output parsers (JSON, lines, key-value)
│   ├── errors.js         # Error classes
│   ├── note.js           # Note class (lazy content + metadata)
│   ├── query.js          # QueryBuilder class
│   ├── batch.js          # batch() and each() helpers
│   ├── markdown.js       # Frontmatter, wikilinks, headings, note builder
│   └── commands/         # Thin CLI wrappers
│       ├── file.js
│       ├── search.js
│       ├── link.js
│       ├── property.js
│       ├── tag.js
│       ├── task.js
│       ├── daily.js
│       ├── template.js
│       ├── plugin.js
│       ├── theme.js
│       ├── vault.js
│       ├── workspace.js
│       ├── base.js
│       ├── bookmark.js
│       ├── history.js
│       ├── sync.js
│       ├── publish.js
│       ├── dev.js
│       └── general.js
├── test/
└── examples/
    ├── vault-health.js
    ├── zk-file.js
    ├── morning-automation.js
    └── inbox-triage.js
```

---

## Comparison: obsidian-ts vs obsidian.js

**obsidian-ts** is a good 1:1 CLI wrapper. It mirrors the CLI faithfully — if you know the CLI,
you know the library. But it offers no abstractions above "call this command, get typed output."

**obsidian.js** uses obsidian-ts's approach as the bottom layer, then adds:

1. **Simplified top-level API** — `obs.read(name)` instead of `file.read({ file: name })`
2. **Note object** — content + metadata + links in one lazy-loaded object
3. **Query builder** — composable search instead of separate function calls per filter
4. **Batch operations** — `for await` iteration with progress tracking
5. **Markdown toolkit** — parse frontmatter, extract wikilinks, build note content
6. **Plain JS** — no TypeScript build step needed, works in any Node 18+ script

The philosophy: **the CLI wrapper is infrastructure; the scripting API is the product.**

---

## Open Design Questions

1. **Caching strategy** — How aggressively to cache? File list is stale quickly if scripts
   create notes. Could use TTL-based cache or explicit `obs.invalidate()`.

2. **Concurrency** — The CLI is sequential, but we could queue commands internally.
   Worth adding an internal command queue with configurable concurrency (default 1)?

3. **Streaming large results** — `obs.files()` returns all paths at once. For vaults with
   50k files, should we offer a streaming/paginated variant?

4. **Plugin system** — Should obsidian.js support user plugins that add methods?
   e.g. `obs.use(zkPlugin)` that adds `obs.zk.file()`, `obs.zk.triage()`.

5. **Watch mode** — The CLI doesn't support file watching, but we could poll.
   Worth adding `obs.watch(folder, callback, { interval: 5000 })`?
