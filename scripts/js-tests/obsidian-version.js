const obs = require("obsidian");

(async () => {
  const version = await obs.version();
  return JSON.stringify({
    version,
  });
})()
