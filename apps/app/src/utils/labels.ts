export const pricingRuleTypeLabel: Record<string, string> = {
  OCCUPANCY: "Ocupacao",
  LEAD_TIME: "Antecedencia",
  DOW: "Dia da semana",
  SEASON: "Sazonal",
};

export const pricingScopeLabel: Record<string, string> = {
  GLOBAL: "Global",
  ROUTE: "Rota",
  TRIP: "Viagem",
};

export const fareModeLabel: Record<string, string> = {
  AUTO: "Automatico",
  AUTOMATIC: "Automatico",
  MANUAL: "Manual",
};

export const tripStatusLabel: Record<string, string> = {
  SCHEDULED: "Programada",
  IN_PROGRESS: "Em andamento",
  COMPLETED: "Concluida",
  CANCELLED: "Cancelada",
};

export const tripOperationalStatusLabel: Record<string, string> = {
  REQUESTED: "Solicitada",
  PASSENGERS_READY: "Passageiros prontos",
  ITINERARY_READY: "Roteiro pronto",
  DISPATCH_VALIDATED: "Escalacao validada",
  AUTHORIZED: "Autorizada",
  IN_PROGRESS: "Em viagem",
  RETURNED: "Retornada",
  RETURN_CHECKED: "Retorno conferido",
  SETTLED: "Acerto concluido",
  CLOSED: "Fechada",
};

export const pricingRuleTypeHelp: Record<string, string> = {
  OCCUPANCY: "Ocupacao: quanto mais cheio o onibus, maior o valor.",
  LEAD_TIME: "Antecedencia: descontos para compras antecipadas.",
  DOW: "Dia da semana: ajuste por dias mais buscados.",
  SEASON: "Sazonal: varia por periodo (ex: feriados).",
};
