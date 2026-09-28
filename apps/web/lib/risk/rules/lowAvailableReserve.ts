import type { RiskRule } from "@/lib/risk/types";

export const findLowAvailableReserve: RiskRule = (account) => {
  if (account.nativeBalance <= 0) return null;

  const availableRatio = account.availableBalance / account.nativeBalance;
  if (availableRatio >= 0.1) return null;

  return {
    id: "LOW_AVAILABLE_BALANCE",
    severity: "medium",
    title: "Reserva operativa al límite",
    description: `Solo el ${(availableRatio * 100).toFixed(1)}% del balance XLM permanece disponible después de la reserva estimada. La cuenta tiene poco margen para comisiones, nuevas subentradas u otras operaciones.`,
  };
};
