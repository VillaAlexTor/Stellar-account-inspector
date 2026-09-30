# Product

<!-- impeccable:product-schema 1 -->

## Platform

Web frontend.

## Stack

- Next.js App Router, React and TypeScript.
- Direct browser requests to Horizon Testnet.
- Deterministic local rules for the Risk Score.
- No proprietary backend or database.

## Users

The primary user is a student or technically proficient Stellar user who needs to inspect a Testnet account and understand its control configuration without manually interpreting a raw Horizon response.

## Product Purpose

Stellar Account Inspector turns a public key into two related views: a current account inspection and a heuristic security assessment. Success means the user can connect every judgment to observable account data.

## Operating Context

- Users paste a Stellar Testnet public key without connecting a wallet.
- The browser reads Horizon directly.
- The product communicates in Spanish while preserving protocol field names where they improve precision.
- The evaluated scope contains Inspector and Risk Score only.

## Capabilities and Constraints

- Public-key validation uses `^G[A-Z2-7]{55}$`.
- Inspector shows balances, estimated reserve, trustlines, signers, thresholds, authorization flags and real multisig status.
- Risk Score is implemented as independent pure rules and remains in the `0–100` range.
- `sha256_hash` hash-locks and `preauth_tx` preauthorized transactions are distinct signer types and must never share a misleading label.
- Testnet fixtures are provisioned on demand because the network can reset.
- No wallet secrets or seed phrases are requested, accepted or stored.

## Optional Portfolio Extension

The former Sentinel backend is outside the evaluated scope. Its complete implementation is preserved in the `Retro` branch as an optional portfolio extension and is not required by this product version.

## Brand Commitments

- Product name: Stellar Account Inspector.
- Voice: precise, educational, security-conscious and calm.
- The interface behaves as an operational audit console, not a marketing dashboard.

## Product Principles

1. Explain every security judgment with observable Stellar data.
2. Keep the fixed Testnet context and data freshness unambiguous.
3. Reveal protocol complexity progressively.
4. Treat findings as heuristic signals, not absolute verdicts.
5. Keep the evaluated solution proportional to its frontend-only purpose.

## Accessibility & Inclusion

The interface should meet WCAG 2.2 AA expectations, remain keyboard operable, avoid color-only status communication and provide clear Spanish labels and recovery messages.
