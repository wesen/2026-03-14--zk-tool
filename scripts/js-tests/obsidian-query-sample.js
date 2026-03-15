const obs = require("obsidian");

(async () => {
  const rows = await obs
    .query()
    .withExtension("md")
    .limit(3)
    .run();

  return JSON.stringify({
    count: rows.length,
    sample: rows.map((row) => ({
      path: row.path,
      title: row.title,
      tags: (row.tags || []).slice(0, 3),
      headings: (row.headings || []).slice(0, 3),
    })),
  });
})()
