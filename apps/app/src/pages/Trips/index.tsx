import { useMemo, useState } from "react";
import DataTable, { type DataTableColumn } from "../../components/table/DataTable";
import EmptyState from "../../components/EmptyState";
import InlineAlert from "../../components/InlineAlert";
import Modal from "../../components/overlay/Modal";
import PageHeader from "../../components/PageHeader";
import SearchToolbar from "../../components/input/SearchToolbar";
import { useRoutes, type RouteItem } from "../../hooks/useRoutes";
import { useRouteStops, type RouteStopItem } from "../../hooks/useRouteStops";
import { useTrips, type TripItem } from "../../hooks/useTrips";
import { formatDateTime, formatShortId } from "../../utils/format";

type RouteTripSummary = {
  activeCount: number;
  nextTrip: TripItem | null;
};

const normalizeTripStatus = (status: string | null | undefined) =>
  String(status ?? "")
    .trim()
    .toUpperCase();

const isActiveTripStatus = (status: string | null | undefined) => {
  const normalized = normalizeTripStatus(status);
  return normalized === "ATIVO" || normalized === "ACTIVE" || normalized === "SCHEDULED" || normalized === "IN_PROGRESS";
};

const isTemplateTripStatus = (status: string | null | undefined) => normalizeTripStatus(status) === "TEMPLATE";

const isCancelledTripStatus = (status: string | null | undefined) => {
  const normalized = normalizeTripStatus(status);
  return normalized === "CANCELLED" || normalized === "CANCELED";
};

export default function Trips() {
  const [search, setSearch] = useState("");
  const [selectedRouteId, setSelectedRouteId] = useState<string | null>(null);

  const routesQuery = useRoutes(500, 0, { search, status: "active" });
  const routes = routesQuery.data ?? [];
  const tripsQuery = useTrips(500, 0);
  const trips = tripsQuery.data ?? [];

  const routeTripSummary = useMemo(() => {
    const tripsByRoute = new Map<string, TripItem[]>();
    for (const trip of trips) {
      if (isTemplateTripStatus(trip.status)) continue;
      const existing = tripsByRoute.get(trip.route_id);
      if (existing) {
        existing.push(trip);
      } else {
        tripsByRoute.set(trip.route_id, [trip]);
      }
    }

    const now = Date.now();
    const summary = new Map<string, RouteTripSummary>();
    for (const [routeID, routeTrips] of tripsByRoute.entries()) {
      routeTrips.sort((a, b) => Date.parse(a.departure_at) - Date.parse(b.departure_at));
      const activeTrips = routeTrips.filter((item) => isActiveTripStatus(item.status));
      const referenceTrips = activeTrips.length > 0
        ? activeTrips
        : routeTrips.filter((item) => !isCancelledTripStatus(item.status));
      const nextTrip =
        referenceTrips.find((item) => {
          const departureTime = Date.parse(item.departure_at);
          return !Number.isNaN(departureTime) && departureTime >= now;
        }) ??
        referenceTrips[referenceTrips.length - 1] ??
        null;

      summary.set(routeID, {
        activeCount: activeTrips.length,
        nextTrip,
      });
    }

    return summary;
  }, [trips]);

  const selectedRoute = useMemo(
    () => routes.find((item) => item.id === selectedRouteId) ?? null,
    [routes, selectedRouteId]
  );

  const stopsQuery = useRouteStops(selectedRouteId);
  const stops = stopsQuery.data ?? [];

  const loadError =
    (routesQuery.error as Error | undefined)?.message ||
    (tripsQuery.error as Error | undefined)?.message ||
    (stopsQuery.error as Error | undefined)?.message ||
    null;

  const activeTripsLabel = (routeID: string) => String(routeTripSummary.get(routeID)?.activeCount ?? 0);

  const nextCapacityLabel = (routeID: string) => {
    const trip = routeTripSummary.get(routeID)?.nextTrip;
    if (!trip) return "-";
    const seatsTotal = Number.isFinite(trip.seats_total) ? Math.max(0, Number(trip.seats_total)) : 0;
    const seatsAvailable = Number.isFinite(trip.seats_available)
      ? Math.max(0, Number(trip.seats_available))
      : 0;
    return seatsTotal > 0 ? `${seatsAvailable}/${seatsTotal}` : "-";
  };

  const nextTripLabel = (routeID: string) => {
    const trip = routeTripSummary.get(routeID)?.nextTrip;
    return trip ? formatDateTime(trip.departure_at) : "-";
  };

  const routeColumns: DataTableColumn<RouteItem>[] = [
    { label: "ID da rota", accessor: (item) => item.id, width: "180px" },
    { label: "Rota", accessor: (item) => item.name },
    {
      label: "Trecho",
      accessor: (item) => `${item.origin_city} -> ${item.destination_city}`,
      width: "340px",
    },
    {
      label: "Paradas",
      accessor: (item) => String(item.stop_count ?? 0),
      width: "110px",
      align: "center",
    },
    {
      label: "Viagens ativas",
      accessor: (item) => activeTripsLabel(item.id),
      width: "130px",
      align: "center",
    },
    {
      label: "Vagas/Capacidade",
      accessor: (item) => nextCapacityLabel(item.id),
      width: "150px",
      align: "center",
    },
    {
      label: "Proxima viagem",
      accessor: (item) => nextTripLabel(item.id),
      width: "180px",
    },
    {
      label: "Status",
      accessor: () => "ATIVA",
      width: "120px",
    },
  ];

  const stopColumns: DataTableColumn<RouteStopItem>[] = [
    { label: "Ordem", accessor: (item) => String(item.stop_order), width: "100px", align: "center" },
    { label: "Parada", accessor: (item) => item.city },
    { label: "ID", accessor: (item) => formatShortId(item.id), width: "180px" },
  ];

  return (
    <section className="page">
      <PageHeader
        title="Viagens"
        subtitle="Cada linha representa uma rota ativa, com lotacao da proxima viagem."
        meta={<span className="badge">Gestao</span>}
      />

      <div className="section">
        <SearchToolbar value={search} onChange={setSearch} resultCount={routes.length} />

        {loadError ? <InlineAlert tone="error">{loadError}</InlineAlert> : null}

        <DataTable
          columns={routeColumns}
          rows={routes}
          rowKey={(item) => item.id}
          actions={(item) => (
            <button className="button secondary sm" type="button" onClick={() => setSelectedRouteId(item.id)}>
              Ver paradas
            </button>
          )}
          emptyState={
            routesQuery.isLoading ? (
              <EmptyState title="Carregando rotas ativas" description="Aguarde alguns segundos." />
            ) : (
              <EmptyState title="Nenhuma rota ativa encontrada" description="Tente ajustar os filtros." />
            )
          }
        />
      </div>

      <Modal
        open={selectedRouteId !== null}
        onClose={() => setSelectedRouteId(null)}
        title="Paradas da rota"
        description={
          selectedRoute
            ? `${selectedRoute.name} - ${selectedRoute.origin_city} -> ${selectedRoute.destination_city}`
            : "Detalhes das paradas da rota selecionada."
        }
        footer={
          <button className="button secondary" type="button" onClick={() => setSelectedRouteId(null)}>
            Fechar
          </button>
        }
      >
        <DataTable
          columns={stopColumns}
          rows={stops}
          rowKey={(item) => item.id}
          emptyState={
            selectedRouteId ? (
              stopsQuery.isLoading ? (
                <EmptyState title="Carregando paradas" description="Aguarde alguns segundos." />
              ) : (
                <EmptyState title="Sem paradas" description="Nao ha paradas cadastradas para esta rota." />
              )
            ) : (
              <EmptyState title="Selecione uma rota" description="Escolha uma rota para ver as paradas." />
            )
          }
        />
      </Modal>
    </section>
  );
}
