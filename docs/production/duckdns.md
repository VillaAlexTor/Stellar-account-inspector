# Subdominio gratuito con DuckDNS

DuckDNS es el proveedor DNS oficial de este despliegue. El resultado será una dirección pública como `stellar-villa.duckdns.org`, sin comprar un dominio ni contratar Route 53.

## 1. Reservar el nombre

1. Entra en <https://www.duckdns.org/> e inicia sesión con uno de los proveedores admitidos.
2. En **domains**, escribe un nombre corto, por ejemplo `stellar-villa`, y selecciona **add domain**.
3. Copia la IP estática asignada a la instancia de Lightsail.
4. Escribe esa IP en el campo **current ip** del subdominio y selecciona **update ip**.

No pegues el token de DuckDNS en el repositorio, `.env.production`, capturas ni chats. Con la IP estática de Lightsail no hace falta guardar ese token en el servidor: sólo se necesita para cambiar el registro desde DuckDNS.

## 2. Verificar DNS

Desde PowerShell:

```powershell
Resolve-DnsName stellar-villa.duckdns.org
```

Desde Linux:

```bash
getent ahostsv4 stellar-villa.duckdns.org
```

La dirección devuelta debe coincidir con la IP estática de Lightsail. Si acabas de actualizarla, espera unos minutos y repite la consulta.

## 3. Configurar Sentinel

En `.env.production` escribe el nombre completo, sin protocolo ni barra final:

```dotenv
DOMAIN=stellar-villa.duckdns.org
ACME_EMAIL=tu-correo@example.com
```

No uses `https://` dentro de `DOMAIN`. Caddy construye el sitio HTTPS, solicita el certificado y conserva las renovaciones en sus volúmenes.

## 4. Abrir sólo los puertos públicos

En el firewall de Lightsail permite:

- TCP 80 desde Internet, para redirección y validación ACME.
- TCP 443 desde Internet, para HTTPS y SSE.
- UDP 443 desde Internet, para HTTP/3.
- TCP 22 únicamente desde tu IP administrativa.

No abras 5432, 8080, 3000, 9090 ni 12345.

## 5. Validar y publicar

```bash
docker compose --env-file .env.production -f compose.prod.yaml config --quiet
docker compose --env-file .env.production -f compose.prod.yaml up -d --build
curl -I "https://${DOMAIN}/healthz"
curl -I "https://${DOMAIN}/readyz"
```

Si Caddy no obtiene el certificado, comprueba primero que DuckDNS resuelva a la IP correcta y que los puertos 80 y 443 estén abiertos. Después revisa:

```bash
docker compose --env-file .env.production -f compose.prod.yaml logs --tail=100 caddy
```

## Si cambia la IP

Una IP estática correctamente adjunta a Lightsail no debería cambiar al reiniciar la instancia. Si reemplazas la instancia o liberas la IP, actualiza manualmente el valor en DuckDNS antes de levantar Caddy. Este flujo evita almacenar permanentemente el token de DuckDNS en el servidor.

Referencia oficial: [API y actualización DNS de DuckDNS](https://www.duckdns.org/spec.jsp).
