import { useEffect, useMemo, useState } from "react";
import CRUDListPage, {
  type ColumnConfig,
  type FormFieldConfig,
  type VisibilityOption,
} from "../../components/layout/CRUDListPage";
import StatusBadge, { type StatusTone } from "../../components/StatusBadge";
import { apiGet, apiPatch, apiPost } from "../../services/api";
import { formatDateTime, formatShortId } from "../../utils/format";
import { tripStatusLabel } from "../../utils/labels";

type TripItem = {
  id: string;
  route_id: string;
  bus_id: string;
  driver_id?: string;
  departure_at: string;
  status: string;
};

type RouteItem = { id: string; name: string; origin_city: string; destination_city: string };

type BusItem = { id: string; name: string };

type DriverItem = { id: string; name: string };

type TripForm = {
  route_id: string;
  bus_id: string;
  driver_id: string;
  departure_at: string;
};

const statusTone = (status: string): StatusTone => {
  if (status === "IN_PROGRESS") return "info";
  if (status === "SCHEDULED") return "success";
  if (status === "CANCELLED") return "danger";
  return "neutral";
};

export default function Trips() {
  const [routes, setRoutes] = useState<RouteItem[]>([]);
  const [buses, setBuses] = useState<BusItem[]>([]);
  const [drivers, setDrivers] = useState<DriverItem[]>([]);
  const [statusFilter, setStatusFilter] = useState("ALL");

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      try {
        const [r, b, d] = await Promise.all([
          apiGet<RouteItem[]>("/routes?limit=200&offset=0"),
          apiGet<BusItem[]>("/buses?limit=200&offset=0"),
          apiGet<DriverItem[]>("/drivers?limit=200&offset=0"),
        ]);
        if (cancelled) return;
        setRoutes(r);
        setBuses(b);
        setDrivers(d);
      } catch {
        if (cancelled) return;
        setRoutes([]);
        setBuses([]);
        setDrivers([]);
      }
    };
    load();
    return () => {
      cancelled = true;
    };
  }, []);

  const routeLabel = (id: string) => {
    const route = routes.find((item) => item.id === id);
    return route ? `${route.origin_city} → ${route.destination_city}` : formatShortId(id);
  };

  const busLabel = (id: string) => {
    const bus = buses.find((item) => item.id === id);
    return bus ? bus.name : formatShortId(id);
  };

  const driverOptions = useMemo(
    () => drivers.map((driver) => ({ label: driver.name, value: driver.id })),
    [drivers]
  );

  const formFields: FormFieldConfig<TripForm>[] = [
    {
      key: "route_id",
      label: "Rota",
      type: "select",
      required: true,
      options: [
        { label: "Selecione a rota", value: "" },
        ...routes.map((route) => ({
          label: `${route.origin_city} → ${route.destination_city}`,
          value: route.id,
        })),
      ],
    },
    {
      key: "bus_id",
      label: "Ônibus",
      type: "select",
      required: true,
      options: [
        { label: "Selecione o ônibus", value: "" },
        ...buses.map((bus) => ({
          label: bus.name,
          value: bus.id,
        })),
      ],
    },
    {
      key: "driver_id",
      label: "Motorista",
      type: "select",
      options: [{ label: "Selecionar depois", value: "" }, ...driverOptions],
    },
    {
      key: "departure_at",
      label: "Data e hora de saída",
      type: "datetime",
      required: true,
    },
  ];

  const columns: ColumnConfig<TripItem>[] = [
    { label: "Rota", accessor: (item) => routeLabel(item.route_id) },
    { label: "Ônibus", accessor: (item) => busLabel(item.bus_id) },
    { label: "Saída", accessor: (item) => formatDateTime(item.departure_at) },
    {
      label: "Status",
      render: (item) => (
        <StatusBadge tone={statusTone(item.status)}>
          {tripStatusLabel[item.status] ?? item.status}
        </StatusBadge>
      ),
    },
  ];

  const visibilityOptions: VisibilityOption<TripItem>[] = [
    { label: "Ativas", value: "active", predicate: (item) => item.status !== "CANCELLED" },
    { label: "Canceladas", value: "cancelled", predicate: (item) => item.status === "CANCELLED" },
    { label: "Todas", value: "all", predicate: () => true },
  ];

  return (
    <CRUDListPage<TripItem, TripForm>
      title="Viagens"
      subtitle="Cadastro de viagens, datas e vínculo com rotas e ônibus."
      meta={<span className="badge">MVP</span>}
      formTitle="Nova viagem"
      listTitle="Viagens cadastradas"
      createLabel="Criar viagem"
      updateLabel="Salvar viagem"
      emptyState={{
        title: "Nenhuma viagem encontrada",
        description: "Tente ajustar a busca ou cadastre uma nova viagem.",
      }}
      formFields={formFields}
      columns={columns}
      initialForm={{ route_id: "", bus_id: "", driver_id: "", departure_at: "" }}
      mapItemToForm={(item) => ({
        route_id: item.route_id,
        bus_id: item.bus_id,
        driver_id: item.driver_id ?? "",
        departure_at: item.departure_at ? item.departure_at.slice(0, 16) : "",
      })}
      getId={(item) => item.id}
      fetchItems={async ({ page, pageSize }) =>
        apiGet<TripItem[]>(`/trips?limit=${pageSize}&offset=${page * pageSize}`)
      }
      createItem={(form) =>
        apiPost("/trips", {
          route_id: form.route_id,
          bus_id: form.bus_id,
          driver_id: form.driver_id || undefined,
          departure_at: new Date(form.departure_at).toISOString(),
        })
      }
      updateItem={(id, form) =>
        apiPatch(`/trips/${id}`, {
          route_id: form.route_id,
          bus_id: form.bus_id,
          driver_id: form.driver_id || undefined,
          departure_at: new Date(form.departure_at).toISOString(),
        })
      }
      softDeleteItem={(item) => apiPatch(`/trips/${item.id}`, { status: "CANCELLED" })}
      restoreItem={(item) => apiPatch(`/trips/${item.id}`, { status: "SCHEDULED" })}
      getIsActive={(item) => item.status !== "CANCELLED"}
      searchFilter={(item, term) => {
        const matchesStatus = statusFilter === "ALL" ? true : item.status === statusFilter;
        if (!matchesStatus) return false;
        if (!term) return true;
        const route = routeLabel(item.route_id).toLowerCase();
        const bus = busLabel(item.bus_id).toLowerCase();
        return route.includes(term) || bus.includes(term) || item.id.toLowerCase().includes(term);
      }}
      visibilityOptions={visibilityOptions}
      visibilityDefault="active"
      layout="split"
      extraFilters={
        <select
          className="input"
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          aria-label="Filtrar status"
        >
          <option value="ALL">Todos os status</option>
          <option value="SCHEDULED">Programada</option>
          <option value="IN_PROGRESS">Em andamento</option>
          <option value="COMPLETED">Concluída</option>
          <option value="CANCELLED">Cancelada</option>
        </select>
      }
    />
  );
}
