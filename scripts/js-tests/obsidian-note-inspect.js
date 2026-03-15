const obs = require("obsidian");

(async () => {
  const files = await obs.files({ ext: "md" });
  const candidates = files.filter(
    (path) => !path.startsWith("Attachments/") && !path.endsWith(".excalidraw.md"),
  );
  const path = candidates[0] || files[0];

  if (!path) {
    throw new Error("No markdown files found in the vault");
  }

  const note = await obs.note(path);
  return JSON.stringify({
    path: note.path,
    title: note.title,
    frontmatterKeys: Object.keys(note.frontmatter || {}).slice(0, 10),
    headings: (note.headings || []).slice(0, 10),
    tags: (note.tags || []).slice(0, 10),
    wikilinks: (note.wikilinks || []).slice(0, 10),
    tasks: (note.tasks || []).slice(0, 10),
  });
})()
