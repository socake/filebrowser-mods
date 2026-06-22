// Share-related time helpers.
//
// dayjs plugins (relativeTime / localizedFormat) are registered globally in
// main.ts; importing the dayjs singleton here reuses that configuration.

import dayjs from "dayjs";

/** True when a share has a non-zero expire timestamp (unix seconds) in the past. */
export function isShareExpired(expire?: number): boolean {
  return !!expire && expire !== 0 && expire * 1000 < Date.now();
}

/** Locale date-time string for a unix-seconds timestamp (pure formatting). */
export function formatTimestamp(ts: number): string {
  return new Date(ts * 1000).toLocaleString();
}

/** Shares-table expire column text: 永久 / 已过期 / relative time. */
export function formatShareExpire(expire?: number): string {
  if (!expire || expire === 0) return "永久";
  if (isShareExpired(expire)) return "已过期";
  return dayjs(expire * 1000).fromNow();
}

/** Shares-table created column text: 未知 / formatted date. */
export function formatCreatedTime(createdAt?: number): string {
  if (!createdAt) return "未知";
  return dayjs(createdAt * 1000).format("YYYY-MM-DD HH:mm");
}
