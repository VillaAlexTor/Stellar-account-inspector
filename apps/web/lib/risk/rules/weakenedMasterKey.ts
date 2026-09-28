import type { RiskRule } from "@/lib/risk/types";

export const findWeakenedMasterKey: RiskRule = (account) => {
  const masterSigner = account.signers.find((signer) => signer.key === account.accountId);
  if ((masterSigner?.weight ?? 0) !== 0) return null;

  const backupWeight = account.signers
    .filter((signer) => signer.key !== account.accountId)
    .reduce((total, signer) => total + signer.weight, 0);
  const mediumThreshold = account.thresholds.med_threshold;

  if (backupWeight >= mediumThreshold) return null;

  return {
    id: "MASTER_KEY_NO_BACKUP",
    severity: "critical",
    title: "Master key deshabilitada sin respaldo suficiente",
    description: `La master key tiene peso 0 y los demás firmantes suman ${backupWeight}, por debajo del umbral medio (${mediumThreshold}). La cuenta puede quedar bloqueada para operaciones que requieren ese nivel de autorización.`,
  };
};
