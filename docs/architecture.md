# Arquitectura de Sentinel

```text
Horizon SSE ──> sesión Go por cuenta/red ──> evaluador de reglas
                         │                         │
                         │                         └─> alertas SSE al navegador
                         └─> GORM / PostgreSQL
                              ├─ monitored_accounts
                              ├─ relevant_operations
                              └─ sentinel_alerts
```

El `Manager` deduplica sesiones usando `network:publicKey`. Cada sesión parte de un estado conocido de firmantes, umbrales y emisores. Después de cada operación, guarda en una sola transacción el cursor, el estado siguiente, la operación relevante y sus alertas.

Al reiniciar, Sentinel restaura `LastCursor` y `StateJSON`, por lo que puede reanudar el stream sin comparar operaciones antiguas contra un snapshot actual. Si todavía no existe checkpoint, obtiene el estado de la cuenta desde Horizon y comienza en `cursor=now`.

## Decisiones

- SSE sobre WebSocket: el flujo es unidireccional y SSE aporta reconexión nativa con menor superficie operativa.
- Sesiones compartidas: varios navegadores consumen una sola conexión persistente hacia Horizon.
- Estado persistido: evita falsos positivos después de reinicios y permite interpretar cambios respecto al estado anterior.
- Alertas explicativas: ninguna severidad depende sólo del color; cada evento conserva regla, mensaje, operación y fecha.
