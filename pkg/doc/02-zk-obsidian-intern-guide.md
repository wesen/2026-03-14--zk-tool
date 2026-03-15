---
Title: ZK Obsidian Intern Guide
Slug: zk-obsidian-intern-guide
Short: Detailed onboarding guide for interns working on the ZK Obsidian integration.
Topics:
- zk
- obsidian
- onboarding
- tutorial
- intern
Commands:
- obsidian
- run-script
- help
Flags:
- binary
- vault
- output
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

This guide shows a new intern how to orient themselves in the codebase, run the current tooling safely, and add new read-only scripts without breaking the integration. It is written for someone who may understand Go and JavaScript separately but does not yet understand how this repository joins them.

The key practical idea is simple: start with read-only scripts, observe the real vault behavior, and only then touch higher-level APIs or write operations. That workflow keeps the feedback loop tight and reduces the chance of accidental vault mutations while you are still learning the system.

## What You Are Looking At

The project is a Go CLI that uses `go-go-goja` to run JavaScript files. Those JavaScript files can call a local native module named `obsidian`, which is implemented in Go and delegates to the Obsidian CLI wrapper.

When you are getting started, think in these layers:

1. User types a CLI command.
2. The command runs a JS file.
3. The JS file calls the `obsidian` module.
4. The module calls Go client code.
5. The client calls the local `obsidian` wrapper.
6. The wrapper talks to the running Obsidian app and active vault.

## First-Day Checklist

This section covers the first things you should verify on a new machine and why each one matters.

- Confirm the wrapper exists:
  - `~/.local/bin/obsidian help`
- Confirm the CLI command works:
  - `go run ./cmd/zk obsidian run-script scripts/js-tests/obsidian-version.js`
- Confirm JSON output works:
  - `go run ./cmd/zk obsidian run-script scripts/js-tests/obsidian-version.js --output json`
- Confirm help pages load:
  - `go run ./cmd/zk help`
- Confirm tests pass:
  - `go test ./...`

If any of those fail, stop and fix the environment before writing code. Otherwise you risk chasing symptoms from a broken setup instead of real code behavior.

## Safe Working Loop

This section explains the recommended day-to-day workflow and why it is safe. The important constraint is that you should prove behavior with read-only scripts before expanding the API.

```text
pick one narrow question
    ->
write or edit a read-only JS script
    ->
run it through `zk obsidian run-script`
    ->
inspect the JSON result
    ->
if behavior is wrong, debug in the smallest layer possible
    ->
only after the script is reliable, adjust higher-level Go code
```

Use this loop for most tasks:

```bash
go run ./cmd/zk obsidian run-script scripts/js-tests/obsidian-query-sample.js --output json
go test ./...
go run ./cmd/zk help zk-obsidian-js-api-reference
```

## Important Files To Read In Order

This section gives a reading order. It matters because reading files in dependency order is faster than bouncing randomly around the repo.

1. `/home/manuel/code/wesen/2026-03-14--zk-tool/cmd/zk/main.go`
   - entrypoint
   - help system setup
2. `/home/manuel/code/wesen/2026-03-14--zk-tool/cmd/zk/cmds/obsidian/run_script.go`
   - user-facing command definition
   - Glazed row output
3. `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidianjs/runner.go`
   - runtime bootstrap
   - Promise settlement
4. `/home/manuel/code/wesen/2026-03-14--zk-tool/modules/obsidian/module.go`
   - actual JavaScript API surface
5. `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidian/client.go`
   - note and query operations
6. `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidiancli/runner.go`
   - raw subprocess execution
   - stdout sanitization
7. `/home/manuel/code/wesen/2026-03-14--zk-tool/scripts/js-tests`
   - concrete usage examples

## How To Write A New Read-Only Script

This section covers the normal script-authoring pattern and why each step exists.

- Start from one of the existing scripts in `/home/manuel/code/wesen/2026-03-14--zk-tool/scripts/js-tests`
- Prefer dynamic discovery over hardcoded note names
- Return JSON strings so the output is easy to inspect and stable in docs
- Avoid `create`, `append`, `prepend`, `move`, `rename`, and `delete` while learning

Recommended template:

```javascript
const obs = require("obsidian");

(async () => {
  const files = await obs.files({ ext: "md" });
  const path = files[0];
  const note = await obs.note(path);
  return JSON.stringify({
    path: note.path,
    title: note.title,
  });
})()
```

## How To Debug A Bad Result

This section explains how to narrow down bugs by layer instead of guessing. That discipline matters more than any single debugging trick.

If a script fails:

- Check the JS script first.
  - Is the Promise chain correct?
  - Are you parsing `obs.exec()` output correctly?
- Then check the native module.
  - Does the method exist in `/home/manuel/code/wesen/2026-03-14--zk-tool/modules/obsidian/module.go`?
  - Does it map arguments correctly?
- Then check the high-level client.
  - Does the Go client send the right parameters?
- Then check the transport.
  - Does the CLI spec name match the real `obsidian help` output?
  - Is startup noise contaminating stdout?

A practical debugging sequence looks like this:

```text
JS output wrong
  -> check script JSON shaping
  -> check module export method
  -> check Go client parameter names
  -> run ~/.local/bin/obsidian help <command>
  -> compare raw stdout with sanitizer behavior
```

## Current Read-Only Smoke Scripts

The safest way to learn the system is to run the existing scripts and inspect their JSON output.

| Script | What it proves |
|---|---|
| `obsidian-version.js` | The wrapper and module are reachable |
| `obsidian-sample-files.js` | `files()` returns note paths |
| `obsidian-read-first-note.js` | `read()` returns note content |
| `obsidian-note-inspect.js` | `note()` hydrates parsed fields |
| `obsidian-query-sample.js` | `query()` returns note objects |
| `obsidian-exec-vault.js` | `exec()` can read raw non-JSON CLI output |
| `obsidian-exec-tags.js` | `exec()` can read JSON CLI output |
| `obsidian-batch-sample.js` | `batch()` can map over note results |
| `obsidian-markdown-helpers.js` | `obs.md.*` helpers work without vault reads |

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| `No markdown files found` | The vault is empty or the wrapper targets the wrong vault | Run `~/.local/bin/obsidian vault` and confirm the active vault |
| `JSON.parse` fails in a script | `obs.exec()` returned plain text, not JSON | Check the underlying command format and parse manually |
| Help pages do not show your new doc | The doc file frontmatter is invalid or not embedded | Re-run `go run ./cmd/zk help` and inspect `pkg/doc/doc.go` |
| Query results include attachment notes | The script did not filter them out | Filter paths in JS before choosing samples |

## See Also

- `zk help zk-obsidian-system-overview`
- `zk help zk-obsidian-js-api-reference`
- `zk help zk-obsidian-smoke-tests-playbook`
- `zk help zk-obsidian-implementation-diary`
