"use client";

import Link from "next/link";
import {
  Activity,
  ArrowLeft,
  ArrowRight,
  Check,
  ChevronRight,
  FileText,
  FlaskConical,
  Gauge,
  Maximize2,
  Minimize2,
  SearchCheck,
  ShieldCheck,
  Target,
  X,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { AppHeader } from "@/components/layout/AppHeader";

type Slide = {
  id: string;
  label: string;
  title: string;
  summary: string;
  seconds: number;
  notes: string;
};

const slides: Slide[] = [
  {
    id: "cover",
    label: "Stellar Account Inspector",
    title: "De datos crudos a decisiones de seguridad.",
    summary: "Una lectura clara y verificable de cualquier cuenta Stellar Testnet.",
    seconds: 10,
    notes: "Stellar expone mucha información útil, pero entender si una cuenta está bien configurada todavía exige leer datos técnicos. Nosotros convertimos esa complejidad en una inspección comprensible.",
  },
  {
    id: "problem",
    label: "El problema",
    title: "Horizon informa. No interpreta.",
    summary: "Balances, firmantes, pesos, umbrales y flags llegan como datos separados. Un error importante puede pasar desapercibido.",
    seconds: 13,
    notes: "Hoy un estudiante o desarrollador debe interpretar manualmente la respuesta de Horizon. La información existe, pero está fragmentada y requiere conocimiento del protocolo para convertirla en una decisión.",
  },
  {
    id: "solution",
    label: "La solución",
    title: "Una inspección que explica cada señal.",
    summary: "Pegas una clave pública. El sistema organiza la cuenta, calcula un Risk Score y conecta cada hallazgo con evidencia observable.",
    seconds: 16,
    notes: "Nuestra solución es una herramienta de solo lectura. El usuario pega una clave pública y recibe dos vistas: el estado de la cuenta y una evaluación heurística explicada. No conectamos una wallet ni pedimos secretos.",
  },
  {
    id: "flow",
    label: "Cómo funciona",
    title: "Tres pasos. Ningún secreto.",
    summary: "La consulta ocurre directamente desde el navegador y el análisis se calcula de forma local.",
    seconds: 14,
    notes: "Validamos la clave, consultamos Horizon Testnet y ejecutamos reglas deterministas en el navegador. La arquitectura es deliberadamente simple: sin backend propio, sin base de datos y sin custodia.",
  },
  {
    id: "score",
    label: "Risk Score",
    title: "¿Qué es el Risk Score?",
    summary: "Una puntuación creada por la aplicación para resumir posibles señales de riesgo en una cuenta Stellar.",
    seconds: 16,
    notes: "El Risk Score va de cero a cien y resume señales detectadas por reglas automáticas. Un hallazgo bajo suma 5 puntos, uno medio 10, uno alto 25 y uno crítico 40. Es una puntuación heurística, no una probabilidad de pérdida ni una auditoría profesional.",
  },
  {
    id: "traction",
    label: "Tracción técnica",
    title: "El instrumento ya funciona en Testnet.",
    summary: "Consulta cuentas reales, distingue mecanismos de autorización y prueba cada regla de riesgo de forma independiente.",
    seconds: 14,
    notes: "La evidencia actual es técnica y verificable: el prototipo consulta Testnet, tiene escenarios reproducibles y pruebas para validación, reservas, multifirma, umbrales y tipos de firmante.",
  },
  {
    id: "difference",
    label: "¿Por qué esto?",
    title: "No reemplaza al explorador. Añade criterio.",
    summary: "Un explorador muestra qué hay en la cuenta. Stellar Account Inspector ayuda a entender qué significa para su configuración de seguridad.",
    seconds: 13,
    notes: "Nuestra diferencia no es competir con Horizon o con los exploradores. Los usamos como fuente y contraste. El valor está en la capa educativa: ordenar, evaluar y explicar sin ocultar la evidencia.",
  },
  {
    id: "risks",
    label: "Riesgos y límites",
    title: "Útil porque es honesto sobre lo que no hace.",
    summary: "Opera sólo en Testnet, depende de Horizon y ofrece señales heurísticas: no es una auditoría profesional ni firma transacciones.",
    seconds: 13,
    notes: "No presentamos un prototipo como producto de producción. Testnet puede reiniciarse, Horizon puede no estar disponible y el score no sustituye una auditoría. Esos límites se muestran en la propia experiencia.",
  },
  {
    id: "ask",
    label: "El pedido",
    title: "Validemos el criterio, no sólo el código.",
    summary: "Buscamos probar el instrumento con equipos que usan Stellar y revisar sus reglas con especialistas para priorizar el siguiente ciclo.",
    seconds: 14,
    notes: "Nuestro siguiente paso requiere validación. Pedimos acceso a usuarios técnicos y retroalimentación experta para comprobar si las explicaciones ayudan y qué reglas aportan más valor antes de ampliar el alcance.",
  },
];

function SlideVisual({ id }: { id: string }) {
  if (id === "cover") {
    return (
      <div className="pitch-orbit" aria-hidden="true">
        <div className="pitch-orbit__ring pitch-orbit__ring--outer" />
        <div className="pitch-orbit__ring pitch-orbit__ring--inner" />
        <div className="pitch-orbit__account">G...</div>
        <div className="pitch-orbit__signal"><span>RISK</span><strong>24</strong><small>/ 100</small></div>
      </div>
    );
  }

  if (id === "problem") {
    return (
      <div className="pitch-raw" aria-label="Ejemplo de datos técnicos fragmentados">
        <span>thresholds.low_threshold</span><strong>2</strong>
        <span>signers[2].type</span><strong>sha256_hash</strong>
        <span>flags.auth_revocable</span><strong>true</strong>
        <span>subentry_count</span><strong>7</strong>
        <div className="pitch-raw__question">¿Qué requiere atención?</div>
      </div>
    );
  }

  if (id === "solution") {
    return (
      <div className="pitch-verdict">
        <div><ShieldCheck size={30} /><span>Lectura de cuenta</span><strong>Configuración visible</strong></div>
        <ChevronRight aria-hidden="true" />
        <div><Gauge size={30} /><span>Risk Score</span><strong>Hallazgos explicados</strong></div>
      </div>
    );
  }

  if (id === "flow") {
    return (
      <ol className="pitch-flow">
        <li><SearchCheck /><span>01</span><strong>Clave pública</strong><small>Validación local</small></li>
        <li><Activity /><span>02</span><strong>Horizon Testnet</strong><small>Lectura en vivo</small></li>
        <li><ShieldCheck /><span>03</span><strong>Evaluación</strong><small>Reglas deterministas</small></li>
      </ol>
    );
  }

  if (id === "score") {
    return (
      <div className="pitch-risk-scale" aria-label="Escala y pesos del Risk Score">
        <div className="pitch-risk-scale__readout">
          <span>Escala de riesgo</span>
          <strong>0—100</strong>
        </div>
        <div className="pitch-risk-scale__head" aria-hidden="true">
          <span>Rango</span><span>Nivel</span><span>Peso por hallazgo</span>
        </div>
        {[
          ["00—09", "Bajo", "+5"],
          ["10—24", "Medio", "+10"],
          ["25—39", "Alto", "+25"],
          ["40—100", "Crítico", "+40"],
        ].map(([range, level, weight]) => (
          <div className="pitch-risk-scale__row" key={level}>
            <span>{range}</span>
            <strong>{level}</strong>
            <b>{weight}</b>
          </div>
        ))}
      </div>
    );
  }

  if (id === "traction") {
    return (
      <div className="pitch-checks">
        {[
          "Consulta directa a Horizon",
          "Risk Score de 0 a 100",
          "Reglas independientes",
          "Escenarios Testnet reproducibles",
          "Pruebas automatizadas",
        ].map((item) => <div key={item}><Check aria-hidden="true" /><span>{item}</span></div>)}
      </div>
    );
  }

  if (id === "difference") {
    return (
      <div className="pitch-compare">
        <div><span>Explorador</span><strong>¿Qué datos existen?</strong></div>
        <div className="pitch-compare__plus">+</div>
        <div><span>Inspector</span><strong>¿Qué significan?</strong></div>
      </div>
    );
  }

  if (id === "risks") {
    return (
      <div className="pitch-limits">
        <div><FlaskConical /><span>Red</span><strong>Sólo Testnet</strong></div>
        <div><Activity /><span>Fuente</span><strong>Depende de Horizon</strong></div>
        <div><ShieldCheck /><span>Alcance</span><strong>Heurístico, no auditoría</strong></div>
      </div>
    );
  }

  return (
    <div className="pitch-ask">
      <Target size={76} />
      <div><span>Siguiente ciclo</span><strong>Usuarios técnicos</strong><strong>Revisión experta</strong><strong>Prioridades validadas</strong></div>
    </div>
  );
}

export function PitchDeck() {
  const [index, setIndex] = useState(0);
  const [notesOpen, setNotesOpen] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const slide = slides[index];

  const goTo = useCallback((nextIndex: number) => {
    setIndex(Math.min(Math.max(nextIndex, 0), slides.length - 1));
  }, []);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      const target = event.target;
      if (target instanceof HTMLElement && target.matches("button, a, input, textarea, select")) return;

      if (event.key === "ArrowRight" || event.key === "PageDown" || event.key === " ") {
        event.preventDefault();
        goTo(index + 1);
      }
      if (event.key === "ArrowLeft" || event.key === "PageUp") {
        event.preventDefault();
        goTo(index - 1);
      }
      if (event.key === "Home") goTo(0);
      if (event.key === "End") goTo(slides.length - 1);
      if (event.key.toLowerCase() === "n") setNotesOpen((open) => !open);
    };

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [goTo, index]);

  useEffect(() => {
    const onFullscreenChange = () => setIsFullscreen(Boolean(document.fullscreenElement));
    document.addEventListener("fullscreenchange", onFullscreenChange);
    return () => document.removeEventListener("fullscreenchange", onFullscreenChange);
  }, []);

  const toggleFullscreen = async () => {
    if (document.fullscreenElement) {
      await document.exitFullscreen();
    } else {
      await document.documentElement.requestFullscreen();
    }
  };

  return (
    <main className="pitch-page">
      <AppHeader statusLabel="Pitch · 1 min 59 s" active="pitch" />

      <section className="pitch-deck" aria-label="Presentación de Stellar Account Inspector">
        <div className="pitch-stage" aria-live="polite" aria-atomic="true">
          <div className="pitch-stage__copy" key={`copy-${slide.id}`}>
            <h1>{slide.title}</h1>
            <p className="pitch-summary">{slide.summary}</p>
            {slide.id === "cover" && (
              <div className="pitch-cover-actions">
                <button type="button" onClick={() => goTo(1)}>Iniciar pitch <ArrowRight size={18} /></button>
                <Link href="/">Abrir instrumento</Link>
              </div>
            )}
          </div>
          <div className="pitch-stage__visual" key={`visual-${slide.id}`}><SlideVisual id={slide.id} /></div>
          <div className="pitch-stage__meta">
            <span className="pitch-stage__section">{slide.label}</span>
            <span>{String(index + 1).padStart(2, "0")} / {String(slides.length).padStart(2, "0")}</span>
            <span>{slide.seconds} s</span>
          </div>
        </div>

        <div className="pitch-navigation-panel">
          <div className="pitch-progress" aria-label={`Diapositiva ${index + 1} de ${slides.length}`}>
            {slides.map((item, itemIndex) => (
              <button
                type="button"
                key={item.id}
                className={itemIndex === index ? "is-current" : itemIndex < index ? "is-complete" : ""}
                onClick={() => goTo(itemIndex)}
                aria-label={`Ir a ${item.label}`}
                aria-current={itemIndex === index ? "step" : undefined}
              />
            ))}
          </div>

          <div className="pitch-controls">
            <div className="pitch-controls__navigation">
              <button type="button" onClick={() => goTo(index - 1)} disabled={index === 0} aria-label="Diapositiva anterior"><ArrowLeft /></button>
              <button type="button" onClick={() => goTo(index + 1)} disabled={index === slides.length - 1} aria-label="Diapositiva siguiente"><ArrowRight /></button>
            </div>
            <div className="pitch-controls__tools">
              <button type="button" onClick={() => setNotesOpen((open) => !open)} aria-expanded={notesOpen}><FileText size={18} /> Guion</button>
              <button type="button" onClick={toggleFullscreen}>{isFullscreen ? <Minimize2 size={18} /> : <Maximize2 size={18} />} {isFullscreen ? "Salir" : "Pantalla completa"}</button>
            </div>
          </div>
        </div>

        {notesOpen && (
          <aside className="pitch-notes" aria-label="Guion del presentador">
            <div><span>Guion · {slide.seconds} segundos</span><button type="button" onClick={() => setNotesOpen(false)} aria-label="Cerrar guion"><X size={19} /></button></div>
            <p>{slide.notes}</p>
            <small>Teclas: ← → navegar · N mostrar guion · Inicio/Fin saltar</small>
          </aside>
        )}
      </section>
    </main>
  );
}
