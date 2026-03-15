---
Title: ZK Obsidian System Overview
Slug: zk-obsidian-system-overview
Short: Architecture map for the ZK repository's Obsidian JavaScript integration.
Topics:
- zk
- obsidian
- architecture
- goja
- glazed
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
SectionType: GeneralTopic
---

This page explains what the ZK Obsidian subsystem is, how requests move through it, and why the code is split across multiple packages. Read this first if you are new to the repository and want a mental model before reading individual files.

The subsystem exists to let JavaScript code talk to a running Obsidian instance through the local `obsidian` wrapper while keeping the command-line surface stable and scriptable. The important design choice is that JavaScript evaluation, Obsidian command transport, markdown parsing, and user-facing CLI help are separate concerns. That separation is what makes the system debuggable.

## What The System Does

At a high level, the repository provides a Glazed/Cobra command that runs JavaScript files. Those JavaScript files can `require("obsidian")` and call a Go-backed module that exposes read and write operations against the local Obsidian CLI.

The current most important user-facing command is:

```bash
go run ./cmd/zk obsidian run-script scripts/js-tests/obsidian-version.js
```

That command:

- creates a goja runtime using `go-go-goja`
- registers the local `obsidian` native module
- injects CLI configuration such as `~/.local/bin/obsidian`
- executes the script
- waits for Promise results
- returns structured output through Glazed

## Architecture Diagram

The full request path looks like this:

```text
+-----------------------------+
| zk CLI (Cobra + Glazed)     |
| cmd/zk/main.go              |
+-------------+---------------+
              |
              v
+-----------------------------+
| run-script command          |
| cmd/zk/cmds/obsidian/...    |
+-------------+---------------+
              |
              v
+-----------------------------+
| JS runtime bridge           |
| pkg/obsidianjs/runner.go    |
+-------------+---------------+
              |
              v
+-----------------------------+
| goja runtime + require()    |
| go-go-goja/engine           |
+-------------+---------------+
              |
              v
+-----------------------------+
| local native module         |
| modules/obsidian/module.go  |
+-------------+---------------+
              |
              v
+-----------------------------+
| high-level Obsidian client  |
| pkg/obsidian/*.go           |
+-------------+---------------+
              |
              v
+-----------------------------+
| CLI transport + parsers     |
| pkg/obsidiancli/*.go        |
+-------------+---------------+
              |
              v
+-----------------------------+
| ~/.local/bin/obsidian       |
| Flatpak wrapper into app    |
+-------------+---------------+
              |
              v
+-----------------------------+
| Running Obsidian vault      |
+-----------------------------+
```

## Package Map

This section covers the package layout, how each package works in practice, and why it matters for maintenance. If you skip this split, the code starts to look like one large integration blob and you lose track of where to debug failures.

| Package | Responsibility | Why it exists |
|---|---|---|
| `/home/manuel/code/wesen/2026-03-14--zk-tool/cmd/zk` | Cobra root and command registration | Keeps the user-facing CLI thin |
| `/home/manuel/code/wesen/2026-03-14--zk-tool/cmd/zk/cmds/obsidian` | Glazed command definitions | Separates CLI parsing from business logic |
| `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidianjs` | Runtime bootstrap and Promise settlement | Keeps goja ownership logic out of command handlers |
| `/home/manuel/code/wesen/2026-03-14--zk-tool/modules/obsidian` | JS-facing native module | Defines the `require("obsidian")` API |
| `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidian` | High-level note/query client | Gives scripts a friendlier model than raw CLI strings |
| `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidiancli` | Raw CLI invocation and output parsing | Contains the wrapper-specific normalization logic |
| `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidianmd` | Markdown helper functions | Lets scripts reuse markdown parsing without reimplementing it |
| `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/doc` | Embedded Glazed help pages | Makes the docs discoverable from `zk help` |
| `/home/manuel/code/wesen/2026-03-14--zk-tool/scripts/js-tests` | Read-only smoke scripts | Gives interns safe, concrete examples |

## Runtime Responsibilities

This section explains which layer owns which behavior. That ownership split is critical because many bugs only make sense when you know which layer is supposed to solve them.

- `cmd/zk/cmds/obsidian/run_script.go`
  - decodes flags
  - calls the JS runner
  - emits Glazed rows
- `pkg/obsidianjs/runner.go`
  - resolves script path
  - configures the Obsidian module
  - builds the runtime
  - settles Promise values before returning
- `modules/obsidian/module.go`
  - exposes JS functions like `version()`, `files()`, `note()`, `query()`, `batch()`, and `exec()`
  - maps JS objects into Go option structs
  - converts results back into JS-friendly values
- `pkg/obsidian/client.go`
  - performs note resolution
  - caches note contents
  - implements the fluent query behavior
- `pkg/obsidiancli/runner.go`
  - serializes subprocess access
  - runs the actual `obsidian` command
  - strips Flatpak/Electron startup noise from stdout

## Data Flow Pseudocode

This pseudocode shows the steady-state happy path. It matters because it gives you a compact model to compare against when a real run diverges.

```text
function runScript(scriptPath):
    cfg = resolveCLIConfig()
    runtime = buildGojaRuntime()
    registerLocalObsidianModule(runtime)
    runtime.eval('require("obsidian").configure(cfg)')
    result = runtime.eval(scriptContents)
    if result is Promise:
        wait until settled
    return normalizedOutput(result)
```

## Common Failure Modes

This section covers the failures you are most likely to see first and why they happen.

| Problem | Cause | Solution |
|---|---|---|
| Script output includes Flatpak noise | The Obsidian wrapper prints startup chatter before payload data | Inspect `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidiancli/runner.go` and update the sanitizer |
| Promise never resolves | The JS code returned a pending Promise or an async path deadlocked | Reproduce with a minimal script and inspect `pkg/obsidianjs/runner.go` plus `modules/obsidian/module.go` |
| `require("obsidian")` fails | The local module is not imported into the runtime bridge | Verify the blank import in `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/obsidianjs/runner.go` |
| Help page does not appear | The markdown file is not embedded or frontmatter is invalid | Check `/home/manuel/code/wesen/2026-03-14--zk-tool/pkg/doc/doc.go` and run `go run ./cmd/zk help` |

## See Also

- `zk help zk-obsidian-intern-guide`
- `zk help zk-obsidian-js-api-reference`
- `zk help zk-obsidian-smoke-tests-playbook`
- `zk help zk-obsidian-implementation-diary`
