import type { Adapter } from "./model";

// QQ avatars load straight from the QQ avatar hosts; the host page CSP allows
// HTTPS images. A failed or missing image falls back to the initial.
export const userAvatar = (id: string) =>
  `https://q1.qlogo.cn/g?b=qq&nk=${encodeURIComponent(id)}&s=100`;

export const groupAvatar = (id: string) =>
  `https://p.qlogo.cn/gh/${encodeURIComponent(id)}/${encodeURIComponent(id)}/100`;

export const botAvatar = (adapter: Adapter) =>
  adapter.identity
    ? adapter.identity.avatar_url || userAvatar(adapter.identity.id)
    : undefined;

export const botName = (adapter: Adapter) =>
  adapter.identity?.name || adapter.display_name;

const states: Record<string, string> = {
  idle: "未启动",
  listening: "等待连接",
  connecting: "连接中",
  connected: "已连接",
  auth_failed: "鉴权失败",
  reconnecting: "重连中",
  stopped: "已停止",
};

export function botState(adapter: Adapter): {
  label: string;
  tone: "ok" | "warn" | "error";
} {
  if (!adapter.enabled) return { label: "已停用", tone: "error" };
  if (!adapter.identity) return { label: "身份未确认", tone: "warn" };
  const label = states[adapter.state] ?? adapter.state;
  if (adapter.state === "connected") return { label, tone: "ok" };
  if (adapter.state === "auth_failed" || adapter.state === "stopped")
    return { label, tone: "error" };
  return { label, tone: "warn" };
}

export const usableBot = (adapter: Adapter) =>
  adapter.enabled && !!adapter.identity;
