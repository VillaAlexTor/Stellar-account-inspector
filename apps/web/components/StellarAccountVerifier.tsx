"use client";

import { FormEvent, useCallback, useEffect, useRef, useState } from "react";
import {
  Activity,
  ArrowUpRight,
  BellRing,
  Check,
  CircleAlert,
  Copy,
  Database,
  Gauge,
  KeyRound,
  LoaderCircle,
  Radio,
  Search,
  ShieldCheck,
  SlidersHorizontal,
  UsersRound,
  WalletCards,
  X,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { RiskScoreCard } from "@/components/risk/RiskScoreCard";
import {
  fetchStellarAccount,
  formatXlm,
  getStellarExpertUrl,
  isValidStellarPublicKey,
  StellarAccountError,
  type StellarAccountData,
} from "@/lib/stellar";
import { cn } from "@/lib/utils";

function PanelHeading({
  icon: Icon,
  title,
  detail,
}: {
  icon: typeof Gauge;
  title: string;
  detail?: string;
}) {
  return (
    <div className="panel-heading">
      <span className="panel-heading__icon" aria-hidden="true">
        <Icon size={17} strokeWidth={1.8} />
      </span>
      <h2>{title}</h2>
      {detail ? <span className="panel-heading__detail">{detail}</span> : null}
    </div>
  );
}

function Meter({ value, max = 8, label }: { value: number; max?: number; label: string }) {
  const active = Math.min(max, Math.max(0, Math.round(value)));
  return (
    <div className="meter" role="img" aria-label={`${label}: ${value}`}>
      {Array.from({ length: max }, (_, index) => (
        <span key={index} className={cn("meter__segment", index < active && "is-active")} />
      ))}
    </div>
  );
}

function CopyableKey({ value, compact = false }: { value: string; compact?: boolean }) {
  const [copied, setCopied] = useState(false);

  async function copy() {
    await navigator.clipboard.writeText(value);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1600);
  }

  return (
    <button className={cn("copy-key", compact && "copy-key--compact")} onClick={copy} type="button">
      <span>{value}</span>
      {copied ? <Check size={15} aria-label="Copiada" /> : <Copy size={15} aria-label="Copiar" />}
    </button>
  );
}

function AccountOverview({ account }: { account: StellarAccountData }) {
  const reserveRatio = account.nativeBalance
    ? Math.min(8, (account.reserveBalance / account.nativeBalance) * 8)
    : 0;

  return (
    <section className="instrument-panel instrument-panel--overview" aria-labelledby="account-summary-title">
      <PanelHeading icon={Gauge} title="Lectura de cuenta" detail={`Ledger ${account.lastModifiedLedger}`} />

      <div className="account-identity">
        <div>
          <span className="control-label">Clave pública</span>
          <CopyableKey value={account.accountId} />
        </div>
        <div className="identity-badges">
          <Badge variant="active">Testnet</Badge>
          <Badge variant={account.isMultisig ? "safe" : "muted"}>
            <UsersRound size={12} aria-hidden="true" />
            {account.isMultisig ? "Multisig real" : "Firma única"}
          </Badge>
        </div>
      </div>

      <div className="readout-grid">
        <div className="digital-readout">
          <span className="control-label control-label--light">Balance nativo</span>
          <strong>{formatXlm(account.nativeBalance)}</strong>
          <span>XLM total</span>
        </div>
        <div className="balance-channel">
          <div className="balance-channel__line">
            <span>Disponible</span>
            <strong>{formatXlm(account.availableBalance)} XLM</strong>
          </div>
          <Meter value={8 - reserveRatio} label="Proporción disponible" />
          <div className="balance-channel__line balance-channel__line--muted">
            <span>Reserva estimada</span>
            <strong>{formatXlm(account.reserveBalance)} XLM</strong>
          </div>
        </div>
      </div>

      <dl className="spec-strip">
        <div>
          <dt>Subentradas</dt>
          <dd>{account.subentryCount}</dd>
        </div>
        <div>
          <dt>Secuencia</dt>
          <dd title={account.sequence}>{account.sequence}</dd>
        </div>
        <div>
          <dt>Última modificación</dt>
          <dd>{account.lastModifiedTime ? new Date(account.lastModifiedTime).toLocaleString("es-BO") : "No informada"}</dd>
        </div>
      </dl>
    </section>
  );
}

function AccountFlags({ account }: { account: StellarAccountData }) {
  const flags = [
    ["Auth required", account.flags.auth_required],
    ["Auth revocable", account.flags.auth_revocable],
    ["Auth immutable", account.flags.auth_immutable],
    ["Clawback", account.flags.auth_clawback_enabled],
  ] as const;

  return (
    <section className="instrument-panel" aria-labelledby="flags-title">
      <PanelHeading icon={ShieldCheck} title="Controles del emisor" detail="Flags" />
      <div className="switch-bank">
        {flags.map(([label, enabled]) => (
          <div className="switch-row" key={label}>
            <span>{label}</span>
            <span className={cn("toggle-indicator", enabled && "is-on")}>
              <span>{enabled ? "ON" : "OFF"}</span>
              <i aria-hidden="true" />
            </span>
          </div>
        ))}
      </div>
      <p className="panel-note">
        Estos flags controlan cómo la cuenta emisora autoriza, revoca o recupera activos. No afectan al XLM nativo.
      </p>
    </section>
  );
}

function Thresholds({ account }: { account: StellarAccountData }) {
  return (
    <section className="instrument-panel" aria-labelledby="thresholds-title">
      <PanelHeading icon={SlidersHorizontal} title="Umbrales" detail="Pesos requeridos" />
      <div className="threshold-bank">
        {([
          ["Bajo", account.thresholds.low_threshold],
          ["Medio", account.thresholds.med_threshold],
          ["Alto", account.thresholds.high_threshold],
        ] as const).map(([label, value]) => (
          <div className="threshold-row" key={label}>
            <span>{label}</span>
            <Meter value={value} label={`Umbral ${label.toLowerCase()}`} />
            <strong>{value}</strong>
          </div>
        ))}
      </div>
    </section>
  );
}

function Trustlines({ account }: { account: StellarAccountData }) {
  return (
    <section className="instrument-panel instrument-panel--wide" aria-labelledby="trustlines-title">
      <PanelHeading icon={WalletCards} title="Trustlines" detail={`${account.trustlines.length} activas`} />
      {account.trustlines.length === 0 ? (
        <div className="empty-module">
          <Database size={23} strokeWidth={1.6} aria-hidden="true" />
          <div>
            <strong>Sin trustlines</strong>
            <p>Esta cuenta solo mantiene el activo nativo XLM.</p>
          </div>
        </div>
      ) : (
        <div className="table-scroll" tabIndex={0} aria-label="Trustlines de la cuenta">
          <table>
            <thead>
              <tr>
                <th>Activo</th>
                <th>Balance</th>
                <th>Límite</th>
                <th>Emisor</th>
                <th>Estado</th>
              </tr>
            </thead>
            <tbody>
              {account.trustlines.map((trustline) => (
                <tr key={`${trustline.assetCode}-${trustline.assetIssuer}`}>
                  <td><strong>{trustline.assetCode}</strong><small>{trustline.assetType.replace("credit_", "")}</small></td>
                  <td className="numeric">{formatXlm(trustline.balance)}</td>
                  <td className="numeric">{formatXlm(trustline.limit)}</td>
                  <td><CopyableKey value={trustline.assetIssuer} compact /></td>
                  <td>
                    <div className="trustline-status">
                      <Badge variant={trustline.authorized ? "safe" : "warning"}>
                        {trustline.authorized ? "Autorizada" : "Restringida"}
                      </Badge>
                      {trustline.issuerFlags?.auth_revocable ? (
                        <Badge variant="medium">Revocable</Badge>
                      ) : null}
                      {trustline.issuerLookupStatus === "unavailable" ? (
                        <Badge variant="muted">Emisor no verificado</Badge>
                      ) : null}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

function Signers({ account }: { account: StellarAccountData }) {
  return (
    <section className="instrument-panel instrument-panel--wide" aria-labelledby="signers-title">
      <PanelHeading icon={KeyRound} title="Firmantes" detail={`${account.signers.length} configurados`} />
      <div className="signer-list">
        {account.signers.map((signer) => {
          const isMaster = signer.key === account.accountId;
          return (
            <article className="signer-row" key={signer.key}>
              <div className="signer-row__identity">
                <div className="signer-type-icon" aria-hidden="true"><KeyRound size={17} /></div>
                <div>
                  <div className="signer-labels">
                    <strong>{isMaster ? "Master key" : signer.type.replaceAll("_", " ")}</strong>
                    {isMaster ? <Badge variant="active">Principal</Badge> : null}
                    {signer.weight === 0 ? <Badge variant="warning">Sin peso</Badge> : null}
                  </div>
                  <CopyableKey value={signer.key} compact />
                </div>
              </div>
              <div className="weight-readout">
                <span>Peso</span>
                <strong>{signer.weight}</strong>
              </div>
            </article>
          );
        })}
      </div>
    </section>
  );
}

function LoadingConsole() {
  return (
    <section className="loading-console" aria-live="polite" aria-busy="true">
      <div className="loading-console__display">
        <LoaderCircle className="animate-spin" size={22} aria-hidden="true" />
        <span>Sincronizando con Horizon</span>
      </div>
      <div className="loading-bars" aria-hidden="true">
        {Array.from({ length: 22 }, (_, index) => <i key={index} style={{ animationDelay: `${index * 45}ms` }} />)}
      </div>
    </section>
  );
}

export function StellarAccountVerifier() {
  const [publicKey, setPublicKey] = useState("");
  const [account, setAccount] = useState<StellarAccountData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const controllerRef = useRef<AbortController | null>(null);

  const loadAccount = useCallback(async (normalizedKey: string) => {
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;
    setLoading(true);
    setError(null);

    try {
      const nextAccount = await fetchStellarAccount(normalizedKey, controller.signal);
      setAccount(nextAccount);
    } catch (reason) {
      if (reason instanceof DOMException && reason.name === "AbortError") return;
      setAccount(null);
      setError(
        reason instanceof StellarAccountError
          ? reason.message
          : "Ocurrió un error inesperado al inspeccionar la cuenta.",
      );
    } finally {
      if (controllerRef.current === controller) setLoading(false);
    }
  }, []);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const requestedKey = params.get("account")?.trim().toUpperCase() ?? "";
    if (!isValidStellarPublicKey(requestedKey)) return;
    const frame = window.requestAnimationFrame(() => {
      setPublicKey(requestedKey);
      void loadAccount(requestedKey);
    });

    return () => {
      window.cancelAnimationFrame(frame);
      controllerRef.current?.abort();
    };
  }, [loadAccount]);

  async function inspect(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const normalizedKey = publicKey.trim().toUpperCase();
    if (!isValidStellarPublicKey(normalizedKey)) {
      setError("La clave debe comenzar con G y contener exactamente 56 caracteres válidos.");
      setAccount(null);
      return;
    }

    await loadAccount(normalizedKey);
  }

  return (
    <div className="inspector-shell">
      <section className="search-console" aria-labelledby="inspector-title">
        <div className="console-intro">
          <div>
            <h1 id="inspector-title">Inspecciona el control real de una cuenta Stellar.</h1>
            <p>
              Consulta balance, reserva, trustlines y firmantes directamente desde Horizon. Solo necesitas una clave pública; nunca una seed phrase.
            </p>
          </div>
          <div className="signal-display" aria-label="Red fija: Stellar Testnet">
            <Radio size={18} aria-hidden="true" />
            <span>Red de pruebas</span>
            <strong>TESTNET</strong>
          </div>
        </div>

        <form onSubmit={inspect} noValidate>
          <label className="search-field">
            <span className="control-label">Clave pública de la cuenta</span>
            <div className="search-field__controls">
              <Input
                aria-describedby={error ? "account-error" : "account-help"}
                aria-invalid={Boolean(error)}
                autoCapitalize="characters"
                autoComplete="off"
                maxLength={56}
                onChange={(event) => {
                  setPublicKey(event.target.value.toUpperCase());
                  if (error) setError(null);
                }}
                placeholder="G…"
                spellCheck={false}
                value={publicKey}
              />
              <Button disabled={loading || !publicKey.trim()} type="submit">
                {loading ? <LoaderCircle className="animate-spin" size={17} /> : <Search size={17} />}
                {loading ? "Leyendo" : "Inspeccionar"}
              </Button>
            </div>
            <span id="account-help" className="field-help">
              56 caracteres · La consulta se realiza desde tu navegador
            </span>
          </label>
        </form>

        {error ? (
          <div className="error-module" id="account-error" role="alert">
            <CircleAlert size={19} aria-hidden="true" />
            <div><strong>No se pudo completar la lectura</strong><p>{error}</p></div>
            <button onClick={() => setError(null)} type="button" aria-label="Cerrar error"><X size={17} /></button>
          </div>
        ) : null}
      </section>

      {loading ? <LoadingConsole /> : null}

      {!loading && !account ? (
        <section className="standby-console">
          <div className="standby-console__signal" aria-hidden="true">
            <span /><span /><span /><span /><span /><span /><span /><span />
          </div>
          <div>
            <h2>Inspector en espera</h2>
            <p>Introduce una cuenta Testnet. La lectura aparecerá aquí organizada por control, activos y capacidad de firma.</p>
          </div>
          <ul>
            <li><Activity size={16} />Estado actual desde Horizon</li>
            <li><ShieldCheck size={16} />Flags de autorización visibles</li>
            <li><UsersRound size={16} />Detección de multisig real</li>
          </ul>
        </section>
      ) : null}

      {!loading && account ? (
        <div className="results-console">
          <div className="results-console__status">
            <div><span className="status-dot is-live" /><strong>Lectura completada</strong><span>Datos actuales de Horizon</span></div>
            <div className="results-console__actions">
              <Button asChild size="sm">
                <a href={`/sentinel/${account.accountId}`}>
                  Activar Sentinel <BellRing size={14} />
                </a>
              </Button>
              <Button asChild variant="outline" size="sm">
                <a href={getStellarExpertUrl(account.accountId)} target="_blank" rel="noreferrer">
                  Ver en Stellar Expert <ArrowUpRight size={14} />
                </a>
              </Button>
            </div>
          </div>
          <div className="results-layout">
            <RiskScoreCard account={account} />
            <AccountOverview account={account} />
            <div className="results-layout__side">
              <AccountFlags account={account} />
              <Thresholds account={account} />
            </div>
            <Trustlines account={account} />
            <Signers account={account} />
          </div>
        </div>
      ) : null}
    </div>
  );
}
