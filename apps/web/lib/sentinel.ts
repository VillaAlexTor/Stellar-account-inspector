import type { StellarNetwork } from "@/lib/stellar";

export type SentinelConnectionState = "idle" | "connecting" | "connected" | "reconnecting" | "down";
export type SentinelSeverity = "info" | "warning" | "critical";

export interface SentinelAlert {
  id: number;
  monitoredAccountId: number;
  ruleId: string;
  severity: SentinelSeverity;
  message: string;
  operationId: string;
  createdAt: string;
}

export interface SentinelStatus {
  state: Exclude<SentinelConnectionState, "idle">;
  detail?: string;
  at: string;
}

export interface MonitoredAccount {
  id: number;
  publicKey: string;
  network: StellarNetwork;
  lastCursor: string;
  createdAt: string;
  updatedAt: string;
}

export const SENTINEL_API_URL = (
  process.env.NEXT_PUBLIC_SENTINEL_API_URL ?? "http://localhost:8081"
).replace(/\/$/, "");

export function sentinelAccountPath(publicKey: string): string {
  return `${SENTINEL_API_URL}/api/v1/monitored-accounts/${encodeURIComponent(publicKey)}`;
}
