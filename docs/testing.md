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

La prueba levanta el PostgreSQL real de Compose, aplica las migraciones, registra una cuenta aislada, simula el stream de Horizon y comprueba de extremo a extremo que la API entrega una alerta por SSE y la persiste junto con su cursor. También verifica autenticación, readiness, métricas protegidas, entrega durable de notificaciones, paginación por cursor y purga de retención. Los datos de la prueba se eliminan al terminar. Para usar otro PostgreSQL define `SENTINEL_INTEGRATION_DATABASE_URL`.

## Seguridad y contenedores

```bash
pnpm audit --prod
docker run --rm -v "${PWD}/services/sentinel-api:/src" -w /src golang:1.25.14-alpine \
  sh -c "go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./..."
docker build -t stellar-sentinel-api:audit services/sentinel-api
docker build -f apps/web/Dockerfile -t stellar-sentinel-web:audit .
```

La CI ejecuta además `go test -race` sobre Linux, CodeQL y la prueba de integración contra un servicio PostgreSQL 17 real.

Para una comprobación manual del servicio completo:

```bash
pnpm sentinel:up
curl http://localhost:8081/healthz
```

Usa una cuenta Testnet activa, abre la ruta Sentinel y ejecuta una operación `set_options` o `change_trust` desde una herramienta autorizada. Nunca introduzcas una seed phrase en Stellar Account Inspector; el producto sólo recibe claves públicas.

Testnet puede reiniciarse. `pnpm --filter @stellar-inspector/web testnet:provision` crea fixtures nuevas para Inspector y Risk Score, pero no conserva ni imprime secretos.
