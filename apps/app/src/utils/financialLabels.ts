import { formatCurrency as baseFormatCurrency } from "./format";

export const advanceStatusLabel: Record<string, string> = {
  PENDING: "Pendente",
  DELIVERED: "Entregue",
  SETTLED: "Acertado",
  CANCELLED: "Cancelado",
};

export const expenseTypeLabel: Record<string, string> = {
  FUEL: "Combustível",
  FOOD: "Alimentação",
  LODGING: "Hospedagem",
  TOLL: "Pedágio",
  MAINTENANCE: "Manutenção",
  OTHER: "Outros",
};

export const paymentMethodLabel: Record<string, string> = {
  ADVANCE: "Adiantamento",
  CARD: "Cartão",
  PERSONAL: "Pessoal",
  COMPANY: "Empresa",
};

export const settlementStatusLabel: Record<string, string> = {
  DRAFT: "Rascunho",
  UNDER_REVIEW: "Em Revisão",
  APPROVED: "Aprovado",
  REJECTED: "Rejeitado",
  COMPLETED: "Concluído",
};

export const cardTypeLabel: Record<string, string> = {
  FUEL: "Combustível",
  MULTIPURPOSE: "Múltiplo Propósito",
  FOOD: "Alimentação",
};

export const cardTransactionLabel: Record<string, string> = {
  CREDIT: "Crédito",
  DEBIT: "Débito",
  ADJUSTMENT: "Ajuste",
  REFUND: "Estorno",
};

export function formatCurrency(value: number): string {
  return baseFormatCurrency(value);
}
