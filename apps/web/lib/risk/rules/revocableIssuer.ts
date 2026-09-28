import type { RiskRule } from "@/lib/risk/types";

export const findRevocableIssuer: RiskRule = (account) => {
  const revocableTrustlines = account.trustlines.filter(
    (trustline) => trustline.issuerFlags?.auth_revocable === true,
  );

  if (revocableTrustlines.length === 0) return null;

  const assets = [...new Set(revocableTrustlines.map((trustline) => trustline.assetCode))];
  return {
    id: "REVOCABLE_ASSET_ISSUER",
    severity: "medium",
    title: "Activo emitido por una cuenta revocable",
    description: `Las trustlines ${assets.join(", ")} dependen de emisores con auth_revocable activo. Esos emisores pueden retirar la autorización y congelar la transferencia de los fondos en cualquier momento.`,
  };
};
