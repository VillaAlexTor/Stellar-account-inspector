# Lista de salida a producción

## Antes de desplegar

- [ ] DNS A/AAAA apunta al host correcto y 80/443 están disponibles.
- [ ] `.env.production` existe fuera de Git y no contiene valores de ejemplo.
- [ ] Token de acceso, secreto de sesión, contraseña PostgreSQL y secretos de notificación son aleatorios e independientes.
- [ ] Webhook y API Telegram usan HTTPS; SMTP usa STARTTLS o TLS implícito.
- [ ] `docker compose --env-file .env.production -f compose.prod.yaml config --quiet` pasa.
- [ ] Lint, pruebas frontend, pruebas Go, integración PostgreSQL/SSE, auditorías y builds de imágenes pasan.
- [ ] Se definieron ventanas de retención acordes a capacidad y obligaciones aplicables.
- [ ] Existe un respaldo cifrado y se probó una restauración.

## Después de desplegar

- [ ] `/healthz` y `/readyz` responden 200 a través de HTTPS.
- [ ] HTTP redirige a HTTPS y el certificado coincide con el dominio.
- [ ] Una sesión inválida responde 401 y el límite devuelve 429 al excederse.
- [ ] `/metrics` no es accesible desde Internet; Prometheus lo obtiene por red interna con autenticación.
- [ ] Se generó una alerta controlada y llegó por cada canal configurado una sola vez.
- [ ] Logs incluyen `request_id` y no contienen tokens, contraseñas ni seeds.
- [ ] Las alertas operativas cubren readiness, errores 5xx, reconexiones Horizon y fallos terminales de notificación.

## Operación continua

- [ ] Dependabot y CodeQL están habilitados en GitHub.
- [ ] Se revisan actualizaciones y hallazgos semanalmente.
- [ ] Se rota acceso tras cambios de personal o sospecha de exposición.
- [ ] Se prueba restauración y recuperación de incidentes periódicamente.
