import { getRiskLevel, SEVERITY_WEIGHTS } from "@/lib/risk/constants";
import { findInconsistentThresholds } from "@/lib/risk/rules/inconsistentThresholds";
import { findLowAvailableReserve } from "@/lib/risk/rules/lowAvailableReserve";
import { findPreauthHighWeight } from "@/lib/risk/rules/preauthHighWeight";
import { findRevocableIssuer } from "@/lib/risk/rules/revocableIssuer";
import { findWeakenedMasterKey } from "@/lib/risk/rules/weakenedMasterKey";
import type { RiskFinding, RiskReport, RiskRule } from "@/lib/risk/types";
import type { StellarAccountData } from "@/lib/stellar";

export type { RiskFinding, RiskLevel, RiskReport, RiskSeverity } from "@/lib/risk/types";
export { RISK_LEVEL_THRESHOLDS, SEVERITY_WEIGHTS } from "@/lib/risk/constants";
export {
  findInconsistentThresholds,
  findLowAvailableReserve,
  findPreauthHighWeight,
  findRevocableIssuer,
  findWeakenedMasterKey,
};

export const RISK_RULES: readonly RiskRule[] = [
  findPreauthHighWeight,
  findInconsistentThresholds,
  findWeakenedMasterKey,
  findRevocableIssuer,
  findLowAvailableReserve,
];

export function calculateRiskScore(account: StellarAccountData): RiskReport {
  const findings = RISK_RULES.map((rule) => rule(account)).filter(
    (finding): finding is RiskFinding => finding !== null,
  );
  const score = Math.min(
    100,
    findings.reduce((total, finding) => total + SEVERITY_WEIGHTS[finding.severity], 0),
  );

  return {
    score,
    level: getRiskLevel(score),
    findings,
  };
}
