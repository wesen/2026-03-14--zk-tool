---
Title: Go implementation plan for obsidian.js API with go-go-goja and Bobatea REPL
Ticket: ZK-001
Status: active
Topics:
    - obsidian
    - goja
    - bobatea
    - repl
    - api-design
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: DESIGN-obsidian-js-api.md
      Note: Primary source design to port
    - Path: PROJ - ZK Tool.md
      Note: Local workflow that should become a first-class consumer
    - Path: scripts/zk_create.py
      Note: Current Python implementation that the guide references for migration context
ExternalSources: []
Summary: |
    Ticket workspace for the obsidian.js Go/goja/Bobatea port planning effort, including the main design guide, diary, and implementation checklist.
LastUpdated: 2026-03-15T15:04:15.687711077-04:00
WhatFor: |
    Landing page for the documentation ticket that turns the local obsidian.js design into a concrete Go/goja/Bobatea implementation plan.
WhenToUse: Start here when orienting yourself to ticket ZK-001, then follow the links to the design guide, diary, tasks, and changelog.
---


# Go implementation plan for obsidian.js API with go-go-goja and Bobatea REPL

## Overview

This ticket captures the design and implementation plan for porting the local `obsidian.js` API design into a Go-first system built on `go-go-goja`, `goja`, and a Bobatea REPL. The main deliverable is an intern-oriented guide that explains the current source design, the relevant runtime and REPL infrastructure that already exists, the gaps that still need code, and a phased file-level plan to implement the module cleanly.

## Key Links

- **Primary design guide**: `design-doc/01-obsidian-js-go-port-analysis-design-and-implementation-guide.md`
- **Investigation diary**: `reference/01-diary.md`
- **Task checklist**: `tasks.md`
- **Changelog**: `changelog.md`
- **Related Files**: See frontmatter `RelatedFiles`
- **External Sources**: See frontmatter `ExternalSources`

## Status

Current status: **active**

## Topics

- obsidian
- goja
- bobatea
- repl
- api-design

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design-doc/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
