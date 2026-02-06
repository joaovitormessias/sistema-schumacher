import { useMemo, useState } from "react";
import CRUDListPage, { type ColumnConfig, type FormFieldConfig } from "../../components/layout/CRUDListPage";
import { apiGet, apiPatch, apiPost } from "../../services/api";
import type { FiscalDocument } from "../../types/financial";
import { formatCurrency } from "../../utils/financialLabels";
import { formatDateTime, formatShortId } from "../../utils/format";

type FiscalDocumentForm = {
  trip_id: string;
  document_type: string;
  document_number: string;
  issue_date: string;
  amount: number;
  recipient_name: string;
  recipient_document: string;
  status: string;
  external_id: string;
  metadata: string;
};

type TripItem = { id: string; route_id: string; departure_at: string };

type RouteItem = { id: string; origin_city: string; destination_city: string };

export default function FiscalDocuments() {
  const [trips, setTrips] = useState<TripItem[]>([]);
  const [routes, setRoutes] = useState<RouteItem[]>([]);

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
      ? `${route.origin_city} → ${route.destination_city}`
      : formatShortId(trip.route_id);
    return `${routeLabel} • ${formatDateTime(trip.departure_at)}`;
  };

  const formFields: FormFieldConfig<FiscalDocumentForm>[] = [
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
      key: "document_type",
      label: "Tipo de documento",
      required: true,
    },
    {
      key: "document_number",
      label: "Número",
    },
    {
      key: "issue_date",
      label: "Data de emissão",
      type: "datetime",
    },
    {
      key: "amount",
      label: "Valor",
      type: "number",
      required: true,
      inputProps: { min: 0, step: 0.01 },
    },
    {
      key: "recipient_name",
      label: "Destinatário",
    },
    {
      key: "recipient_document",
      label: "Documento do destinatário",
    },
    {
      key: "status",
      label: "Status",
    },
    {
      key: "external_id",
      label: "ID externo",
    },
    {
      key: "metadata",
      label: "Metadata (JSON)",
      type: "textarea",
      colSpan: "full",
      hint: "Cole um JSON válido se precisar de dados extras",
    },
  ];

  const columns: ColumnConfig<FiscalDocument>[] = [
    { label: "Viagem", accessor: (item) => tripLabel(item.trip_id) },
    { label: "Tipo", accessor: (item) => item.document_type },
    { label: "Número", accessor: (item) => item.document_number ?? "-" },
    { label: "Valor", accessor: (item) => formatCurrency(item.amount) },
    { label: "Status", accessor: (item) => item.status },
    { label: "Emissão", accessor: (item) => formatDateTime(item.issue_date) },
  ];

  const parseMetadata = (raw: string) => {
    const trimmed = raw.trim();
    if (!trimmed) return undefined;
    return JSON.parse(trimmed);
  };

  return (
    <CRUDListPage<FiscalDocument, FiscalDocumentForm>
      title="Documentos Fiscais"
      subtitle="Registro básico de NFS-e, CT-e e outros documentos."
      formTitle="Novo documento"
      listTitle="Documentos registrados"
      createLabel="Criar documento"
      updateLabel="Salvar documento"
      emptyState={{
        title: "Nenhum documento encontrado",
        description: "Cadastre um documento fiscal para começar.",
      }}
      formFields={formFields}
      columns={columns}
      initialForm={{
        trip_id: "",
        document_type: "",
        document_number: "",
        issue_date: "",
        amount: 0,
        recipient_name: "",
        recipient_document: "",
        status: "PENDING",
        external_id: "",
        metadata: "",
      }}
      mapItemToForm={(item) => ({
        trip_id: item.trip_id,
        document_type: item.document_type,
        document_number: item.document_number ?? "",
        issue_date: item.issue_date ? item.issue_date.slice(0, 16) : "",
        amount: item.amount,
        recipient_name: item.recipient_name ?? "",
        recipient_document: item.recipient_document ?? "",
        status: item.status ?? "",
        external_id: item.external_id ?? "",
        metadata: item.metadata ? JSON.stringify(item.metadata, null, 2) : "",
      })}
      getId={(item) => item.id}
      fetchItems={async ({ page, pageSize }) => {
        const data = await apiGet<FiscalDocument[]>(
          `/fiscal-documents?limit=${pageSize}&offset=${page * pageSize}`
        );
        const [tripsData, routesData] = await Promise.all([
          apiGet<TripItem[]>("/trips?limit=500&offset=0"),
          apiGet<RouteItem[]>("/routes?limit=500&offset=0"),
        ]);
        setTrips(tripsData);
        setRoutes(routesData);
        return data;
      }}
      createItem={async (form) => {
        const metadata = parseMetadata(form.metadata);
        await apiPost("/fiscal-documents", {
          trip_id: form.trip_id,
          document_type: form.document_type,
          document_number: form.document_number || undefined,
          issue_date: form.issue_date ? new Date(form.issue_date).toISOString() : undefined,
          amount: Number(form.amount),
          recipient_name: form.recipient_name || undefined,
          recipient_document: form.recipient_document || undefined,
          status: form.status || undefined,
          external_id: form.external_id || undefined,
          metadata,
        });
      }}
      updateItem={async (id, form) => {
        const metadata = parseMetadata(form.metadata);
        await apiPatch(`/fiscal-documents/${id}`, {
          document_number: form.document_number || undefined,
          issue_date: form.issue_date ? new Date(form.issue_date).toISOString() : undefined,
          amount: Number(form.amount),
          recipient_name: form.recipient_name || undefined,
          recipient_document: form.recipient_document || undefined,
          status: form.status || undefined,
          external_id: form.external_id || undefined,
          metadata,
        });
      }}
      searchFilter={(item, term) => {
        const trip = tripLabel(item.trip_id).toLowerCase();
        return (
          trip.includes(term) ||
          item.document_type.toLowerCase().includes(term) ||
          item.document_number?.toLowerCase().includes(term) ||
          item.status?.toLowerCase().includes(term)
        );
      }}
    />
  );
}
