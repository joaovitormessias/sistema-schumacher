import { useMemo, useState } from "react";
import CRUDListPage, {
  type ColumnConfig,
  type FormFieldConfig,
  type VisibilityOption,
} from "../../components/layout/CRUDListPage";
import StatusBadge from "../../components/StatusBadge";
import useToast from "../../hooks/useToast";
import { apiGet, apiPatch, apiPost } from "../../services/api";
import type { DriverCard, TripExpense } from "../../types/financial";
import {
  expenseTypeLabel,
  formatCurrency,
  paymentMethodLabel,
} from "../../utils/financialLabels";
import { formatDateTime, formatShortId } from "../../utils/format";

type TripExpenseForm = {
  trip_id: string;
  driver_id: string;
  expense_type: string;
  amount: number;
  description: string;
  expense_date: string;
  payment_method: string;
  driver_card_id: string;
  receipt_number: string;
  notes: string;
};

type TripItem = { id: string; route_id: string; departure_at: string };

type RouteItem = { id: string; origin_city: string; destination_city: string };

type DriverItem = { id: string; name: string };

export default function TripExpenses() {
  const toast = useToast();
  const [trips, setTrips] = useState<TripItem[]>([]);
  const [routes, setRoutes] = useState<RouteItem[]>([]);
  const [drivers, setDrivers] = useState<DriverItem[]>([]);
  const [cards, setCards] = useState<DriverCard[]>([]);
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
  const cardMap = useMemo(
    () => new Map(cards.map((card) => [card.id, card.card_number])),
    [cards]
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

  const formFields: FormFieldConfig<TripExpenseForm>[] = useMemo(
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
        key: "expense_type",
        label: "Tipo de despesa",
        type: "select",
        required: true,
        options: [
          { label: "Selecione o tipo", value: "" },
          { label: "Combustível", value: "FUEL" },
          { label: "Alimentação", value: "FOOD" },
          { label: "Hospedagem", value: "LODGING" },
          { label: "Pedágio", value: "TOLL" },
          { label: "Manutenção", value: "MAINTENANCE" },
          { label: "Outros", value: "OTHER" },
        ],
      },
      {
        key: "amount",
        label: "Valor",
        type: "number",
        required: true,
        inputProps: { min: 0, step: 0.01 },
      },
      {
        key: "payment_method",
        label: "Forma de pagamento",
        type: "select",
        required: true,
        options: [
          { label: "Selecione a forma", value: "" },
          { label: "Adiantamento", value: "ADVANCE" },
          { label: "Cartão", value: "CARD" },
          { label: "Pessoal", value: "PERSONAL" },
          { label: "Empresa", value: "COMPANY" },
        ],
      },
      {
        key: "driver_card_id",
        label: "Cartão do motorista",
        type: "select",
        options: [
          { label: "Selecionar cartão (se aplicável)", value: "" },
          ...cards.map((card) => ({
            label: `${card.card_number} • ${card.card_type}`,
            value: card.id,
          })),
        ],
      },
      {
        key: "expense_date",
        label: "Data da despesa",
        type: "datetime",
        required: true,
      },
      {
        key: "description",
        label: "Descrição",
        type: "textarea",
        required: true,
        colSpan: "full",
      },
      {
        key: "receipt_number",
        label: "Número do comprovante",
        type: "text",
      },
      {
        key: "notes",
        label: "Observações",
        type: "textarea",
        colSpan: "full",
      },
    ],
    [trips, drivers, cards, routeMap]
  );

  const columns: ColumnConfig<TripExpense>[] = [
    { label: "Viagem", accessor: (item) => tripLabel(item.trip_id) },
    {
      label: "Motorista",
      accessor: (item) => driverMap.get(item.driver_id) ?? formatShortId(item.driver_id),
    },
    {
      label: "Tipo",
      accessor: (item) => expenseTypeLabel[item.expense_type] ?? item.expense_type,
    },
    { label: "Valor", accessor: (item) => formatCurrency(item.amount) },
    {
      label: "Pagamento",
      accessor: (item) => paymentMethodLabel[item.payment_method] ?? item.payment_method,
    },
    {
      label: "Aprovação",
      render: (item) => (
        <StatusBadge tone={item.is_approved ? "success" : "warning"}>
          {item.is_approved ? "Aprovada" : "Pendente"}
        </StatusBadge>
      ),
    },
    { label: "Data", accessor: (item) => formatDateTime(item.expense_date) },
  ];

  const visibilityOptions: VisibilityOption<TripExpense>[] = [
    { label: "Pendentes", value: "pending", predicate: (item) => !item.is_approved },
    { label: "Aprovadas", value: "approved", predicate: (item) => item.is_approved },
    { label: "Todas", value: "all", predicate: () => true },
  ];

  const handleApprove = async (item: TripExpense) => {
    if (!window.confirm("Confirmar aprovação da despesa?")) return;
    try {
      await apiPost(`/trip-expenses/${item.id}/approve`, {});
      toast.success("Despesa aprovada.");
      setReloadKey((value) => value + 1);
    } catch (err: any) {
      toast.error(err.message || "Erro ao aprovar despesa");
    }
  };

  return (
    <CRUDListPage<TripExpense, TripExpenseForm>
      key={reloadKey}
      title="Despesas de Viagem"
      subtitle="Controle de despesas registradas durante a viagem."
      formTitle="Nova despesa"
      listTitle="Despesas registradas"
      createLabel="Criar despesa"
      updateLabel="Salvar despesa"
      emptyState={{
        title: "Nenhuma despesa encontrada",
        description: "Cadastre uma despesa para começar.",
      }}
      formFields={formFields}
      columns={columns}
      initialForm={{
        trip_id: "",
        driver_id: "",
        expense_type: "",
        amount: 0,
        description: "",
        expense_date: "",
        payment_method: "ADVANCE",
        driver_card_id: "",
        receipt_number: "",
        notes: "",
      }}
      mapItemToForm={(item) => ({
        trip_id: item.trip_id,
        driver_id: item.driver_id,
        expense_type: item.expense_type,
        amount: item.amount,
        description: item.description,
        expense_date: item.expense_date ? item.expense_date.slice(0, 16) : "",
        payment_method: item.payment_method,
        driver_card_id: item.driver_card_id ?? "",
        receipt_number: item.receipt_number ?? "",
        notes: item.notes ?? "",
      })}
      getId={(item) => item.id}
      fetchItems={async ({ page, pageSize }) => {
        const data = await apiGet<TripExpense[]>(
          `/trip-expenses?limit=${pageSize}&offset=${page * pageSize}`
        );
        const [tripsData, routesData, driversData, cardsData] = await Promise.all([
          apiGet<TripItem[]>("/trips?limit=500&offset=0"),
          apiGet<RouteItem[]>("/routes?limit=500&offset=0"),
          apiGet<DriverItem[]>("/drivers?limit=500&offset=0"),
          apiGet<DriverCard[]>("/driver-cards?limit=500&offset=0"),
        ]);
        setTrips(tripsData);
        setRoutes(routesData);
        setDrivers(driversData);
        setCards(cardsData);
        return data;
      }}
      createItem={(form) =>
        apiPost("/trip-expenses", {
          trip_id: form.trip_id,
          driver_id: form.driver_id,
          expense_type: form.expense_type,
          amount: Number(form.amount),
          description: form.description,
          expense_date: new Date(form.expense_date).toISOString(),
          payment_method: form.payment_method,
          driver_card_id: form.driver_card_id || undefined,
          receipt_number: form.receipt_number || undefined,
          notes: form.notes || undefined,
        })
      }
      updateItem={(id, form) =>
        apiPatch(`/trip-expenses/${id}`, {
          amount: Number(form.amount),
          description: form.description,
          expense_date: new Date(form.expense_date).toISOString(),
          receipt_number: form.receipt_number || undefined,
          notes: form.notes || undefined,
        })
      }
      searchFilter={(item, term) => {
        const trip = tripLabel(item.trip_id).toLowerCase();
        const driver = (driverMap.get(item.driver_id) ?? "").toLowerCase();
        const card = item.driver_card_id
          ? (cardMap.get(item.driver_card_id) ?? "").toLowerCase()
          : "";
        return (
          trip.includes(term) ||
          driver.includes(term) ||
          item.description.toLowerCase().includes(term) ||
          card.includes(term)
        );
      }}
      visibilityOptions={visibilityOptions}
      visibilityDefault="pending"
      rowActions={(item) =>
        !item.is_approved ? (
          <button
            className="button success sm"
            type="button"
            onClick={() => handleApprove(item)}
          >
            Aprovar
          </button>
        ) : null
      }
    />
  );
}
