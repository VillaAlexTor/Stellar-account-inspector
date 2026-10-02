import Link from "next/link";
import { Activity, Presentation, SearchCheck } from "lucide-react";

export function AppHeader({
  statusLabel,
  active = "inspector",
}: {
  statusLabel: string;
  active?: "inspector" | "risk" | "pitch";
}) {
  return (
    <header className="app-header">
      <Link className="brand" href="/" aria-label="Stellar Account Inspector, inicio">
        <span className="brand__mark" aria-hidden="true"><i /><i /></span>
        <span><strong>Stellar</strong><small>Account Inspector</small></span>
      </Link>
      <nav className="level-nav" aria-label="Niveles de la aplicación">
        <Link className={active === "inspector" ? "is-active" : "is-enabled"} href="/#inspector">
          <SearchCheck size={18} /><span>Inspector</span><small>Lectura</small>
        </Link>
        <Link className={active === "risk" ? "is-active" : "is-enabled"} href="/#risk-score">
          <Activity size={18} /><span>Risk Score</span><small>Auditoría</small>
        </Link>
        <Link className={active === "pitch" ? "is-active" : "is-enabled"} href="/pitch">
          <Presentation size={18} /><span>Pitch</span><small>Presentación</small>
        </Link>
      </nav>
      <div className="header-status"><span className="status-dot" />{statusLabel}</div>
    </header>
  );
}
