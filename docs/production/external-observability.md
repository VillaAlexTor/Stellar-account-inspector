# Conectar observabilidad externa

La opción más sencilla para este proyecto es **Grafana Cloud Free + Grafana Alloy**. Grafana aloja Prometheus, Loki, dashboards y alertas; Alloy corre junto a Sentinel, lee `/metrics` por la red Docker y envía también los logs de los contenedores. No se publica `/metrics` en Internet.

El plan Free actual ofrece una cuota permanente para proyectos pequeños, con límites de series, logs y retención. Revisa periódicamente **Cost Management and Billing → Usage** para no superar la cuota.

## 1. Crear la cuenta Grafana Cloud

1. Regístrate en <https://grafana.com/auth/sign-up/create-user>.
2. Crea un stack gratuito.
3. En **Connections → Add new connection**, busca **Hosted Prometheus / Send metrics**.
4. Copia:
   - Remote Write URL → `GRAFANA_CLOUD_PROM_URL`.
   - Username / Instance ID → `GRAFANA_CLOUD_PROM_USER`.
5. Abre la configuración de **Hosted Logs / Loki** y copia:
   - URL de escritura Loki → `GRAFANA_CLOUD_LOKI_URL`.
   - Username / Instance ID → `GRAFANA_CLOUD_LOKI_USER`.
6. Crea un access policy token con permisos de escritura de métricas y logs. Guárdalo como `GRAFANA_CLOUD_API_KEY`; sólo se muestra una vez.

## 2. Preparar la autenticación de métricas

Genera un token independiente y agrégalo tanto a la lista aceptada por Sentinel como a Alloy:

```dotenv
SENTINEL_AUTH_TOKENS=token-del-operador,token-exclusivo-de-metricas
SENTINEL_METRICS_TOKEN=token-exclusivo-de-metricas
```

Completa en `.env.production` las seis variables de Grafana. No incluyas comillas ni espacios alrededor de `=`.

## 3. Validar Alloy

El repositorio incluye `deploy/observability/config.alloy` y `compose.observability.yaml`. Antes de iniciarlo:

```bash
docker run --rm \
  -e SENTINEL_METRICS_TOKEN=dummy \
  -e GRAFANA_CLOUD_PROM_URL=https://example.invalid/api/prom/push \
  -e GRAFANA_CLOUD_PROM_USER=dummy \
  -e GRAFANA_CLOUD_LOKI_URL=https://example.invalid/loki/api/v1/push \
  -e GRAFANA_CLOUD_LOKI_USER=dummy \
  -e GRAFANA_CLOUD_API_KEY=dummy \
  -v "$PWD/deploy/observability/config.alloy:/etc/alloy/config.alloy:ro" \
  grafana/alloy:latest validate /etc/alloy/config.alloy
```

## 4. Iniciar producción con Alloy

```bash
docker compose --env-file .env.production \
  -f compose.yaml -f compose.observability.yaml \
  config --quiet

docker compose --env-file .env.production \
  -f compose.yaml -f compose.observability.yaml \
  up -d
```

Comprueba el recolector:

```bash
docker compose --env-file .env.production \
  -f compose.yaml -f compose.observability.yaml \
  logs --tail=100 alloy
```

No publiques el puerto 12345; sirve sólo para diagnóstico interno.

## 5. Verificar datos en Grafana

En **Explore → Metrics**, ejecuta:

```promql
sentinel_database_ready
```

Después:

```promql
sum by (status) (rate(sentinel_http_requests_total[5m]))
```

En **Explore → Logs** selecciona Loki y consulta:

```logql
{application="stellar-account-inspector"}
```

Los logs de la API son JSON, por lo que puedes filtrar errores con:

```logql
{service_name=~".*sentinel-api.*"} | json | level="ERROR"
```

## 6. Crear alertas mínimas

Crea reglas administradas para:

- `sentinel_database_ready == 0` durante 2 minutos.
- porcentaje de HTTP 5xx mayor al 2% durante 10 minutos.
- incremento de `sentinel_horizon_stream_events_total{event="stream_error"}`.
- incremento de `sentinel_notification_deliveries_total{result="failed"}`.
- ausencia de la métrica `up{job="stellar-sentinel"}` durante 5 minutos.

Configura el punto de contacto de Grafana con un correo o Telegram diferente del canal que vigila Sentinel, para detectar también fallos de ese proveedor.

## Seguridad y límites

- El montaje `/var/run/docker.sock` permite descubrir y leer logs, pero también amplía el impacto de una vulnerabilidad en Alloy. Mantén la imagen actualizada y el socket en modo sólo lectura.
- No coloques claves públicas Stellar, tokens o URLs con credenciales como etiquetas Prometheus: producirían cardinalidad o exposición.
- Usa un token de Grafana dedicado y rótalo si aparece en logs o terminales compartidos.
- El volumen `alloy_data` conserva la WAL para reenviar métricas después de una interrupción.

Referencias oficiales: [Grafana Cloud Free](https://grafana.com/docs/grafana/latest/introduction/grafana-cloud/), [`prometheus.remote_write`](https://grafana.com/docs/alloy/latest/reference/components/prometheus/prometheus.remote_write/), [`prometheus.scrape`](https://grafana.com/docs/alloy/latest/reference/components/prometheus/prometheus.scrape/) y [monitorización de Docker con Alloy](https://grafana.com/docs/alloy/latest/monitor/monitor-docker-containers/).
