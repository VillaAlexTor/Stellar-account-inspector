# Observabilidad de Sentinel

Sentinel combina logs JSON correlacionados, métricas Prometheus y sondas separadas de vida y disponibilidad. Las rutas de métricas usan plantillas como `GET /api/v1/monitored-accounts/{publicKey}` para evitar cardinalidad ilimitada y no copiar claves públicas a las etiquetas.

## Endpoints operativos

- `GET /healthz`: confirma que el proceso HTTP está vivo. No consulta dependencias.
- `GET /readyz`: ejecuta un `PingContext` contra PostgreSQL; devuelve `503` si la instancia no debe recibir tráfico.
- `GET /metrics`: exposición Prometheus/OpenMetrics. Queda protegida cuando `SENTINEL_AUTH_TOKENS` está configurado.

Ejemplo de scrape manual:

```bash
curl -H "Authorization: Bearer $SENTINEL_TOKEN" http://localhost:8081/metrics
```

Configuración mínima de Prometheus:

```yaml
scrape_configs:
  - job_name: stellar-sentinel
    metrics_path: /metrics
    static_configs:
      - targets: ["sentinel-api:8080"]
    authorization:
      type: Bearer
      credentials_file: /run/secrets/sentinel_metrics_token
```

No guardes el token directamente en el YAML ni lo incluyas en una URL.

## Métricas propias

| Métrica | Tipo | Uso |
| --- | --- | --- |
| `sentinel_http_requests_total` | Counter | Tráfico por método, ruta normalizada y estado. |
| `sentinel_http_request_duration_seconds` | Histogram | Latencia HTTP por método y ruta. |
| `sentinel_http_requests_in_flight` | Gauge | Solicitudes actualmente en curso. |
| `sentinel_sse_connections` | Gauge | Navegadores con un stream SSE activo. |
| `sentinel_auth_attempts_total` | Counter | Autenticaciones exitosas, inválidas, malformadas o rechazadas. |
| `sentinel_rate_limit_rejections_total` | Counter | Rechazos por alcance `ip` o `account`. |
| `sentinel_horizon_stream_events_total` | Counter | Conexiones, operaciones y errores de Horizon. |
| `sentinel_alerts_total` | Counter | Alertas emitidas por regla y severidad. |
| `sentinel_notification_deliveries_total` | Counter | Envíos, reintentos y fallos terminales por canal. |
| `sentinel_monitored_sessions` | Gauge | Sesiones únicas administradas por el proceso. |
| `sentinel_database_ready` | Gauge | Último resultado de readiness de PostgreSQL. |

También se publican métricas estándar del runtime Go y del proceso.

## Logs y correlación

Los logs se escriben como JSON en `stdout`; Docker u otro runtime debe encargarse de recolectarlos. Cada respuesta HTTP incluye `X-Request-ID`. Un identificador válido enviado por el proxy se conserva; valores inválidos se reemplazan para impedir inyección de campos.

```bash
docker compose logs -f sentinel-api
```

Configura `SENTINEL_LOG_LEVEL` como `debug`, `info`, `warn` o `error`. `SENTINEL_ENVIRONMENT` y `SENTINEL_VERSION` aparecen en todos los registros para separar despliegues.

## Alertas iniciales sugeridas

- Tasa sostenida de respuestas `5xx` mayor al 2% durante 10 minutos.
- Percentil 95 de latencia HTTP mayor a 2 segundos durante 10 minutos.
- `sentinel_database_ready == 0` durante 2 minutos.
- Incremento continuo de `horizon_stream_events_total{event="stream_error"}`.
- Aumento de `rate_limit_rejections_total` o `auth_attempts_total{result="invalid"}` fuera del patrón habitual.

Los umbrales deben ajustarse con tráfico real; no son garantías universales.
