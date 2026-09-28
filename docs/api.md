# Sentinel API

Base local: `http://localhost:8080`

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
