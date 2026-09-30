import type { RiskRule } from "@/lib/risk/types";

export const findHashLockHighWeight: RiskRule = (account) => {
  const highThreshold = account.thresholds.high_threshold;
  if (highThreshold <= 0) return null;

  const riskySigners = account.signers.filter(
    (signer) => signer.type === "sha256_hash" && signer.weight >= highThreshold,
  );

  if (riskySigners.length === 0) return null;

  const weights = riskySigners.map((signer) => signer.weight).join(", ");
  return {
    id: "SIGNER_HASH_LOCK_HIGH_WEIGHT",
    severity: "high",
    title: "Hash-lock con control alto",
    description: `Se detectaron ${riskySigners.length} firmante(s) sha256_hash con peso ${weights}, suficiente para alcanzar el umbral alto (${highThreshold}). Este mecanismo autoriza a quien revele una preimagen cuyo SHA-256 coincida con el hash configurado; no representa una transacción preautorizada.`,
  };
};
