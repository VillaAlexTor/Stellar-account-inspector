import type { StellarAccountData, StellarSigner } from "@/lib/stellar";

export const MASTER_KEY = `G${"A".repeat(55)}`;
export const BACKUP_KEY = `G${"B".repeat(55)}`;
export const ISSUER_KEY = `G${"C".repeat(55)}`;

const masterSigner: StellarSigner = {
  key: MASTER_KEY,
  weight: 1,
  type: "ed25519_public_key",
};

const baseAccount: StellarAccountData = {
  accountId: MASTER_KEY,
  network: "testnet",
  sequence: "100",
  subentryCount: 1,
  lastModifiedLedger: 12345,
  nativeBalance: 100,
  reserveBalance: 1.5,
  availableBalance: 98.5,
  thresholds: {
    low_threshold: 1,
    med_threshold: 1,
    high_threshold: 1,
  },
  flags: {
    auth_required: false,
    auth_revocable: false,
    auth_immutable: false,
    auth_clawback_enabled: false,
  },
  signers: [masterSigner],
  trustlines: [],
  isMultisig: false,
};

export function makeTestnetAccount(
  overrides: Partial<StellarAccountData> = {},
): StellarAccountData {
  return {
    ...baseAccount,
    ...overrides,
    thresholds: overrides.thresholds ?? { ...baseAccount.thresholds },
    flags: overrides.flags ?? { ...baseAccount.flags },
    signers: overrides.signers ?? baseAccount.signers.map((signer) => ({ ...signer })),
    trustlines: overrides.trustlines ?? [],
  };
}

export const HEALTHY_TESTNET_ACCOUNT = makeTestnetAccount({
  signers: [
    masterSigner,
    { key: BACKUP_KEY, weight: 1, type: "ed25519_public_key" },
  ],
  thresholds: { low_threshold: 1, med_threshold: 2, high_threshold: 2 },
  isMultisig: true,
});

export const INCONSISTENT_THRESHOLDS_TESTNET_ACCOUNT = makeTestnetAccount({
  thresholds: { low_threshold: 3, med_threshold: 2, high_threshold: 1 },
});

export const DISABLED_MASTER_TESTNET_ACCOUNT = makeTestnetAccount({
  signers: [
    { ...masterSigner, weight: 0 },
    { key: BACKUP_KEY, weight: 1, type: "ed25519_public_key" },
  ],
  thresholds: { low_threshold: 1, med_threshold: 2, high_threshold: 2 },
  isMultisig: false,
});
