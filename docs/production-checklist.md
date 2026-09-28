# Lista para la demostración local

## Antes de desplegar

- [ ] `.env.production` existe fuera de Git y no contiene valores de ejemplo.
- [ ] Token de acceso, secreto de sesión, contraseña PostgreSQL y secretos de notificación son aleatorios e independientes.
- [ ] Webhook y API Telegram usan HTTPS; SMTP usa STARTTLS o TLS implícito.
- [ ] `docker compose --env-file .env.production config --quiet` pasa.
- [ ] Lint, pruebas frontend, pruebas Go, integración PostgreSQL/SSE, auditorías y builds de imágenes pasan.
- [ ] Se definieron ventanas de retención acordes a capacidad y obligaciones aplicables.
- [ ] Existe un respaldo cifrado y se probó una restauración.

## Después de desplegar

- [ ] `/healthz` y `/readyz` responden 200 desde `localhost`.
- [ ] La interfaz responde en el puerto local configurado y no está publicada mediante el router.
- [ ] Una sesión inválida responde 401 y el límite devuelve 429 al excederse.
- [ ] `/metrics` exige autenticación.
- [ ] Se generó una alerta controlada y llegó por cada canal configurado una sola vez.
- [ ] Logs incluyen `request_id` y no contienen tokens, contraseñas ni seeds.
- [ ] Las alertas operativas cubren readiness, errores 5xx, reconexiones Horizon y fallos terminales de notificación.

## Operación continua

- [ ] Dependabot y CodeQL están habilitados en GitHub.
- [ ] Se revisan actualizaciones y hallazgos semanalmente.
- [ ] Se rota acceso tras cambios de personal o sospecha de exposición.
- [ ] Se prueba restauración y recuperación de incidentes periódicamente.
