Nivel 1 y 2
Browser ──directo──> Horizon
   │
   ├─ lib/stellar.ts
   └─ lib/riskScore.ts (sin efectos de red)

Nivel 3
Browser ──SSE──> Go Sentinel API ──SSE──> Horizon
                       │
                       └─ GORM ──> PostgreSQL