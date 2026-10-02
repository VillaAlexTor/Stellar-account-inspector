import type { Metadata } from "next";
import "@fontsource/barlow-condensed/400.css";
import "@fontsource/barlow-condensed/500.css";
import "@fontsource/barlow-condensed/600.css";
import "@fontsource/ibm-plex-mono/400.css";
import "@fontsource/ibm-plex-mono/500.css";
import "@fontsource/ibm-plex-sans/400.css";
import "@fontsource/ibm-plex-sans/500.css";
import "@fontsource/ibm-plex-sans/600.css";
import "./globals.css";

export const metadata: Metadata = {
  title: "Stellar Account Inspector",
  description: "Inspección y análisis de seguridad para cuentas Stellar.",
};

const designContract = `
THESIS: Una consola de telemetría convierte la configuración Stellar y su presentación en evidencia legible; rechaza el dashboard y el deck de tarjetas genéricas.
OWN-WORLD: Aluminio cálido, plástico beige, tinta negra y LED ámbar; paneles rectos, etiquetas grabadas y lecturas digitales.
STORY: Inspeccionar una cuenta sin conectar una wallet y explicar el instrumento mediante problema, solución, evidencia, límites y pedido.
FIRST VIEWPORT: En el Inspector, la consulta domina; en Pitch, una consola 50/50 enfrenta argumento y visual orbital con la acción Iniciar pitch.
FORM: Consola de telemetría heredada de la opción challenger cassette, seed verificable 79bf05b2.
FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance
`;

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="es" data-scroll-behavior="smooth">
      <body>
        <span hidden data-design-contract={designContract} aria-hidden="true" />
        {children}
      </body>
    </html>
  );
}
