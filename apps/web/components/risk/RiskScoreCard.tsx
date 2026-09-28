import {
  CheckCircle2,
  ChevronDown,
  CircleGauge,
  ShieldAlert,
  TriangleAlert,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { calculateRiskScore, type RiskFinding, type RiskLevel } from "@/lib/riskScore";
import type { StellarAccountData } from "@/lib/stellar";
import { cn } from "@/lib/utils";

const LEVEL_COPY: Record<RiskLevel, { label: string; summary: string }> = {
  bajo: {
    label: "Riesgo bajo",
    summary: "Las cinco comprobaciones automáticas no detectaron señales de riesgo.",
  },
  medio: {
    label: "Riesgo medio",
    summary: "Hay condiciones que conviene validar antes de operar con esta cuenta.",
  },
  alto: {
    label: "Riesgo alto",
    summary: "La configuración presenta señales que pueden afectar el control de la cuenta.",
  },
  crítico: {
    label: "Riesgo crítico",
    summary: "Se detectó al menos una condición capaz de bloquear o comprometer la cuenta.",
  },
};

const SEVERITY_LABELS: Record<RiskFinding["severity"], string> = {
  low: "Baja",
  medium: "Media",
  high: "Alta",
  critical: "Crítica",
};

export function RiskScoreCard({ account }: { account: StellarAccountData }) {
  const report = calculateRiskScore(account);
  const activeSegments = Math.ceil(report.score / 10);
  const unavailableIssuers = account.trustlines.filter(
    (trustline) => trustline.issuerLookupStatus === "unavailable",
  ).length;

  return (
    <section
      className={cn("risk-console", `risk-console--${report.level.replace("í", "i")}`)}
      id="risk-score"
      aria-labelledby="risk-score-title"
    >
      <div className="risk-console__heading">
        <span className="panel-heading__icon" aria-hidden="true">
          <CircleGauge size={18} strokeWidth={1.8} />
        </span>
        <h2 id="risk-score-title">Risk Score</h2>
        <span>5 reglas · análisis local</span>
      </div>

      <div className="risk-console__readout">
        <div className="risk-score-display" aria-label={`Puntaje de riesgo ${report.score} de 100`}>
          <span>Puntaje</span>
          <strong>{report.score.toString().padStart(2, "0")}</strong>
          <small>/ 100</small>
        </div>

        <div className="risk-console__interpretation">
          <Badge variant={report.level}>{LEVEL_COPY[report.level].label}</Badge>
          <p>{LEVEL_COPY[report.level].summary}</p>
          <div className="risk-meter" role="img" aria-label={`${report.score}% de riesgo calculado`}>
            {Array.from({ length: 10 }, (_, index) => (
              <i key={index} className={cn(index < activeSegments && "is-active")} />
            ))}
          </div>
          <div className="risk-scale" aria-hidden="true">
            <span>0 · Bajo</span>
            <span>100 · Crítico</span>
          </div>
        </div>
      </div>

      <div className="risk-findings">
        <div className="risk-findings__header">
          <strong>Hallazgos</strong>
          <span>{report.findings.length} detectado{report.findings.length === 1 ? "" : "s"}</span>
        </div>

        {report.findings.length === 0 ? (
          <div className="risk-clear-state">
            <CheckCircle2 size={20} aria-hidden="true" />
            <div>
              <strong>Sin señales automáticas</strong>
              <p>La lectura actual no activa ninguna regla. Esto no sustituye una auditoría de operaciones ni de custodia.</p>
            </div>
          </div>
        ) : (
          <div className="risk-findings__list">
            {report.findings.map((finding) => (
              <details key={finding.id} open={finding.severity === "critical"}>
                <summary>
                  <span className="finding-signal" aria-hidden="true">
                    {finding.severity === "critical" ? <ShieldAlert size={17} /> : <TriangleAlert size={17} />}
                  </span>
                  <span className="finding-title">
                    <strong>{finding.title}</strong>
                    <small>{finding.id}</small>
                  </span>
                  <Badge variant={finding.severity}>{SEVERITY_LABELS[finding.severity]}</Badge>
                  <ChevronDown className="finding-chevron" size={17} aria-hidden="true" />
                </summary>
                <p>{finding.description}</p>
              </details>
            ))}
          </div>
        )}

        {unavailableIssuers > 0 ? (
          <div className="risk-coverage-note">
            <TriangleAlert size={16} aria-hidden="true" />
            <span>
              Cobertura parcial: no se pudieron verificar los flags de {unavailableIssuers} emisor{unavailableIssuers === 1 ? "" : "es"}.
            </span>
          </div>
        ) : null}
      </div>
    </section>
  );
}
