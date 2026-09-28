# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

- Frontend: Next.js App Router, React, TypeScript, Tailwind CSS, shadcn/ui, and Zustand.
- Backend: Go, introduced for Sentinel monitoring and persistence.
- Database: PostgreSQL through GORM.
- Stellar data: Horizon Mainnet and Testnet APIs; Inspector and Risk Score query Horizon directly from the browser.

## Users

The primary user is an information-security student or technically proficient Stellar user who needs to inspect an account, understand its control configuration, and recognize risky changes without manually interpreting raw Horizon responses.

## Product Purpose

Stellar Account Inspector turns a Stellar public key into three progressively deeper views: a current account inspection, a heuristic security assessment, and continuous monitoring for sensitive account-control changes. Success means the user can move from raw account state to an understandable security judgment and, later, to actionable alerts.

## Positioning

The product connects Stellar-native account mechanics—reserves, trustlines, signers, thresholds, authorization flags, and operations—to explicit security findings and takeover-oriented monitoring in one learning-friendly tool.

## Operating Context

- Users paste a Stellar public key, choose Mainnet or Testnet, and inspect the account without connecting a wallet.
- Results must remain useful for both quick review and deeper security study.
- Sentinel maintains one server-side Horizon stream per account and network, persists relevant operations, alerts, cursors, and comparison state, and delivers live events to the browser.
- The product communicates in Spanish while preserving protocol field names where they improve technical precision.

## Capabilities and Constraints

- Public-key validation uses `^G[A-Z2-7]{55}$`.
- The Inspector shows native balance, available balance, reserve, trustlines, signers, thresholds, authorization flags, and real multisig status.
- Risk Score is deterministic and implemented as independent pure rules. Network enrichment happens before evaluation.
- Score range is 0–100, where a higher value means higher risk.
- Sentinel uses a Go proxy, GORM, PostgreSQL, and server-sent events.
- Testnet examples must be provisioned reproducibly because Stellar Testnet accounts can be reset.
- No wallet secrets or seed phrases are requested, accepted, or stored.

## Brand Commitments

- Product name: Stellar Account Inspector.
- Voice: precise, educational, security-conscious, and calm; warnings explain evidence rather than dramatizing it.
- The interface is an operational audit console, not a marketing dashboard.

## Evidence on Hand

- The project brief in `promtp.md` defines the product levels, technical stack, risk rules, Sentinel alerts, and implementation order.
- No existing application UI, logo, customer claims, production metrics, or visual assets exist yet and none should be fabricated.

## Product Principles

1. Explain every security judgment with observable Stellar data.
2. Keep network identity and data freshness unambiguous.
3. Reveal complexity progressively without hiding protocol details.
4. Treat heuristic findings as evidence-informed signals, not absolute verdicts.
5. Keep the architecture extensible from one-time inspection to continuous monitoring.

## Accessibility & Inclusion

The web interface should meet WCAG 2.2 AA expectations, remain fully keyboard operable, avoid color-only status communication, and provide clear Spanish labels and error messages.
