import manifest from "../../info.json";

export type Rules = typeof manifest.default_config.rules;
export type Replies = typeof manifest.default_config.replies;
export type RuleKey = keyof Rules;
export type ReplyKey = keyof Replies;
export interface GroupOverride {
  source_adapter: string;
  bot_id: string;
  group_id: string;
  group_name?: string;
  rules: Partial<Rules>;
  replies: Partial<Replies>;
}
export interface Settings {
  trigger_commands: string[];
  rules: Rules;
  replies: Replies;
  group_overrides: GroupOverride[];
}
export interface Issue {
  field: string;
  message: string;
}
export interface Adapter {
  id: string;
  protocol: string;
  display_name: string;
  enabled: boolean;
  state: string;
  identity?: { id: string; name: string; avatar_url?: string };
}
export interface Group {
  target_id: string;
  target_name: string;
}
export function clone<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}
export function defaults(): Settings {
  return clone(manifest.default_config);
}
const object = (v: unknown): v is Record<string, unknown> =>
  !!v && typeof v === "object" && !Array.isArray(v);

// Preserve invalid declared values so administrators can see and correct them.
// Undeclared values are dropped in the same way as the backend decoder.
export function normalize(raw: Record<string, unknown>): Settings {
  const d = defaults();
  const known = <T extends object>(base: T, v: unknown): T =>
    Object.fromEntries(
      Object.keys(base).map((k) => [
        k,
        object(v) && k in v ? v[k] : base[k as keyof T],
      ]),
    ) as T;
  const patch = <T extends object>(base: T, v: unknown): Partial<T> =>
    object(v)
      ? (Object.fromEntries(
          Object.keys(base)
            .filter((k) => k in v)
            .map((k) => [k, v[k]]),
        ) as Partial<T>)
      : {};
  return {
    trigger_commands: Array.isArray(raw.trigger_commands)
      ? (raw.trigger_commands as string[])
      : d.trigger_commands,
    rules: known(d.rules, raw.rules),
    replies: known(d.replies, raw.replies),
    group_overrides: Array.isArray(raw.group_overrides)
      ? raw.group_overrides.map((v: unknown) => {
          const g = object(v) ? v : {};
          return {
            source_adapter: String(g.source_adapter ?? ""),
            bot_id: String(g.bot_id ?? ""),
            group_id: String(g.group_id ?? ""),
            group_name: String(g.group_name ?? ""),
            rules: patch(d.rules, g.rules),
            replies: patch(d.replies, g.replies),
          };
        })
      : [],
  };
}

export const replyLabels: Record<ReplyKey, string> = {
  start: "开局",
  miss: "空枪",
  hit: "命中并禁言",
  immune: "命中但权限豁免",
  end: "子弹耗尽",
  all_loaded: "剩余全是实弹",
  timeout: "超时结束",
  timeout_hit: "超时并禁言发起人",
  timeout_immune: "超时但发起人豁免",
  stopped: "管理员停止",
};
export const replyKeys = Object.keys(replyLabels) as ReplyKey[];
export const tokenLabels: Record<string, string> = {
  "<bullet>": "初始子弹数",
  "<chamber>": "初始弹膛数",
  "<remain-bullet>": "剩余子弹数",
  "<remain-chamber>": "剩余弹膛数",
  "<mute-s>": "禁言秒数",
  "<mute-f>": "禁言时长",
  "<timeout-s>": "超时秒数",
  "<timeout-f>": "超时时长",
  "<target>": "提及参与者或发起人",
};
export function effective(s: Settings, g?: GroupOverride) {
  return {
    rules: { ...s.rules, ...g?.rules },
    replies: { ...s.replies, ...g?.replies },
  };
}
export function groupKey(g: GroupOverride): string {
  return JSON.stringify([g.source_adapter, g.bot_id, g.group_id]);
}

export function issueLabel(s: Settings, field: string): string {
  const parts = field.split(".");
  let scope = "默认规则";
  if (parts[0] === "group_overrides" && parts.length > 1) {
    const group = s.group_overrides[Number(parts[1])];
    scope = group ? group.group_name || `群 ${group.group_id}` : "单独设置的群";
    parts.splice(0, 2);
  }
  const labels: Record<string, string> = {
    trigger_commands: "命令触发词",
    group_overrides: "单独设置的群",
    source_adapter: "机器人实例",
    bot_id: "机器人账号",
    group_id: "群号",
    group_name: "群名",
    random_chambers: "弹膛生成方式",
    chambers: "弹膛数量",
    chambers_max: "弹膛上限",
    random_bullets: "子弹生成方式",
    bullets: "子弹数量",
    bullets_max: "子弹上限",
    random_mute: "禁言生成方式",
    mute_seconds: "禁言时长",
    mute_max_seconds: "禁言上限",
    timeout_seconds: "超时时长",
    timeout_mute: "超时禁言",
    end_when_all_loaded: "全实弹提前结束",
  };
  const key = parts.at(-1) ?? "";
  const label =
    labels[key] ??
    (key in replyLabels ? `${replyLabels[key as ReplyKey]}台词` : undefined);
  return label ? `${scope} · ${label}` : scope;
}

export function validate(s: Settings, maxTimeout = 3570): Issue[] {
  const issues: Issue[] = [];
  const add = (field: string, message: string) =>
    issues.push({ field, message });
  if (!s.trigger_commands.length || s.trigger_commands.length > 20)
    add("trigger_commands", "请设置 1–20 个触发词。");
  if (
    s.trigger_commands.some(
      (v, i, all) =>
        typeof v !== "string" ||
        !v ||
        [...v].length > 32 ||
        /[\s/\\]/u.test(v) ||
        v === "停止轮盘" ||
        all.indexOf(v) !== i,
    )
  )
    add(
      "trigger_commands",
      "触发词不可为空、重复、包含空白或前缀，也不能使用停止轮盘。",
    );
  const check = (r: Rules, replies: Replies, prefix: string) => {
    const bounds: [RuleKey, number, number][] = [
      ["chambers", 2, 100],
      ["chambers_max", 2, 100],
      ["bullets", 1, 99],
      ["bullets_max", 1, 99],
      ["mute_seconds", 1, 2592000],
      ["mute_max_seconds", 1, 2592000],
      ["timeout_seconds", 1, maxTimeout],
    ];
    for (const [key, min, max] of bounds)
      if (
        !Number.isInteger(r[key]) ||
        Number(r[key]) < min ||
        Number(r[key]) > max
      )
        add(
          `${prefix}rules.${key}`,
          `${key === "timeout_seconds" ? "超时时长" : "数值"}必须为 ${min}–${max} 内的整数。`,
        );
    for (const key of [
      "random_chambers",
      "random_bullets",
      "random_mute",
      "timeout_mute",
      "end_when_all_loaded",
    ] as const)
      if (typeof r[key] !== "boolean")
        add(`${prefix}rules.${key}`, "开关值必须为布尔值。");
    if (r.random_chambers && r.chambers_max < r.chambers)
      add(`${prefix}rules.chambers_max`, "弹膛上限不能小于下限。");
    if (r.random_bullets && r.bullets_max < r.bullets)
      add(`${prefix}rules.bullets_max`, "子弹上限不能小于下限。");
    if (r.random_mute && r.mute_max_seconds < r.mute_seconds)
      add(`${prefix}rules.mute_max_seconds`, "禁言上限不能小于下限。");
    if ((r.random_bullets ? r.bullets_max : r.bullets) >= r.chambers)
      add(
        `${prefix}rules.bullets`,
        Number.isInteger(r.chambers) && r.chambers > 1
          ? `最大子弹数必须小于最小弹膛数：最少 ${r.chambers} 个弹膛时，子弹最多 ${r.chambers - 1} 发。`
          : "最大子弹数必须小于最小弹膛数。",
      );
    for (const key of replyKeys) {
      const text = replies[key];
      if (typeof text !== "string" || !text.trim() || [...text].length > 1000) {
        add(`${prefix}replies.${key}`, "台词需要 1–1000 字。");
        continue;
      }
      for (const token of text.match(/<[^<>\s]+>/g) ?? [])
        if (!(token in tokenLabels))
          add(`${prefix}replies.${key}`, `未知占位符 ${token}`);
    }
  };
  check(s.rules, s.replies, "");
  if (s.group_overrides.length > 500)
    add("group_overrides", "最多设置 500 个群覆盖。");
  const seen = new Set<string>();
  s.group_overrides.forEach((g, i) => {
    const prefix = `group_overrides.${i}.`;
    if (
      !g.source_adapter.trim() ||
      g.source_adapter.length > 128 ||
      !/^[1-9][0-9]{0,19}$/.test(g.bot_id) ||
      !/^[1-9][0-9]{0,19}$/.test(g.group_id)
    )
      add(prefix.slice(0, -1), "请选择有效的机器人账号和群。");
    if ([...(g.group_name ?? "")].length > 100)
      add(`${prefix}group_name`, "群名不能超过 100 字。");
    if (seen.has(groupKey(g)))
      add(prefix.slice(0, -1), "同一机器人账号和群不能重复覆盖。");
    seen.add(groupKey(g));
    const e = effective(s, g);
    check(e.rules, e.replies, prefix);
  });
  return issues;
}

export function duration(seconds: number): string {
  let rest = seconds,
    result = "";
  for (const [size, label] of [
    [86400, "天"],
    [3600, "小时"],
    [60, "分"],
    [1, "秒"],
  ] as const)
    if (rest >= size) {
      result += `${Math.floor(rest / size)}${label}`;
      rest %= size;
    }
  return result || "0秒";
}
export function preview(text: string, r: Rules): string {
  const values: Record<string, string> = {
    "<bullet>": String(r.bullets),
    "<chamber>": String(r.chambers),
    "<remain-bullet>": String(r.bullets),
    "<remain-chamber>": String(Math.max(0, r.chambers - 1)),
    "<mute-s>": String(r.mute_seconds),
    "<mute-f>": duration(r.mute_seconds),
    "<timeout-s>": String(r.timeout_seconds),
    "<timeout-f>": duration(r.timeout_seconds),
    "<target>": "@参与者",
  };
  return typeof text === "string"
    ? text.replace(/<[^<>\s]+>/g, (token) => values[token] ?? token)
    : "";
}
