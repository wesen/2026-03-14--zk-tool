# Changelog

## 2026-03-15

- Initial workspace created.
- Initialized `docmgr` in this repository (`.ttmp.yaml` + `ttmp/`) so the ticket workspace could exist.
- Created ticket `ZK-001` for the obsidian.js Go/goja/Bobatea implementation effort.
- Added a detailed design/implementation guide that maps the local source design onto `go-go-goja` and `bobatea`.
- Added a diary entry capturing the commands run, the initial `docmgr` setup failure, and the validation/upload workflow.
- Implemented `pkg/obsidiancli` in `go-go-goja` and committed it as `229c9f7` (`feat(obsidiancli): add Obsidian CLI transport package`).
- Implemented `pkg/obsidianmd` in `go-go-goja` and committed it as `f7961ef` (`feat(obsidianmd): add Obsidian markdown parsing helpers`).
- Implemented `pkg/obsidian` in `go-go-goja` and committed it as `d7a8dc1` (`feat(obsidian): add high-level Obsidian client layer`).
- Implemented `modules/obsidian` in `go-go-goja` and committed it as `4faf260` (`feat(obsidian): add goja native module`).
- Implemented REPL evaluator Promise-settlement and expression-style top-level `await` support in `go-go-goja` and committed it as `8d7ab9d` (`feat(repl): settle promise results and support top-level await expressions`).
- Confirmed that repo-local testing and commit hooks must run with `GOWORK=off` because the parent `go.work` does not include `go-go-goja`.
