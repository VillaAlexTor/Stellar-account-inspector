# PROMPT — Stellar Account Inspector (Inspector + Risk Score + Sentinel)

Pega este prompt completo en Claude Code (o el agente que uses) para arrancar el proyecto de punta a punta, en fases.

---

## CONTEXTO

Soy estudiante de Informática (UMSA, mención Seguridad de la Información). Estoy construyendo **Stellar Account Inspector**, una herramienta para la red Stellar que evoluciona en tres niveles:

1. **Nivel 1 — Inspector**: consulta el estado de una cuenta Stellar (saldos, trustlines, reserva, firmantes).
2. **Nivel 2 — Risk Score**: audita esa misma cuenta y calcula un puntaje de riesgo/seguridad con reglas heurísticas.
3. **Nivel 3 — Sentinel**: monitorea una cuenta en tiempo real vía streaming y genera alertas ante eventos sospechosos (patrón de account takeover / drainer).

Quiero que me ayudes a construir esto de forma incremental, empezando por el Nivel 1 y dejando la arquitectura lista para no tener que reescribir nada al llegar al Nivel 3.

## STACK TÉCNICO (usar exactamente este, no proponer alternativas)

**Frontend**
- Next.js (App Router) + React + TypeScript
- Tailwind CSS + shadcn/ui
- Zustand para estado global (cuenta activa, historial de alertas en memoria, estado de conexión del Sentinel)

**Backend**
- Go (Golang) como lenguaje principal del backend
- Python solo si hace falta un script puntual de tooling/automatización (no como servicio productivo)

**Base de datos**
- PostgreSQL
- GORM como ORM en Go

**Red Stellar**
- Horizon Testnet: `https://horizon-testnet.stellar.org`
- Sin backend propio para el Nivel 1 (fetch directo desde el cliente); el backend en Go se introduce recién en el Nivel 3, para persistencia de alertas y el proxy del stream (ver más abajo).

---

## NIVEL 1 — Inspector (base, ya diseñado, que quede como cimiento)

Endpoint fuente: `GET /accounts/{account_id}`

Módulos esperados:
- `lib/stellar.ts`: validación de formato de clave pública (`^G[A-Z2-7]{55}$`), fetch a Horizon, cálculo de reserva bloqueada `(2 + subentry_count) * 0.5 XLM`, extracción de trustlines y balance disponible.
- `components/StellarAccountVerifier.tsx`: formulario + tarjetas de resultado (resumen, trustlines, firmantes) usando shadcn/ui.

Pendientes de esta fase que quiero que completes:
- Selector Mainnet / Testnet (parametrizar la URL base de Horizon).
- Mostrar flags de la cuenta: `auth_required`, `auth_revocable`, `auth_clawback_enabled`.
- Detectar y marcar visualmente si la cuenta es **multisig real** (más de un signer con `weight > 0`).

---

## NIVEL 2 — Stellar Risk Score

Objetivo: convertir los mismos datos de Horizon en un **puntaje de riesgo (0–100, o una escala Bajo/Medio/Alto/Crítico)**, con el detalle de qué reglas se dispararon y por qué, como un mini-auditor.

Implementar como un módulo puro y testeable: `lib/riskScore.ts`, que recibe un `StellarAccountData` y devuelve:

```ts
interface RiskFinding {
  id: string;            // ej. "SIGNER_PREAUTH_HIGH_WEIGHT"
  severity: "low" | "medium" | "high" | "critical";
  title: string;
  description: string;   // explicación en lenguaje humano de por qué es riesgoso
}

interface RiskReport {
  score: number;          // 0-100, más alto = más riesgo
  level: "bajo" | "medio" | "alto" | "crítico";
  findings: RiskFinding[];
}
```

Reglas a implementar (cada una es una función independiente que retorna `RiskFinding | null`, para poder agregar más reglas después sin tocar las existentes):

1. **Signer preauth con weight alto**: existe un signer `type: "sha256_hash"` (preauth tx) con `weight` alto relativo al `high_threshold` → severidad `high`. Es una configuración inusual fuera de flujos de multifirma predefinidos.
2. **Umbrales inconsistentes**: `high_threshold < med_threshold`, o `med_threshold < low_threshold`, o cualquiera de ellos en `0` de forma no intencionada → severidad `high`. Umbral en 0 sin justificación documentada suele indicar cuenta "vaciada" de control deliberadamente.
3. **Master key debilitada sin respaldo**: el signer cuya `key` coincide con el `account_id` (master key) tiene `weight = 0`, y la suma de weights de los demás signers no alcanza el `med_threshold` → severidad `critical`. La cuenta queda huérfana/bloqueada.
4. **Trustline hacia emisor revocable**: alguna trustline apunta a un activo cuyo emisor tiene `auth_revocable = true` (requiere una consulta adicional a `GET /accounts/{issuer}` para leer sus flags) → severidad `medium`. El emisor puede congelar esos fondos en cualquier momento.
5. **Reserva al límite**: el saldo disponible (`getAvailableBalance`) es menor al 10% del saldo nativo total → severidad `medium`. La cuenta tiene poco margen operativo.

El score final se calcula sumando pesos por severidad (ej. `critical=40, high=25, medium=10, low=5`, cap en 100) — deja esto como constante fácil de ajustar.

UI: nueva sección/tarjeta "Risk Score" en el mismo `StellarAccountVerifier`, con un badge de color según `level` y la lista de `findings` expandible (título + descripción de cada uno).

---

## NIVEL 3 — Stellar Sentinel (monitoreo en tiempo real)

Objetivo: pasar de una foto puntual a **monitoreo continuo** de una cuenta, con alertas cuando ocurren eventos sensibles.

### Fuente de datos
Horizon Streaming API (Server-Sent Events) sobre:
`GET /accounts/{account_id}/operations?cursor=now` con header `Accept: text/event-stream`.

### Arquitectura para este nivel (aquí sí entra el backend en Go)

```
[Browser] --SSE--> [Next.js frontend]
                         |
                         | (opción A: EventSource directo a Horizon desde el cliente)
                         | (opción B: proxy en Go que reenvía el stream + persiste alertas)
                         v
                  [Backend Go] --GORM--> [PostgreSQL]
```

Quiero la **opción B**: un servicio en Go que:
1. Abre y mantiene el stream SSE hacia Horizon para cada cuenta que un usuario esté monitoreando.
2. Evalúa cada operación entrante contra las reglas de alerta (abajo).
3. Persiste en PostgreSQL: la cuenta monitoreada, el historial de operaciones relevantes y las alertas generadas.
4. Reenvía las alertas al frontend (vía su propio SSE/WebSocket, o polling simple a un endpoint REST — decide el más simple de implementar bien).

Esto me deja practicar Go + GORM + Postgres en un caso real, y evita que cada cliente abra su propia conexión persistente a Horizon.

### Reglas de alerta (`internal/sentinel/rules.go` en Go)

Cada regla analiza una operación (`type` del payload de Horizon) contra el estado previo de la cuenta:

1. **Firmante agregado o removido** — operación `set_options` donde cambia el array de signers respecto al estado anterior conocido.
2. **Cambio de umbrales o de weight de un firmante** — operación `set_options` que modifica `low_threshold`, `med_threshold`, `high_threshold`, o el `weight` de un signer existente.
3. **Trustline nueva hacia emisor no visto antes** — operación `change_trust` hacia un `asset_issuer` que no estaba en el historial de trustlines conocido para esa cuenta.
4. **Posible account takeover** — operación `set_options` que reduce el `weight` de la master key a `0`. Severidad máxima, notificación inmediata.

### Modelo de datos sugerido (Postgres vía GORM)

```go
type MonitoredAccount struct {
    ID          uint   `gorm:"primaryKey"`
    PublicKey   string `gorm:"uniqueIndex"`
    Network     string // "testnet" | "mainnet"
    LastCursor  string
    CreatedAt   time.Time
}

type SentinelAlert struct {
    ID                  uint `gorm:"primaryKey"`
    MonitoredAccountID  uint
    RuleID              string // ej. "MASTER_KEY_ZEROED"
    Severity            string // "info" | "warning" | "critical"
    Message             string
    OperationID         string // id de la operación de Horizon que la disparó
    CreatedAt           time.Time
}
```

### UI del Sentinel
- Vista `/sentinel/[publicKey]`: estado de conexión (conectado / reconectando / caído), línea de tiempo de alertas (más reciente arriba), badge de severidad por alerta.
- Zustand store dedicado para el estado de la sesión de monitoreo activa (no mezclar con el store del Inspector).

---

## CÓMO QUIERO QUE TRABAJES

1. Antes de escribir código, propón la estructura de carpetas completa del monorepo (frontend Next.js + backend Go) y espera mi confirmación.
2. Implementa por fases en este orden: (a) pendientes del Nivel 1, (b) Nivel 2 completo y testeado con al menos 3 cuentas Testnet de ejemplo (una "sana", una con umbrales inconsistentes, una con master key en 0), (c) backend Go mínimo (conexión a Postgres + modelo `MonitoredAccount`), (d) integración del stream SSE, (e) reglas de alerta, (f) UI del Sentinel.
3. En cada fase, dame el código completo de los archivos nuevos/modificados y un resumen breve de cómo probarlo (comandos, cuentas de prueba, curl de ejemplo si aplica al backend Go).
4. Señálame explícitamente cualquier decisión de diseño donde haya más de una opción razonable, en vez de asumir en silencio.