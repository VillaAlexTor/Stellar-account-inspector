# Sentinel API

Base local: `http://localhost:8081`

## Autenticación

Si `SENTINEL_AUTH_TOKENS` contiene uno o más tokens separados por comas, todos los endpoints salvo `/healthz` y la gestión de sesión quedan protegidos. `SENTINEL_SESSION_SECRET` debe tener al menos 32 caracteres.

El navegador intercambia el token por una cookie firmada, `HttpOnly` y `SameSite=Strict`:

```http
POST /api/v1/auth/session
Content-Type: application/json

{"token":"TOKEN_CONFIGURADO"}
```

`GET /api/v1/auth/session` consulta el estado y `DELETE /api/v1/auth/session` cierra la sesión. Los clientes de API también pueden usar `Authorization: Bearer TOKEN_CONFIGURADO`.

## Límites

Los límites se aplican por IP y por clave pública Stellar durante la ventana configurada. Las respuestas incluyen `RateLimit-Limit` y `RateLimit-Remaining`; al exceder el límite responden `429 Too Many Requests` con `Retry-After`. `X-Forwarded-For` sólo se utiliza cuando la conexión procede de un CIDR incluido explícitamente en `SENTINEL_TRUSTED_PROXY_CIDRS`.

## Salud

`GET /healthz`

`GET /readyz` comprueba que PostgreSQL acepte conexiones y responde `503` cuando el proceso no está listo para recibir tráfico.

## Métricas

`GET /metrics` expone métricas Prometheus y requiere la misma autenticación que la API cuando está habilitada. Consulta [observability.md](observability.md) para nombres, ejemplos de scrape y alertas sugeridas.

## Iniciar o recuperar monitoreo

`POST /api/v1/monitored-accounts`

```json
{
  "publicKey": "G...",
  "network": "testnet"
}
```

La respuesta `202 Accepted` contiene la cuenta persistida y el estado actual de la sesión.

## Consultar cuenta

`GET /api/v1/monitored-accounts/{publicKey}?network=testnet`

## Historial de alertas

`GET /api/v1/monitored-accounts/{publicKey}/alerts?network=testnet&limit=100&cursor=123`

Las alertas se entregan de la más reciente a la más antigua. `limit` acepta de 1 a 200. Cuando quedan resultados, la respuesta incluye `nextCursor`; úsalo como `cursor` en la siguiente petición. El cursor es estable porque corresponde al ID descendente de la última alerta de la página.

La retención se ejecuta en segundo plano. Por defecto conserva alertas 90 días y operaciones relevantes 30 días; `SENTINEL_ALERT_RETENTION`, `SENTINEL_OPERATION_RETENTION` y `SENTINEL_RETENTION_INTERVAL` permiten cambiar estas políticas.

## Eventos en vivo

`GET /api/v1/monitored-accounts/{publicKey}/events?network=testnet`

Tipos SSE:

- `status`: `connecting`, `connected`, `reconnecting` o `down`.
- `alert`: alerta persistida con `ruleId`, `severity`, `message`, `operationId` y `createdAt`.

El servidor emite heartbeats cada 15 segundos y desactiva buffering de proxies compatibles.
