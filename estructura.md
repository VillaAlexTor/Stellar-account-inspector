stellar-account-inspector/
├─ apps/
│  └─ web/
│     ├─ app/
│     │  ├─ layout.tsx
│     │  ├─ page.tsx
│     │  ├─ globals.css
│     │  └─ sentinel/
│     │     └─ [publicKey]/
│     │        └─ page.tsx
│     │
│     ├─ components/
│     │  ├─ ui/                         # Componentes generados por shadcn/ui
│     │  ├─ layout/
│     │  │  ├─ AppHeader.tsx
│     │  │  └─ NetworkIndicator.tsx
│     │  ├─ inspector/
│     │  │  ├─ StellarAccountVerifier.tsx
│     │  │  ├─ AccountSearchForm.tsx
│     │  │  ├─ AccountSummary.tsx
│     │  │  ├─ AccountFlags.tsx
│     │  │  ├─ TrustlinesTable.tsx
│     │  │  ├─ SignersTable.tsx
│     │  │  └─ MultisigBadge.tsx
│     │  ├─ risk/
│     │  │  ├─ RiskScoreCard.tsx
│     │  │  ├─ RiskLevelBadge.tsx
│     │  │  └─ RiskFindingItem.tsx
│     │  └─ sentinel/
│     │     ├─ SentinelHeader.tsx
│     │     ├─ ConnectionStatus.tsx
│     │     ├─ AlertTimeline.tsx
│     │     ├─ AlertTimelineItem.tsx
│     │     └─ EmptyAlertState.tsx
│     │
│     ├─ lib/
│     │  ├─ stellar.ts                  # Fachada pública solicitada
│     │  ├─ stellar/
│     │  │  ├─ client.ts               # Fetch a Horizon y errores HTTP
│     │  │  ├─ networks.ts             # Testnet y URL de Horizon
│     │  │  ├─ types.ts
│     │  │  ├─ accountMapper.ts
│     │  │  └─ balances.ts
│     │  ├─ riskScore.ts                # API pública pura y testeable
│     │  ├─ risk/
│     │  │  ├─ types.ts
│     │  │  ├─ constants.ts
│     │  │  └─ rules/
│     │  │     ├─ preauthHighWeight.ts
│     │  │     ├─ inconsistentThresholds.ts
│     │  │     ├─ weakenedMasterKey.ts
│     │  │     ├─ revocableIssuer.ts
│     │  │     └─ lowAvailableReserve.ts
│     │  ├─ sentinelApi.ts
│     │  └─ utils.ts
│     │
│     ├─ stores/
│     │  ├─ inspectorStore.ts
│     │  └─ sentinelStore.ts
│     │
│     ├─ tests/
│     │  ├─ unit/
│     │  │  ├─ stellar.test.ts
│     │  │  ├─ riskScore.test.ts
│     │  │  └─ riskRules.test.ts
│     │  ├─ integration/
│     │  │  └─ horizon.test.ts
│     │  └─ fixtures/
│     │     ├─ healthy-account.json
│     │     ├─ inconsistent-thresholds.json
│     │     └─ disabled-master-key.json
│     │
│     ├─ scripts/
│     │  └─ provision-testnet-accounts.ts
│     ├─ public/
│     ├─ components.json
│     ├─ next.config.ts
│     ├─ tailwind.config.ts
│     ├─ tsconfig.json
│     ├─ vitest.config.ts
│     └─ package.json
│
├─ services/
│  └─ sentinel-api/
│     ├─ cmd/
│     │  └─ api/
│     │     └─ main.go
│     ├─ internal/
│     │  ├─ config/
│     │  │  └─ config.go
│     │  ├─ database/
│     │  │  ├─ postgres.go
│     │  │  └─ migrate.go
│     │  ├─ model/
│     │  │  ├─ monitored_account.go
│     │  │  ├─ sentinel_alert.go
│     │  │  └─ relevant_operation.go
│     │  ├─ repository/
│     │  │  ├─ account_repository.go
│     │  │  ├─ alert_repository.go
│     │  │  └─ operation_repository.go
│     │  ├─ horizon/
│     │  │  ├─ client.go
│     │  │  ├─ stream.go
│     │  │  └─ types.go
│     │  ├─ sentinel/
│     │  │  ├─ manager.go
│     │  │  ├─ session.go
│     │  │  ├─ state.go
│     │  │  ├─ evaluator.go
│     │  │  └─ rules.go
│     │  └─ http/
│     │     ├─ router.go
│     │     ├─ middleware/
│     │     │  └─ cors.go
│     │     └─ handler/
│     │        ├─ health.go
│     │        ├─ monitored_accounts.go
│     │        ├─ alerts.go
│     │        └─ alert_stream.go
│     ├─ go.mod
│     └─ go.sum
│
├─ docs/
│  ├─ project-brief.md
│  ├─ architecture.md
│  ├─ api.md
│  └─ testing.md
├─ PRODUCT.md
├─ DESIGN.md
├─ compose.yaml
├─ package.json
├─ pnpm-workspace.yaml
├─ .env.example
├─ .editorconfig
├─ .gitignore
└─ README.md
