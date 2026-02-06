export const pricingRuleTypeLabel: Record<string, string> = {
  OCCUPANCY: "Ocupação",
  LEAD_TIME: "Antecedência",
  DOW: "Dia da semana",
  SEASON: "Sazonal",
};

export const pricingScopeLabel: Record<string, string> = {
  GLOBAL: "Global",
  ROUTE: "Rota",
  TRIP: "Viagem",
};

export const fareModeLabel: Record<string, string> = {
  AUTO: "Automático",
  AUTOMATIC: "Automático",
  MANUAL: "Manual",
};

export const tripStatusLabel: Record<string, string> = {
  SCHEDULED: "Programada",
  IN_PROGRESS: "Em andamento",
  COMPLETED: "Concluída",
  CANCELLED: "Cancelada",
};

export const pricingRuleTypeHelp: Record<string, string> = {
  OCCUPANCY: "Ocupação: quanto mais cheio o ônibus, maior o valor.",
  LEAD_TIME: "Antecedência: descontos para compras antecipadas.",
  DOW: "Dia da semana: ajuste por dias mais buscados.",
  SEASON: "Sazonal: varia por período (ex: feriados).",
};
