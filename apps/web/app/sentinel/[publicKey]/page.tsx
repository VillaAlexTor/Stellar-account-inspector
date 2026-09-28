import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { AppHeader } from "@/components/layout/AppHeader";
import { SentinelConsole } from "@/components/sentinel/SentinelConsole";
import { isValidStellarPublicKey, type StellarNetwork } from "@/lib/stellar";

export const metadata: Metadata = {
  title: "Sentinel · Stellar Account Inspector",
  description: "Monitoreo continuo de cambios sensibles en una cuenta Stellar.",
};

export default async function SentinelPage({
  params,
  searchParams,
}: {
  params: Promise<{ publicKey: string }>;
  searchParams: Promise<{ network?: string }>;
}) {
  const { publicKey: rawPublicKey } = await params;
  const { network: rawNetwork } = await searchParams;
  const publicKey = rawPublicKey.toUpperCase();
  if (!isValidStellarPublicKey(publicKey)) notFound();
  const network: StellarNetwork = rawNetwork === "mainnet" ? "mainnet" : "testnet";

  return (
    <main>
      <AppHeader
        active="sentinel"
        sentinelHref={`/sentinel/${publicKey}?network=${network}`}
        statusLabel="Proxy SSE activo"
      />
      <div className="page-frame page-frame--sentinel">
        <SentinelConsole publicKey={publicKey} initialNetwork={network} />
      </div>
      <footer className="app-footer">
        <span>Monitoreo público · Estado persistido · Sin custodia</span>
        <a href="https://developers.stellar.org/docs/data/apis/horizon/api-reference/resources/operations" target="_blank" rel="noreferrer">
          Operaciones de Horizon
        </a>
      </footer>
    </main>
  );
}
