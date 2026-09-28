"use client";

import { useEffect, useMemo, useState } from "react";
import {
  BellRing,
  CircleAlert,
  KeyRound,
  Pause,
  Play,
  Radio,
  RefreshCw,
  ShieldAlert,
  SlidersHorizontal,
  UserRoundPlus,
  WalletCards,
  Wifi,
  WifiOff,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { SentinelConnectionState, SentinelSeverity } from "@/lib/sentinel";
import type { StellarNetwork } from "@/lib/stellar";
import { cn } from "@/lib/utils";
import { useSentinelStore } from "@/stores/sentinel";

const STATUS_COPY: Record<SentinelConnectionState, { label: string; description: string }> = {
  idle: { label: "Detenido", description: "El navegador no recibe alertas en este momento." },
  connecting: { label: "Conectando", description: "Preparando la sesión y el estado base de la cuenta." },
  connected: { label: "Conectado", description: "Horizon, Sentinel y este navegador mantienen un canal activo." },
  reconnecting: { label: "Reconectando", description: "El canal se interrumpió y está intentando recuperarse." },
  down: { label: "Caído", description: "Sentinel no está disponible. Revisa el backend y PostgreSQL." },
};

const SEVERITY_COPY: Record<SentinelSeverity, string> = {
  info: "Info",
  warning: "Advertencia",
  critical: "Crítica",
};

const RULES = [
  { id: "SIGNER_ADDED / REMOVED", label: "Cambios de firmantes", icon: UserRoundPlus },
  { id: "SIGNER_WEIGHT / THRESHOLDS", label: "Pesos y umbrales", icon: SlidersHorizontal },
  { id: "NEW_TRUSTLINE_ISSUER", label: "Emisores nuevos", icon: WalletCards },
  { id: "MASTER_KEY_ZEROED", label: "Posible account takeover", icon: ShieldAlert },
] as const;

export function SentinelConsole({
  publicKey,
  initialNetwork,
}: {
  publicKey: string;
  initialNetwork: StellarNetwork;
}) {
  const [selectedNetwork, setSelectedNetwork] = useState(initialNetwork);
  const status = useSentinelStore((state) => state.status);
  const statusDetail = useSentinelStore((state) => state.statusDetail);
  const alerts = useSentinelStore((state) => state.alerts);
  const error = useSentinelStore((state) => state.error);
  const start = useSentinelStore((state) => state.start);
  const disconnect = useSentinelStore((state) => state.disconnect);

  useEffect(() => {
    void start(publicKey, initialNetwork);
    return () => disconnect();
  }, [disconnect, initialNetwork, publicKey, start]);

  const criticalCount = useMemo(
    () => alerts.filter((alert) => alert.severity === "critical").length,
    [alerts],
  );

  function chooseNetwork(network: StellarNetwork) {
    if (network === selectedNetwork) return;
    setSelectedNetwork(network);
    window.history.replaceState(null, "", `/sentinel/${publicKey}?network=${network}`);
    void start(publicKey, network);
  }

  const statusCopy = STATUS_COPY[status];
  return (
    <div className="sentinel-shell">
      <section className="sentinel-command" aria-labelledby="sentinel-title">
        <div className="sentinel-command__intro">
          <div>
            <h1 id="sentinel-title">Vigila el control de esta cuenta mientras cambia.</h1>
            <p>
              Sentinel mantiene un stream único en el backend, compara cada operación con el estado anterior y conserva evidencia de cambios sensibles en PostgreSQL.
            </p>
          </div>
          <div className={cn("sentinel-signal", `is-${status}`)} aria-live="polite">
            {status === "down" || status === "idle" ? <WifiOff size={19} /> : <Wifi size={19} />}
            <span>Estado del canal</span>
            <strong>{statusCopy.label}</strong>
          </div>
        </div>

        <div className="sentinel-controls">
          <div className="sentinel-account">
            <span className="control-label">Cuenta monitoreada</span>
            <div><KeyRound size={16} aria-hidden="true" /><code title={publicKey}>{publicKey}</code></div>
          </div>
          <div className="network-switch" aria-label="Red que monitorea Sentinel">
            {(["testnet", "mainnet"] as const).map((network) => (
              <button
                aria-pressed={selectedNetwork === network}
                className={cn(selectedNetwork === network && "is-active")}
                key={network}
                onClick={() => chooseNetwork(network)}
                type="button"
              >
                <span className="status-dot" aria-hidden="true" />
                {network === "testnet" ? "Testnet" : "Mainnet"}
              </button>
            ))}
          </div>
          {status === "idle" ? (
            <Button onClick={() => void start(publicKey, selectedNetwork)}>
              <Play size={16} /> Reanudar
            </Button>
          ) : status === "down" ? (
            <Button onClick={() => void start(publicKey, selectedNetwork)}>
              <RefreshCw size={16} /> Reintentar
            </Button>
          ) : (
            <Button variant="outline" onClick={disconnect}>
              <Pause size={16} /> Detener vista
            </Button>
          )}
        </div>

        {error ? (
          <div className="error-module sentinel-error" role="alert">
            <CircleAlert size={19} aria-hidden="true" />
            <div><strong>No se pudo abrir el canal Sentinel</strong><p>{error}</p></div>
          </div>
        ) : null}
      </section>

      <section className="sentinel-dashboard" aria-label="Sesión Sentinel">
        <div className="sentinel-dashboard__status">
          <div className="connection-readout">
            <span className="control-label control-label--light">Conexión</span>
            <strong>{statusCopy.label}</strong>
            <small>{statusDetail ?? statusCopy.description}</small>
          </div>
          <dl className="sentinel-counters">
            <div><dt>Alertas registradas</dt><dd>{alerts.length.toString().padStart(2, "0")}</dd></div>
            <div><dt>Críticas</dt><dd>{criticalCount.toString().padStart(2, "0")}</dd></div>
            <div><dt>Red</dt><dd>{selectedNetwork === "testnet" ? "TEST" : "MAIN"}</dd></div>
          </dl>
        </div>

        <div className="sentinel-workspace">
          <section className="alert-timeline" aria-labelledby="alert-timeline-title">
            <div className="sentinel-panel-heading">
              <span><BellRing size={17} aria-hidden="true" /></span>
              <h2 id="alert-timeline-title">Línea de tiempo de alertas</h2>
              <small>Más reciente primero</small>
            </div>

            {alerts.length === 0 ? (
              <div className="sentinel-empty">
                <div className="sentinel-empty__scope" aria-hidden="true">
                  {Array.from({ length: 18 }, (_, index) => <i key={index} />)}
                </div>
                <div>
                  <Radio size={22} aria-hidden="true" />
                  <h3>{status === "connected" ? "Escuchando operaciones" : "Sin alertas registradas"}</h3>
                  <p>Los cambios de firmantes, umbrales, trustlines y master key aparecerán aquí con su evidencia.</p>
                </div>
              </div>
            ) : (
              <ol className="alert-list">
                {alerts.map((alert) => (
                  <li className={cn("alert-entry", `is-${alert.severity}`)} key={`${alert.id}-${alert.ruleId}`}>
                    <div className="alert-entry__rail" aria-hidden="true"><span /></div>
                    <article>
                      <header>
                        <Badge variant={alert.severity}>{SEVERITY_COPY[alert.severity]}</Badge>
                        <strong>{formatRule(alert.ruleId)}</strong>
                        <time dateTime={alert.createdAt}>{formatDate(alert.createdAt)}</time>
                      </header>
                      <p>{alert.message}</p>
                      <div className="alert-entry__evidence">
                        <span>Operación</span>
                        <code title={alert.operationId}>{alert.operationId}</code>
                        <span>{alert.ruleId}</span>
                      </div>
                    </article>
                  </li>
                ))}
              </ol>
            )}
          </section>

          <aside className="sentinel-rules" aria-labelledby="sentinel-rules-title">
            <div className="sentinel-panel-heading">
              <span><ShieldAlert size={17} aria-hidden="true" /></span>
              <h2 id="sentinel-rules-title">Detectores activos</h2>
            </div>
            <div className="detector-bank">
              {RULES.map(({ id, label, icon: Icon }) => (
                <div className="detector-row" key={id}>
                  <Icon size={17} aria-hidden="true" />
                  <div><strong>{label}</strong><small>{id}</small></div>
                  <span><i /> ON</span>
                </div>
              ))}
            </div>
            <p>
              Las reglas son deterministas y comparan cada operación con el último estado confirmado. Sentinel no firma ni bloquea transacciones.
            </p>
          </aside>
        </div>
      </section>
    </div>
  );
}

function formatRule(ruleId: string): string {
  const labels: Record<string, string> = {
    SIGNER_ADDED: "Firmante agregado",
    SIGNER_REMOVED: "Firmante removido",
    SIGNER_WEIGHT_CHANGED: "Peso de firmante modificado",
    THRESHOLDS_CHANGED: "Umbrales modificados",
    NEW_TRUSTLINE_ISSUER: "Trustline hacia emisor nuevo",
    MASTER_KEY_ZEROED: "Master key reducida a cero",
  };
  return labels[ruleId] ?? ruleId.replaceAll("_", " ");
}

function formatDate(value: string): string {
  return new Intl.DateTimeFormat("es-BO", {
    dateStyle: "medium",
    timeStyle: "medium",
  }).format(new Date(value));
}
