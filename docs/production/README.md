# Guías operativas

La demostración se ejecuta solamente en `localhost`; no necesita proveedor cloud, dominio ni configuración del router. Empieza por [la guía local](../local-demo.md).

Guías complementarias:

1. [Variables y secretos](environment.md): administra el archivo local `.env.production`.
2. [Canales de notificación](notifications-setup.md): configura webhook, Telegram o SMTP si deseas demostrar alertas externas.
3. [Observabilidad externa](external-observability.md): conecta métricas y logs a Grafana Cloud Free de forma opcional.
4. Ejecuta el [checklist general](../production-checklist.md).

El archivo `.env.production` es deliberadamente un **archivo**, no una carpeta: Docker Compose lo consume con `--env-file`. Está ignorado por Git y nunca debe contenerse en un commit.
