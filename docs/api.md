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

`GET /api/v1/monitored-accounts/{publicKey}/alerts?network=testnet&limit=100`

Las alertas se entregan de la más reciente a la más antigua.

## Eventos en vivo

`GET /api/v1/monitored-accounts/{publicKey}/events?network=testnet`

Tipos SSE:

- `status`: `connecting`, `connected`, `reconnecting` o `down`.
- `alert`: alerta persistida con `ruleId`, `severity`, `message`, `operationId` y `createdAt`.

El servidor emite heartbeats cada 15 segundos y desactiva buffering de proxies compatibles.
