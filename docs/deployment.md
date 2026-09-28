# Despliegue HTTPS

La configuración de producción usa PostgreSQL, la API Go, el servidor standalone de Next.js y Caddy. Sólo Caddy publica puertos; termina TLS automáticamente y enruta `/api/*`, `/healthz` y `/readyz` a la API. `/metrics` permanece en la red interna.

## Preparación

1. Apunta un registro DNS A/AAAA del dominio al servidor y abre TCP 80/443 y UDP 443.
2. Copia `.env.production.example` a `.env.production` y reemplaza todos los marcadores. Usa una contraseña PostgreSQL URL-safe.
3. Genera tokens y secretos con un generador criptográfico; no los confirmes en Git.
4. Ejecuta `docker compose --env-file .env.production -f compose.prod.yaml config` para validar.
5. Inicia con `docker compose --env-file .env.production -f compose.prod.yaml up -d --build`.

Comprueba `https://DOMINIO/healthz` y `https://DOMINIO/readyz`. Revisa después `docker compose --env-file .env.production -f compose.prod.yaml logs -f caddy sentinel-api`.

## Seguridad operativa

- Caddy añade HSTS, CSP, protección contra framing y otras cabeceras defensivas.
- La API exige autenticación, cookies seguras y CORS de origen exacto cuando `SENTINEL_ENVIRONMENT=production`.
- La red Docker usa `172.30.0.0/24`; sólo ese proxy puede aportar `X-Forwarded-For` a los límites por IP.
- API y web no publican puertos del host. PostgreSQL tampoco se expone.
- Respalda el volumen `stellar_postgres_data` y prueba restauraciones periódicamente.
- Para actualizar: descarga o construye imágenes, ejecuta las pruebas de CI y recrea servicios con `docker compose ... up -d`.

Los certificados y la configuración de ACME se conservan en `caddy_data` y `caddy_config`.
