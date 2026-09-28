# Arquitectura de Sentinel

```text
Horizon SSE ──> sesión Go por cuenta/red ──> evaluador de reglas
                         │                         │
                         │                         └─> alertas SSE al navegador
                         └─> GORM / PostgreSQL
                              ├─ monitored_accounts
                              ├─ relevant_operations
                              ├─ sentinel_alerts
                              └─ notification_deliveries ──> webhook / Telegram / SMTP
```

El `Manager` deduplica sesiones usando `network:publicKey`. Cada sesión parte de un estado conocido de firmantes, umbrales y emisores. Después de cada operación, guarda en una sola transacción el cursor, el estado siguiente, la operación relevante, sus alertas y la bandeja de notificaciones. El despachador procesa esa bandeja fuera de la transacción con reintentos idempotentes.

Al reiniciar, Sentinel restaura `LastCursor` y `StateJSON`, por lo que puede reanudar el stream sin comparar operaciones antiguas contra un snapshot actual. Si todavía no existe checkpoint, obtiene el estado de la cuenta desde Horizon y comienza en `cursor=now`.

## Decisiones

- SSE sobre WebSocket: el flujo es unidireccional y SSE aporta reconexión nativa con menor superficie operativa.
- Sesiones compartidas: varios navegadores consumen una sola conexión persistente hacia Horizon.
- Estado persistido: evita falsos positivos después de reinicios y permite interpretar cambios respecto al estado anterior.
- Alertas explicativas: ninguna severidad depende sólo del color; cada evento conserva regla, mensaje, operación y fecha.
- Paginación por cursor: el historial usa IDs descendentes para evitar saltos cuando entran nuevas alertas.
- Retención separada: alertas y operaciones tienen ventanas configurables; las entregas dependientes se eliminan antes de su alerta.
