import { useMemo, useState } from "react";
import CRUDListPage, {
  type ColumnConfig,
  type FormFieldConfig,
} from "../../components/layout/CRUDListPage";
import StatusBadge from "../../components/StatusBadge";
import Timeline from "../../components/data-display/Timeline";
import useToast from "../../hooks/useToast";
import { useFinancialFiltersOptional } from "../Financial/FinancialContext";
import { apiGet, apiPost } from "../../services/api";
import type { TripSettlement } from "../../types/financial";
import { formatCurrency, settlementStatusLabel } from "../../utils/financialLabels";
import { formatDateTime, formatShortId } from "../../utils/format";

type SettlementForm = {
  trip_id: string;
  notes: string;
};

type TripItem = { id: string; route_id: string; departure_at: string };
type RouteItem = { id: string; origin_city: string; destination_city: string };

type TripSettlementsProps = {
  embedded?: boolean;
};

export default function TripSettlements({ embedded = false }: TripSettlementsProps) {
  const toast = useToast();
  const financialFilters = useFinancialFiltersOptional();
  const tripFilter = embedded ? financialFilters?.tripFilter ?? "" : "";
  const [trips, setTrips] = useState<TripItem[]>([]);
  const [routes, setRoutes] = useState<RouteItem[]>([]);
  const [reloadKey, setReloadKey] = useState(0);

  const routeMap = useMemo(
    () => new Map(routes.map((route) => [route.id, route])),
    [routes]
  );
  const tripMap = useMemo(
    () => new Map(trips.map((trip) => [trip.id, trip])),
    [trips]
  );

  const tripLabel = (tripId: string) => {
    const trip = tripMap.get(tripId);
    if (!trip) return formatShortId(tripId);
    const route = routeMap.get(trip.route_id);
    const routeLabel = route
      ? `${route.origin_city} -> ${route.destination_city}`
      : formatShortId(trip.route_id);
    return `${routeLabel} - ${formatDateTime(trip.departure_at)}`;
  };

  const formFields: FormFieldConfig<SettlementForm>[] = [
    {
      key: "trip_id",
      label: "Viagem",
      type: "select",
      required: true,
      options: [
        { label: "Selecione a viagem", value: "" },
        ...trips.map((trip) => ({
          label: tripLabel(trip.id),
          value: trip.id,
        })),
      ],
    },
    {
      key: "notes",
      label: "Observacoes",
      type: "textarea",
      colSpan: "full",
    },
  ];

  const columns: ColumnConfig<TripSettlement>[] = [
    { label: "Viagem", accessor: (item) => tripLabel(item.trip_id) },
    { label: "Adiantamento", accessor: (item) => formatCurrency(item.advance_amount) },
    { label: "Despesas", accessor: (item) => formatCurrency(item.expenses_total) },
    {
      label: "Saldo",
      render: (item) => (
        <span style={{ color: item.balance >= 0 ? "green" : "red" }}>
          {formatCurrency(item.balance)}
        </span>
      ),
    },
    { label: "A Devolver", accessor: (item) => formatCurrency(item.amount_to_return) },
    { label: "A Reembolsar", accessor: (item) => formatCurrency(item.amount_to_reimburse) },
    {
      label: "Status",
      render: (item) => (
        <StatusBadge tone={getSettlementStatusTone(item.status)}>
          {settlementStatusLabel[item.status] ?? item.status}
        </StatusBadge>
      ),
    },
    {
      label: "Timeline",
      hideOnMobile: true,
      render: (item) => (
        <Timeline
          compact
          items={buildSettlementTimeline(item).map((event) => ({
            id: event.label,
            title: event.label,
            timestamp: event.date ? formatDateTime(event.date) : "aguardando",
            tone: event.tone,
          }))}
        />
      ),
    },
  ];

  const runAction = async (id: string, action: string, successMessage: string) => {
    try {
      await apiPost(`/trip-settlements/${id}/${action}`, {});
      toast.success(successMessage);
      setReloadKey((value) => value + 1);
    } catch (err: any) {
      toast.error(err.message || "Erro ao atualizar acerto");
    }
  };

  return (
    <CRUDListPage<TripSettlement, SettlementForm>
      key={`${reloadKey}-${tripFilter}`}
      hidePageHeader={embedded}
      title="Acertos de Viagem"
      subtitle="Reconciliacao financeira pos-viagem."
      formTitle="Novo acerto"
      listTitle="Acertos registrados"
      createLabel="Criar acerto"
      updateLabel="Salvar acerto"
      emptyState={{
        title: "Nenhum acerto encontrado",
        description: "Crie um acerto para consolidar a viagem.",
      }}
      formFields={formFields}
      columns={columns}
      initialForm={{ trip_id: "", notes: "" }}
      mapItemToForm={(item) => ({ trip_id: item.trip_id, notes: item.notes ?? "" })}
      getId={(item) => item.id}
      fetchItems={async ({ page, pageSize }) => {
        const tripFilterQuery = tripFilter ? `&trip_id=${encodeURIComponent(tripFilter)}` : "";
        const data = await apiGet<TripSettlement[]>(
          `/trip-settlements?limit=${pageSize}&offset=${page * pageSize}${tripFilterQuery}`
        );
        const [tripsData, routesData] = await Promise.all([
          apiGet<TripItem[]>("/trips?limit=500&offset=0"),
          apiGet<RouteItem[]>("/routes?limit=500&offset=0"),
        ]);
        setTrips(tripsData);
        setRoutes(routesData);
        return data;
      }}
      createItem={(form) =>
        apiPost("/trip-settlements", {
          trip_id: form.trip_id,
          notes: form.notes || undefined,
        })
      }
      updateItem={undefined}
      searchFilter={(item, term) => {
        const trip = tripLabel(item.trip_id).toLowerCase();
        return trip.includes(term) || item.id.toLowerCase().includes(term);
      }}
      rowActions={(item) => {
        switch (item.status) {
          case "DRAFT":
            return (
              <button
                className="button ghost sm"
                type="button"
                onClick={() => runAction(item.id, "review", "Acerto enviado para revisao.")}
              >
                Enviar revisao
              </button>
            );
          case "UNDER_REVIEW":
            return (
              <>
                <button
                  className="button success sm"
                  type="button"
                  onClick={() => runAction(item.id, "approve", "Acerto aprovado.")}
                >
                  Aprovar
                </button>
                <button
                  className="button danger sm"
                  type="button"
                  onClick={() => runAction(item.id, "reject", "Acerto rejeitado.")}
                >
                  Rejeitar
                </button>
              </>
            );
          case "APPROVED":
            return (
              <button
                className="button success sm"
                type="button"
                onClick={() => runAction(item.id, "complete", "Acerto concluido.")}
              >
                Concluir
              </button>
            );
          default:
            return null;
        }
      }}
    />
  );
}

function getSettlementStatusTone(status: string) {
  switch (status) {
    case "DRAFT":
      return "neutral";
    case "UNDER_REVIEW":
      return "info";
    case "APPROVED":
      return "success";
    case "REJECTED":
      return "danger";
    case "COMPLETED":
      return "success";
    default:
      return "neutral";
  }
}

function buildSettlementTimeline(item: TripSettlement) {
  return [
    { label: "Criado", date: item.created_at, tone: "neutral" },
    { label: "Revisado", date: item.reviewed_at, tone: "info" },
    { label: "Aprovado", date: item.approved_at, tone: "success" },
    { label: "Concluido", date: item.completed_at, tone: "success" },
  ] as const;
}
