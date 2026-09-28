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
pnpm sentinel:test:integration
```

La prueba levanta el PostgreSQL real de Compose, aplica las migraciones, registra una cuenta aislada, simula el stream de Horizon y comprueba de extremo a extremo que la API entrega una alerta por SSE y la persiste junto con su cursor. Los datos de la prueba se eliminan al terminar. Para usar otro PostgreSQL define `SENTINEL_INTEGRATION_DATABASE_URL`.

Para una comprobación manual del servicio completo:

```bash
pnpm sentinel:up
curl http://localhost:8081/healthz
```

Usa una cuenta Testnet activa, abre la ruta Sentinel y ejecuta una operación `set_options` o `change_trust` desde una herramienta autorizada. Nunca introduzcas una seed phrase en Stellar Account Inspector; el producto sólo recibe claves públicas.

Testnet puede reiniciarse. `pnpm --filter @stellar-inspector/web testnet:provision` crea fixtures nuevas para Inspector y Risk Score, pero no conserva ni imprime secretos.
