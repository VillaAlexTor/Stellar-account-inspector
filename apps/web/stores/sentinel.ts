import { create } from "zustand";
import {
  SENTINEL_API_URL,
  sentinelAccountPath,
  type MonitoredAccount,
  type SentinelAlert,
  type SentinelConnectionState,
  type SentinelStatus,
} from "@/lib/sentinel";
import type { StellarNetwork } from "@/lib/stellar";

interface SentinelStore {
  publicKey: string | null;
  network: StellarNetwork;
  account: MonitoredAccount | null;
  status: SentinelConnectionState;
  statusDetail: string | null;
  alerts: SentinelAlert[];
  error: string | null;
  start: (publicKey: string, network: StellarNetwork) => Promise<void>;
  disconnect: () => void;
}

let activeSource: EventSource | null = null;
let activeGeneration = 0;

function closeSource() {
  activeSource?.close();
  activeSource = null;
}

function newestFirst(alerts: SentinelAlert[]): SentinelAlert[] {
  return [...alerts].sort((left, right) => {
    const timeDifference = new Date(right.createdAt).getTime() - new Date(left.createdAt).getTime();
    return timeDifference || right.id - left.id;
  });
}

export const useSentinelStore = create<SentinelStore>((set, get) => ({
  publicKey: null,
  network: "testnet",
  account: null,
  status: "idle",
  statusDetail: null,
  alerts: [],
  error: null,

  start: async (publicKey, network) => {
    closeSource();
    const generation = ++activeGeneration;
    set({
      publicKey,
      network,
      account: null,
      status: "connecting",
      statusDetail: "Registrando la sesión de monitoreo",
      alerts: [],
      error: null,
    });

    try {
      const monitorResponse = await fetch(`${SENTINEL_API_URL}/api/v1/monitored-accounts`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify({ publicKey, network }),
      });
      if (!monitorResponse.ok) {
        const payload = (await monitorResponse.json().catch(() => null)) as { error?: string } | null;
        throw new Error(payload?.error ?? `Sentinel respondió con estado ${monitorResponse.status}.`);
      }
      const monitor = (await monitorResponse.json()) as {
        account: MonitoredAccount;
        status: SentinelStatus;
      };

      const alertsResponse = await fetch(`${sentinelAccountPath(publicKey)}/alerts?network=${network}&limit=100`, {
        headers: { Accept: "application/json" },
      });
      if (!alertsResponse.ok) {
        throw new Error("La sesión inició, pero el historial de alertas no pudo cargarse.");
      }
      const history = (await alertsResponse.json()) as { alerts: SentinelAlert[] };
      if (generation !== activeGeneration) return;
      set({
        account: monitor.account,
        alerts: newestFirst(history.alerts),
        status: monitor.status.state,
        statusDetail: monitor.status.detail ?? null,
      });

      const source = new EventSource(`${sentinelAccountPath(publicKey)}/events?network=${network}`);
      activeSource = source;
      source.addEventListener("status", (event) => {
        if (generation !== activeGeneration) return;
        const status = JSON.parse((event as MessageEvent<string>).data) as SentinelStatus;
        set({ status: status.state, statusDetail: status.detail ?? null, error: null });
      });
      source.addEventListener("alert", (event) => {
        if (generation !== activeGeneration) return;
        const alert = JSON.parse((event as MessageEvent<string>).data) as SentinelAlert;
        const current = get().alerts.filter((item) => item.id !== alert.id);
        set({ alerts: newestFirst([alert, ...current]), error: null });
      });
      source.onerror = () => {
        if (generation !== activeGeneration || get().status === "down") return;
        set({ status: "reconnecting", statusDetail: "Restableciendo el canal con Sentinel" });
      };
    } catch (reason) {
      if (generation !== activeGeneration) return;
      closeSource();
      set({
        status: "down",
        statusDetail: null,
        error: reason instanceof Error ? reason.message : "No fue posible iniciar Sentinel.",
      });
    }
  },

  disconnect: () => {
    activeGeneration += 1;
    closeSource();
    set({ status: "idle", statusDetail: "Monitoreo detenido en este navegador", error: null });
  },
}));
