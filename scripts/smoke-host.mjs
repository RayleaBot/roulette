// Local validation host: serves built assets and sends management operations to
// the real plugin process. All bot identities and OneBot responses are fixtures.
import http from "node:http";
import { spawn } from "node:child_process";
import { readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { createInterface } from "node:readline";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const manifest = JSON.parse(
  await readFile(path.join(root, "info.json"), "utf8"),
);
let config = structuredClone(manifest.default_config);
const executable =
  process.argv[2] ?? path.join(root, "dist", "smoke", "roulette.exe");
const child = spawn(executable, [], {
  cwd: root,
  stdio: ["pipe", "pipe", "pipe"],
  windowsHide: true,
});
const pending = new Map();
let sequence = 0;
const write = (frame) => child.stdin.write(`${JSON.stringify(frame)}\n`);
const ready = new Promise((resolve, reject) => {
  child.once("error", reject);
  createInterface({ input: child.stdout }).on("line", (line) => {
    const frame = JSON.parse(line);
    if (frame.type === "init_ack") {
      resolve();
      return;
    }
    if (frame.type === "action") {
      let data = {};
      if (frame.action === "config.write") {
        config = { ...config, ...frame.data.values };
        data = { changed_keys: Object.keys(frame.data.values) };
      } else if (frame.action === "group.member.get")
        data = { role: frame.data.user_id === "9001" ? "admin" : "member" };
      else if (frame.action === "event.detach")
        data = { deadline_at_ms: Date.now() + 900000 };
      else if (frame.action === "message.send")
        data = { message_id: String(++sequence) };
      write({
        type: "result",
        request_id: frame.request_id,
        status: "success",
        data,
      });
      if (frame.action === "config.write")
        write({
          type: "event",
          request_id: `config-${++sequence}`,
          deadline_at_ms: Date.now() + 60000,
          event: {
            event_id: `config-${sequence}`,
            event_type: "config.changed",
            source_protocol: "platform",
            source_adapter: "management.internal",
            timestamp: Date.now(),
            payload: { config, changed_keys: Object.keys(frame.data.values) },
          },
        });
      return;
    }
    const job = pending.get(frame.request_id);
    if (job) {
      pending.delete(frame.request_id);
      clearTimeout(job.timer);
      job.resolve(frame);
    }
  });
});
child.stderr.on("data", (data) => process.stderr.write(data));
child.on("exit", () => {
  for (const job of pending.values()) {
    clearTimeout(job.timer);
    job.reject(new Error("plugin exited"));
  }
  pending.clear();
});
write({
  type: "init",
  request_id: "init",
  protocol_version: "4",
  plugin_id: manifest.id,
  timezone: "Asia/Shanghai",
  concurrency: 4,
  config,
  bots: [
    { source_adapter: "fixture-a", source_protocol: "onebot11", id: "9001" },
  ],
  super_admins: [],
  command_prefixes: ["/"],
});
await ready;
const invoke = (action, payload) =>
  new Promise((resolve, reject) => {
    const id = `manage-${++sequence}`;
    const timer = setTimeout(() => {
      pending.delete(id);
      reject(new Error("management timeout"));
    }, 10000);
    pending.set(id, { resolve, reject, timer });
    write({
      type: "event",
      request_id: id,
      deadline_at_ms: Date.now() + 60000,
      event: {
        event_id: id,
        event_type: "management.action",
        source_protocol: "platform",
        source_adapter: "management.internal",
        timestamp: Date.now(),
        payload: { action, payload },
      },
    });
  });
const server = http.createServer(async (req, res) => {
  try {
    const url = new URL(req.url, "http://localhost");
    const json = (data, status = 200) => {
      res.writeHead(status, {
        "content-type": "application/json; charset=utf-8",
        "X-Raylea-CSRF": "local-fixture",
      });
      res.end(JSON.stringify(data));
    };
    if (url.pathname === "/api/config")
      return json({
        config: {
          runtime: {
            plugin_detached_event_timeout_seconds: 900,
            max_detached_events_per_plugin: 8,
          },
        },
      });
    if (url.pathname === "/api/adapters")
      return json({
        adapters: [
          {
            id: "fixture-a",
            protocol: "onebot11",
            display_name: "测试机器人 A",
            enabled: true,
            state: "connected",
            identity: { id: "9001", name: "测试机器人" },
          },
          {
            id: "fixture-b",
            protocol: "onebot11",
            display_name: "测试机器人 B",
            enabled: true,
            state: "connected",
            identity: { id: "9002", name: "测试机器人" },
          },
        ],
      });
    if (url.pathname.endsWith("/onebot11/targets"))
      return json({
        available: true,
        groups: [
          { target_id: "1001", target_name: "轮盘测试群" },
          { target_id: "1002", target_name: "周末聊天群" },
        ],
        issues: [],
      });
    if (url.pathname === `/api/plugins/${manifest.id}`)
      return json({
        plugin: {
          id: manifest.id,
          name: manifest.name,
          version: manifest.version,
          state: "running",
        },
      });
    if (url.pathname.endsWith("/settings")) return json({ values: config });
    if (url.pathname.endsWith("/secrets")) return json({ configured: {} });
    if (url.pathname.endsWith("/management/actions")) {
      if (req.method !== "POST")
        return json(
          {
            error: {
              code: "platform.invalid_request",
              message: "POST required",
            },
          },
          405,
        );
      let raw = "";
      for await (const chunk of req) {
        raw += chunk;
        if (raw.length > 1024 * 1024) throw new Error("request too large");
      }
      const { action, payload } = JSON.parse(raw);
      const result = await invoke(action, payload);
      if (result.type === "error")
        return json(
          {
            error: {
              code: result.code,
              message: result.message,
              details: result.details,
            },
          },
          400,
        );
      return json({ result: result.data });
    }
    const prefix = `/plugin-ui/${manifest.id}/`;
    if (!url.pathname.startsWith(prefix)) {
      res.writeHead(302, { location: `${prefix}index.html?page=settings` });
      res.end();
      return;
    }
    const relative =
      decodeURIComponent(url.pathname.slice(prefix.length)) || "index.html";
    const assets = path.join(root, "ui", "dist"),
      file = path.resolve(assets, relative);
    if (!file.startsWith(assets + path.sep)) {
      res.writeHead(404);
      res.end();
      return;
    }
    const content = await readFile(file);
    const contentType =
      {
        ".html": "text/html; charset=utf-8",
        ".js": "text/javascript; charset=utf-8",
        ".css": "text/css; charset=utf-8",
      }[path.extname(file)] ?? "application/octet-stream";
    res.writeHead(200, {
      "content-type": contentType,
      "Content-Security-Policy": `default-src 'none'; script-src http://${req.headers.host}${prefix}; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; media-src 'self' data: https:; font-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; form-action 'self'; frame-ancestors 'self'`,
    });
    res.end(content);
  } catch (error) {
    res.writeHead(500, { "content-type": "application/json" });
    res.end(
      JSON.stringify({
        error: { code: "fixture.failed", message: error.message },
      }),
    );
  }
});
server.listen(0, "127.0.0.1", () =>
  console.log(
    `SMOKE_URL=http://127.0.0.1:${server.address().port}/plugin-ui/${manifest.id}/index.html?page=settings`,
  ),
);
const close = () => {
  server.close();
  child.kill();
};
process.once("SIGINT", close);
process.once("SIGTERM", close);
