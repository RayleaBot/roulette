import {
  duration,
  groupKey,
  replyKeys,
  replyLabels,
  type GroupOverride,
  type Replies,
  type RuleKey,
  type Rules,
  type Settings,
} from "./model";

export type Topic =
  | "chambers"
  | "bullets"
  | "mute"
  | "timeout"
  | "triggers"
  | "loaded"
  | "timeoutMute";
export type UnitId =
  | "chambers"
  | "bullets"
  | "mute"
  | "timeout"
  | "timeoutMute"
  | "endAllLoaded";

export interface RangeUnit {
  id: "chambers" | "bullets" | "mute";
  title: string;
  random: RuleKey;
  min: RuleKey;
  max: RuleKey;
  low: number;
  high: number;
  unit: string;
}

export const rangeUnits: readonly RangeUnit[] = [
  {
    id: "chambers",
    title: "弹膛",
    random: "random_chambers",
    min: "chambers",
    max: "chambers_max",
    low: 2,
    high: 100,
    unit: "个",
  },
  {
    id: "bullets",
    title: "子弹",
    random: "random_bullets",
    min: "bullets",
    max: "bullets_max",
    low: 1,
    high: 99,
    unit: "发",
  },
  {
    id: "mute",
    title: "禁言",
    random: "random_mute",
    min: "mute_seconds",
    max: "mute_max_seconds",
    low: 1,
    high: 2592000,
    unit: "秒",
  },
];

// Group overrides inherit per rule unit; a unit counts as overridden when any of
// its keys is present, so older single-field overrides stay visible and intact.
// Membership uses `in` because Vue tracks it on reactive drafts, unlike Object.hasOwn.
export const unitKeys: Record<UnitId, RuleKey[]> = {
  chambers: ["random_chambers", "chambers", "chambers_max"],
  bullets: ["random_bullets", "bullets", "bullets_max"],
  mute: ["random_mute", "mute_seconds", "mute_max_seconds"],
  timeout: ["timeout_seconds"],
  timeoutMute: ["timeout_mute"],
  endAllLoaded: ["end_when_all_loaded"],
};

export const unitTitles: Record<UnitId, string> = {
  chambers: "弹膛",
  bullets: "子弹",
  mute: "禁言",
  timeout: "时限",
  timeoutMute: "超时禁言发起人",
  endAllLoaded: "全实弹提前结束",
};

const unitOf = Object.fromEntries(
  (Object.entries(unitKeys) as [UnitId, RuleKey[]][]).flatMap(([unit, keys]) =>
    keys.map((key) => [key, unit]),
  ),
) as Record<RuleKey, UnitId>;

export function isOverridden(
  overrides: Partial<Rules> | undefined,
  unit: UnitId,
): boolean {
  return !!overrides && unitKeys[unit].some((k) => k in overrides);
}

const amount = (u: RangeUnit, v: unknown) =>
  typeof v !== "number" || !Number.isFinite(v)
    ? "—"
    : u.id === "mute"
      ? duration(v)
      : `${v} ${u.unit}`;

export function describeRange(u: RangeUnit, r: Rules): string {
  return r[u.random]
    ? `随机 ${amount(u, r[u.min])} – ${amount(u, r[u.max])}`
    : `固定 ${amount(u, r[u.min])}`;
}

export function describeUnit(unit: UnitId, r: Rules): string {
  const range = rangeUnits.find((u) => u.id === unit);
  if (range) return describeRange(range, r);
  if (unit === "timeout") return `${r.timeout_seconds} 秒`;
  if (unit === "timeoutMute") return r.timeout_mute ? "开启" : "关闭";
  return r.end_when_all_loaded ? "开启" : "关闭";
}

export function overriddenUnits(g: GroupOverride): UnitId[] {
  return (Object.keys(unitKeys) as UnitId[]).filter((u) =>
    isOverridden(g.rules, u),
  );
}

export function overrideSummary(g: GroupOverride): string {
  const parts: string[] = overriddenUnits(g).map((u) => unitTitles[u]);
  const replies = replyKeys.filter((k) => k in g.replies);
  if (replies.length === 1) parts.push(`${replyLabels[replies[0]]}台词`);
  else if (replies.length > 1) parts.push(`${replies.length} 条台词`);
  return parts.join("、");
}

export function overrideCount(g: GroupOverride): number {
  return (
    overriddenUnits(g).length +
    replyKeys.filter((k) => k in g.replies).length
  );
}

const numeric = (v: unknown) => typeof v === "number" && Number.isFinite(v);

// First-shot hit probability across the configured ranges, or null while the
// counts are not valid numbers.
export function firstShotOdds(r: Rules): [number, number] | null {
  const cMin = r.chambers,
    cMax = r.random_chambers ? r.chambers_max : r.chambers;
  const bMin = r.bullets,
    bMax = r.random_bullets ? r.bullets_max : r.bullets;
  if (![cMin, cMax, bMin, bMax].every(numeric) || cMin < 1 || cMax < 1)
    return null;
  return [Math.min(bMin / cMax, 1), Math.min(bMax / cMin, 1)];
}

export const percent = (v: number) =>
  `${(v * 100).toFixed(1).replace(/\.0$/, "")}%`;

export const groupTitle = (g: GroupOverride) =>
  g.group_name || `群 ${g.group_id}`;

// Readable list of what the draft changes relative to the last saved values;
// default rule units carry their before and after values.
export function changes(draft: Settings, saved: Settings | null): string[] {
  if (!saved) return [];
  const out: string[] = [];
  const same = (a: unknown, b: unknown) =>
    JSON.stringify(a) === JSON.stringify(b);
  const units = (a: Partial<Rules>, b: Partial<Rules>) => [
    ...new Set(
      (Object.keys({ ...a, ...b }) as RuleKey[])
        .filter((k) => !same(a[k], b[k]))
        .map((k) => unitOf[k]),
    ),
  ];
  const replies = (a: Partial<Replies>, b: Partial<Replies>) =>
    replyKeys.filter((k) => !same(a[k], b[k]));
  if (!same(draft.trigger_commands, saved.trigger_commands))
    out.push("开局口令");
  for (const unit of units(draft.rules, saved.rules))
    out.push(
      `${unitTitles[unit]}：${describeUnit(unit, saved.rules)} → ${describeUnit(unit, draft.rules)}`,
    );
  for (const key of replies(draft.replies, saved.replies))
    out.push(`默认${replyLabels[key]}台词`);
  const before = new Map(saved.group_overrides.map((g) => [groupKey(g), g]));
  const after = new Set<string>();
  for (const g of draft.group_overrides) {
    const key = groupKey(g);
    after.add(key);
    const old = before.get(key);
    if (!old) {
      out.push(`新增 ${groupTitle(g)}`);
      continue;
    }
    const parts = [
      ...units(g.rules, old.rules).map((u) => unitTitles[u]),
      ...replies(g.replies, old.replies).map((k) => `${replyLabels[k]}台词`),
    ];
    if (parts.length) out.push(`${groupTitle(g)}：${parts.join("、")}`);
  }
  for (const [key, g] of before)
    if (!after.has(key)) out.push(`移除 ${groupTitle(g)}`);
  return out;
}
