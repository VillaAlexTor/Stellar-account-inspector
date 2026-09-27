import { describe, expect, it } from "vitest";
import {
  getAvailableBalance,
  getLockedReserve,
  hasRealMultisig,
  isValidStellarPublicKey,
} from "@/lib/stellar";

const VALID_PUBLIC_KEY = `G${"A".repeat(55)}`;

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
});
