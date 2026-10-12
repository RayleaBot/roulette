// QQ avatars load straight from the QQ avatar hosts; the host page CSP allows
// HTTPS images. A failed or missing image falls back to the initial.
export const userAvatar = (id: string) =>
  `https://q1.qlogo.cn/g?b=qq&nk=${encodeURIComponent(id)}&s=100`;

export const groupAvatar = (id: string) =>
  `https://p.qlogo.cn/gh/${encodeURIComponent(id)}/${encodeURIComponent(id)}/100`;
