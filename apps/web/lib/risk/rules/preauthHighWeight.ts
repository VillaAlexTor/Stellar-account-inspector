import type { RiskRule } from "@/lib/risk/types";

export const findPreauthHighWeight: RiskRule = (account) => {
  const highThreshold = account.thresholds.high_threshold;
  if (highThreshold <= 0) return null;

  const riskySigners = account.signers.filter(
    (signer) => signer.type === "sha256_hash" && signer.weight >= highThreshold,
  );

  if (riskySigners.length === 0) return null;

  const weights = riskySigners.map((signer) => signer.weight).join(", ");
  return {
    id: "SIGNER_PREAUTH_HIGH_WEIGHT",
    severity: "high",
    title: "Transacción preautorizada con control alto",
    description: `Se detectaron ${riskySigners.length} firmante(s) sha256_hash con peso ${weights}, suficiente para alcanzar el umbral alto (${highThreshold}). Una transacción preautorizada puede ejecutar cambios sensibles cuando se publique y es inusual fuera de flujos multifirma planificados.`,
  };
};
