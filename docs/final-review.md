# Revisión de alcance para la entrega formal

Fecha de ajuste: 2026-09-30.

## Observaciones atendidas

| Observación | Ajuste en `main` | Evidencia |
| --- | --- | --- |
| El entregable excedía el alcance frontend. | Se retiraron API Go, PostgreSQL, Docker, proxy, notificaciones y UI Sentinel. | Inspector y Risk Score consultan Horizon directamente desde el navegador. |
| Sentinel debía quedar fuera de evaluación. | La implementación completa se preservó en la rama `Retro`. | [`sentinel-extension.md`](sentinel-extension.md) documenta la separación. |
| `sha256_hash` se describía como `preauth_tx`. | Se crearon reglas y textos distintos para ambos tipos. | `hashLockHighWeight.ts`, `preauthHighWeight.ts` y sus pruebas unitarias. |

## Entregable evaluado

- Inspector de cuentas Stellar Testnet.
- Risk Score explicativo.
- Consulta directa a Horizon.
- Frontend Next.js/React/TypeScript.
- Pruebas, lint y build automatizados.

## Fuera del alcance evaluado

Sentinel es una extensión opcional de portafolio. No se necesita para instalar, ejecutar, demostrar ni evaluar la rama `main`.
