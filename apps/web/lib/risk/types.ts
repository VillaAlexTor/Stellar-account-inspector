import type { StellarAccountData } from "@/lib/stellar";

export type RiskSeverity = "low" | "medium" | "high" | "critical";
export type RiskLevel = "bajo" | "medio" | "alto" | "crítico";

export interface RiskFinding {
  id: string;
  severity: RiskSeverity;
  title: string;
  description: string;
}

export interface RiskReport {
  score: number;
  level: RiskLevel;
  findings: RiskFinding[];
}

export type RiskRule = (account: StellarAccountData) => RiskFinding | null;
