# Variables y secretos de producción

El repositorio ya contiene un `.env.production` local, vacío e ignorado por Git. En otro servidor créalo así:

```bash
cp .env.production.example .env.production
chmod 600 .env.production
```

## Valores obligatorios

| Variable | Cómo obtenerla |
| --- | --- |
| `DOMAIN` | Subdominio gratuito registrado en DuckDNS, por ejemplo `stellar-villa.duckdns.org`. |
| `ACME_EMAIL` | Correo real para avisos de certificados. |
| `POSTGRES_PASSWORD` | Secreto aleatorio URL-safe de 32 bytes o más. |
| `SENTINEL_AUTH_TOKENS` | Uno o más tokens aleatorios separados por coma. |
| `SENTINEL_SESSION_SECRET` | Secreto aleatorio independiente de al menos 32 caracteres. |
| `SENTINEL_VERSION` | Tag o commit desplegado, por ejemplo `v1.0.0`. |

Genera tres valores distintos en Linux:

```bash
openssl rand -hex 32
openssl rand -hex 32
openssl rand -hex 32
```

O en PowerShell:

```powershell
1..3 | ForEach-Object {
  [Convert]::ToHexString([Security.Cryptography.RandomNumberGenerator]::GetBytes(32)).ToLower()
}
```

Usa uno para PostgreSQL, otro para el token de acceso y otro para la sesión. No reutilices secretos.

## Token dedicado de métricas

Genera un cuarto token. Inclúyelo dentro de `SENTINEL_AUTH_TOKENS` y repítelo en `SENTINEL_METRICS_TOKEN`:

```dotenv
SENTINEL_AUTH_TOKENS=token-del-operador,token-exclusivo-de-metricas
SENTINEL_METRICS_TOKEN=token-exclusivo-de-metricas
```

Así Alloy puede leer `/metrics` sin reutilizar el token con el que entras al navegador.

## Imágenes

Mientras no exista un release GHCR conserva:

```dotenv
SENTINEL_API_IMAGE=stellar-sentinel-api:local
SENTINEL_WEB_IMAGE=stellar-sentinel-web:local
```

Después de publicar un tag sustituye esos valores por las imágenes generadas por `release.yml`.

## Validación segura

```bash
docker compose --env-file .env.production -f compose.prod.yaml config --quiet
```

No pegues la salida completa de `docker compose config` en issues o chats: contiene secretos expandidos. Antes de cualquier commit ejecuta `git status --ignored --short` y confirma que `.env.production` aparece ignorado.
