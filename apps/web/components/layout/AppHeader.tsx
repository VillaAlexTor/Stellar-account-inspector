import Link from "next/link";
import { Activity, BellRing, SearchCheck } from "lucide-react";
import { cn } from "@/lib/utils";

type AppLevel = "inspector" | "sentinel";

export function AppHeader({
  active,
  sentinelHref = "#inspector",
  statusLabel,
}: {
  active: AppLevel;
  sentinelHref?: string;
  statusLabel: string;
}) {
  return (
    <header className="app-header">
      <Link className="brand" href="/" aria-label="Stellar Account Inspector, inicio">
        <span className="brand__mark" aria-hidden="true"><i /><i /></span>
        <span><strong>Stellar</strong><small>Account Inspector</small></span>
      </Link>
      <nav className="level-nav" aria-label="Niveles de la aplicación">
        <Link className={cn(active === "inspector" ? "is-active" : "is-enabled")} href="/#inspector">
          <SearchCheck size={18} /><span>Inspector</span><small>Lectura</small>
        </Link>
        <Link className="is-enabled" href="/#risk-score">
          <Activity size={18} /><span>Risk Score</span><small>Auditoría</small>
        </Link>
        <Link className={cn(active === "sentinel" ? "is-active" : "is-enabled")} href={sentinelHref}>
          <BellRing size={18} /><span>Sentinel</span><small>{active === "sentinel" ? "Activo" : "Monitorear"}</small>
        </Link>
      </nav>
      <div className="header-status"><span className="status-dot" />{statusLabel}</div>
    </header>
  );
}
