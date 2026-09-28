# Auditoría de seguridad

Fecha: 2026-09-27. Alcance: frontend Next.js, API Go, PostgreSQL/GORM, SSE, notificaciones, contenedores, proxy HTTPS y pipelines.

## Resultado

No quedan hallazgos críticos o altos conocidos en el alcance revisado. Se corrigieron durante la auditoría:

- crecimiento no acotado del mapa de rate limiting ante identidades rotatorias; ahora usa un bucket de desborde y mantiene memoria acotada;
- redirecciones automáticas en webhook y Telegram, que podían cambiar el destino configurado;
- índices de deduplicación globales; ahora operación y alerta se deduplican dentro de cada cuenta;
- respuestas de sesión potencialmente cacheables; ahora incluyen `Cache-Control: no-store`;
- URLs HTTP para Horizon, webhook o Telegram en producción; la configuración las rechaza;
- imágenes no fijadas a una revisión Go corregida; el build usa Go 1.25.14 y usuario distroless no privilegiado.
- saludo SSE real de Horizon (`"hello"`) no reconocido, que provocaba reconexiones; el parser y su prueba de regresión aceptan ambas variantes.

## Evidencia automatizada

- `pnpm audit --prod`: sin vulnerabilidades conocidas.
- `govulncheck` v1.7.0 ejecutado con Go 1.25.14: sin vulnerabilidades alcanzables.
- `go vet ./...`: correcto.
- suites frontend: 14 pruebas correctas; lint y build Next.js correctos.
- suite Go e integración real PostgreSQL/SSE/notificaciones: correctas.
- imágenes API y web: construidas; la imagen web respondió HTTP 200 como usuario `node`.
- ambos archivos Compose: configuración válida.

El equipo local tenía Go 1.26.5, afectado por avisos recientes de biblioteca estándar. No se usa para la imagen productiva. La base oficial confirma que las ramas 1.25 anteriores a 1.25.13 eran vulnerables; el build quedó fijado en 1.25.14. Referencias: <https://pkg.go.dev/vuln/GO-2026-6218> y <https://pkg.go.dev/vuln/GO-2026-6090>.

## Riesgos residuales aceptados

- El rate limiting vive en memoria por instancia. Un despliegue horizontal necesita un limitador compartido en Redis o en el borde.
- La autenticación es de operador, no multi-tenant: cualquier token válido ve todas las cuentas monitorizadas.
- `AutoMigrate` es apropiado para esta escala, pero una instalación con gran volumen debe migrar índices mediante una ventana controlada.
- Las credenciales se inyectan por variables de entorno Compose. En un orquestador, usa su almacén nativo de secretos.
- SMTP no firma DKIM por sí mismo; esa responsabilidad corresponde al proveedor de correo.

Estos riesgos no impiden el despliegue de instancia única documentado, pero deben reevaluarse antes de ofrecer el servicio a usuarios mutuamente no confiables.
