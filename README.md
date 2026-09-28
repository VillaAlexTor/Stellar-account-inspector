# Stellar Account Inspector

Herramienta educativa para inspeccionar cuentas Stellar, explicar su configuración de seguridad y monitorear cambios sensibles en tiempo real.

## Estado

- Nivel 1 — Inspector: implementado.
- Nivel 2 — Risk Score: implementado.
- Nivel 3 — Sentinel: implementado.

## Requisitos

- Node.js 22.12 o superior.
- pnpm 10.

## Desarrollo

```bash
pnpm install
pnpm dev
```

Abre `http://localhost:3000`. El Inspector consulta Horizon directamente desde el navegador, calcula el riesgo localmente y no solicita claves secretas. También acepta enlaces directos con `?account=G…&network=testnet`.

## Sentinel

Sentinel usa Go, GORM y PostgreSQL. Mantiene un stream SSE compartido hacia Horizon, conserva checkpoints y reenvía estados y alertas al navegador mediante otro canal SSE.

```bash
pnpm sentinel:up
pnpm dev
```

Después de inspeccionar una cuenta, selecciona **Activar Sentinel**. También puedes abrir directamente:

```text
http://localhost:3000/sentinel/CLAVE_PUBLICA?network=testnet
```

Comprobaciones útiles:

```bash
curl http://localhost:8081/healthz
curl -X POST http://localhost:8081/api/v1/monitored-accounts \
  -H "Content-Type: application/json" \
	-H "Authorization: Bearer $SENTINEL_TOKEN" \
  -d '{"publicKey":"CLAVE_PUBLICA","network":"testnet"}'
curl -N "http://localhost:8081/api/v1/monitored-accounts/CLAVE_PUBLICA/events?network=testnet"
```

Los comandos `pnpm sentinel:logs` y `pnpm sentinel:down` muestran los logs y detienen el entorno, respectivamente.

### Protección de acceso

Sentinel admite tokens configurables, sesiones de navegador mediante cookie firmada `HttpOnly` y límites separados por IP y cuenta Stellar. La autenticación queda desactivada únicamente cuando `SENTINEL_AUTH_TOKENS` está vacío, para conservar el flujo de desarrollo local. Copia `.env.example` a `.env`, define `SENTINEL_AUTH_TOKENS` y un `SENTINEL_SESSION_SECRET` aleatorio de al menos 32 caracteres para activarla. En despliegues HTTPS usa `SENTINEL_COOKIE_SECURE=true`.

### Observabilidad

La API emite logs JSON con `request_id`, ruta normalizada, estado, bytes y duración; ofrece liveness en `/healthz`, readiness de PostgreSQL en `/readyz` y métricas Prometheus protegidas en `/metrics`. La guía de métricas, consultas y alertas está en [`docs/observability.md`](docs/observability.md).

### Historial, retención y notificaciones

El historial se pagina por cursor y las políticas de retención eliminan automáticamente alertas y operaciones vencidas. Sentinel puede entregar cada alerta por webhook firmado, Telegram y correo SMTP; la bandeja persistente conserva reintentos entre reinicios. Consulta [`docs/notifications.md`](docs/notifications.md).

### Producción y CI/CD

[`compose.yaml`](compose.yaml) despliega PostgreSQL, la API Go y la interfaz Next.js para la demostración local. GitHub Actions ejecuta lint, pruebas, integración real, builds de contenedores y CodeQL; los tags `v*` publican imágenes en GHCR.

Para mostrar el sistema sin contratar dominio ni servidor usa la guía de [`docs/local-demo.md`](docs/local-demo.md). Levanta PostgreSQL, la API y la interfaz con Docker y abre el proyecto directamente en `localhost`.

### Conexión desde DBeaver

Crea una conexión **PostgreSQL** con estos valores:

- Host: `localhost`
- Puerto: `5433`
- Base de datos: `stellar_inspector`
- Usuario: `stellar`
- Contraseña: `stellar`
- SSL: desactivado

Las tablas creadas por GORM son `monitored_accounts`, `relevant_operations`, `sentinel_alerts` y `notification_deliveries`. Los puertos externos pueden cambiarse con `POSTGRES_HOST_PORT` y `SENTINEL_HOST_PORT`; dentro de Docker permanecen en `5432` y `8080`.

## Verificación

```bash
pnpm test
pnpm lint
pnpm build
pnpm sentinel:test
pnpm sentinel:test:integration
```

La reserva bloqueada se calcula según el alcance académico definido para esta versión: `(2 + subentry_count) × 0.5 XLM`.

Para crear y validar tres escenarios nuevos en Stellar Testnet (cuenta saludable, umbrales inconsistentes y master key con peso 0):

```bash
pnpm --filter @stellar-inspector/web testnet:provision
```

El script usa Friendbot, imprime únicamente claves públicas y omite deliberadamente todas las claves secretas. Testnet puede reiniciarse, por lo que los ejemplos se generan bajo demanda.

## Estructura

- `apps/web`: frontend Next.js.
- `services/sentinel-api`: backend Go de monitoreo, reglas, persistencia y SSE.
- `PRODUCT.md`: contexto y principios duraderos del producto.
- `prompt.md`: brief original del proyecto.
- `docs/final-review.md`: trazabilidad de fases A–F y endurecimiento 1–8.
- `SECURITY.md`: política de reporte y límites de seguridad conocidos.
