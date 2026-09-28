# Guías de producción

Estas guías son opcionales y sólo aplican si en el futuro se decide publicar el sistema en Internet. Para la demostración actual sin costo usa [la guía local](../local-demo.md); no necesitas AWS ni DuckDNS.

Sigue estas guías en orden:

1. [AWS con créditos](aws-free-tier.md): crea el servidor y reserva una IP estática.
2. [DuckDNS gratuito](duckdns.md): registra el subdominio y apunta la IP sin comprar un dominio.
3. [Variables y secretos](environment.md): completa el archivo local `.env.production`.
4. [Canales de notificación](notifications-setup.md): configura webhook, Telegram o Amazon SES.
5. [Observabilidad externa](external-observability.md): conecta métricas y logs a Grafana Cloud Free.
6. Ejecuta el [checklist general](../production-checklist.md).

El archivo `.env.production` es deliberadamente un **archivo**, no una carpeta: Docker Compose lo consume con `--env-file`. Está ignorado por Git y nunca debe contenerse en un commit.
