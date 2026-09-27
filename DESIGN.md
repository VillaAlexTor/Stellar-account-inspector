---
name: Stellar Account Inspector
description: Consola de telemetría para leer configuración y riesgo de cuentas Stellar.
colors:
  canvas: "#d2c9b9"
  panel: "#e7dfcf"
  panel-strong: "#f2ecdf"
  steel: "#b8b8b2"
  ink: "#1b1b19"
  amber: "#efa51f"
  amber-dark: "#8a5708"
  display: "#151713"
  display-foreground: "#ffc14d"
  display-muted: "#9b7f4a"
  safe: "#2f6951"
  safe-dark: "#184c39"
  warning: "#a8462d"
  warning-dark: "#7d2f1d"
typography:
  display:
    fontFamily: "Barlow Condensed, sans-serif"
    fontSize: "clamp(3.25rem, 7vw, 6rem)"
    fontWeight: 520
    lineHeight: 0.88
    letterSpacing: "-0.025em"
  headline:
    fontFamily: "Barlow Condensed, sans-serif"
    fontSize: "2.25rem"
    fontWeight: 550
    lineHeight: 1
    letterSpacing: "0.04em"
  title:
    fontFamily: "Barlow Condensed, sans-serif"
    fontSize: "1rem"
    fontWeight: 650
    letterSpacing: "0.14em"
  body:
    fontFamily: "IBM Plex Sans, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.65
  label:
    fontFamily: "Barlow Condensed, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 650
    letterSpacing: "0.16em"
  readout:
    fontFamily: "IBM Plex Mono, monospace"
    fontSize: "clamp(2rem, 4vw, 3.2rem)"
    fontWeight: 500
    lineHeight: 1.15
    letterSpacing: "-0.035em"
rounded:
  square: "0px"
  round: "9999px"
spacing:
  xs: "8px"
  sm: "12px"
  md: "18px"
  lg: "24px"
  panel: "36px"
  frame: "64px"
components:
  button-primary:
    backgroundColor: "{colors.ink}"
    textColor: "{colors.panel}"
    typography: "{typography.label}"
    rounded: "{rounded.square}"
    padding: "12px 20px"
    height: "44px"
  button-primary-hover:
    backgroundColor: "{colors.amber}"
    textColor: "{colors.ink}"
  button-outline:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    typography: "{typography.label}"
    rounded: "{rounded.square}"
    padding: "10px 16px"
    height: "44px"
  input-console:
    backgroundColor: "{colors.display}"
    textColor: "{colors.display-foreground}"
    typography: "{typography.readout}"
    rounded: "{rounded.square}"
    padding: "12px 16px"
    height: "48px"
  badge-active:
    backgroundColor: "{colors.amber}"
    textColor: "{colors.ink}"
    typography: "{typography.label}"
    rounded: "{rounded.square}"
    padding: "2px 8px"
    height: "24px"
  readout-dark:
    backgroundColor: "{colors.display}"
    textColor: "{colors.amber}"
    typography: "{typography.readout}"
    rounded: "{rounded.square}"
    padding: "21px"
---

# Design System: Stellar Account Inspector

## Overview

**Creative North Star: "Consola de telemetría"**

Stellar Account Inspector se comporta como un instrumento de auditoría construido para durar: una fascia de casetera combina aluminio cálido, ABS beige, tinta negra y LED ámbar. La interfaz debe sentirse operacional, precisa y táctil, no como un dashboard de tarjetas genéricas. Paneles rectos, divisiones técnicas, etiquetas grabadas y lecturas digitales convierten la configuración de una cuenta en señales que se pueden escanear y contrastar.

La densidad es informativa pero calmada. El contraste material separa controles, superficies y displays; la jerarquía tipográfica distingue instrucciones, rótulos y datos sin depender del color. Los estados de seguridad siempre conservan texto o iconografía además de su tono, y los detalles de protocolo permanecen legibles para estudio técnico.

**Key Characteristics:**

- Superficies cálidas con textura física contenida.
- Geometría recta, divisiones visibles y controles compactos.
- Ámbar reservado para actividad, foco y señal positiva de instrumento.
- Tipografía condensada para rótulos y monoespaciada para datos.
- Profundidad estructural mediante biseles, líneas e iluminación interna.

## Colors

La paleta combina neutros de equipo analógico con una única señal ámbar, más verde y óxido para estados semánticos explícitos.

### Primary

- **LED ámbar:** señal de actividad, selección, medidores y foco; debe ser escasa para conservar su autoridad.
- **Ámbar de control:** contorno de foco y borde de componentes activos sobre superficies claras.

### Secondary

- **Verde de estado:** evidencia segura o configuración favorable; nunca sustituye la etiqueta textual.
- **Óxido de advertencia:** error o condición de atención, con lenguaje explicativo y no alarmista.

### Neutral

- **Aluminio cálido:** lienzo general y atmósfera exterior del instrumento.
- **ABS beige:** superficie principal de consolas, controles y contenedores.
- **Marfil de panel:** capa clara para controles secundarios y realces.
- **Acero mate:** detalle metálico y neutral intermedio.
- **Tinta negra:** texto, trazos estructurales y acción primaria.
- **Cristal de display:** fondo casi negro para datos electrónicos.
- **Fósforo claro:** lectura principal dentro de un display.
- **Fósforo apagado:** metadatos y unidades dentro de un display.

### Named Rules

**The Signal Rarity Rule.** El ámbar identifica energía, selección, foco o lectura activa; no se usa como relleno decorativo general.

**The Evidence Beyond Color Rule.** Verde, ámbar y óxido siempre aparecen con etiqueta, icono o dato que explique el estado.

## Typography

**Display Font:** Barlow Condensed (with sans-serif fallback)  
**Body Font:** IBM Plex Sans (with sans-serif fallback)  
**Label/Mono Font:** IBM Plex Mono (with monospace fallback)

**Character:** Barlow Condensed aporta el ritmo industrial de las leyendas grabadas; IBM Plex Sans mantiene la explicación serena y legible; IBM Plex Mono convierte claves, secuencias, pesos y balances en lecturas instrumentales estables.

### Hierarchy

- **Display:** condensada, mayúscula y de impacto compacto; se reserva para la identidad de una superficie principal.
- **Headline:** condensada y mayúscula; introduce estados o módulos importantes sin competir con los datos.
- **Title:** condensada, seminegrita y espaciada; rotula paneles e instrumentos.
- **Body:** sans humanista de lectura continua; los párrafos explicativos se mantienen alrededor de 55–68 caracteres por línea.
- **Label:** condensada, seminegrita, mayúscula y ampliamente espaciada; identifica controles, unidades y metadatos.
- **Readout:** monoespaciada con numerales tabulares; se usa para claves, saldos, secuencias, umbrales y valores técnicos.

### Named Rules

**The Instrument Type Rule.** Si el contenido se mide, copia o compara carácter por carácter, usa la voz monoespaciada; si nombra un control, usa la voz condensada.

## Layout

El sistema usa una retícula de paneles contiguos, con líneas compartidas en lugar de colecciones de tarjetas flotantes. El contenido vive en un marco centrado de hasta 1560px y emplea padding fluido; la separación habitual entre consolas es compacta, mientras el interior de paneles respira más. Los grupos de datos se organizan en canales, bancos, tiras de especificación y tablas, con alineación estricta para favorecer comparación.

Dos puntos de adaptación gobiernan la densidad: cerca de 1050px los conjuntos laterales se reorganizan en una sola columna o dos paneles pares; cerca de 760px navegación, formularios, displays y lecturas pasan a una columna. En móvil se preservan los bordes, rótulos y estados, pero desaparece el texto auxiliar de navegación que no cabe. Las tablas técnicas mantienen desplazamiento horizontal en vez de comprimir datos hasta hacerlos ilegibles.

**The Shared Chassis Rule.** Los módulos relacionados comparten bordes y ejes; no deben convertirse en tarjetas aisladas con espacios arbitrarios.

## Elevation & Depth

La profundidad es híbrida y material. Los paneles principales reciben una sombra ambiental baja y un filo interior claro; los displays usan sombras internas profundas que simulan cristal hundido. El resto permanece plano y obtiene jerarquía por tono, trazo y bisel. No se apilan sombras decorativas en controles secundarios.

### Shadow Vocabulary

- **Chasis ambiental:** `0 16px 36px rgba(41, 35, 25, 0.13), inset 0 1px 0 rgba(255, 255, 255, 0.68)`; eleva una consola completa sobre el lienzo.
- **Fascia metálica:** `inset 0 1px 0 #f8f8f3, 0 10px 24px rgba(38, 32, 23, 0.12)`; separa la navegación como pieza de hardware.
- **Cristal hundido:** `inset 0 0 28px rgba(0, 0, 0, 0.76), 0 1px 0 rgba(255, 255, 255, 0.65)`; se reserva para lecturas digitales.

**The Structural Depth Rule.** La sombra pertenece al chasis o al cristal del instrumento; divisores, tono y alineación resuelven el resto de la jerarquía.

## Shapes

La forma dominante es rectangular y sin radio: botones, inputs, badges, paneles, medidores y tablas se ensamblan como piezas de una fascia. Los bordes oscuros definen estructura; los bordes suaves separan información interna. Los círculos son excepciones funcionales para marca, LED de estado o tornillería visual, no una decoración repetida. Los paneles pueden incluir un filete interior claro o puntos de fijación, siempre subordinados al contenido.

**The Square Hardware Rule.** Conserva esquinas rectas en toda pieza de interacción o lectura; usa círculos sólo cuando el objeto representado es físicamente circular.

## Components

### Buttons

- **Shape:** bloque rectangular sin radio, altura mínima táctil y rótulo condensado en mayúsculas.
- **Primary:** tinta negra sobre ABS beige, con padding horizontal compacto y bisel interior sutil.
- **Hover / Focus:** cambia a LED ámbar con tinta negra; el foco visible usa un aro ámbar oscuro desplazado y el estado activo desciende un píxel.
- **Outline / Ghost:** mantienen el mismo lenguaje tipográfico; el outline usa panel translúcido y el ghost sólo un baño tenue de tinta.

### Chips

- **Style:** badges rectos de 24px de alto, borde visible y rótulo condensado; los activos usan ámbar, los seguros verde y las advertencias óxido.
- **State:** todo estado conserva texto explícito y contraste de borde; el neutral puede permanecer transparente.

### Cards / Containers

- **Corner Style:** chasis recto sin radio.
- **Background:** ABS beige o marfil sobre el lienzo de aluminio cálido; los displays internos usan cristal casi negro.
- **Shadow Strategy:** sólo la consola completa utiliza elevación ambiental; los paneles internos se separan con divisores.
- **Border:** trazo de tinta semitransparente para estructura y versión suave para subdivisiones.
- **Internal Padding:** fluido, de compacto en filas de datos a generoso en paneles principales.

### Inputs / Fields

- **Style:** display oscuro, borde de tinta, texto monoespaciado de fósforo y placeholder apagado; sin radio.
- **Focus:** borde ámbar y halo exterior translúcido, sin desplazar el layout.
- **Error / Disabled:** los errores viven en un módulo óxido con explicación; los campos deshabilitados reducen opacidad y bloquean interacción.

### Navigation

La navegación es una fascia segmentada con rótulos condensados, iconos lineales y divisores verticales. El estado activo combina un baño claro y una barra ámbar inferior; los niveles no disponibles se muestran atenuados y etiquetados, no ocultos. En móvil cada segmento apila icono y nombre y omite el subtítulo auxiliar.

### Digital Readout

El display distintivo usa fondo casi negro, lectura ámbar monoespaciada, unidades condensadas y sombra interna profunda. Sus valores deben ser tabulares, truncarse con criterio cuando sean extensos y conservar una etiqueta visible fuera del valor.

### Segment Meter

Los medidores se construyen con segmentos rectos, espaciado estrecho y activación ámbar. Siempre exponen una etiqueta accesible y, cuando el valor exacto importe, lo acompañan con una cifra monoespaciada.

## Do's and Don'ts

### Do:

- **Do** organiza datos relacionados como canales, bancos o tiras alineadas dentro de un chasis común.
- **Do** reserva el ámbar para actividad, selección, foco y lectura activa.
- **Do** usa Barlow Condensed para rótulos e IBM Plex Mono para valores técnicos comparables.
- **Do** acompaña cada estado semántico con texto, icono o evidencia observable.
- **Do** conserva foco visible, operación por teclado y reducción de movimiento.

### Don't:

- **Don't** conviertas la interfaz en un mosaico de tarjetas redondeadas y flotantes.
- **Don't** uses degradados de neón, vidrio borroso o brillo futurista ajeno al equipo analógico cálido.
- **Don't** uses el ámbar como decoración de fondo o para contenido sin estado.
- **Don't** presentes un juicio de seguridad sólo por color o sin explicar la evidencia.
- **Don't** comprimas claves, secuencias o tablas técnicas hasta volverlas ilegibles; permite truncado consciente o desplazamiento.
