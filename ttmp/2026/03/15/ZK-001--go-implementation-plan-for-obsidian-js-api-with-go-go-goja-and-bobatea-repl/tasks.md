# Tasks

## TODO

- [x] Read `DESIGN-obsidian-js-api.md` and identify the user-facing contract to preserve
- [x] Inspect `go-go-goja` for runtime, module, async, and REPL integration surfaces
- [x] Inspect `bobatea` for REPL evaluator and UI capabilities relevant to the requested experience
- [x] Inspect the local ZK workflow to identify the first practical consumer of the new API
- [x] Create a `docmgr` ticket in this repository and store the design deliverables there
- [x] Write a detailed intern-oriented analysis / design / implementation guide
- [x] Write a chronological diary entry with exact commands and failures

## Phase 1: `pkg/obsidiancli`

- [x] Add `pkg/obsidiancli/config.go` with binary/vault/working-dir/timeout configuration
- [x] Add `pkg/obsidiancli/spec.go` with command specs for version, file CRUD, search, links, tags, properties, tasks, daily, templates, plugins, themes, vaults, and eval
- [x] Add `pkg/obsidiancli/args.go` for deterministic CLI argument serialization
- [x] Add `pkg/obsidiancli/parse.go` for JSON, line-list, and key/value parsing
- [x] Add `pkg/obsidiancli/errors.go` with typed execution/parse/not-found/ambiguous/version errors
- [x] Add `pkg/obsidiancli/runner.go` with subprocess execution and sequential locking
- [x] Add transport-focused tests for arg serialization, parsing, and error classification
- [x] Run focused tests for `pkg/obsidiancli`
- [x] Commit the `pkg/obsidiancli` slice in `go-go-goja` (`229c9f7`)

## Phase 2: `pkg/obsidianmd`

- [x] Add `pkg/obsidianmd/frontmatter.go`
- [x] Add `pkg/obsidianmd/wikilinks.go`
- [x] Add `pkg/obsidianmd/headings.go`
- [x] Add `pkg/obsidianmd/tags.go`
- [x] Add `pkg/obsidianmd/tasks.go`
- [x] Add `pkg/obsidianmd/note_builder.go`
- [x] Add markdown helper tests covering ZK-style notes with and without frontmatter
- [x] Run focused tests for `pkg/obsidianmd`
- [x] Commit the `pkg/obsidianmd` slice in `go-go-goja` (`f7961ef`)

## Phase 3: `pkg/obsidian`

- [x] Add `pkg/obsidian/types.go` with note/query/batch/config types
- [x] Add `pkg/obsidian/client.go` with low-level methods backed by `pkg/obsidiancli`
- [x] Add `pkg/obsidian/note.go` with lazy loading and cache invalidation
- [x] Add `pkg/obsidian/query.go` with native-filter vs post-filter planning
- [x] Add `pkg/obsidian/batch.go` with sequential default execution
- [x] Add `pkg/obsidian/cache.go`
- [x] Add client/query/note tests using a fake CLI runner
- [x] Run focused tests for `pkg/obsidian`
- [x] Commit the `pkg/obsidian` slice in `go-go-goja` (`d7a8dc1`)

## Phase 4: `modules/obsidian`

- [x] Add `modules/obsidian/module.go` and register the native module
- [x] Export `configure`, `version`, `files`, `read`, `create`, `append`, `prepend`, `move`, `rename`, `delete`, `note`, `query`, `batch`, `exec`
- [x] Export `md` namespace from `pkg/obsidianmd`
- [x] Use `runtimeowner.Runner` for Promise settlement when an owner is injected; fall back to synchronous settlement otherwise
- [x] Add JS-facing tests for Promise resolution/rejection and fluent query chaining
- [x] Run focused tests for `modules/obsidian` and the JS evaluator
- [x] Commit the `modules/obsidian` slice in `go-go-goja` (`4faf260`)

## Phase 5: Evaluator And REPL Runtime Behavior

- [x] Extend the JS evaluator so Promise-returning expressions can settle into transcript output
- [x] Decide and implement expression-style top-level `await` handling strategy for REPL input
- [x] Add evaluator tests for Promise settlement and async error rendering
- [x] Commit evaluator/runtime updates in `go-go-goja` (`8d7ab9d`)

## Phase 6: Dedicated Obsidian REPL

- [ ] Add `cmd/obsidian-js-repl/main.go` based on `cmd/js-repl/main.go`
- [ ] Wire vault/binary/path flags into evaluator/module configuration
- [ ] Set Obsidian-specific title, placeholder, and help text
- [ ] Add smoke-level tests or scripted validation instructions
- [ ] Commit the dedicated REPL command

## Phase 7: Local ZK Workflow Proof Of Concept

- [ ] Port the vault tree-index reading workflow into Go or JS on top of the new API
- [ ] Recreate the current `zk_create.py` read/classify/write loop as a proof-of-concept script
- [ ] Validate against the local vault workflow described in `PROJ - ZK Tool.md`
- [ ] Commit the workflow proof-of-concept

## Ongoing Documentation And Validation

- [ ] Update the diary after each implementation slice with exact commands, failures, and commit hashes
- [ ] Update ticket changelog after each committed slice
- [ ] Re-run `docmgr doctor --ticket ZK-001 --stale-after 30 --fail-on warning` after each documentation update
- [ ] Upload a refreshed reMarkable bundle when a meaningful implementation milestone is complete
