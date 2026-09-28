import type { RiskRule } from "@/lib/risk/types";

export const findInconsistentThresholds: RiskRule = (account) => {
  const { low_threshold: low, med_threshold: medium, high_threshold: high } = account.thresholds;
  const problems: string[] = [];

  if (high < medium) problems.push(`alto (${high}) es menor que medio (${medium})`);
  if (medium < low) problems.push(`medio (${medium}) es menor que bajo (${low})`);

  const zeroThresholds = [
    ["bajo", low],
    ["medio", medium],
    ["alto", high],
  ].filter(([, value]) => value === 0);

  if (zeroThresholds.length > 0) {
    problems.push(`${zeroThresholds.map(([name]) => name).join(", ")} configurado(s) en 0`);
  }

  if (problems.length === 0) return null;

  return {
    id: "INCONSISTENT_THRESHOLDS",
    severity: "high",
    title: "Umbrales de autorización inconsistentes",
    description: `La configuración presenta: ${problems.join("; ")}. Horizon no registra la intención del propietario, por lo que un valor cero se trata como señal heurística y debe revisarse manualmente.`,
  };
};
