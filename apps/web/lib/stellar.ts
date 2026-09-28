export const STELLAR_PUBLIC_KEY_PATTERN = /^G[A-Z2-7]{55}$/;

export const HORIZON_TESTNET_URL = "https://horizon-testnet.stellar.org";

export type StellarNetwork = "testnet";

export interface StellarThresholds {
  low_threshold: number;
  med_threshold: number;
  high_threshold: number;
}

export interface StellarFlags {
  auth_required: boolean;
  auth_revocable: boolean;
  auth_immutable: boolean;
  auth_clawback_enabled: boolean;
}

export interface StellarSigner {
  key: string;
  weight: number;
  type: string;
  sponsor?: string;
}

export interface StellarTrustline {
  assetType: string;
  assetCode: string;
  assetIssuer: string;
  balance: number;
  limit: number;
  authorized: boolean;
  authorizedToMaintainLiabilities: boolean;
  clawbackEnabled: boolean;
  issuerFlags?: StellarFlags;
  issuerLookupStatus: "loaded" | "unavailable";
}

export interface StellarAccountData {
  accountId: string;
  network: StellarNetwork;
  sequence: string;
  subentryCount: number;
  lastModifiedLedger: number;
  lastModifiedTime?: string;
  nativeBalance: number;
  reserveBalance: number;
  availableBalance: number;
  thresholds: StellarThresholds;
  flags: StellarFlags;
  signers: StellarSigner[];
  trustlines: StellarTrustline[];
  isMultisig: boolean;
}

interface HorizonBalance {
  asset_type: string;
  asset_code?: string;
  asset_issuer?: string;
  balance: string;
  limit?: string;
  is_authorized?: boolean;
  is_authorized_to_maintain_liabilities?: boolean;
  is_clawback_enabled?: boolean;
}

interface HorizonAccount {
  account_id: string;
  sequence: string;
  subentry_count: number;
  last_modified_ledger: number;
  last_modified_time?: string;
  thresholds: StellarThresholds;
  flags: StellarFlags;
  balances: HorizonBalance[];
  signers: StellarSigner[];
}

const ISSUER_LOOKUP_CONCURRENCY = 6;

export class StellarAccountError extends Error {
  constructor(
    message: string,
    public readonly code: "INVALID_KEY" | "NOT_FOUND" | "HORIZON_ERROR" | "NETWORK_ERROR",
  ) {
    super(message);
    this.name = "StellarAccountError";
  }
}

export function isValidStellarPublicKey(value: string): boolean {
  return STELLAR_PUBLIC_KEY_PATTERN.test(value.trim().toUpperCase());
}

export function getLockedReserve(subentryCount: number): number {
  return (2 + subentryCount) * 0.5;
}

export function getAvailableBalance(nativeBalance: number, subentryCount: number): number {
  return Math.max(0, nativeBalance - getLockedReserve(subentryCount));
}

export function hasRealMultisig(signers: StellarSigner[]): boolean {
  return signers.filter((signer) => signer.weight > 0).length > 1;
}

function normalizeFlags(flags: Partial<StellarFlags> | undefined): StellarFlags {
  return {
    auth_required: flags?.auth_required ?? false,
    auth_revocable: flags?.auth_revocable ?? false,
    auth_immutable: flags?.auth_immutable ?? false,
    auth_clawback_enabled: flags?.auth_clawback_enabled ?? false,
  };
}

async function fetchIssuerFlags(
  publicKey: string,
  signal?: AbortSignal,
): Promise<StellarFlags | undefined> {
  try {
    const response = await fetch(`${HORIZON_TESTNET_URL}/accounts/${publicKey}`, {
      headers: { Accept: "application/json" },
      signal,
    });
    if (!response.ok) return undefined;
    const issuer = (await response.json()) as Pick<HorizonAccount, "flags">;
    return normalizeFlags(issuer.flags);
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") throw error;
    return undefined;
  }
}

async function fetchIssuerFlagsMap(
  issuers: string[],
  accountId: string,
  accountFlags: StellarFlags,
  signal?: AbortSignal,
): Promise<Map<string, StellarFlags | undefined>> {
  const flagsByIssuer = new Map<string, StellarFlags | undefined>();
  const uniqueIssuers = [...new Set(issuers)];

  for (let index = 0; index < uniqueIssuers.length; index += ISSUER_LOOKUP_CONCURRENCY) {
    const batch = uniqueIssuers.slice(index, index + ISSUER_LOOKUP_CONCURRENCY);
    const results = await Promise.all(
      batch.map(async (issuer) => {
        if (issuer === accountId) return [issuer, accountFlags] as const;
        return [issuer, await fetchIssuerFlags(issuer, signal)] as const;
      }),
    );
    results.forEach(([issuer, flags]) => flagsByIssuer.set(issuer, flags));
  }

  return flagsByIssuer;
}

export async function fetchStellarAccount(
  publicKey: string,
  signal?: AbortSignal,
): Promise<StellarAccountData> {
  const normalizedKey = publicKey.trim().toUpperCase();

  if (!isValidStellarPublicKey(normalizedKey)) {
    throw new StellarAccountError(
      "La clave debe comenzar con G y contener 56 caracteres válidos de Stellar.",
      "INVALID_KEY",
    );
  }

  let response: Response;
  try {
    response = await fetch(`${HORIZON_TESTNET_URL}/accounts/${normalizedKey}`, {
      headers: { Accept: "application/json" },
      signal,
    });
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") throw error;
    throw new StellarAccountError(
      "No fue posible conectar con Horizon. Revisa tu conexión e intenta otra vez.",
      "NETWORK_ERROR",
    );
  }

  if (response.status === 404) {
    throw new StellarAccountError(
      "La cuenta no existe en Stellar Testnet.",
      "NOT_FOUND",
    );
  }

  if (!response.ok) {
    throw new StellarAccountError(
      `Horizon respondió con el estado ${response.status}. Intenta nuevamente en unos segundos.`,
      "HORIZON_ERROR",
    );
  }

  const account = (await response.json()) as HorizonAccount;
  const accountFlags = normalizeFlags(account.flags);
  const nativeBalance = Number(
    account.balances.find((balance) => balance.asset_type === "native")?.balance ?? 0,
  );
  const reserveBalance = getLockedReserve(account.subentry_count);
  const creditBalances = account.balances.filter((balance) => balance.asset_type !== "native");
  const issuerFlags = await fetchIssuerFlagsMap(
    creditBalances.flatMap((balance) => (balance.asset_issuer ? [balance.asset_issuer] : [])),
    account.account_id,
    accountFlags,
    signal,
  );

  return {
    accountId: account.account_id,
    network: "testnet",
    sequence: account.sequence,
    subentryCount: account.subentry_count,
    lastModifiedLedger: account.last_modified_ledger,
    lastModifiedTime: account.last_modified_time,
    nativeBalance,
    reserveBalance,
    availableBalance: getAvailableBalance(nativeBalance, account.subentry_count),
    thresholds: account.thresholds,
    flags: accountFlags,
    signers: account.signers,
    trustlines: creditBalances.map((balance) => {
      const flags = balance.asset_issuer ? issuerFlags.get(balance.asset_issuer) : undefined;
      return {
        assetType: balance.asset_type,
        assetCode: balance.asset_code ?? "—",
        assetIssuer: balance.asset_issuer ?? "—",
        balance: Number(balance.balance),
        limit: Number(balance.limit ?? 0),
        authorized: balance.is_authorized ?? false,
        authorizedToMaintainLiabilities:
          balance.is_authorized_to_maintain_liabilities ?? false,
        clawbackEnabled: balance.is_clawback_enabled ?? false,
        issuerFlags: flags,
        issuerLookupStatus: flags ? "loaded" : "unavailable",
      };
    }),
    isMultisig: hasRealMultisig(account.signers),
  };
}

export function formatXlm(value: number): string {
  return new Intl.NumberFormat("es-BO", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 7,
  }).format(value);
}

export function getStellarExpertUrl(accountId: string): string {
  return `https://stellar.expert/explorer/testnet/account/${accountId}`;
}
