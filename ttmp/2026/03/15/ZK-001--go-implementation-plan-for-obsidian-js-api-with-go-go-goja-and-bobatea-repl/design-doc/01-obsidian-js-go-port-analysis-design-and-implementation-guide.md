---
Title: obsidian.js Go port analysis design and implementation guide
Ticket: ZK-001
Status: active
Topics:
    - obsidian
    - goja
    - bobatea
    - repl
    - api-design
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../corporate-headquarters/bobatea/docs/repl.md
      Note: Higher-level REPL usage and embedding model
    - Path: ../../../../../../../corporate-headquarters/bobatea/pkg/repl/evaluator.go
      Note: Streaming evaluator contract and event kinds
    - Path: ../../../../../../../corporate-headquarters/bobatea/pkg/repl/model.go
      Note: |-
        REPL shell composition, feature toggles, and evaluator capability wiring
        Evidence for existing REPL shell capabilities
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/cmd/goja-jsdoc/main.go
      Note: Example of Glazed/Cobra command wiring pattern in go-go-goja
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/cmd/js-repl/main.go
      Note: Existing Bobatea-based JS REPL command to clone/extend
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/engine/factory.go
      Note: |-
        Runtime factory composition and lifecycle entrypoint
        Runtime composition and lifecycle design
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/engine/module_specs.go
      Note: Explicit module and runtime initializer extension points
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/modules/common.go
      Note: Native module registration contract
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/pkg/doc/03-async-patterns.md
      Note: Recommended Promise settlement model using runtimeowner.Runner
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/pkg/repl/adapters/bobatea/javascript.go
      Note: Existing bridge from go-go-goja evaluator to Bobatea REPL contracts
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/pkg/repl/evaluators/javascript/evaluator.go
      Note: |-
        Current JS evaluator behavior and current lack of Promise-awaiting in REPL output
        Evidence for current Promise handling gap
    - Path: DESIGN-obsidian-js-api.md
      Note: |-
        Source design to port into Go/goja/Bobatea form
        Source API contract and implementation requirements
    - Path: PROJ - ZK Tool.md
      Note: Current local ZK filing workflow and target user workflows
    - Path: scripts/build_tree_index.py
      Note: Current vault indexing logic that may migrate into JS automation helpers later
    - Path: scripts/zk_create.py
      Note: |-
        Existing Python note-filing pipeline that the new JS API should eventually replace
        Concrete local workflow targeted by the future module
ExternalSources:
    - https://help.obsidian.md/cli
    - https://github.com/kitschpatrol/obsidian-ts
Summary: |
    Detailed implementation guide for porting the local obsidian.js design into a Go-first architecture built on go-go-goja and goja, with a Bobatea REPL. The guide explains the current state, identifies the major gaps, proposes a package/module layout, covers async and REPL implications, and gives a phased file-level plan for an intern to execute.
LastUpdated: 2026-03-15T15:04:15.77280481-04:00
WhatFor: |
    Use this as the primary design and onboarding document for implementing the obsidian.js API in Go while preserving the ergonomics of the source design and fitting naturally into the existing go-go-goja runtime and Bobatea REPL architecture.
WhenToUse: Use this before writing code in go-go-goja or bobatea for Obsidian automation, when onboarding a new intern to the project, or when reviewing whether a proposed implementation still matches the source obsidian.js design.
---


# obsidian.js Go port analysis design and implementation guide

## Executive Summary

The local source design in `DESIGN-obsidian-js-api.md` describes a JavaScript-first scripting toolkit for Obsidian that is deliberately richer than a thin CLI wrapper. It aims to combine concise vault operations, richer note objects, markdown utilities, a fluent query builder, batch helpers, and an escape hatch for raw CLI execution. The design also assumes a pleasant interactive scripting experience, with examples built around `await`, cached note state, and composable higher-level operations.

The two target libraries already provide most of the runtime substrate needed to build this:

- `go-go-goja` already has explicit runtime composition, native module registration, Promise/event-loop guidance, and a Bobatea adapter for a JavaScript evaluator.
- `bobatea` already has a generic streaming REPL shell, autocomplete/help surfaces, history, and timeline rendering.

The missing work is not the TUI. The missing work is the Obsidian integration layer itself: a robust Go package that shells out to the Obsidian CLI, parses its output, enforces sequential access, exposes a high-level scripting API, and then maps that API into goja in a Promise-friendly way.

The recommended architecture is a layered design:

1. `pkg/obsidiancli` in `go-go-goja` for low-level CLI execution, config, parsing, and errors.
2. `pkg/obsidianmd` for markdown/frontmatter/wikilink/task parsing and note-building helpers.
3. `pkg/obsidian` for the higher-level typed client, note objects, query builder, batch runner, and caching.
4. `modules/obsidian` as the native goja bridge that exports an `obsidian` / `obsidian.js` module with Promise-based methods.
5. A dedicated Bobatea command, ideally `cmd/obsidian-js-repl`, that reuses the existing JavaScript evaluator and REPL shell but starts with the Obsidian module preconfigured and documented.

The largest technical design issue is not shelling out. It is preserving the source design's asynchronous ergonomics. The current `go-go-goja` evaluator executes input with `runtime.RunString(code)` and returns the stringified result immediately, which means returned Promises are not automatically awaited in the REPL. That is acceptable for the current generic JS REPL, but it is a gap relative to the desired `await obs.files()` experience and must be addressed explicitly in the implementation plan.

## Problem Statement And Scope

### What the user asked for

The user asked for a ticket-backed implementation guide that turns `DESIGN-obsidian-js-api.md` into a concrete Go design using:

- `~/code/wesen/corporate-headquarters/go-go-goja`
- `goja`
- `~/code/wesen/corporate-headquarters/bobatea`

The requested deliverable is documentation, not code. This ticket therefore defines what should be built, where it should live, why each piece exists, and how to sequence the implementation so an intern can execute it safely.

### What the source design requires

Observed in `DESIGN-obsidian-js-api.md`:

- It is not only a 1:1 CLI wrapper. It adds richer behavior such as `obs.note()`, `obs.query()`, `obs.batch()`, and `md.note(...)`.
- It expects composable return types, lazy loading, and caching (`DESIGN-obsidian-js-api.md:15-20`, `DESIGN-obsidian-js-api.md:98-133`, `DESIGN-obsidian-js-api.md:174-176`, `DESIGN-obsidian-js-api.md:494-524`).
- It assumes an async JS user experience via `await` throughout the API examples.
- It includes workflows that matter directly to this repository, especially the "ZK Filing Pipeline (Replaces zk_create.py)" (`DESIGN-obsidian-js-api.md:396-433`).

### What is in scope

In scope:

- A Go implementation strategy for the low-level Obsidian CLI bridge.
- A Go package structure for the high-level API.
- A goja module design matching the source API as closely as practical.
- A Bobatea REPL strategy for interactive scripting.
- A file-by-file implementation plan.
- A testing and validation plan.

Out of scope for this ticket:

- Writing the production code itself.
- Settling every future feature question in the source design.
- Building a full MCP server or an Obsidian plugin.
- Designing a stable public semver release strategy.

## Current-State Analysis

### 1. The local repository already has a concrete user workflow to improve

This repo is not a purely abstract design sandbox. It already contains a Python tool for filing ZK notes:

- `PROJ - ZK Tool.md:1-4` defines the product as a quick note-filing CLI for an Obsidian vault.
- `PROJ - ZK Tool.md:64-85` describes the note structure and explicitly notes that ZK claims do not use YAML frontmatter.
- `PROJ - ZK Tool.md:108-114` describes the current pipeline as tree indexing, classification, note generation, confirmation, and writing.
- `scripts/build_tree_index.py:51-160` implements the vault tree scan and text/JSON renderers used by the current LLM pipeline.
- `scripts/zk_create.py:32-43`, `scripts/zk_create.py:46-111`, and `scripts/zk_create.py:244-275` show that the current tool builds prompts, asks an LLM to classify content, then writes notes directly to the vault.

This matters because the new `obsidian` module should not be designed in isolation. One of its first real consumers should be a future replacement for `scripts/zk_create.py`. That gives the project a grounded acceptance target:

- Read vault structure.
- Find likely links and branch locations.
- Generate or update note content.
- Write the note through the Obsidian CLI so link/index behavior remains aligned with Obsidian itself.

### 2. The source design is opinionated, not vague

The design doc is detailed enough to act as a functional contract.

Observed major surfaces:

- Architecture sketch for `core/`, thin command wrappers, `note.js`, `query.js`, `batch.js`, and `markdown.js` (`DESIGN-obsidian-js-api.md:24-48`).
- Core file CRUD and file listing (`DESIGN-obsidian-js-api.md:69-97`).
- A rich `Note` object with metadata, parsed content, backlinks, outgoing links, and mutating methods (`DESIGN-obsidian-js-api.md:98-133`).
- A query builder with both native-CLI and post-filter behavior (`DESIGN-obsidian-js-api.md:135-176`).
- Batch operations with explicit sequential safety (`DESIGN-obsidian-js-api.md:178-214`).
- Markdown utilities and a note-builder (`DESIGN-obsidian-js-api.md:216-250`).
- Daily, tags, properties, tasks, links, plugins, themes, templates, eval, and vault helpers (`DESIGN-obsidian-js-api.md:252-354`).
- Implementation notes explicitly stating that Obsidian CLI acts as remote control over a running app and that commands must be sequential (`DESIGN-obsidian-js-api.md:494-512`).

This is enough to define a first implementation without guessing the user-facing surface.

### 3. go-go-goja already provides the right extension seams

Observed in `go-go-goja`:

- `engine/factory.go:15-179` provides an explicit runtime builder and factory with module registration and per-runtime initializers.
- `engine/module_specs.go:14-82` defines `ModuleSpec`, `RuntimeInitializer`, and the explicit `DefaultRegistryModules()` pattern.
- `modules/common.go:29-102` defines the native module contract and global registration model.
- `engine/runtime.go:12-18` shows that blank imports are used only to populate the default registry; actual runtime exposure remains explicit.
- `engine/module_roots.go:11-118` already handles layered module-root resolution for script-based loading.

This means the Obsidian implementation should not bypass the engine. It should fit into the existing explicit composition model. That is the safest way to avoid special-case code and to keep the module usable from both file-based scripts and a REPL.

### 4. Async patterns are already documented and standardized

Observed in `pkg/doc/03-async-patterns.md`:

- `pkg/doc/03-async-patterns.md:18-30` says JS VM interactions must return to the owner goroutine and recommends `runtimeowner.Runner`.
- `pkg/doc/03-async-patterns.md:32-54` gives the preferred Promise-settlement pattern.
- `pkg/doc/03-async-patterns.md:65-191` gives Promise-based API examples.
- `pkg/runtimeowner/runner.go:62-159` shows the concrete `Call` and `Post` scheduling primitives.

This is important because Obsidian CLI calls involve subprocess execution, file IO, and possibly longer-running operations. The implementation should not settle Promise results directly from worker goroutines. It should follow the repository's documented runner pattern.

### 5. The current JS evaluator is close, but not yet aligned with the desired API ergonomics

Observed in `pkg/repl/evaluators/javascript/evaluator.go`:

- `pkg/repl/evaluators/javascript/evaluator.go:87-143` can create an evaluator backed by a fully composed `go-go-goja` runtime.
- `pkg/repl/evaluators/javascript/evaluator.go:201-228` executes code with `runtime.RunString(code)` and immediately stringifies the result.
- `pkg/repl/evaluators/javascript/evaluator.go:231-239` simply wraps that output in a Bobatea REPL event.
- `pkg/repl/evaluators/javascript/evaluator.go:242-420` already supports completions, help-bar context, and help-drawer context.

The evaluator is therefore already a strong host for an Obsidian module, but it has a real limitation:

- If a module method returns a Promise, the current evaluator will print the Promise representation rather than await it.
- The source design examples are overwhelmingly `await`-based.

This does not block the architecture, but it does require an implementation phase specifically for "top-level async ergonomics".

### 6. The Bobatea REPL is already capable of hosting this experience

Observed in `bobatea`:

- `bobatea/pkg/repl/evaluator.go:5-38` defines a streaming evaluator interface based on semantic events rather than a single text blob.
- `bobatea/pkg/repl/model.go:21-179` shows the REPL shell already supports history, timeline transcript, autocomplete providers, help bar, help drawer, and command palette wiring.
- `bobatea/docs/repl.md:20-36` and `bobatea/docs/repl.md:173-194` document the generic evaluator model.
- `go-go-goja/cmd/js-repl/main.go:48-91` already composes the existing JS evaluator with Bobatea, event bus wiring, and an alternate-screen TUI.
- `go-go-goja/pkg/repl/adapters/bobatea/javascript.go:10-76` shows the adapter layer is intentionally thin.

This means a dedicated Obsidian REPL should be built by reusing the existing JS REPL stack, not by inventing a second REPL framework.

## Gap Analysis

The implementation gap is easiest to understand by comparing the source design to what already exists.

### Gap 1: No Obsidian CLI bridge package exists

Missing today:

- command argument serialization,
- Obsidian binary resolution,
- vault selection config,
- output parsers,
- error typing for CLI failures,
- sequential execution guarantees per vault,
- compatibility/version checks.

This is the foundational package that everything else depends on.

### Gap 2: No high-level `obsidian` domain package exists

Missing today:

- a typed `Client`,
- a stateful `Note`,
- a query builder,
- batch helpers,
- caching/invalidation,
- markdown parsing and note builders tailored to Obsidian workflows.

Without this package, a goja module would either become a giant, unmaintainable shell-out layer or would expose raw CLI primitives that do not match the source design.

### Gap 3: No Obsidian goja module exists

Missing today:

- `require("obsidian")` or `require("obsidian.js")`,
- Promise-returning module exports,
- JS objects that wrap Go note/query state,
- error mapping from Go to JS,
- JS-facing config and per-call option handling.

### Gap 4: No dedicated REPL flow exists for Obsidian scripting

The current `cmd/js-repl` is generic. It has no:

- vault flags,
- Obsidian-specific help text,
- common snippets or startup bindings,
- direct support for top-level `await`-style ergonomics,
- explicit module preconfiguration.

### Gap 5: The current local ZK toolchain is disconnected from go-go-goja

Today the ZK workflow is implemented in Python and prompt files. There is no shared Go package that:

- reads vault structure for classification,
- creates notes via the Obsidian CLI,
- exposes that capability to JS scripts in the REPL,
- provides a migration path off `scripts/zk_create.py`.

## Proposed Architecture

### Design goals

The port should preserve these source-level properties:

- concise scripts for common operations,
- composable return values,
- built-in markdown helpers,
- lazy note loading and cache reuse,
- raw escape hatch to the underlying CLI,
- safe sequential execution.

The port should also fit these host-repo properties:

- explicit runtime composition via `engine.NewBuilder()`,
- native module registration via `modules.NativeModule`,
- Promise settlement using `runtimeowner.Runner`,
- REPL integration through the existing Bobatea adapter.

### High-level package layout

Recommended target layout inside `~/code/wesen/corporate-headquarters/go-go-goja`:

```text
go-go-goja/
  pkg/
    obsidiancli/
      config.go
      runner.go
      args.go
      parse.go
      errors.go
      spec.go
      transport.go
    obsidianmd/
      frontmatter.go
      wikilinks.go
      headings.go
      tags.go
      tasks.go
      note_builder.go
    obsidian/
      client.go
      note.go
      query.go
      batch.go
      types.go
      cache.go
      options.go
  modules/
    obsidian/
      module.go
      exports_client.go
      exports_note.go
      exports_query.go
      exports_markdown.go
      errors.go
  pkg/
    repl/
      adapters/
        bobatea/
          obsidian.go
  cmd/
    obsidian-js-repl/
      main.go
```

Optional future additions:

- `cmd/obsidian-run/` for file-based script execution with explicit vault flags.
- `pkg/obsidianzk/` for higher-level ZK helpers specific to this repository.

### Layer responsibilities

#### Layer 1: `pkg/obsidiancli`

This layer is intentionally boring. It should know how to:

- find the Obsidian CLI binary,
- build subprocess arguments,
- inject default or per-call vault configuration,
- serialize option maps to CLI flags/`key=value`,
- run the command,
- normalize stdout/stderr,
- parse JSON, line-list, and key/value outputs,
- return typed Go errors.

It should not know:

- how a `Note` caches state,
- how markdown parsing works,
- how ZK filing works,
- how goja or Bobatea work.

#### Layer 2: `pkg/obsidianmd`

This layer handles text processing and should remain usable from both Go and JS-facing layers:

- parse frontmatter if present,
- split body from frontmatter,
- extract wikilinks,
- extract tags,
- extract headings,
- extract markdown checkbox tasks,
- build common note templates.

Important local nuance:

- `PROJ - ZK Tool.md:64-85` explicitly says the local ZK claims do not use YAML frontmatter.
- Therefore the markdown package should treat frontmatter as optional. It should not assume Obsidian notes always use properties.

#### Layer 3: `pkg/obsidian`

This is the real API layer. It should present the source design in Go terms:

- `Client` for configuration and command dispatch.
- `Note` as a lazy, cached wrapper for one note.
- `Query` as a mutable builder that can translate partly to native CLI calls and partly to post-filters.
- `Batch` as an orchestrator that enforces sequential execution by default.

This layer is also where caching and invalidation belong.

#### Layer 4: `modules/obsidian`

This layer bridges `pkg/obsidian` into goja:

- export a module named both `obsidian` and `obsidian.js`,
- return Promise-like results for IO-bound methods,
- expose a nested `md` namespace,
- create JS note/query objects backed by Go state,
- translate Go errors into rejected JS Promises.

#### Layer 5: REPL command

The REPL command should:

- reuse the existing Bobatea JS REPL architecture,
- build an evaluator with the Obsidian module enabled,
- set a title and placeholder focused on vault automation,
- accept flags such as `--vault`, `--bin`, and optionally `--cwd`,
- improve Promise/top-level-await handling so the interaction matches the source design more closely.

### Architecture diagram

```text
             +-----------------------------------------------+
             |          Bobatea REPL / Script Entry          |
             |  cmd/obsidian-js-repl or future script cmd    |
             +---------------------------+-------------------+
                                         |
                                         v
             +-----------------------------------------------+
             |         go-go-goja JS Evaluator Layer         |
             |  pkg/repl/evaluators/javascript + adapter     |
             +---------------------------+-------------------+
                                         |
                              require("obsidian")
                                         |
                                         v
             +-----------------------------------------------+
             |          modules/obsidian (goja bridge)       |
             |  Promise exports, JS objects, error mapping   |
             +---------------------------+-------------------+
                                         |
                                         v
             +-----------------------------------------------+
             |             pkg/obsidian API layer            |
             | Client, Note, Query, Batch, Cache            |
             +---------------------------+-------------------+
                                         |
               +-------------------------+-------------------------+
               |                                                   |
               v                                                   v
    +------------------------------+                 +------------------------------+
    |       pkg/obsidiancli        |                 |        pkg/obsidianmd        |
    | process execution + parsing  |                 | markdown parsing/building    |
    +------------------------------+                 +------------------------------+
               |
               v
    +------------------------------+
    |      Official Obsidian CLI   |
    |      running against vault   |
    +------------------------------+
```

## Core Design Decisions

### Decision 1: Keep the public JS API async even if the CLI call is internally synchronous

Recommendation:

- Use Promise-returning exports in the goja module for IO-bound methods.
- Let `pkg/obsidiancli` do the blocking work in Go goroutines.
- Settle Promises back on the owner thread via `runtimeowner.Runner`.

Why:

- The source design is Promise/`await` oriented.
- It keeps future room for long-running or concurrent orchestration.
- It matches the repo's documented async guidance in `pkg/doc/03-async-patterns.md:18-30`.

Tradeoff:

- It requires extra REPL support for nice display of results.
- It increases implementation complexity compared with a purely synchronous module.

Why this is still the right choice:

- Choosing sync-only now would bake in the wrong ergonomics at the exact boundary the user cares about.
- The async surface is already an intentional part of `go-go-goja`'s documented model.

### Decision 2: Separate low-level CLI transport from the user-facing API

Recommendation:

- Build `pkg/obsidiancli` first.
- Build `pkg/obsidian` on top of it.

Why:

- The same low-level transport will be needed by Go callers, the JS module, tests, and possible future commands.
- It keeps CLI parsing bugs from being spread across note/query/batch logic.
- It lets you write transport-focused tests with fixture outputs.

### Decision 3: Treat the source design as the JS contract, but translate ESM examples into CommonJS for goja

Observed constraint:

- The source design uses ESM-style examples like `import obs from 'obsidian.js'`.
- The existing `go-go-goja` runtime is CommonJS/`require()`-oriented (`cmd/repl/main.go:28-35`, `pkg/doc/02-creating-modules.md:18-24`).

Recommendation:

- Expose both `require("obsidian")` and `require("obsidian.js")`.
- Document REPL examples in CommonJS form:

```javascript
const obs = require("obsidian")
const { md } = obs
```

- Keep method names and object shapes as close as possible to the source design.

This is an implementation translation, not a product change.

### Decision 4: Enforce sequential command execution at the client layer

Observed in the source design:

- `DESIGN-obsidian-js-api.md:501-503` explicitly says commands are sequential and should not run in parallel against the same vault.

Recommendation:

- Put a per-client mutex or queue inside `pkg/obsidian.Client`.
- Make `Batch` default to concurrency `1`.
- Require an explicit opt-in flag for any future concurrent mode, and document it as experimental.

This is a correctness constraint, not a performance problem.

### Decision 5: Implement note/query objects in Go, not as a pile of JS helper files

Recommendation:

- Keep business logic in Go packages.
- Use the goja module to expose objects whose methods call back into those Go packages.

Why:

- Easier testing from Go.
- Easier reuse from non-JS tools.
- Better fit with the user's explicit request to implement in Go.
- Avoids duplicating logic between Go and embedded JS wrappers.

### Decision 6: Add a dedicated Obsidian REPL command rather than overload the generic one

Recommendation:

- Create `cmd/obsidian-js-repl/main.go`.
- Reuse the existing generic JS REPL internals.
- Keep `cmd/js-repl` generic.

Why:

- This avoids destabilizing existing users of the generic JS REPL.
- It gives room for Obsidian-specific defaults, flags, snippets, and startup docs.

## Detailed API Design

### Go API sketch

The exact names can move, but the first implementation should be close to this:

```go
package obsidian

type Config struct {
    Vault      string
    BinaryPath string
    WorkingDir string
    Timeout    time.Duration
}

type Client struct {
    cfg    Config
    runner *obsidiancli.Runner
    cache  *Cache
    mu     sync.Mutex
}

func New(cfg Config, opts ...Option) (*Client, error)
func (c *Client) Version(ctx context.Context) (string, error)
func (c *Client) Configure(cfg Config)
func (c *Client) Files(ctx context.Context, opts FilesOptions) ([]string, error)
func (c *Client) Read(ctx context.Context, ref NoteRef) (string, error)
func (c *Client) Create(ctx context.Context, title string, opts CreateOptions) (CreateResult, error)
func (c *Client) Append(ctx context.Context, ref NoteRef, text string) error
func (c *Client) Prepend(ctx context.Context, ref NoteRef, text string) error
func (c *Client) Move(ctx context.Context, ref NoteRef, dst string, opts MoveOptions) error
func (c *Client) Delete(ctx context.Context, ref NoteRef, opts DeleteOptions) error
func (c *Client) Note(ctx context.Context, ref NoteRef) (*Note, error)
func (c *Client) Query() *Query
func (c *Client) Batch(ctx context.Context, q *Query, fn BatchFunc, opts BatchOptions) ([]any, error)
func (c *Client) Exec(ctx context.Context, command string, args map[string]any, flags []string) (any, error)
```

Supporting types:

```go
type NoteRef struct {
    Name string // wikilink-style resolution
    Path string // exact path
}

type Note struct {
    client *Client
    ref    NoteRef

    loadedContent bool
    content       string

    loadedMeta bool
    meta       NoteMetadata
}

type Query struct {
    client     *Client
    filters    QueryFilters
    searchText string
    withCtx    bool
}
```

This is intentionally conventional Go, even though the public JS face will look more fluent.

### JS module sketch

The goja module should export one main object:

```javascript
const obs = require("obsidian")

await obs.configure({ vault: "My Vault" })
const files = await obs.files({ folder: "ZK/Claims" })
const note = await obs.note("ZK - 2a0 - Systems thinking")
const q = obs.query().withTag("software").modifiedAfter("2026-01-01")
const rows = await q.run()
```

Recommended export shape:

```javascript
module.exports = {
  configure,
  version,
  files,
  read,
  create,
  append,
  prepend,
  move,
  rename,
  delete,
  note,
  query,
  batch,
  exec,
  daily,
  tags,
  properties,
  tasks,
  links,
  plugins,
  themes,
  templates,
  vaults,
  md,
  errors,
}
```

This keeps the source-design organization without forcing exact JS source-file parity.

### Query builder design

The query builder is one of the most important places to be explicit about translation boundaries.

Recommendation:

- Treat the builder as a mutable struct in Go with fluent JS wrappers.
- Separate filters into:
  - native CLI filters,
  - post-filters,
  - output transforms such as sort/limit/context.

Pseudo-structure:

```go
type QueryPlan struct {
    Native NativeCriteria
    Post   PostCriteria
    Output OutputCriteria
}
```

Execution algorithm:

```text
1. Start from the most selective native command possible.
2. Ask the CLI for the smallest useful result set.
3. Convert raw rows into typed intermediate rows.
4. Apply post-filters in Go.
5. Apply sort/limit.
6. Materialize Notes only if the caller requested note objects.
```

This matters because the source design explicitly allows post-filtering when a filter cannot be expressed as a CLI argument (`DESIGN-obsidian-js-api.md:174-176`).

### Note object design

The `Note` object should be lazy and invalidation-aware.

Recommended behavior:

- `obs.note(ref)` returns a handle immediately after resolving the note reference.
- `note.content`, `note.frontmatter`, `note.body`, `note.headings`, `note.wikilinks`, `note.tags`, and `note.tasks` should compute lazily on first access.
- mutating operations (`append`, `moveTo`, `rename`, `setProperty`, and so on) should invalidate cached derived fields.

Pseudo-flow:

```text
note.content access
  -> if cached, return cached content
  -> else client.read(note.ref)
  -> cache raw content
  -> derive markdown helpers lazily or memoize after first parse

note.append("...")
  -> client.Append(...)
  -> invalidate content/body/headings/wikilinks/tags/tasks cache
  -> keep stable identity of note object
```

### Markdown helper design

Do not mix this logic into the CLI runner.

Recommended helpers:

- `Parse(raw string) (frontmatter map[string]any, body string, err error)`
- `Wikilinks(raw string) []string`
- `Tags(raw string) []string`
- `Headings(raw string) []Heading`
- `Tasks(raw string) []Task`
- `BuildNote(input NoteTemplate) string`

Local-repo note:

- Because the ZK system in `PROJ - ZK Tool.md` is largely frontmatter-free, the note builder should support templates where the second line is wiki-tag text rather than YAML properties.

### Error model

Use typed Go errors and attach machine-readable codes.

Recommended Go errors:

- `ErrBinaryNotFound`
- `ErrVaultRequired`
- `ErrCommandFailed`
- `ErrParseOutput`
- `ErrUnsupportedVersion`
- `ErrAmbiguousReference`
- `ErrNotFound`

JS mapping recommendation:

- Rejected Promise should carry:
  - `name`
  - `message`
  - `code`
  - `command` if relevant
  - `stdout` and `stderr` when useful for debugging

If custom JS error subclasses are too much for phase 1, it is acceptable to use standard `Error` objects with extra fields.

## Async And REPL Design

### Why this section matters

The source design is written like Node scripting. The current evaluator behaves like a synchronous expression evaluator that happens to support Promises in the runtime. Those are not the same user experience.

### Current gap

Observed:

- `pkg/repl/evaluators/javascript/evaluator.go:211-225` evaluates code directly and stringifies the immediate result.
- If the result is a Promise, the user sees the Promise object, not the settled value.

### Recommended phase-1 REPL behavior

Add explicit support for "awaiting the returned Promise".

Recommended rules:

1. Evaluate the input.
2. If the result is not a Promise, stream it exactly as today.
3. If the result is a Promise, emit an initial progress event such as "Promise pending".
4. Attach resolution/rejection handlers on the owner thread.
5. Emit the final resolved or rejected value into the transcript.

Optional phase-2 enhancement:

- detect bare top-level `await` in user input and rewrite it to an async IIFE before evaluation.

Example transform:

```javascript
await obs.files({ folder: "Inbox" })
```

could become internally:

```javascript
(async () => await obs.files({ folder: "Inbox" }))()
```

This is important because it lets the REPL more closely match the source design without requiring the user to type:

```javascript
obs.files({ folder: "Inbox" }).then(console.log)
```

### REPL command structure

Recommended new command: `cmd/obsidian-js-repl/main.go`

Responsibilities:

- parse flags:
  - `--vault`
  - `--bin`
  - `--cwd`
  - `--log-level`
  - maybe `--no-alt-screen`
- build a configured evaluator,
- enable autocomplete/help via existing JS evaluator support,
- display Obsidian-specific title/help text,
- optionally seed startup globals such as:
  - `const obs = require("obsidian")`

Pseudo-main:

```go
func main() {
    cfg := parseFlags()

    evaluator, err := bobateaobsidian.NewEvaluator(cfg)
    if err != nil { log.Fatal(err) }
    defer evaluator.Close()

    replCfg := repl.DefaultConfig()
    replCfg.Title = "Obsidian JS REPL"
    replCfg.Placeholder = "await obs.files(), obs.note(...), const { md } = require(\"obsidian\")"
    replCfg.Autocomplete.Enabled = true
    replCfg.HelpBar.Enabled = true
    replCfg.HelpDrawer.Enabled = true

    bus := eventbus.NewInMemoryBus()
    repl.RegisterReplToTimelineTransformer(bus)

    model := repl.NewModel(evaluator, replCfg, bus.Publisher)
    runBubbleTea(model, bus)
}
```

## File-Level Implementation Plan

This section is the intern's build order. Follow the phases in sequence. Do not start with the REPL.

### Phase 0: Confirm the external contract

Create a short fixture pack before writing code.

Tasks:

1. Record the local source API from `DESIGN-obsidian-js-api.md`.
2. Record the subset of official Obsidian CLI commands needed for phase 1.
3. Decide the first supported set:
   - configure/version
   - files/read/create/append/prepend/move/rename/delete
   - note
   - query basics
   - markdown helpers
   - exec escape hatch

Deliverables:

- test fixtures for representative CLI outputs,
- a small command matrix document or test table.

### Phase 1: Build `pkg/obsidiancli`

Suggested files:

- `pkg/obsidiancli/config.go`
- `pkg/obsidiancli/runner.go`
- `pkg/obsidiancli/args.go`
- `pkg/obsidiancli/parse.go`
- `pkg/obsidiancli/errors.go`
- `pkg/obsidiancli/spec.go`

What to implement:

- config defaults,
- command specs,
- output parser helpers,
- subprocess invocation,
- typed errors,
- per-client sequential lock.

Pseudocode:

```go
func (r *Runner) Run(ctx context.Context, spec CommandSpec, opts CallOptions) (*Result, error) {
    r.mu.Lock()
    defer r.mu.Unlock()

    argv := buildArgs(r.cfg, spec, opts)
    cmd := exec.CommandContext(ctx, r.binaryPath(), argv...)
    stdout, stderr, err := run(cmd)
    if err != nil {
        return nil, wrapCommandError(spec, argv, stdout, stderr, err)
    }
    parsed, err := parseOutput(spec.OutputKind, stdout)
    if err != nil {
        return nil, wrapParseError(spec, stdout, err)
    }
    return &Result{Stdout: stdout, Stderr: stderr, Parsed: parsed}, nil
}
```

Testing:

- unit tests for arg serialization,
- unit tests for output parsing,
- unit tests for error classification,
- unit tests for sequential access semantics.

### Phase 2: Build `pkg/obsidianmd`

Suggested files:

- `pkg/obsidianmd/frontmatter.go`
- `pkg/obsidianmd/wikilinks.go`
- `pkg/obsidianmd/headings.go`
- `pkg/obsidianmd/tags.go`
- `pkg/obsidianmd/tasks.go`
- `pkg/obsidianmd/note_builder.go`

What to implement:

- standalone parsers,
- note-content builder matching the design doc,
- fixtures using examples from the source design and the local ZK format.

Testing focus:

- frontmatter present vs absent,
- wikilinks with aliases,
- headings across markdown levels,
- checkbox tasks,
- ZK note-builder format from `PROJ - ZK Tool.md:64-85`.

### Phase 3: Build `pkg/obsidian`

Suggested files:

- `pkg/obsidian/client.go`
- `pkg/obsidian/types.go`
- `pkg/obsidian/note.go`
- `pkg/obsidian/query.go`
- `pkg/obsidian/batch.go`
- `pkg/obsidian/cache.go`
- `pkg/obsidian/options.go`

What to implement first:

1. `Client`
2. `Files`, `Read`, `Create`, `Append`, `Prepend`, `Move`, `Rename`, `Delete`
3. `Note` lazy loading
4. basic `Query`
5. `Batch`

Do not start with every category of API. Get the spine working first.

Important behavioral rules:

- string note references mean wikilink-style resolution,
- `{ path: ... }` means exact path,
- note mutations invalidate cache,
- batch defaults to sequential mode.

### Phase 4: Build `modules/obsidian`

Suggested files:

- `modules/obsidian/module.go`
- `modules/obsidian/exports_client.go`
- `modules/obsidian/exports_note.go`
- `modules/obsidian/exports_query.go`
- `modules/obsidian/exports_markdown.go`
- `modules/obsidian/errors.go`

Implementation advice:

- keep one Go `Client` per runtime/module instance,
- use `runtimeowner.Runner` for Promise settlement,
- expose small JS objects built with `vm.NewObject()` and bound closures,
- register aliases for `obsidian` and `obsidian.js`.

Pseudo-loader skeleton:

```go
func (m *Module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)
    client := obsidian.New(...)
    owner := runtimeowner.NewRunner(vm, eventloop.Get(vm), ...)

    exports.Set("configure", m.promiseFunc(owner, func(ctx context.Context, call goja.FunctionCall) (any, error) {
        cfg := decodeConfig(call.Arguments)
        client.Configure(cfg)
        return goja.Undefined(), nil
    }))

    exports.Set("files", m.promiseFunc(owner, func(ctx context.Context, call goja.FunctionCall) (any, error) {
        opts := decodeFilesOptions(call.Arguments)
        return client.Files(ctx, opts)
    }))

    exports.Set("query", func() *goja.Object {
        return newQueryObject(vm, owner, client)
    })

    exports.Set("md", newMarkdownNamespace(vm))
}
```

### Phase 5: Add dedicated evaluator/adapter support where needed

Possible files:

- `pkg/repl/adapters/bobatea/obsidian.go`
- `pkg/repl/evaluators/javascript/evaluator.go`

What may need to change:

- ability to inject runtime initializers or module config cleanly,
- optional Promise-await behavior,
- optional top-level-await wrapping,
- richer help text describing the Obsidian module.

Be conservative here. Reuse the generic evaluator as much as possible.

### Phase 6: Add `cmd/obsidian-js-repl`

Use `cmd/js-repl/main.go:48-91` as the structural template.

Do:

- customize flags,
- customize title/help text,
- wire in the obsidian-configured evaluator,
- add tests or at least smoke-run instructions.

Do not:

- fork Bobatea internals,
- copy large parts of the generic evaluator unless unavoidable.

### Phase 7: Migrate the local ZK workflow

This phase is optional for the first PR, but it is the first concrete consumer worth targeting.

Migration ideas:

- replace direct Python vault writes with `obsidian` API calls,
- port `build_tree_index.py` logic into a Go or JS helper later,
- implement a JS script in the new REPL/script runner that performs the current classify-generate-write flow.

The immediate goal is not to fully port the LLM prompt system. The immediate goal is to prove that the new API can support the workflow described in `PROJ - ZK Tool.md:87-123`.

## Pseudocode For Key Flows

### Flow 1: `obs.files({ folder: "ZK/Claims" })`

```text
JS caller
  -> modules/obsidian.files()
  -> Promise created on VM thread
  -> goroutine runs client.Files()
  -> client serializes execution through runner mutex
  -> obsidiancli.Run(commandSpecFilesList, options)
  -> parse stdout into []string
  -> runtimeowner.Post(...) settles Promise with []string
  -> REPL prints resolved result or script continues
```

### Flow 2: `const note = await obs.note("ZK - 2a0 - Systems thinking")`

```text
1. JS module resolves note reference via client
2. client constructs Go Note handle with unresolved lazy fields
3. module wraps Go Note in JS object with property/method closures
4. first access to note.content triggers client.Read(...)
5. note cache is populated
6. note.append(...) invalidates cached derived fields
```

### Flow 3: `obs.query().withTag("software").modifiedAfter("2026-01-01").run()`

```text
1. query() returns JS object backed by Go Query state
2. each fluent method mutates the Go query state and returns the same JS object
3. run() compiles QueryPlan
4. plan chooses native CLI search/list path
5. results are post-filtered in Go if required
6. final rows are encoded back to JS-friendly arrays/objects
```

### Flow 4: REPL Promise resolution

```text
1. evaluator runs code
2. if result is ordinary value:
   -> emit immediate result event
3. if result is Promise:
   -> emit pending/progress event
   -> attach resolve/reject handlers
   -> on settle, emit final markdown/stdout/stderr event
```

## Testing And Validation Strategy

### Unit tests

For `pkg/obsidiancli`:

- args serialization,
- parser correctness,
- error classification,
- binary lookup,
- sequential locking.

For `pkg/obsidianmd`:

- frontmatter parsing,
- wikilink extraction,
- heading extraction,
- task extraction,
- note builder formatting.

For `pkg/obsidian`:

- note cache invalidation,
- exact-path vs name resolution behavior,
- query plan compilation,
- batch sequential behavior,
- escape hatch path.

For `modules/obsidian`:

- Promise resolution and rejection,
- JS object shape,
- query chaining,
- note methods mutating underlying state,
- markdown namespace exposure.

For REPL/evaluator:

- Promise-returning command settles to final rendered value,
- top-level-await transform if implemented,
- autocomplete/help still works with the new module available.

### Integration tests

Recommended:

- create a fake Obsidian CLI binary that prints deterministic outputs,
- point `pkg/obsidiancli` at that fake binary,
- run end-to-end tests through `pkg/obsidian` and `modules/obsidian`.

This is safer than depending on a real desktop Obsidian instance in CI.

### Manual smoke tests

At minimum:

```bash
go test ./...
go run ./cmd/obsidian-js-repl --vault "My Vault"
```

REPL manual script checklist:

```javascript
const obs = require("obsidian")
const { md } = obs

await obs.version()
await obs.files({ folder: "Inbox" })
const note = await obs.note("Some Existing Note")
await note.append("\n## Test\n")
md.wikilinks("See [[Foo]] and [[Bar]]")
```

### Documentation tests

The intern should also treat code snippets in the final README/help docs as test cases. This project already benefits from docs that are close to executable examples.

## Risks, Alternatives, And Open Questions

### Risk 1: Top-level `await` is underspecified in the current evaluator

This is the biggest UX risk.

Mitigation:

- explicitly implement Promise result awaiting in the evaluator,
- optionally add an input transform for bare `await`.

### Risk 2: The official Obsidian CLI output formats may not be uniform

The source design assumes JSON, key-value, and line-list outputs. If actual commands vary more than expected, parsing becomes a maintenance surface.

Mitigation:

- centralize command specs and parsers in `pkg/obsidiancli`,
- capture fixtures from the real CLI early,
- prefer command-specific parsers over one overly generic parser.

### Risk 3: Note reference resolution may be ambiguous

The design allows `obs.read("name")` as a convenience, but vaults can have duplicate names.

Mitigation:

- return `ErrAmbiguousReference`,
- provide an exact-path option,
- document that scripts should switch to exact paths when collisions appear.

### Risk 4: REPL result rendering may be noisy for large objects

Large arrays of files or note objects can flood the transcript.

Mitigation:

- add pretty-print helpers,
- consider truncated preview plus explicit `JSON.stringify(...)` guidance,
- keep raw escape hatch available.

### Alternative A: Keep everything synchronous

Rejected because:

- it diverges from the source design,
- it makes later async evolution more disruptive,
- it does not use the runtime patterns already documented in `go-go-goja`.

### Alternative B: Implement only a raw CLI wrapper module first

Rejected as the primary design because:

- it recreates the `obsidian-ts` problem the design is trying to move beyond,
- it does not help replace the local Python workflow cleanly,
- it delays the hard but necessary API decisions.

### Alternative C: Put the high-level API in JS files and keep Go only for shelling out

Rejected for phase 1 because:

- the user explicitly asked for a Go implementation,
- Go testing and reuse would be worse,
- the repository already has good Go-level runtime composition patterns worth using.

### Open questions to defer until after phase 1

These are valid future questions, but they should not block the first implementation:

- Should the package support a plugin system such as `obs.use(...)`?
- Should cache invalidation be TTL-based, manual, or operation-scoped?
- Should the REPL auto-preload `const obs = require("obsidian")`?
- Should ZK-specific helpers live in the core package or in a separate extension package?
- Should a script-runner command be built alongside the REPL in the same PR?

## Recommended First PR Shape

Do not try to land everything at once.

Recommended first PR:

1. `pkg/obsidiancli`
2. `pkg/obsidianmd`
3. `pkg/obsidian` with a minimal client and note support
4. `modules/obsidian` exposing:
   - `configure`
   - `version`
   - `files`
   - `read`
   - `create`
   - `append`
   - `note`
   - `md`
   - `exec`
5. tests for the above

Recommended second PR:

1. query builder
2. batch operations
3. REPL Promise-await improvements
4. `cmd/obsidian-js-repl`

Recommended third PR:

1. daily/tags/properties/tasks/links/plugins/templates/vaults
2. local ZK workflow migration proof-of-concept

## References

### Primary source design

- `DESIGN-obsidian-js-api.md:1-20` for goals and principles.
- `DESIGN-obsidian-js-api.md:24-48` for the intended module architecture.
- `DESIGN-obsidian-js-api.md:52-354` for the public API surface.
- `DESIGN-obsidian-js-api.md:396-433` for the ZK filing workflow that directly matters to this repository.
- `DESIGN-obsidian-js-api.md:494-524` for implementation constraints and the intended value-add over `obsidian-ts`.

### Current local workflow

- `PROJ - ZK Tool.md:1-4` for product intent.
- `PROJ - ZK Tool.md:64-85` for note structure.
- `PROJ - ZK Tool.md:108-123` for the pipeline and file map.
- `scripts/build_tree_index.py:51-160` for vault indexing.
- `scripts/zk_create.py:32-43`, `scripts/zk_create.py:46-111`, and `scripts/zk_create.py:244-275` for classification and write-path behavior.

### go-go-goja runtime and module system

- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja/engine/factory.go:15-179`
- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja/engine/module_specs.go:14-82`
- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja/modules/common.go:29-102`
- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja/engine/module_roots.go:11-118`
- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/doc/02-creating-modules.md:18-24`
- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/doc/03-async-patterns.md:18-54`

### REPL and evaluator surfaces

- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/repl/evaluators/javascript/evaluator.go:201-239`
- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/repl/adapters/bobatea/javascript.go:10-76`
- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja/cmd/js-repl/main.go:48-91`
- `/home/manuel/code/wesen/corporate-headquarters/bobatea/pkg/repl/evaluator.go:5-38`
- `/home/manuel/code/wesen/corporate-headquarters/bobatea/pkg/repl/model.go:21-179`
- `/home/manuel/code/wesen/corporate-headquarters/bobatea/docs/repl.md:20-36`

### External references

- Official Obsidian CLI docs: `https://help.obsidian.md/cli`
- `obsidian-ts` repository: `https://github.com/kitschpatrol/obsidian-ts`
