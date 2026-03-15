const obs = require("obsidian");

(async () => {
  const raw = await obs.exec("tags", { format: "json", counts: true });
  const parsed = JSON.parse(raw || "[]");
  return JSON.stringify({
    count: parsed.length,
    sample: parsed.slice(0, 10),
  });
})()
