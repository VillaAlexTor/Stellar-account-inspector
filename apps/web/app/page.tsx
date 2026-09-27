import { Activity, BellRing, SearchCheck } from "lucide-react";
import { StellarAccountVerifier } from "@/components/StellarAccountVerifier";

export default function Home() {
  return (
    <main>
      <header className="app-header">
        <a className="brand" href="#top" aria-label="Stellar Account Inspector, inicio">
          <span className="brand__mark" aria-hidden="true"><i /><i /></span>
          <span><strong>Stellar</strong><small>Account Inspector</small></span>
        </a>
        <nav className="level-nav" aria-label="Niveles de la aplicación">
          <a className="is-active" href="#inspector"><SearchCheck size={18} /><span>Inspector</span><small>Activo</small></a>
          <span aria-disabled="true"><Activity size={18} /><span>Risk Score</span><small>Siguiente</small></span>
          <span aria-disabled="true"><BellRing size={18} /><span>Sentinel</span><small>Próximamente</small></span>
        </nav>
        <div className="header-status"><span className="status-dot is-live" />Horizon disponible</div>
      </header>

      <div id="top" className="page-frame">
        <section id="inspector" aria-label="Inspector de cuentas Stellar">
          <StellarAccountVerifier />
        </section>
      </div>

      <footer className="app-footer">
        <span>Lectura pública · Sin custodia · Sin seed phrases</span>
        <a href="https://developers.stellar.org/docs/data/apis/horizon" target="_blank" rel="noreferrer">Documentación de Horizon</a>
      </footer>
    </main>
  );
}
