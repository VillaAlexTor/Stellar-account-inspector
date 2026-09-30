import Link from "next/link";
import { Activity, SearchCheck } from "lucide-react";

export function AppHeader({
  statusLabel,
}: {
  statusLabel: string;
}) {
  return (
    <header className="app-header">
      <Link className="brand" href="/" aria-label="Stellar Account Inspector, inicio">
        <span className="brand__mark" aria-hidden="true"><i /><i /></span>
        <span><strong>Stellar</strong><small>Account Inspector</small></span>
      </Link>
      <nav className="level-nav" aria-label="Niveles de la aplicación">
        <Link className="is-active" href="/#inspector">
          <SearchCheck size={18} /><span>Inspector</span><small>Lectura</small>
        </Link>
        <Link className="is-enabled" href="/#risk-score">
          <Activity size={18} /><span>Risk Score</span><small>Auditoría</small>
        </Link>
      </nav>
      <div className="header-status"><span className="status-dot" />{statusLabel}</div>
    </header>
  );
}
