# Seguridad

## Reporte responsable

No abras un issue público si encuentras una vulnerabilidad explotable o un secreto. Envía al mantenedor una descripción, versión o commit afectado, pasos mínimos de reproducción e impacto estimado mediante un canal privado del repositorio. No incluyas seeds, claves privadas Stellar, tokens reales ni datos de terceros.

Se intentará confirmar la recepción en 72 horas y publicar una corrección coordinada según severidad. Las versiones sin soporte pueden requerir actualizar a la última versión.

## Alcance y modelo de confianza

- Inspector sólo acepta claves públicas y consulta Horizon; nunca necesita una seed phrase.
- Sentinel es un servicio para uno o varios operadores de confianza. Un token válido puede consultar todas las cuentas registradas; no implementa aislamiento multi-tenant ni roles.
- Quien administra el host, Docker o PostgreSQL es de confianza y puede acceder a variables de entorno y datos persistidos.
- Los receptores de webhook deben validar HMAC y deduplicar `X-Sentinel-Delivery-ID`.

## Requisitos de producción

Usa únicamente HTTPS, tokens aleatorios, un secreto de sesión independiente, PostgreSQL sin puerto público y copias de seguridad probadas. Mantén Go, Node, imágenes base y dependencias actualizados; ejecuta CI, CodeQL, `pnpm audit --prod` y `govulncheck` antes de publicar.
