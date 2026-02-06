export function formatCurrency(value: number, currency = "BRL") {
  return new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency,
    minimumFractionDigits: 2,
  }).format(value);
}

export function formatDateTime(value: string | number | Date) {
  const date = value instanceof Date ? value : new Date(value);
  return date.toLocaleString("pt-BR");
}

export function formatDate(value: string | number | Date) {
  const date = value instanceof Date ? value : new Date(value);
  return date.toLocaleDateString("pt-BR");
}

export function formatShortId(id: string, size = 8) {
  if (!id) return "-";
  return id.slice(0, size);
}
