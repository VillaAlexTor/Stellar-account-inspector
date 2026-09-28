import { StellarAccountVerifier } from "@/components/StellarAccountVerifier";
import { AppHeader } from "@/components/layout/AppHeader";

export default function Home() {
  return (
    <main>
      <AppHeader active="inspector" statusLabel="Consulta vía Horizon" />

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
