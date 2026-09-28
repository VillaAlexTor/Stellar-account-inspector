# AWS con créditos y DuckDNS gratuito

## Aclaración importante

AWS no ofrece un dominio registrable gratuito. Route 53 cobra el registro anual según el TLD y una zona alojada cuesta aparte; además, los créditos promocionales no pagan el registro de dominios. AWS sí asigna un hostname público a EC2, pero no conviene usarlo como dominio definitivo de Caddy porque no controlas la zona DNS.

El plan oficial de este proyecto no usa Route 53. Para mantener costo inicial cero usa:

- **Servidor:** Amazon Lightsail o EC2 cubierto por los créditos de una cuenta AWS nueva.
- **Dominio:** un subdominio gratuito de DuckDNS, por ejemplo `stellar-villa.duckdns.org`.
- **HTTPS:** Caddy, ya incluido en `compose.prod.yaml`.

La modalidad gratuita actual de AWS es temporal: las cuentas nuevas reciben créditos y el plan gratuito termina al agotarlos o al cumplirse su periodo. Configura alertas de presupuesto desde el primer día.

## 1. Crear la cuenta y proteger costos

1. Crea una cuenta AWS y elige el plan Free si está disponible para tu cuenta.
2. Activa MFA para el usuario raíz.
3. En **Billing and Cost Management → Budgets**, crea alertas al 50%, 80% y 100% del crédito disponible.
4. No crees NAT Gateway, Load Balancer ni RDS para esta instalación pequeña: pueden consumir el crédito rápidamente.

## 2. Crear el servidor

La opción más sencilla es Lightsail:

1. Abre **Lightsail → Create instance**.
2. Elige Linux/Unix, **OS Only → Ubuntu 24.04 LTS**.
3. Selecciona una instancia con al menos **2 GB de RAM**. PostgreSQL, Next.js, Go, Caddy y Alloy no deberían compartir una máquina de 512 MB.
4. Crea y adjunta una IP estática de Lightsail.
5. En Networking permite TCP 22 sólo desde tu IP y TCP 80/443 desde Internet. No abras 5432, 8080, 3000, 9090 ni 12345.

En EC2 aplica la misma regla con un Security Group. Recuerda que AWS cobra las direcciones IPv4 públicas; una IP dinámica puede cambiar al detener la instancia y una Elastic IP también tiene cargo.

## 3. Conseguir el subdominio gratuito

Sigue [duckdns.md](duckdns.md). Elige un nombre disponible, apunta la IP estática y verifica la resolución antes de iniciar Caddy.

## 4. Preparar Ubuntu

Conéctate por SSH y ejecuta:

```bash
sudo apt update
sudo apt upgrade -y
sudo apt install -y git ca-certificates curl
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker "$USER"
```

Cierra y abre la sesión SSH, clona el repositorio y entra en él:

```bash
git clone https://github.com/VillaAlexTor/Stellar-account-inspector.git
cd Stellar-account-inspector
```

Completa `.env.production` siguiendo [environment.md](environment.md). Después valida y despliega:

```bash
docker compose --env-file .env.production -f compose.prod.yaml config --quiet
docker compose --env-file .env.production -f compose.prod.yaml up -d --build
docker compose --env-file .env.production -f compose.prod.yaml ps
```

Comprueba:

```bash
curl -I "https://${DOMAIN}/healthz"
curl -I "https://${DOMAIN}/readyz"
```

## 5. Evitar cargos inesperados

- Revisa **Billing → Free Tier** semanalmente.
- Conserva una sola IP pública y una sola instancia.
- No expongas PostgreSQL ni copies backups a buckets sin política de retención.
- Al terminar la prueba, elimina snapshots, discos, IPs y recursos que ya no uses; detener una instancia no elimina todos los cargos.

Fuentes oficiales: [AWS Free Tier](https://aws.amazon.com/free/free-tier-faqs/), [precios Route 53](https://aws.amazon.com/route53/pricing/), [direcciones EC2](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-instance-addressing.html) y [primeros pasos Lightsail](https://docs.aws.amazon.com/lightsail/latest/userguide/getting-started-with-amazon-lightsail.html).
