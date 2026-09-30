import { StellarAccountVerifier } from "@/components/StellarAccountVerifier";
import { AppHeader } from "@/components/layout/AppHeader";

export default function Home() {
  return (
    <main>
      <AppHeader statusLabel="Frontend directo a Horizon" />

      <div id="top" className="page-frame">
        <section id="inspector" aria-label="Inspector de cuentas Stellar">
          <StellarAccountVerifier />
        </section>
      </div>

    </main>
  );
}
