import { afterEach, describe, expect, it, vi } from "vitest";
import {
  fetchStellarAccount,
  getAvailableBalance,
  getLockedReserve,
  hasRealMultisig,
  isValidStellarPublicKey,
} from "@/lib/stellar";

const VALID_PUBLIC_KEY = `G${"A".repeat(55)}`;
const ISSUER_PUBLIC_KEY = `G${"B".repeat(55)}`;

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("Stellar account helpers", () => {
  it("validates the expected Stellar public-key shape", () => {
    expect(isValidStellarPublicKey(VALID_PUBLIC_KEY)).toBe(true);
    expect(isValidStellarPublicKey(`S${"A".repeat(55)}`)).toBe(false);
    expect(isValidStellarPublicKey("GABC")).toBe(false);
    expect(isValidStellarPublicKey(`G${"0".repeat(55)}`)).toBe(false);
  });

  it("calculates the locked reserve from base entries and subentries", () => {
    expect(getLockedReserve(0)).toBe(1);
    expect(getLockedReserve(4)).toBe(3);
    expect(getAvailableBalance(12.5, 4)).toBe(9.5);
  });

  it("does not return a negative available balance", () => {
    expect(getAvailableBalance(0.5, 2)).toBe(0);
  });

  it("detects real multisig only when multiple signers have weight", () => {
    expect(
      hasRealMultisig([
        { key: VALID_PUBLIC_KEY, weight: 1, type: "ed25519_public_key" },
        { key: `G${"B".repeat(55)}`, weight: 1, type: "ed25519_public_key" },
      ]),
    ).toBe(true);

    expect(
      hasRealMultisig([
        { key: VALID_PUBLIC_KEY, weight: 1, type: "ed25519_public_key" },
        { key: `G${"B".repeat(55)}`, weight: 0, type: "ed25519_public_key" },
      ]),
    ).toBe(false);
  });

  it("deduplicates issuer lookups and attaches issuer flags to trustlines", async () => {
    const fetchMock = vi.fn(async (input: string | URL | Request) => {
      const url = String(input);
      const payload = url.endsWith(ISSUER_PUBLIC_KEY)
        ? {
            flags: {
              auth_required: true,
              auth_revocable: true,
              auth_immutable: false,
              auth_clawback_enabled: false,
            },
          }
        : {
            account_id: VALID_PUBLIC_KEY,
            sequence: "100",
            subentry_count: 2,
            last_modified_ledger: 123,
            thresholds: { low_threshold: 1, med_threshold: 1, high_threshold: 1 },
            flags: {},
            signers: [{ key: VALID_PUBLIC_KEY, weight: 1, type: "ed25519_public_key" }],
            balances: [
              { asset_type: "native", balance: "100" },
              {
                asset_type: "credit_alphanum4",
                asset_code: "USD",
                asset_issuer: ISSUER_PUBLIC_KEY,
                balance: "10",
                limit: "100",
                is_authorized: true,
              },
              {
                asset_type: "credit_alphanum4",
                asset_code: "EUR",
                asset_issuer: ISSUER_PUBLIC_KEY,
                balance: "5",
                limit: "100",
                is_authorized: true,
              },
            ],
          };

      return new Response(JSON.stringify(payload), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    });
    vi.stubGlobal("fetch", fetchMock);

    const account = await fetchStellarAccount(VALID_PUBLIC_KEY, "testnet");

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(account.trustlines).toHaveLength(2);
    expect(account.trustlines.every((trustline) => trustline.issuerLookupStatus === "loaded")).toBe(true);
    expect(account.trustlines.every((trustline) => trustline.issuerFlags?.auth_revocable)).toBe(true);
  });
});
