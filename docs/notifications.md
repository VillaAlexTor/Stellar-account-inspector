# Notificaciones Sentinel

Sentinel crea las entregas en la misma transacción PostgreSQL que la alerta. Un despachador reclama trabajos con bloqueo `SKIP LOCKED`, registra intentos y reintenta con espera exponencial. Así, reiniciar la API no pierde las notificaciones pendientes.

## Webhook

Configura `SENTINEL_WEBHOOK_URL` y, recomendado, `SENTINEL_WEBHOOK_SECRET`. El receptor recibe un `POST application/json`, un ID idempotente en `X-Sentinel-Delivery-ID` y la firma `sha256=<hex>` en `X-Sentinel-Signature`. La firma es HMAC-SHA256 del cuerpo exacto. El receptor debe deduplicar por ID y comparar la firma en tiempo constante.

En producción la URL debe usar HTTPS y Sentinel no sigue redirecciones.

## Telegram

Configura juntos `SENTINEL_TELEGRAM_BOT_TOKEN` y `SENTINEL_TELEGRAM_CHAT_ID`. El texto usa HTML escapado y se envía mediante la API oficial de bots. `SENTINEL_TELEGRAM_API_URL` sólo se necesita para pruebas o un proxy compatible; en producción debe ser HTTPS.

## Correo

Configura `SENTINEL_SMTP_HOST`, `SENTINEL_SMTP_FROM` y `SENTINEL_SMTP_TO`. Los destinatarios se separan por comas. Las credenciales son opcionales; `SENTINEL_SMTP_TLS` acepta `starttls` (predeterminado), `implicit` o `none`. Producción rechaza `none`.

## Reintentos y operación

- `SENTINEL_NOTIFICATION_TIMEOUT`: tiempo máximo por entrega, 10 segundos por defecto.
- `SENTINEL_NOTIFICATION_POLL`: intervalo del despachador, 2 segundos por defecto.
- `SENTINEL_NOTIFICATION_MAX_ATTEMPTS`: intentos antes del estado terminal `failed`, 5 por defecto.
- `sentinel_notification_deliveries_total{channel,result}`: resultados `sent`, `retry` y `failed` para alertas operativas.

Los errores de proveedores quedan truncados a 2.000 caracteres en `notification_deliveries.last_error`; no se registran tokens ni contraseñas.
