# Pruebas

## Verificación completa

```bash
pnpm lint
pnpm test
pnpm build
```

## Cobertura funcional

La suite comprueba:

- validación de claves públicas;
- cálculo de reserva y balance disponible;
- detección de multisig real;
- normalización de cuentas Horizon;
- master key debilitada;
- umbrales inconsistentes;
- emisores revocables;
- balance disponible bajo;
- firmantes `sha256_hash` como hash-locks;
- firmantes `preauth_tx` como transacciones preautorizadas;
- suma de severidades y límite máximo del Risk Score.

## Escenarios Testnet

```bash
pnpm --filter @stellar-inspector/web testnet:provision
```

El script crea escenarios reproducibles mediante Friendbot e imprime únicamente claves públicas. Testnet puede reiniciarse, por lo que las cuentas deben generarse bajo demanda.

## Criterio de aceptación

La entrega es válida cuando lint, pruebas y build finalizan correctamente, una cuenta Testnet activa puede inspeccionarse desde el navegador y cada hallazgo identifica su evidencia sin confundir tipos de firmante.
