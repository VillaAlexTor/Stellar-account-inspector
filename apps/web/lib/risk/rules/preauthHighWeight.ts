import type { RiskRule } from "@/lib/risk/types";

export const findPreauthHighWeight: RiskRule = (account) => {
  const highThreshold = account.thresholds.high_threshold;
  if (highThreshold <= 0) return null;

  const riskySigners = account.signers.filter(
    (signer) => signer.type === "preauth_tx" && signer.weight >= highThreshold,
  );

  if (riskySigners.length === 0) return null;

  const weights = riskySigners.map((signer) => signer.weight).join(", ");
  return {
    id: "SIGNER_PREAUTH_HIGH_WEIGHT",
    severity: "high",
    title: "Transacción preautorizada con control alto",
    description: `Se detectaron ${riskySigners.length} firmante(s) preauth_tx con peso ${weights}, suficiente para alcanzar el umbral alto (${highThreshold}). Cada firmante autoriza una transacción concreta por su hash y puede ejecutarla cuando se publique, por lo que conviene verificar su propósito y vigencia.`,
  };
};
