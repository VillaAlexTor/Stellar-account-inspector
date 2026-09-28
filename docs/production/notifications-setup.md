# Configuración de notificaciones externas

Puedes activar uno o varios canales. Cada alerta se guarda primero en PostgreSQL y después se entrega con reintentos.

## Telegram

1. Abre Telegram y conversa únicamente con **@BotFather**.
2. Ejecuta `/newbot`, asigna nombre y usuario, y guarda el token como una contraseña.
3. Abre el bot recién creado y envíale un mensaje, por ejemplo `inicio`.
4. Consulta `https://api.telegram.org/bot<TOKEN>/getUpdates` y busca `message.chat.id`. Evita guardar esa URL en historial compartido.
5. Configura:

```dotenv
SENTINEL_TELEGRAM_BOT_TOKEN=123456:token-entregado-por-botfather
SENTINEL_TELEGRAM_CHAT_ID=123456789
```

Para enviar a un grupo, agrega el bot, envía un mensaje y usa el `chat.id` negativo del grupo. Si el token se expone, revócalo inmediatamente en BotFather.

Guía oficial: <https://core.telegram.org/bots/tutorial>.

## Correo con Amazon SES

1. En la misma región de AWS abre **Amazon SES → Verified identities**.
2. Verifica el correo o dominio que usarás como remitente.
3. Abre **SMTP settings → Create SMTP credentials** y descarga el usuario y contraseña una sola vez.
4. Si SES permanece en sandbox, verifica también cada destinatario. Solicita acceso de producción cuando quieras enviar a destinatarios no verificados.
5. Copia el endpoint SMTP mostrado por SES y configura:

```dotenv
SENTINEL_SMTP_HOST=email-smtp.REGION.amazonaws.com
SENTINEL_SMTP_PORT=587
SENTINEL_SMTP_USERNAME=usuario-smtp-de-ses
SENTINEL_SMTP_PASSWORD=contrasena-smtp-de-ses
SENTINEL_SMTP_FROM=sentinel@tu-dominio-o-correo-verificado
SENTINEL_SMTP_TO=tu-correo@example.com
SENTINEL_SMTP_TLS=starttls
```

No uses las access keys normales de AWS como credenciales SMTP. Consulta la [guía oficial de credenciales SMTP de SES](https://docs.aws.amazon.com/ses/latest/dg/smtp-credentials.html).

## Webhook

Necesitas un endpoint HTTPS propio que acepte `POST application/json` y responda con `2xx` rápidamente. Genera un secreto nuevo:

```bash
openssl rand -hex 32
```

Configura:

```dotenv
SENTINEL_WEBHOOK_URL=https://tu-servicio.example/hooks/stellar
SENTINEL_WEBHOOK_SECRET=secreto-hmac-independiente
```

El receptor debe calcular HMAC-SHA256 sobre el cuerpo exacto y compararlo con `X-Sentinel-Signature`. Debe deduplicar por `X-Sentinel-Delivery-ID`, porque un timeout puede provocar reintentos aunque el primer envío haya sido procesado.

## Prueba final

1. Reinicia la API tras cambiar variables:

```bash
docker compose --env-file .env.production -f compose.prod.yaml up -d --force-recreate sentinel-api
docker compose --env-file .env.production -f compose.prod.yaml logs -f sentinel-api
```

2. Usa una cuenta **Testnet desechable** y genera un cambio controlado de firmante, umbral o trustline. No pruebes reduciendo la master key en una cuenta importante.
3. Comprueba una sola entrega por canal y revisa en PostgreSQL `notification_deliveries.status`, `attempts` y `last_error`.
4. Si falla, busca `notification_delivery_failed` en logs y `sentinel_notification_deliveries_total` en métricas.

La estructura del payload, firma y reintentos está documentada también en [notifications.md](../notifications.md).
