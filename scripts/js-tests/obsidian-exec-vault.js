const obs = require("obsidian");

function parseKeyValue(raw) {
  const out = {};
  for (const line of raw.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    const tabParts = trimmed.split(/\t+/);
    if (tabParts.length >= 2) {
      const key = tabParts[0].trim();
      const value = tabParts.slice(1).join(" ").trim();
      if (key) out[key] = value;
      continue;
    }
    const idx = trimmed.search(/[:=]/);
    if (idx === -1) continue;
    const key = trimmed.slice(0, idx).trim();
    const value = trimmed.slice(idx + 1).trim();
    out[key] = value;
  }
  return out;
}

(async () => {
  const raw = await obs.exec("vault");
  const info = parseKeyValue(raw);
  return JSON.stringify({
    name: info.name || null,
    path: info.path || null,
    files: Number(info.files || 0),
    folders: Number(info.folders || 0),
  });
})()
