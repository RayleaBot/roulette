import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import App from "../src/App.vue";
import { defaults, effective, normalize, validate } from "../src/model";
import { buildReplay, sampleValues, middleSample } from "../src/replay";

let wrapper: VueWrapper | undefined;
let saved = defaults();
let failSave = "";
const writes: Array<Record<string, unknown>> = [];

beforeEach(() => {
  saved = defaults();
  failSave = "";
  writes.length = 0;
  vi.stubGlobal(
    "fetch",
    vi.fn(async (path: string, options?: RequestInit) => {
      let body: unknown = {};
      let status = 200;
      if (path === "/api/plugins/raylea.roulette")
        body = {
          plugin: {
            id: "raylea.roulette",
            name: "轮盘",
            version: "0.1.0",
            state: "running",
          },
        };
      else if (path.endsWith("/settings")) body = { values: saved };
      else if (path.endsWith("/secrets")) body = { configured: {} };
      else if (path === "/api/config")
        body = {
          config: {
            runtime: {
              plugin_detached_event_timeout_seconds: 900,
              max_detached_events_per_plugin: 8,
            },
          },
        };
      else if (path === "/api/adapters")
        body = {
          adapters: [
            {
              id: "a",
              display_name: "测试机器人",
              protocol: "onebot11",
              enabled: true,
              state: "connected",
              identity: { id: "9001", name: "测试" },
            },
          ],
        };
      else if (path.endsWith("/targets"))
        body = {
          available: true,
          groups: [{ target_id: "1001", target_name: "测试群" }],
          issues: [],
        };
      else if (path.endsWith("/management/actions")) {
        const request = JSON.parse(String(options?.body)) as {
          action: string;
          payload: { values: typeof saved };
        };
        writes.push(request);
        if (failSave) {
          status = 500;
          body = {
            error: {
              code: failSave,
              message: "模拟保存失败",
              details:
                failSave === "plugin.settings_apply_failed"
                  ? { committed: true }
                  : {},
            },
          };
        } else {
          saved = request.payload.values;
          body = { result: { values: saved } };
        }
      }
      return new Response(JSON.stringify(body), {
        status,
        headers: {
          "content-type": "application/json",
          "X-Raylea-CSRF": "fixture",
        },
      });
    }),
  );
  vi.spyOn(window, "confirm").mockReturnValue(true);
});
afterEach(() => {
  wrapper?.unmount();
  wrapper = undefined;
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

async function page() {
  wrapper = mount(App, { attachTo: document.body });
  await flushPromises();
  return wrapper;
}
function button(text: string) {
  const item = wrapper!.findAll("button").find((b) => b.text() === text);
  if (!item) throw new Error(`missing button ${text}`);
  return item;
}
async function addTrigger(word: string) {
  if (!wrapper!.find("#trigger-input").exists())
    await wrapper!.get("#trigger-add").trigger("click");
  const input = wrapper!.get("#trigger-input");
  await input.setValue(word);
  await input.trigger("keydown", { key: "Enter" });
}
function row(label: string) {
  const item = wrapper!
    .findAll(".row")
    .find((r) => r.find(".row-label").text() === label);
  if (!item) throw new Error(`missing row ${label}`);
  return item;
}
async function openPanel() {
  await wrapper!.get("#add-trigger").trigger("click");
  await flushPromises();
}
async function addGroup() {
  await openPanel();
  const card = wrapper!.findAll(".group-card").find((c) => c.text().includes("测试群"));
  if (!card) throw new Error("missing group card");
  await card.trigger("click");
  await button("添加这个群").trigger("click");
  await flushPromises();
}

describe("settings behavior", () => {
  it("saves through the validated management action and keeps the draft on failure", async () => {
    const w = await page();
    await w.get('[aria-label="删除口令 俄罗斯轮盘"]').trigger("click");
    await w.get('[aria-label="删除口令 roulette"]').trigger("click");
    await addTrigger("来一枪");
    failSave = "platform.internal_error";
    await button("保存设置").trigger("click");
    await flushPromises();
    expect(writes[0]).toMatchObject({
      action: "settings.save",
      payload: { values: { trigger_commands: ["轮盘", "来一枪"] } },
    });
    expect(w.find('[aria-label="删除口令 来一枪"]').exists()).toBe(true);
    expect(w.text()).toContain("草稿已保留");
    failSave = "";
    await button("保存设置").trigger("click");
    await flushPromises();
    expect(button("保存设置").attributes("disabled")).toBeDefined();
  });
  it("rejects a trigger with a prefix before it reaches the draft", async () => {
    const w = await page();
    await addTrigger("/开枪");
    expect(w.text()).toContain("口令不能包含空格或斜杠");
    expect(button("保存设置").attributes("disabled")).toBeDefined();
  });
  it("distinguishes committed but unapplied settings and allows a same-value retry", async () => {
    const w = await page();
    await addTrigger("新轮盘");
    failSave = "plugin.settings_apply_failed";
    await button("保存设置").trigger("click");
    await flushPromises();
    expect(w.text()).toContain("已保存，待应用");
    expect(button("重试应用").attributes("disabled")).toBeUndefined();
    failSave = "platform.internal_error";
    await button("重试应用").trigger("click");
    await flushPromises();
    await button("重新读取").trigger("click");
    await button("丢弃并重新读取").trigger("click");
    await flushPromises();
    expect(w.text()).toContain("已保存，待应用");
    expect(button("重试应用").attributes("disabled")).toBeUndefined();
    failSave = "";
    await button("重试应用").trigger("click");
    await flushPromises();
    expect(writes).toHaveLength(3);
  });
  it("adds a bot-scoped group, overrides one rule unit, and restores inheritance", async () => {
    const w = await page();
    await addGroup();
    const chambers = 'input[id="group_overrides.0.rules.chambers"]';
    expect(w.find(chambers).exists()).toBe(false);
    await row("弹膛").get(".row-scope .link").trigger("click");
    await w.get(chambers).setValue(4);
    expect(w.get(".scope-groups").text()).toContain("弹膛");
    await button("保存设置").trigger("click");
    await flushPromises();
    expect(saved.group_overrides[0]).toMatchObject({
      source_adapter: "a",
      bot_id: "9001",
      group_id: "1001",
      rules: { chambers: 4 },
    });
    expect(saved.group_overrides[0].rules).not.toHaveProperty("timeout_seconds");
    await row("弹膛").get(".row-scope .link").trigger("click");
    await button("保存设置").trigger("click");
    await flushPromises();
    expect(saved.group_overrides[0].rules).toEqual({});
    await button("删除本群设置").trigger("click");
    await button("确认删除").trigger("click");
    await button("保存设置").trigger("click");
    await flushPromises();
    expect(saved.group_overrides).toEqual([]);
  });
  it("edits a group-only reply from the replay without touching the default", async () => {
    const w = await page();
    await addGroup();
    await w.get('[aria-label="编辑命中并禁言台词"]').trigger("click");
    const text = w.get("#reply-hit");
    expect(text.attributes("disabled")).toBeDefined();
    await w.get('.editor [role="switch"]').trigger("click");
    await text.setValue("<target> 中弹。");
    await button("保存设置").trigger("click");
    await flushPromises();
    expect(saved.group_overrides[0].replies).toEqual({ hit: "<target> 中弹。" });
    expect(saved.replies.hit).toBe(defaults().replies.hit);
  });
  it("blocks invalid cross-field settings and shows the runtime timeout limit", async () => {
    const w = await page();
    expect(w.text()).toContain("870 秒");
    await w.get('input[id="rules.bullets"]').setValue(6);
    expect(w.text()).toContain("最大子弹数必须小于最小弹膛数");
    expect(button("保存设置").attributes("disabled")).toBeDefined();
    expect(writes).toHaveLength(0);
  });
});

describe("add group panel", () => {
  it("shows QQ avatars for the only bot and its groups", async () => {
    const w = await page();
    await openPanel();
    expect(w.get(".bot-card").attributes("aria-pressed")).toBe("true");
    expect(w.get(".bot-card img").attributes("src")).toBe(
      "https://q1.qlogo.cn/g?b=qq&nk=9001&s=100",
    );
    expect(w.get(".group-card img").attributes("src")).toBe(
      "https://p.qlogo.cn/gh/1001/1001/100",
    );
  });
  it("opens an existing group instead of adding it twice", async () => {
    const w = await page();
    await addGroup();
    await w.get(".scope-nav > button.scope").trigger("click");
    await openPanel();
    const card = w.findAll(".group-card").find((c) => c.text().includes("测试群"))!;
    expect(card.text()).toContain("已单独设置");
    await card.trigger("click");
    await button("打开已有设置").trigger("click");
    await flushPromises();
    expect(w.findAll(".scope-groups li")).toHaveLength(1);
    expect(w.get(".inspector-title h2").text()).toBe("测试群");
  });
});

describe("effective rules", () => {
  it("inherits unselected fields and catches a global change invalidating a group", () => {
    const s = defaults();
    s.group_overrides = [
      {
        source_adapter: "a",
        bot_id: "9001",
        group_id: "1001",
        rules: { chambers: 3 },
        replies: {},
      },
    ];
    expect(effective(s, s.group_overrides[0]).rules.timeout_seconds).toBe(600);
    expect(validate(s)).toEqual([]);
    s.rules.bullets = 3;
    expect(
      validate(s).some((i) => i.field === "group_overrides.0.rules.bullets"),
    ).toBe(true);
  });
  it("keeps offline group overrides and drops unknown nested fields", () => {
    const s = normalize({
      group_overrides: [
        {
          source_adapter: "offline",
          bot_id: "9001",
          group_id: "1001",
          rules: { timeout_seconds: 200, future: 1 },
          replies: { future: "unused" },
        },
      ],
    });
    expect(s.group_overrides[0]).toMatchObject({
      source_adapter: "offline",
      rules: { timeout_seconds: 200 },
      replies: {},
    });
    expect(Object.keys(s.group_overrides[0].rules)).toEqual([
      "timeout_seconds",
    ]);
  });
});

describe("replay", () => {
  const replay = (
    scenario: Parameters<typeof buildReplay>[0],
    patch: Partial<ReturnType<typeof defaults>["rules"]>,
  ) => {
    const rules = { ...defaults().rules, ...patch };
    return buildReplay(scenario, {
      rules,
      replies: defaults().replies,
      values: sampleValues(rules, middleSample),
      trigger: "轮盘",
      ending: "timeout",
    });
  };
  const keys = (lines: ReturnType<typeof replay>) =>
    lines.flatMap((l) => (l.kind === "bot" ? [l.key] : []));

  it("never ends a hit round early and shows the drawn mute duration", () => {
    const lines = replay("hit", {
      random_chambers: false,
      chambers: 2,
      random_mute: false,
      mute_seconds: 90,
    });
    expect(keys(lines)).toEqual(["start", "hit", "end"]);
    const text = (key: string) =>
      lines
        .flatMap((l) => (l.kind === "bot" && l.key === key ? l.segments : []))
        .map((s) => s.text)
        .join("");
    expect(text("hit")).toContain("1分30秒");
    expect(text("start")).toContain("2 个弹膛");
  });
  it("follows the early-end switch when only live rounds remain", () => {
    const on = replay("loaded", { random_chambers: false, chambers: 6 });
    expect(keys(on).at(-1)).toBe("all_loaded");
    const off = replay("loaded", {
      random_chambers: false,
      chambers: 6,
      end_when_all_loaded: false,
    });
    expect(keys(off).slice(-2)).toEqual(["hit", "end"]);
  });
});
