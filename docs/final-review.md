# Revisión final del proyecto

Fecha de revisión: 2026-09-28.

## Fases A–F

| Fase | Resultado | Evidencia principal |
| --- | --- | --- |
| A · scaffold, diseño y Nivel 1 | Completa | Next.js App Router, selector Testnet/Mainnet, validación de clave pública, flags, balances, trustlines, firmantes y detección multisig. |
| B · Risk Score | Completa | Cinco reglas puras, pesos configurables, UI explicativa, fixtures unitarias y script que provisiona tres escenarios reproducibles en Testnet sin imprimir seeds. |
| C · Go, PostgreSQL, GORM | Completa | Servicio Go, migraciones, `MonitoredAccount`, índices por cuenta/red y Compose local compatible con DBeaver. |
| D · Horizon SSE | Completa | Una sesión compartida por cuenta/red, cursor persistido, reconexión exponencial, parser SSE acotado y restauración tras reinicio. |
| E · estado, reglas y alertas | Completa | Estado previo persistido, reglas de firmantes, pesos, umbrales, trustlines y master key; operación, alerta, cursor y outbox se guardan transaccionalmente. |
| F · SSE navegador y UI | Completa | Ruta dinámica Sentinel, store Zustand independiente, estados de conexión, historial, timeline, severidad, autenticación y reconexión del navegador. |

## Endurecimiento 1–8

| Punto | Resultado | Verificación |
| --- | --- | --- |
| 1. Integración PostgreSQL y SSE | Completo | Prueba real contra PostgreSQL 17 con Horizon SSE simulado, navegador SSE, persistencia y cursor. |
| 2. Autenticación y límites | Completo | Bearer/cookie HMAC `HttpOnly`/`SameSite=Strict`, modo producción obligatorio, límites por IP/cuenta y proxies explícitos. |
| 3. Paginación y retención | Completo tras revisión | Cursor descendente estable, límites 1–200, purga configurable de alertas/operaciones y limpieza de entregas asociadas. |
| 4. Métricas, logs y observabilidad | Completo | Logs JSON y request ID; `/healthz`, `/readyz`, `/metrics`; métricas HTTP, auth, límites, SSE, Horizon, DB, alertas y notificaciones. |
| 5. Notificaciones | Completo | Outbox PostgreSQL atómico, reintentos, recuperación de trabajos, webhook HMAC, Telegram y SMTP TLS. |
| 6. Despliegue | Completo para demostración local | Imágenes no privilegiadas, Docker Compose, autenticación y guía de ejecución en `localhost`. La publicación HTTPS externa queda fuera del alcance actual. |
| 7. CI/CD | Completo en código | GitHub Actions para lint, tests, race, integración, auditoría, contenedores, CodeQL y publicación GHCR por tag. Se activará al subir al repositorio. |
| 8. Seguridad y documentación | Completo | Auditoría, política de reporte, checklist productivo, escáneres sin hallazgos conocidos y riesgos residuales documentados. |

## Pruebas ejecutadas

- `pnpm lint`, `pnpm test` (14 pruebas) y `pnpm build`: correctos.
- `go vet ./...` y `go test ./...`: correctos.
- `go test -tags=integration ./internal/integration -count=1 -v`: correcto contra el PostgreSQL de Docker.
- `pnpm audit --prod`: sin vulnerabilidades conocidas.
- `govulncheck` con el runtime productivo Go 1.25.14: sin vulnerabilidades alcanzables.
- Builds Docker de API y web: correctos; web standalone respondió HTTP 200.
- `docker compose config --quiet` local y producción: correctos.

## Revisión transversal

La arquitectura conserva la separación original: Inspector y Risk Score funcionan en el navegador; Sentinel concentra streaming, estado, persistencia, seguridad y operación en Go. No se almacenan claves privadas. Las escrituras sensibles son transaccionales, los streams tienen límites y reconexión, el historial no depende de offsets inestables y las métricas evitan claves públicas como etiquetas.

Quedan como acciones de operación, no como código pendiente: cargar secretos reales, apuntar DNS, comprobar la emisión ACME, configurar un proveedor de notificaciones, habilitar Actions/CodeQL en GitHub y ejecutar el checklist de producción. Para un futuro SaaS multi-tenant harían falta identidades individuales, autorización por propietario y rate limiting compartido.
