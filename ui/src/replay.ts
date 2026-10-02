import { duration, type Replies, type ReplyKey, type Rules } from "./model";
import type { Topic } from "./rules";

// A sample round rendered from the current rules and replies. Members, group
// names and drawn values are illustrative; nothing here reaches a real group.

export type Scenario = "hit" | "immune" | "loaded" | "timeout" | "stop";
export type TimeoutEnding = "timeout" | "timeout_hit" | "timeout_immune";

export interface Segment {
  text: string;
  topic?: Topic;
  mention?: boolean;
}
export interface Member {
  name: string;
  tone: number;
  admin?: boolean;
}
// A line-level topic lights the whole bubble or notice, for rules such as the
// end switches that change which message appears rather than a number in it.
export type Line =
  | { kind: "time"; segments: Segment[] }
  | { kind: "member"; member: Member; segments: Segment[] }
  | { kind: "bot"; key: ReplyKey; segments: Segment[]; topic?: Topic }
  | { kind: "notice"; segments: Segment[]; topic?: Topic }
  | { kind: "fold"; segments: Segment[] };

export const scenarios: { id: Scenario; label: string }[] = [
  { id: "hit", label: "命中禁言" },
  { id: "immune", label: "命中豁免" },
  { id: "loaded", label: "提前结束" },
  { id: "timeout", label: "超时" },
  { id: "stop", label: "管理员停止" },
];

export const endings: { id: TimeoutEnding; label: string; muted: boolean }[] = [
  { id: "timeout", label: "不惩罚", muted: false },
  { id: "timeout_hit", label: "禁言发起人", muted: true },
  { id: "timeout_immune", label: "发起人豁免", muted: true },
];

const lin: Member = { name: "小林", tone: 1 };
const jie: Member = { name: "阿杰", tone: 2 };
const wang: Member = { name: "老王", tone: 3 };
const kiki: Member = { name: "Kiki", tone: 4, admin: true };

// "middle" shows the midpoint of each random range; a reroll stores positions
// in [0, 1) so the drawn values follow later edits to the range.
export type Sample =
  | "middle"
  | { chambers: number; bullets: number; mute: number };
export const middleSample: Sample = "middle";

export interface Values {
  chambers: number;
  bullets: number;
  mute: number;
  timeout: number;
}

function draw(
  random: boolean,
  min: number,
  max: number,
  at: number | "middle",
  step = 1,
) {
  if (!random || !(max > min)) return min;
  const span = max - min;
  const value =
    at === "middle"
      ? min + Math.floor(span / 2)
      : min + Math.min(span, Math.floor(at * (span + 1)));
  // Mute samples land on whole tens of seconds so durations stay readable.
  if (step === 1 || span < step) return value;
  return Math.min(max, Math.max(min, Math.round(value / step) * step));
}

export function sampleValues(r: Rules, s: Sample): Values {
  const at = (key: "chambers" | "bullets" | "mute") =>
    s === "middle" ? "middle" : s[key];
  return {
    chambers: draw(r.random_chambers, r.chambers, r.chambers_max, at("chambers")),
    bullets: draw(r.random_bullets, r.bullets, r.bullets_max, at("bullets")),
    mute: draw(r.random_mute, r.mute_seconds, r.mute_max_seconds, at("mute"), 10),
    timeout: r.timeout_seconds,
  };
}

interface Context {
  target: string;
  chamber: number;
  bullet: number;
  remainChamber: number;
  remainBullet: number;
  mute: number;
  timeout: number;
}

function token(name: string, c: Context): Segment | undefined {
  switch (name) {
    case "<target>":
      return { text: `@${c.target}`, mention: true };
    case "<chamber>":
      return { text: String(c.chamber), topic: "chambers" };
    case "<remain-chamber>":
      return { text: String(c.remainChamber), topic: "chambers" };
    case "<bullet>":
      return { text: String(c.bullet), topic: "bullets" };
    case "<remain-bullet>":
      return { text: String(c.remainBullet), topic: "bullets" };
    case "<mute-s>":
      return { text: String(c.mute), topic: "mute" };
    case "<mute-f>":
      return { text: duration(c.mute), topic: "mute" };
    case "<timeout-s>":
      return { text: String(c.timeout), topic: "timeout" };
    case "<timeout-f>":
      return { text: duration(c.timeout), topic: "timeout" };
  }
}

// Unknown placeholders stay as typed so the validation message has context.
export function fill(text: string, c: Context): Segment[] {
  const out: Segment[] = [];
  let last = 0;
  for (const match of text.matchAll(/<[^<>\s]+>/g)) {
    const seg = token(match[0], c);
    if (!seg) continue;
    if (match.index > last) out.push({ text: text.slice(last, match.index) });
    out.push(seg);
    last = match.index + match[0].length;
  }
  if (last < text.length) out.push({ text: text.slice(last) });
  return out;
}

export interface ReplayInput {
  rules: Rules;
  replies: Replies;
  values: Values;
  trigger: string;
  ending: TimeoutEnding;
}

export function buildReplay(scenario: Scenario, input: ReplayInput): Line[] {
  const { rules, replies, values } = input;
  const c = values.chambers,
    b = values.bullets;
  const lines: Line[] = [{ kind: "time", segments: [{ text: "21:04" }] }];
  const command = (text: string): Segment[] => [{ text, topic: "triggers" }];
  const say = (m: Member, text = `/${input.trigger}`) =>
    lines.push({ kind: "member", member: m, segments: command(text) });
  const bot = (
    key: ReplyKey,
    target: Member,
    rc: number,
    rb: number,
    topic?: Topic,
  ) =>
    lines.push({
      kind: "bot",
      key,
      topic,
      segments: fill(replies[key], {
        target: target.name,
        chamber: c,
        bullet: b,
        remainChamber: rc,
        remainBullet: rb,
        mute: values.mute,
        timeout: values.timeout,
      }),
    });
  const muted = (m: Member) =>
    lines.push({
      kind: "notice",
      segments: [
        { text: `${m.name} 被禁言 ` },
        { text: duration(values.mute), topic: "mute" },
      ],
    });
  const fold = (segments: Segment[]) => lines.push({ kind: "fold", segments });
  const finishRemaining = (last: Member, rc: number, rb: number) => {
    if (rb > 0)
      fold([
        { text: "之后连续 " },
        { text: String(rb), topic: "bullets" },
        { text: " 枪命中，" },
        { text: String(rb), topic: "bullets" },
        { text: " 人被禁言" },
      ]);
    bot("end", last, rc - rb, 0);
  };

  if (scenario === "hit" || scenario === "immune") {
    // Keep at least one empty chamber after the misses so the round neither
    // ends early nor runs out before the first hit.
    const misses = Math.min(2, c - b - 1);
    const order =
      scenario === "immune"
        ? [lin, jie].slice(0, misses).concat(kiki)
        : [lin, jie, wang];
    say(order[0]);
    bot("start", order[0], c, b);
    let rc = c;
    for (let i = 0; i < misses; i++) {
      if (i > 0) say(order[i]);
      rc--;
      bot("miss", order[i], rc, b);
    }
    const hitter = order[misses];
    if (misses > 0) say(hitter);
    rc--;
    bot(scenario === "immune" ? "immune" : "hit", hitter, rc, b - 1);
    if (scenario === "hit") muted(hitter);
    finishRemaining(hitter, rc, b - 1);
    return lines;
  }

  say(lin);
  bot("start", lin, c, b);

  if (scenario === "loaded") {
    // Live rounds sit in the last chambers, so every shot before them misses.
    const misses = c - b;
    let last = lin;
    bot("miss", lin, c - 1, b);
    if (misses >= 3)
      fold([
        { text: "又空了 " },
        { text: String(misses - 2), topic: "chambers" },
        { text: " 枪" },
      ]);
    if (misses >= 2) {
      say(jie);
      bot("miss", jie, b, b);
      last = jie;
    }
    if (rules.end_when_all_loaded) {
      bot("all_loaded", last, b, b, "loaded");
      return lines;
    }
    lines.push({
      kind: "notice",
      topic: "loaded",
      segments: [{ text: `剩余 ${b} 个弹膛全是实弹；当前设置不提前结束` }],
    });
    say(wang);
    bot("hit", wang, b - 1, b - 1);
    muted(wang);
    finishRemaining(wang, b - 1, b - 1);
    return lines;
  }

  const rc = c - 1;
  bot("miss", lin, rc, b);
  if (rules.end_when_all_loaded && rc === b) {
    bot("all_loaded", lin, rc, b, "loaded");
    lines.push({
      kind: "notice",
      topic: "loaded",
      segments: [{ text: "按当前设置，第一枪后就提前结束" }],
    });
    return lines;
  }
  if (scenario === "stop") {
    say(kiki, "/停止轮盘");
    bot("stopped", kiki, rc, b);
    return lines;
  }
  lines.push({
    kind: "time",
    segments: [
      { text: "开局 " },
      { text: duration(values.timeout), topic: "timeout" },
      { text: " 后，没有人再开枪" },
    ],
  });
  bot(input.ending, lin, rc, b, "timeoutMute");
  if (input.ending === "timeout_hit") muted(lin);
  return lines;
}
