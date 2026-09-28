---
target: ¿Hay que eliminar Risk Score y Sentinel del encabezado?
total_score: 22
max_score: 32
na_heuristics: 9,10
p0_count: 0
p1_count: 1
timestamp: 2026-09-28T16-18-36Z
slug: apps-web-components-layout-appheader-tsx
---
# AppHeader navigation critique

No deben eliminarse Risk Score ni Sentinel: son niveles centrales del producto. La navegación expresa correctamente la progresión Inspector → Risk Score → Sentinel, pero comunica disponibilidad antes de que existan destinos válidos.

## Heurísticas

| Heurística | Puntuación |
| --- | ---: |
| Visibilidad del estado | 2/4 |
| Correspondencia con el dominio | 4/4 |
| Control y libertad | 3/4 |
| Consistencia | 2/4 |
| Prevención de errores | 2/4 |
| Reconocimiento | 4/4 |
| Eficiencia | 2/4 |
| Diseño minimalista | 3/4 |
| Recuperación de errores | n/a |
| Ayuda | n/a |

Total: 22/32, aceptable.

## Prioridades

- P1: Sentinel parece habilitado en inicio, pero conduce al mismo destino que Inspector.
- P2: Risk Score apunta a un ancla que no existe antes de completar una inspección.
- P3: falta `aria-current`, el foco es débil y el ámbar hace que estados disponibles y activos compitan.

## Recomendación

Conservar los tres segmentos. Antes de inspeccionar una cuenta, mostrar Risk Score y Sentinel atenuados, no interactivos y con texto contextual. Habilitarlos sólo cuando exista un destino real. En móvil el tamaño táctil y el ajuste son correctos.
