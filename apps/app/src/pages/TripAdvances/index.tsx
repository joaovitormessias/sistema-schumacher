import { useMemo, useState } from "react";
import CRUDListPage, {
  type ColumnConfig,
  type FormFieldConfig,
} from "../../components/layout/CRUDListPage";
import StatusBadge from "../../components/StatusBadge";
import useToast from "../../hooks/useToast";
import { apiGet, apiPatch, apiPost } from "../../services/api";
import type { TripAdvance } from "../../types/financial";
import { advanceStatusLabel, formatCurrency } from "../../utils/financialLabels";
import { formatDateTime, formatShortId } from "../../utils/format";

type TripAdvanceForm = {
  trip_id: string;
  driver_id: string;
  amount: number;
  purpose?: string;
  notes?: string;
};

type TripItem = { id: string; route_id: string; departure_at: string };

type RouteItem = { id: string; origin_city: string; destination_city: string };

type DriverItem = { id: string; name: string };

export default function TripAdvances() {
  const toast = useToast();
  const [trips, setTrips] = useState<TripItem[]>([]);
  const [routes, setRoutes] = useState<RouteItem[]>([]);
  const [drivers, setDrivers] = useState<DriverItem[]>([]);
  const [reloadKey, setReloadKey] = useState(0);

  const routeMap = useMemo(
    () => new Map(routes.map((route) => [route.id, route])),
    [routes]
  );
  const tripMap = useMemo(
    () => new Map(trips.map((trip) => [trip.id, trip])),
    [trips]
  );
  const driverMap = useMemo(
    () => new Map(drivers.map((driver) => [driver.id, driver.name])),
    [drivers]
  );

  const tripLabel = (tripId: string) => {
    const trip = tripMap.get(tripId);
    if (!trip) return formatShortId(tripId);
    const route = routeMap.get(trip.route_id);
    const routeLabel = route
      ? `${route.origin_city} → ${route.destination_city}`
      : formatShortId(trip.route_id);
    return `${routeLabel} • ${formatDateTime(trip.departure_at)}`;
  };

  const formFields: FormFieldConfig<TripAdvanceForm>[] = useMemo(
    () => [
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
        key: "driver_id",
        label: "Motorista",
        type: "select",
        required: true,
        options: [
          { label: "Selecione o motorista", value: "" },
          ...drivers.map((driver) => ({
            label: driver.name,
            value: driver.id,
          })),
        ],
      },
      {
        key: "amount",
        label: "Valor",
        type: "number",
        required: true,
        hint: "Valor do adiantamento em reais",
        inputProps: { min: 0, step: 0.01 },
      },
      {
        key: "purpose",
        label: "Finalidade",
        type: "textarea",
        hint: "Descreva o propósito do adiantamento",
        colSpan: "full",
      },
      {
        key: "notes",
        label: "Observações",
        type: "textarea",
        colSpan: "full",
      },
    ],
    [trips, drivers, routeMap]
  );

  const columns: ColumnConfig<TripAdvance>[] = [
    { label: "Viagem", accessor: (item) => tripLabel(item.trip_id) },
    {
      label: "Motorista",
      accessor: (item) => driverMap.get(item.driver_id) ?? formatShortId(item.driver_id),
    },
    { label: "Valor", accessor: (item) => formatCurrency(item.amount) },
    {
      label: "Status",
      render: (item) => (
        <StatusBadge tone={getStatusTone(item.status)}>
          {advanceStatusLabel[item.status] ?? item.status}
        </StatusBadge>
      ),
    },
    { label: "Criado em", accessor: (item) => formatDateTime(item.created_at) },
  ];

  const handleDeliver = async (item: TripAdvance) => {
    if (!window.confirm("Confirmar entrega do adiantamento?")) return;
    try {
      await apiPost(`/trip-advances/${item.id}/deliver`, {});
      toast.success("Adiantamento marcado como entregue.");
      setReloadKey((value) => value + 1);
    } catch (err: any) {
      toast.error(err.message || "Erro ao marcar adiantamento");
    }
  };

  return (
    <CRUDListPage<TripAdvance, TripAdvanceForm>
      key={reloadKey}
      title="Adiantamentos de Viagem"
      subtitle="Gestão de adiantamentos para motoristas."
      formTitle="Novo adiantamento"
      listTitle="Adiantamentos registrados"
      createLabel="Criar adiantamento"
      updateLabel="Salvar adiantamento"
      emptyState={{
        title: "Nenhum adiantamento encontrado",
        description: "Cadastre um adiantamento para começar.",
      }}
      formFields={formFields}
      columns={columns}
      initialForm={{ trip_id: "", driver_id: "", amount: 0, purpose: "", notes: "" }}
      mapItemToForm={(item) => ({
        trip_id: item.trip_id,
        driver_id: item.driver_id,
        amount: item.amount,
        purpose: item.purpose ?? "",
        notes: item.notes ?? "",
      })}
      getId={(item) => item.id}
      fetchItems={async ({ page, pageSize }) => {
        const data = await apiGet<TripAdvance[]>(
          `/trip-advances?limit=${pageSize}&offset=${page * pageSize}`
        );
        const [tripsData, routesData, driversData] = await Promise.all([
          apiGet<TripItem[]>("/trips?limit=500&offset=0"),
          apiGet<RouteItem[]>("/routes?limit=500&offset=0"),
          apiGet<DriverItem[]>("/drivers?limit=500&offset=0"),
        ]);
        setTrips(tripsData);
        setRoutes(routesData);
        setDrivers(driversData);
        return data;
      }}
      createItem={(form) =>
        apiPost("/trip-advances", {
          trip_id: form.trip_id,
          driver_id: form.driver_id,
          amount: Number(form.amount),
          purpose: form.purpose || undefined,
          notes: form.notes || undefined,
        })
      }
      updateItem={(id, form) =>
        apiPatch(`/trip-advances/${id}`, {
          amount: Number(form.amount),
          purpose: form.purpose || undefined,
          notes: form.notes || undefined,
        })
      }
      searchFilter={(item, term) => {
        const trip = tripLabel(item.trip_id).toLowerCase();
        const driver = (driverMap.get(item.driver_id) ?? "").toLowerCase();
        return (
          trip.includes(term) ||
          driver.includes(term) ||
          item.id.toLowerCase().includes(term)
        );
      }}
      rowActions={(item) =>
        item.status === "PENDING" ? (
          <button
            className="button ghost sm"
            type="button"
            onClick={() => handleDeliver(item)}
          >
            Marcar entregue
          </button>
        ) : null
      }
    />
  );
}

function getStatusTone(status: string) {
  switch (status) {
    case "PENDING":
      return "warning";
    case "DELIVERED":
      return "info";
    case "SETTLED":
      return "success";
    case "CANCELLED":
      return "danger";
    default:
      return "neutral";
  }
}
