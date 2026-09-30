# Seguridad

## Modelo de confianza

- La aplicación sólo acepta claves públicas Stellar.
- Las consultas salen directamente del navegador hacia Horizon Testnet.
- No existe API propia, base de datos propia, autenticación ni almacenamiento de cuentas.
- El Risk Score se calcula localmente y no firma ni envía transacciones.
- La aplicación nunca necesita una seed phrase o clave privada.

## Uso seguro

1. Verifica que la cuenta pertenezca a Testnet.
2. No pegues secretos en formularios, issues, capturas ni registros.
3. Trata el Risk Score como una ayuda educativa, no como una auditoría definitiva.
4. Mantén Node.js y las dependencias actualizadas.
5. Ejecuta `pnpm audit --prod`, `pnpm lint`, `pnpm test` y `pnpm build` antes de publicar cambios.

## Reporte responsable

No abras un issue público si encuentras una vulnerabilidad explotable o un secreto. Envía al mantenedor una descripción, la versión o commit afectado, pasos mínimos de reproducción y el impacto estimado mediante un canal privado del repositorio.

No incluyas seeds, claves privadas, datos personales ni credenciales de terceros.
