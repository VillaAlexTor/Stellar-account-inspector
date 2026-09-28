# Despliegue

## Demostración local sin costo

El flujo recomendado para demostrar este proyecto es ejecutarlo en la misma computadora y abrirlo desde el navegador mediante `localhost`. No necesita servicios externos, dominio, HTTPS público ni cambios en el router.

Sigue la guía [local-demo.md](local-demo.md). En el puerto predeterminado la aplicación queda disponible en `http://localhost:3000`; si está ocupado, define `WEB_HOST_PORT`, por ejemplo `3001`.

## Seguridad operativa

- La API exige autenticación y CORS de origen exacto para el navegador local.
- No abras ni reenvíes los puertos del proyecto en el router.
- Los puertos publicados por Docker deben usarse únicamente desde la computadora de demostración.
- Respalda el volumen `stellar_postgres_data` y prueba restauraciones periódicamente.
- Para actualizar: descarga o construye imágenes, ejecuta las pruebas de CI y recrea servicios con `docker compose ... up -d`.
