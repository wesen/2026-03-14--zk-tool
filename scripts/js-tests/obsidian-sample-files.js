const obs = require("obsidian");

(async () => {
  const files = await obs.files({ ext: "md" });
  return JSON.stringify({
    count: files.length,
    sample: files.slice(0, 5),
  });
})()
