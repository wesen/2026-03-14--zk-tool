const obs = require("obsidian");

(async () => {
  const rows = await obs.batch(
    { ext: "md", limit: 3 },
    (note) => ({
      path: note.path,
      title: note.title,
      headingCount: (note.headings || []).length,
      tagCount: (note.tags || []).length,
      taskCount: (note.tasks || []).length,
    }),
  );

  return JSON.stringify({
    count: rows.length,
    sample: rows,
  });
})()
