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

  const content = await obs.read(path);
  return JSON.stringify({
    path,
    chars: content.length,
    preview: content.slice(0, 200),
  });
})()
