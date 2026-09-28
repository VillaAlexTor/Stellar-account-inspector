# Pruebas

## Frontend

```bash
pnpm test
pnpm lint
pnpm build
```

## Backend

```bash
pnpm sentinel:test
```

La suite Go cubre los cambios de firmantes, pesos, umbrales, trustlines hacia emisores nuevos, master key en cero y el parser del stream SSE.

## Integración local

```bash
pnpm sentinel:up
curl http://localhost:8080/healthz
```

Usa una cuenta Testnet activa, abre la ruta Sentinel y ejecuta una operación `set_options` o `change_trust` desde una herramienta autorizada. Nunca introduzcas una seed phrase en Stellar Account Inspector; el producto sólo recibe claves públicas.

Testnet puede reiniciarse. `pnpm --filter @stellar-inspector/web testnet:provision` crea fixtures nuevas para Inspector y Risk Score, pero no conserva ni imprime secretos.
