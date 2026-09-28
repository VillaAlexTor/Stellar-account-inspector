# Demostración local sin costo

Este es el flujo recomendado para mostrar Stellar Account Inspector desde el navegador en la misma computadora. No requiere dominio, servidor externo, HTTPS público ni configuración del router.

## Requisitos

- Docker Desktop iniciado.
- Puertos locales 5433 y 8081 disponibles para PostgreSQL y la API.
- Un puerto para la web: 3000 de forma predeterminada o uno alternativo mediante `WEB_HOST_PORT`.

## Iniciar el sistema

En PowerShell, desde la raíz del repositorio:

```powershell
$env:WEB_HOST_PORT = "3001"
$env:SENTINEL_ALLOWED_ORIGINS = "http://localhost:3001"
docker compose --env-file .env.production config --quiet
docker compose --env-file .env.production up -d --build
```

Si el puerto 3000 está libre puedes omitir ambas variables y abrir `http://localhost:3000`.

El archivo `.env.production` aporta los tokens de acceso y el secreto de sesión, pero el Compose local conserva cookies HTTP y ambiente de desarrollo para funcionar correctamente en `localhost`.

## Comprobar servicios

```powershell
docker compose ps
Invoke-RestMethod http://localhost:8081/healthz
Invoke-RestMethod http://localhost:8081/readyz
```

Abre `http://localhost:3001` e inicia sesión con uno de los valores definidos en `SENTINEL_AUTH_TOKENS`.

## Prueba funcional

1. Abre el Inspector y consulta una cuenta pública Stellar Testnet.
2. Comprueba el Risk Score y las reglas activadas.
3. Añade la cuenta a Sentinel.
4. Abre su vista de monitorización.
5. Ejecuta una operación controlada en Testnet y confirma que el evento SSE y la alerta aparezcan sin recargar la página.

## Detener

```powershell
docker compose down
```

No agregues `-v` salvo que quieras borrar también los datos locales de PostgreSQL.
