import { describe, expect, it } from "vitest";
import {
  calculateRiskScore,
  findInconsistentThresholds,
  findLowAvailableReserve,
  findPreauthHighWeight,
  findRevocableIssuer,
  findWeakenedMasterKey,
  SEVERITY_WEIGHTS,
} from "@/lib/riskScore";
import {
  BACKUP_KEY,
  DISABLED_MASTER_TESTNET_ACCOUNT,
  HEALTHY_TESTNET_ACCOUNT,
  INCONSISTENT_THRESHOLDS_TESTNET_ACCOUNT,
  ISSUER_KEY,
  makeTestnetAccount,
  MASTER_KEY,
} from "@/tests/fixtures/accounts";

describe("Stellar Risk Score rules", () => {
  it("reports a healthy Testnet fixture as low risk", () => {
    expect(calculateRiskScore(HEALTHY_TESTNET_ACCOUNT)).toEqual({
      score: 0,
      level: "bajo",
      findings: [],
    });
  });

  it("detects inconsistent Testnet thresholds", () => {
    const finding = findInconsistentThresholds(INCONSISTENT_THRESHOLDS_TESTNET_ACCOUNT);
    expect(finding?.id).toBe("INCONSISTENT_THRESHOLDS");
    expect(finding?.severity).toBe("high");
  });

  it("treats zero thresholds as an explicit heuristic signal", () => {
    const account = makeTestnetAccount({
      thresholds: { low_threshold: 0, med_threshold: 1, high_threshold: 2 },
    });
    expect(findInconsistentThresholds(account)?.description).toContain("configurado(s) en 0");
  });

  it("detects a disabled master key without enough backup weight", () => {
    const finding = findWeakenedMasterKey(DISABLED_MASTER_TESTNET_ACCOUNT);
    expect(finding?.id).toBe("MASTER_KEY_NO_BACKUP");
    expect(finding?.severity).toBe("critical");
  });

  it("does not flag a disabled master key when backups reach the medium threshold", () => {
    const account = makeTestnetAccount({
      thresholds: { low_threshold: 1, med_threshold: 2, high_threshold: 2 },
      signers: [
        { key: MASTER_KEY, weight: 0, type: "ed25519_public_key" },
        { key: BACKUP_KEY, weight: 2, type: "ed25519_public_key" },
      ],
    });
    expect(findWeakenedMasterKey(account)).toBeNull();
  });

  it("detects a high-weight preauthorized transaction signer", () => {
    const account = makeTestnetAccount({
      thresholds: { low_threshold: 1, med_threshold: 2, high_threshold: 3 },
      signers: [
        { key: MASTER_KEY, weight: 1, type: "ed25519_public_key" },
        { key: "A_PREAUTH_HASH", weight: 3, type: "sha256_hash" },
      ],
    });
    expect(findPreauthHighWeight(account)?.id).toBe("SIGNER_PREAUTH_HIGH_WEIGHT");
  });

  it("detects a trustline whose issuer can revoke authorization", () => {
    const account = makeTestnetAccount({
      trustlines: [
        {
          assetType: "credit_alphanum4",
          assetCode: "USDC",
          assetIssuer: ISSUER_KEY,
          balance: 12,
          limit: 1000,
          authorized: true,
          authorizedToMaintainLiabilities: false,
          clawbackEnabled: false,
          issuerFlags: {
            auth_required: true,
            auth_revocable: true,
            auth_immutable: false,
            auth_clawback_enabled: false,
          },
          issuerLookupStatus: "loaded",
        },
      ],
    });
    expect(findRevocableIssuer(account)?.id).toBe("REVOCABLE_ASSET_ISSUER");
  });

  it("detects when less than ten percent of native balance remains available", () => {
    const account = makeTestnetAccount({
      nativeBalance: 10,
      reserveBalance: 9.5,
      availableBalance: 0.5,
    });
    expect(findLowAvailableReserve(account)?.id).toBe("LOW_AVAILABLE_BALANCE");
  });

  it("adds severity weights, derives levels, and caps the final score at 100", () => {
    const account = makeTestnetAccount({
      nativeBalance: 10,
      reserveBalance: 9.5,
      availableBalance: 0.5,
      thresholds: { low_threshold: 3, med_threshold: 2, high_threshold: 1 },
      signers: [
        { key: MASTER_KEY, weight: 0, type: "ed25519_public_key" },
        { key: "A_PREAUTH_HASH", weight: 1, type: "sha256_hash" },
      ],
      trustlines: [
        {
          assetType: "credit_alphanum4",
          assetCode: "RISK",
          assetIssuer: ISSUER_KEY,
          balance: 1,
          limit: 10,
          authorized: true,
          authorizedToMaintainLiabilities: false,
          clawbackEnabled: false,
          issuerFlags: {
            auth_required: false,
            auth_revocable: true,
            auth_immutable: false,
            auth_clawback_enabled: false,
          },
          issuerLookupStatus: "loaded",
        },
      ],
    });

    const report = calculateRiskScore(account);
    expect(report.score).toBe(100);
    expect(report.level).toBe("crítico");
    expect(report.findings).toHaveLength(5);
    expect(SEVERITY_WEIGHTS).toEqual({ critical: 40, high: 25, medium: 10, low: 5 });
  });
});
