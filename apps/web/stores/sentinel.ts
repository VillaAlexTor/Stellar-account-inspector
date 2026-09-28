import { create } from "zustand";
import {
  SENTINEL_API_URL,
  sentinelAccountPath,
  type MonitoredAccount,
  type SentinelAlert,
  type SentinelConnectionState,
  type SentinelStatus,
} from "@/lib/sentinel";

interface SentinelStore {
	 authState: "checking" | "disabled" | "required" | "authenticated";
  publicKey: string | null;
  account: MonitoredAccount | null;
  status: SentinelConnectionState;
  statusDetail: string | null;
  alerts: SentinelAlert[];
  error: string | null;
	 checkAuth: () => Promise<SentinelStore["authState"]>;
	 authenticate: (token: string) => Promise<boolean>;
	 logout: () => Promise<void>;
  start: (publicKey: string) => Promise<void>;
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
	 authState: "checking",
  publicKey: null,
  account: null,
  status: "idle",
  statusDetail: null,
  alerts: [],
  error: null,

	checkAuth: async () => {
		try {
			const response = await fetch(`${SENTINEL_API_URL}/api/v1/auth/session`, {
				headers: { Accept: "application/json" },
				credentials: "include",
			});
			if (!response.ok) throw new Error("No fue posible comprobar la sesión Sentinel.");
			const payload = (await response.json()) as { enabled: boolean; authenticated: boolean };
			const authState = payload.enabled
				? payload.authenticated ? "authenticated" : "required"
				: "disabled";
			set({ authState, error: null });
			return authState;
		} catch (reason) {
			set({
				authState: "required",
				error: reason instanceof Error ? reason.message : "No fue posible comprobar la sesión Sentinel.",
			});
			return "required";
		}
	},

	authenticate: async (token) => {
		try {
			const response = await fetch(`${SENTINEL_API_URL}/api/v1/auth/session`, {
				method: "POST",
				headers: { "Content-Type": "application/json", Accept: "application/json" },
				credentials: "include",
				body: JSON.stringify({ token }),
			});
			if (!response.ok) {
				const payload = (await response.json().catch(() => null)) as { error?: string } | null;
				throw new Error(payload?.error ?? "Sentinel rechazó el token de acceso.");
			}
			const payload = (await response.json()) as { enabled: boolean };
			set({ authState: payload.enabled ? "authenticated" : "disabled", error: null });
			return true;
		} catch (reason) {
			set({
				authState: "required",
				error: reason instanceof Error ? reason.message : "No fue posible iniciar la sesión Sentinel.",
			});
			return false;
		}
	},

	logout: async () => {
		closeSource();
		activeGeneration += 1;
		try {
			await fetch(`${SENTINEL_API_URL}/api/v1/auth/session`, {
				method: "DELETE",
				credentials: "include",
			});
		} finally {
			set({
				authState: "required",
				account: null,
				status: "idle",
				statusDetail: null,
				alerts: [],
				error: null,
			});
		}
	},

  start: async (publicKey) => {
    closeSource();
    const generation = ++activeGeneration;
    set({
      publicKey,
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
		credentials: "include",
        body: JSON.stringify({ publicKey }),
      });
      if (!monitorResponse.ok) {
        const payload = (await monitorResponse.json().catch(() => null)) as { error?: string } | null;
		if (monitorResponse.status === 401) {
			set({ authState: "required" });
		}
		throw new Error(payload?.error ?? `Sentinel respondió con estado ${monitorResponse.status}.`);
      }
      const monitor = (await monitorResponse.json()) as {
        account: MonitoredAccount;
        status: SentinelStatus;
      };

      const loadHistory = async () => {
        const alertsResponse = await fetch(`${sentinelAccountPath(publicKey)}/alerts?limit=100`, {
          headers: { Accept: "application/json" },
		  credentials: "include",
        });
        if (!alertsResponse.ok) {
          throw new Error("La sesión inició, pero el historial de alertas no pudo cargarse.");
        }
        return (await alertsResponse.json()) as { alerts: SentinelAlert[] };
      };
      const history = await loadHistory();
      if (generation !== activeGeneration) return;
      set({
        account: monitor.account,
        alerts: newestFirst(history.alerts),
        status: monitor.status.state,
        statusDetail: monitor.status.detail ?? null,
      });

	  const source = new EventSource(`${sentinelAccountPath(publicKey)}/events`, {
		withCredentials: true,
	  });
      activeSource = source;
      source.addEventListener("status", (event) => {
        if (generation !== activeGeneration) return;
        const status = JSON.parse((event as MessageEvent<string>).data) as SentinelStatus;
        set({ status: status.state, statusDetail: status.detail ?? null, error: null });
        if (status.state === "connected") {
          void loadHistory().then((latest) => {
            if (generation !== activeGeneration) return;
            set({ alerts: newestFirst(latest.alerts) });
          }).catch(() => undefined);
        }
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
      const rawMessage = reason instanceof Error ? reason.message : "";
      set({
        status: "down",
        statusDetail: null,
        error: rawMessage === "Failed to fetch" || rawMessage === "Load failed"
          ? "El servicio Sentinel no responde. Reintenta; si trabajas en local, comprueba la API y PostgreSQL."
          : rawMessage || "No fue posible iniciar Sentinel.",
      });
    }
  },

  disconnect: () => {
    activeGeneration += 1;
    closeSource();
    set({ status: "idle", statusDetail: "Monitoreo detenido en este navegador", error: null });
  },
}));
