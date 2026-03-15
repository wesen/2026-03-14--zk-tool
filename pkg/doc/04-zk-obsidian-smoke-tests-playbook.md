---
Title: ZK Obsidian Smoke Tests Playbook
Slug: zk-obsidian-smoke-tests-playbook
Short: Read-only smoke-test catalog and execution guide for the local Obsidian JS integration.
Topics:
- zk
- obsidian
- testing
- smoke-tests
- javascript
Commands:
- obsidian
- run-script
Flags:
- output
- vault
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Application
---

This playbook explains the read-only smoke tests that ship with the repository, what each one proves, and how to use them to isolate failures. It is the operational companion to the API reference: the API doc explains what should happen, while this page explains how to verify it on a real machine.

The tests are intentionally script-first. That choice matters because scripts are the closest thing to real user behavior. If a script works through `zk obsidian run-script`, the full chain from Cobra to Glazed to goja to the Obsidian wrapper is functioning.

## Script Inventory

All scripts live in:

- `/home/manuel/code/wesen/2026-03-14--zk-tool/scripts/js-tests`

Current read-only set:

| Script | Focus | Notes |
|---|---|---|
| `obsidian-version.js` | wrapper reachability | smallest end-to-end test |
| `obsidian-sample-files.js` | file listing | good first vault-content check |
| `obsidian-read-first-note.js` | note content reads | chooses a candidate note dynamically |
| `obsidian-note-inspect.js` | hydrated note metadata | tests helper-derived fields |
| `obsidian-query-sample.js` | query builder | tests `query().run()` |
| `obsidian-exec-vault.js` | raw plain-text `exec()` | proves custom parsing path |
| `obsidian-exec-tags.js` | raw JSON `exec()` | proves generic command escape hatch |
| `obsidian-batch-sample.js` | `batch()` mapping | tests JS mapper flow |
| `obsidian-markdown-helpers.js` | `obs.md.*` | tests parser helpers without vault I/O |

## Recommended Execution Order

Run tests in this order because it minimizes ambiguity. If the first test fails, there is no point debugging the later ones yet.

```bash
go run ./cmd/zk obsidian run-script scripts/js-tests/obsidian-version.js --output json
go run ./cmd/zk obsidian run-script scripts/js-tests/obsidian-sample-files.js --output json
go run ./cmd/zk obsidian run-script scripts/js-tests/obsidian-read-first-note.js --output json
go run ./cmd/zk obsidian run-script scripts/js-tests/obsidian-note-inspect.js --output json
go run ./cmd/zk obsidian run-script scripts/js-tests/obsidian-query-sample.js --output json
go run ./cmd/zk obsidian run-script scripts/js-tests/obsidian-exec-vault.js --output json
go run ./cmd/zk obsidian run-script scripts/js-tests/obsidian-exec-tags.js --output json
go run ./cmd/zk obsidian run-script scripts/js-tests/obsidian-batch-sample.js --output json
go run ./cmd/zk obsidian run-script scripts/js-tests/obsidian-markdown-helpers.js --output json
```

## How To Interpret Failures

This section covers the common failure clustering pattern and why the order above is useful.

- If `obsidian-version.js` fails:
  - suspect wrapper path
  - suspect runtime bootstrap
  - suspect `require("obsidian")` registration
- If version passes but file listing fails:
  - suspect high-level client or CLI spec mapping
- If read works but note/query fail:
  - suspect markdown helpers or query hydration
- If typed APIs work but `exec()` fails:
  - suspect raw parsing assumptions in the JS script
- If vault-backed scripts fail but markdown helpers pass:
  - suspect the live CLI integration, not the parser helpers

## Review Checklist For A New Script

Use this checklist before you commit a new smoke script.

- Is the script read-only?
- Does it discover sample input dynamically when possible?
- Does it return JSON rather than ad-hoc console text?
- Does it use the highest-level API that fits?
- Does it avoid hardcoding one machine-specific path unless the path is the thing being tested?
- Can another person understand what success looks like from the returned JSON alone?

## Minimal Script Authoring Pattern

```javascript
const obs = require("obsidian");

(async () => {
  const files = await obs.files({ ext: "md" });
  const path = files[0];
  return JSON.stringify({
    path,
  });
})()
```

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| Script hangs longer than expected | A Promise is still pending | Reduce the script to the smallest awaited call and rerun |
| `exec("vault")` parsing is wrong | The output separator assumption is wrong | Inspect the raw string and adjust parsing in the script |
| A script is too brittle across vaults | It hardcodes one note or one folder | Make it discover candidate notes dynamically |
| The JSON result is hard to review | The script returns too much raw content | Slice arrays and trim previews |

## See Also

- `zk help zk-obsidian-intern-guide`
- `zk help zk-obsidian-js-api-reference`
- `zk help zk-obsidian-implementation-diary`
