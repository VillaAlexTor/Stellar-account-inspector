# Despliegue HTTPS

Para una instalación guiada desde cero consulta el índice [production/README.md](production/README.md), que incluye AWS con créditos, DuckDNS gratuito, secretos, notificaciones y Grafana Cloud.

La configuración de producción usa PostgreSQL, la API Go, el servidor standalone de Next.js y Caddy. Sólo Caddy publica puertos; termina TLS automáticamente y enruta `/api/*`, `/healthz` y `/readyz` a la API. `/metrics` permanece en la red interna.

## Preparación

1. Registra un subdominio en DuckDNS y apúntalo a la IP estática del servidor siguiendo [production/duckdns.md](production/duckdns.md).
2. Abre TCP 80/443 y UDP 443. No publiques PostgreSQL, la API ni los puertos internos.
3. Copia `.env.production.example` a `.env.production` y reemplaza todos los marcadores. Usa una contraseña PostgreSQL URL-safe.
4. Genera tokens y secretos con un generador criptográfico; no los confirmes en Git.
5. Ejecuta `docker compose --env-file .env.production -f compose.prod.yaml config --quiet` para validar.
6. Inicia con `docker compose --env-file .env.production -f compose.prod.yaml up -d --build`.

Comprueba `https://TU_NOMBRE.duckdns.org/healthz` y `https://TU_NOMBRE.duckdns.org/readyz`. Revisa después `docker compose --env-file .env.production -f compose.prod.yaml logs -f caddy sentinel-api`.

## Seguridad operativa

- Caddy añade HSTS, CSP, protección contra framing y otras cabeceras defensivas.
- La API exige autenticación, cookies seguras y CORS de origen exacto cuando `SENTINEL_ENVIRONMENT=production`.
- La red Docker usa `172.30.0.0/24`; sólo ese proxy puede aportar `X-Forwarded-For` a los límites por IP.
- API y web no publican puertos del host. PostgreSQL tampoco se expone.
- Respalda el volumen `stellar_postgres_data` y prueba restauraciones periódicamente.
- Para actualizar: descarga o construye imágenes, ejecuta las pruebas de CI y recrea servicios con `docker compose ... up -d`.

Los certificados y la configuración de ACME se conservan en `caddy_data` y `caddy_config`.
