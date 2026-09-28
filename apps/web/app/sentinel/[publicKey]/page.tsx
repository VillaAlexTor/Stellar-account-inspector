import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { AppHeader } from "@/components/layout/AppHeader";
import { SentinelConsole } from "@/components/sentinel/SentinelConsole";
import { isValidStellarPublicKey } from "@/lib/stellar";

export const metadata: Metadata = {
  title: "Sentinel · Stellar Account Inspector",
  description: "Monitoreo continuo de cambios sensibles en una cuenta Stellar.",
};

export default async function SentinelPage({
  params,
}: {
  params: Promise<{ publicKey: string }>;
}) {
  const { publicKey: rawPublicKey } = await params;
  const publicKey = rawPublicKey.toUpperCase();
  if (!isValidStellarPublicKey(publicKey)) notFound();

  return (
    <main>
      <AppHeader
        active="sentinel"
        sentinelHref={`/sentinel/${publicKey}`}
        statusLabel="Canal Sentinel"
      />
      <div className="page-frame page-frame--sentinel">
        <SentinelConsole publicKey={publicKey} />
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
