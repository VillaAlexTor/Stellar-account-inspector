# Arquitectura del entregable

```text
Usuario
  │
  ▼
Next.js en el navegador
  ├── valida la clave pública
  ├── consulta Horizon Testnet
  ├── normaliza la cuenta
  ├── presenta Inspector
  └── ejecuta Risk Score local
```

## Flujo de datos

1. El usuario introduce una clave pública `G...`.
2. El navegador valida formato y longitud.
3. `lib/stellar.ts` solicita la cuenta a Horizon Testnet.
4. Las trustlines se enriquecen con los flags públicos de sus emisores.
5. La respuesta se normaliza en `StellarAccountData`.
6. Inspector presenta los datos y las reglas puras generan el Risk Score.

## Decisiones

- **Frontend puro:** el instrumento no necesita servidor ni persistencia propios.
- **Horizon directo:** evita duplicar una fuente pública y mantiene el estado actual.
- **Reglas independientes:** cada hallazgo puede probarse y explicarse por separado.
- **Sin wallet:** una clave pública basta para inspección de sólo lectura.
- **Testnet fija:** evita confundir la demostración con fondos o cuentas de producción.

## Límites

La disponibilidad depende de Horizon Testnet. El cálculo es heurístico, no modifica la red y no sustituye una auditoría profesional.
