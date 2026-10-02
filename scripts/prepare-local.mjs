import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const core = path.resolve(
  process.argv[2] ?? path.join(root, "../../RayleaBot"),
);
const sdkGo = path.join(core, "sdk/go"),
  sdkVue = path.join(core, "sdk/vue");
await fs.access(path.join(sdkGo, "go.mod"));
await fs.access(path.join(sdkVue, "package.json"));
await fs.mkdir(path.join(root, ".rayleabot/sdk"), { recursive: true });
const link = path.join(root, ".rayleabot/sdk/vue");
try {
  const actual = await fs.realpath(link),
    expected = await fs.realpath(sdkVue);
  if (actual !== expected)
    throw new Error(`Existing SDK link points elsewhere: ${link}`);
} catch (error) {
  if (error.code !== "ENOENT") throw error;
  await fs.symlink(
    sdkVue,
    link,
    process.platform === "win32" ? "junction" : "dir",
  );
}
// Relative use entries avoid Go's Windows path-prefix comparison between
// absolute slash paths and the native working-directory spelling.
const quote = (p) =>
  JSON.stringify(path.relative(root, p).replaceAll("\\", "/") || ".");
await fs.writeFile(
  path.join(root, "go.work"),
  `go 1.27.1\n\nuse (\n ${quote(root)}\n ${quote(sdkGo)}\n)\n\nreplace github.com/RayleaBot/RayleaBot/sdk/go v0.7.0 => ${quote(sdkGo)}\n`,
);
console.log(`Local SDK workspace ready: ${root}`);
