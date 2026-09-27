# Stellar Account Inspector

Herramienta educativa para inspeccionar cuentas Stellar, explicar su configuración de seguridad y, en fases posteriores, monitorear cambios sensibles en tiempo real.

## Estado

- Nivel 1 — Inspector: implementado.
- Nivel 2 — Risk Score: siguiente fase.
- Nivel 3 — Sentinel: planificado.

## Requisitos

- Node.js 20 o superior.
- pnpm 10.

## Desarrollo

```bash
pnpm install
pnpm dev
```

Abre `http://localhost:3000`. El Inspector consulta Horizon directamente desde el navegador y no solicita claves secretas.

## Verificación

```bash
pnpm test
pnpm lint
pnpm build
```

La reserva bloqueada se calcula según el alcance académico definido para esta versión: `(2 + subentry_count) × 0.5 XLM`.

## Estructura

- `apps/web`: frontend Next.js.
- `services/sentinel-api`: backend Go que se incorporará en el Nivel 3.
- `PRODUCT.md`: contexto y principios duraderos del producto.
- `promtp.md`: brief original del proyecto.
