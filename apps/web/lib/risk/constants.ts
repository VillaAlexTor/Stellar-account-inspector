import type { RiskLevel, RiskSeverity } from "@/lib/risk/types";

export const SEVERITY_WEIGHTS: Readonly<Record<RiskSeverity, number>> = {
  critical: 40,
  high: 25,
  medium: 10,
  low: 5,
};

export const RISK_LEVEL_THRESHOLDS = {
  critical: 40,
  high: 25,
  medium: 10,
} as const;

export function getRiskLevel(score: number): RiskLevel {
  if (score >= RISK_LEVEL_THRESHOLDS.critical) return "crítico";
  if (score >= RISK_LEVEL_THRESHOLDS.high) return "alto";
  if (score >= RISK_LEVEL_THRESHOLDS.medium) return "medio";
  return "bajo";
}
