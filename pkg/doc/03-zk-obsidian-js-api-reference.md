---
Title: ZK Obsidian JavaScript API Reference
Slug: zk-obsidian-js-api-reference
Short: Exhaustive reference for the local require("obsidian") JavaScript API.
Topics:
- zk
- obsidian
- javascript
- api
- reference
Commands:
- obsidian
- run-script
Flags:
- binary
- vault
- output
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

This page documents the JavaScript API exposed by `require("obsidian")`. It explains the methods, the expected input shapes, the return values, and the practical differences between the read-only paths and the write-capable paths.

The most important thing to remember is that nearly every method returns a Promise from JavaScript's perspective. The CLI runner waits for that Promise to settle before returning the final result to the user, so the normal pattern is `await obs.method(...)` inside an async IIFE.

## Loading The Module

```javascript
const obs = require("obsidian");
```

Before most scripts do meaningful work, the runner injects a call to `obs.configure(...)` with the configured binary path and optional vault override. You can also call it yourself.

## Top-Level Exports

| Export | Kind | Notes |
|---|---|---|
| `configure(options)` | sync function | Updates module runtime config |
| `version()` | Promise<string> | Read-only |
| `files(options?)` | Promise<string[]> | Read-only |
| `read(ref)` | Promise<string> | Read-only |
| `note(ref)` | Promise<object> | Read-only |
| `query(options?)` | fluent builder | Read-only until `run()` |
| `batch(options, mapper?)` | Promise<any[]> | Read-only if mapper is pure |
| `exec(name, parameters?, flags?)` | Promise<string> | Generic raw CLI adapter |
| `create(...)` | Promise<string> | Mutating |
| `append(...)` | Promise<boolean> | Mutating |
| `prepend(...)` | Promise<boolean> | Mutating |
| `move(...)` | Promise<boolean> | Mutating |
| `rename(...)` | Promise<boolean> | Mutating |
| `delete(...)` | Promise<boolean> | Mutating |
| `md` | object | Markdown helper namespace |

## `configure(options)`

This method updates runtime configuration. Use it when you need to override the default `~/.local/bin/obsidian` path or target a specific vault.

```javascript
obs.configure({
  binaryPath: "/home/manuel/.local/bin/obsidian",
  vault: "obsidian-vault",
  workingDir: "/home/manuel/code/wesen/2026-03-14--zk-tool",
  timeoutMs: 30000,
});
```

Supported option keys:

- `binaryPath`
- `vault`
- `workingDir`
- `timeoutMs`

## `version()`

Returns the version string reported by the local Obsidian wrapper.

```javascript
const version = await obs.version();
```

Observed example:

```json
{"version":"1.12.4 (installer 1.8.9)"}
```

## `files(options?)`

Returns vault file paths, usually filtered by folder and extension.

```javascript
const files = await obs.files({ ext: "md" });
const claimNotes = await obs.files({ folder: "ZK/Claims", ext: "md" });
```

Supported options:

- `folder`
- `ext`
- `vault`
- `limit`

Practical note: the current transport maps `folder` and `ext` to the real CLI. Do not assume `limit` is enforced at the raw CLI level for every command; some limiting is safer to do in JS or the high-level query layer.

## `read(ref)`

Reads note contents by exact path or note-like reference.

```javascript
const text = await obs.read("Animals/Parrots.md");
const maybeByName = await obs.read("Parrots");
```

Reference resolution rules:

- If the string contains `/` or ends with `.md`, it is treated as a path.
- Otherwise the client tries to resolve it against available file basenames.

Failure modes:

- not found
- ambiguous basename match

## `note(ref)`

Returns a hydrated note object with parsed helper fields.

```javascript
const note = await obs.note("Animals/Parrots.md");
```

Observed shape:

```javascript
{
  path: "Animals/Parrots.md",
  title: "Parrots",
  content: "...",
  frontmatter: {},
  headings: ["Parrots", "Key Features"],
  tags: [],
  wikilinks: [],
  tasks: []
}
```

## `query(options?)`

This method creates a fluent query builder. The builder is read-only until you call `run()`.

Example:

```javascript
const rows = await obs
  .query()
  .inFolder("Animals")
  .withExtension("md")
  .limit(5)
  .run();
```

Builder methods:

- `inFolder(folder)`
- `withExtension(ext)`
- `tagged(tag)`
- `contains(term)`
- `limit(n)`
- `inVault(vault)`
- `run()`

The builder works by combining raw CLI expansion with client-side filtering. That means some filters are cheap at the CLI layer and some are applied after notes are read.

Query algorithm pseudocode:

```text
if contains(term) is set:
    run search-oriented expansion
else:
    run files-oriented expansion

hydrate notes
apply tag filter if requested
apply content filter if still needed
apply final limit
return note objects
```

## `batch(options, mapper?)`

Runs a query and optionally maps each note in JavaScript.

```javascript
const rows = await obs.batch(
  { ext: "md", limit: 3 },
  (note) => ({
    path: note.path,
    headingCount: note.headings.length,
  }),
);
```

This is the easiest way to produce compact derived output from note objects without writing a manual `for` loop in every script.

## `exec(name, parameters?, flags?)`

This is the raw escape hatch. Use it when the higher-level API does not yet expose a specific Obsidian CLI command.

JSON example:

```javascript
const raw = await obs.exec("tags", { format: "json", counts: true });
const tags = JSON.parse(raw);
```

Plain text example:

```javascript
const raw = await obs.exec("vault");
console.log(raw);
```

Use `exec()` carefully:

- good for exploration
- good for read-only commands not yet wrapped
- not as stable as the typed API

## Mutating Methods

These methods exist in the module but should not be the first tools an intern reaches for.

- `create(title, options?)`
- `append(ref, content)`
- `prepend(ref, content)`
- `move(ref, destination)`
- `rename(ref, newName)`
- `delete(ref, options?)`

Use them only after you have proven the read-only path and understand the target vault.

## Markdown Helper Namespace: `obs.md`

The `md` namespace works entirely on strings and does not require vault access. That makes it ideal for unit-like exploration.

Available helpers:

- `frontmatter(content)`
- `headings(content)`
- `tags(content)`
- `wikilinks(content)`
- `tasks(content)`

Example:

```javascript
const sample = `
---
title: Demo
---

# Intro

[[Alpha]]
- [ ] Task
#tag
`;

const tags = obs.md.tags(sample);
```

## Source File References

Use these files as the canonical implementation references:

- `/home/manuel/code/wesen/2026-03-14--zk-tool/modules/obsidian/module.go`
- `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidian/client.go`
- `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidian/query.go`
- `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidiancli/spec.go`
- `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidianmd/frontmatter.go`
- `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidianmd/headings.go`
- `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidianmd/tags.go`
- `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidianmd/wikilinks.go`
- `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidianmd/tasks.go`

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| `obs.exec()` returns unexpected text | The target command emits plain text, not JSON | Parse manually or use a higher-level wrapper |
| `note(ref)` fails by basename | More than one file shares the same basename | Use an exact path |
| `query().contains(...)` returns surprising results | Search and file expansion use different paths internally | Inspect `pkg/obsidian/query.go` and compare with a direct script |
| `md.tags()` misses frontmatter tags | It extracts hashtag-style tags, not every possible metadata path | Use `md.frontmatter()` separately |

## See Also

- `zk help zk-obsidian-system-overview`
- `zk help zk-obsidian-intern-guide`
- `zk help zk-obsidian-smoke-tests-playbook`
- `zk help zk-obsidian-implementation-diary`
