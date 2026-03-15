const obs = require("obsidian");

const sample = `---
title: Demo
tags:
  - zk
  - intern
---

# Intro

Link to [[Alpha Note]] and [[Beta Note|Beta]].

- [ ] First task
- [x] Done task

#tag-inline
`;

(async () => {
  return JSON.stringify({
    frontmatter: obs.md.frontmatter(sample),
    headings: obs.md.headings(sample),
    tags: obs.md.tags(sample),
    wikilinks: obs.md.wikilinks(sample),
    tasks: obs.md.tasks(sample),
  });
})()
