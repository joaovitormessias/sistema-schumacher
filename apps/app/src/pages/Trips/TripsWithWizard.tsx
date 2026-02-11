import { useState } from "react";
import { Link } from "react-router-dom";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import StatusBadge, { type StatusTone } from "../../components/StatusBadge";
import PageHeader from "../../components/PageHeader";
import SearchToolbar from "../../components/input/SearchToolbar";
import DataTable, { type DataTableColumn } from "../../components/table/DataTable";
import { apiGet, apiPatch } from "../../services/api";
import { formatDateTime, formatShortId } from "../../utils/format";
import { tripOperationalStatusLabel, tripStatusLabel } from "../../utils/labels";
import { TripWizard } from "./TripWizard";
import useToast from "../../hooks/useToast";
import "./wizard.css";

type TripItem = {
  id: string;
  route_id: string;
  bus_id: string;
  driver_id?: string;
  request_id?: string;
  departure_at: string;
  status: string;
  operational_status: string;
  estimated_km: number;
};

const statusTone = (status: string): StatusTone => {
  if (status === "IN_PROGRESS") return "info";
  if (status === "SCHEDULED") return "success";
  if (status === "CANCELLED") return "danger";
  return "neutral";
};

export default function TripsWithWizard() {
  const [showWizard, setShowWizard] = useState(false);
  const [statusFilter, setStatusFilter] = useState("ALL");
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(0);
  const pageSize = 50;

  const toast = useToast();
  const queryClient = useQueryClient();

  const tripsQuery = useQuery({
    queryKey: ["trips", statusFilter, page, pageSize, searchQuery],
    queryFn: async () => {
      const params = new URLSearchParams({
        limit: String(pageSize),
        offset: String(page * pageSize),
      });
      if (statusFilter !== "ALL") {
        params.set("status", statusFilter);
      }
      if (searchQuery) {
        params.set("search", searchQuery);
      }
      return apiGet<TripItem[]>(`/trips?${params.toString()}`);
    },
  });

  const handleCreateTrip = async (formData: any) => {
    try {
      const response = await fetch('/api/trips', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          route_id: formData.route_id,
          bus_id: formData.bus_id,
          driver_id: formData.driver_id || undefined,
          request_id: formData.request_id || undefined,
          departure_at: new Date(formData.departure_at).toISOString(),
          estimated_km: formData.estimated_km ? Number(formData.estimated_km) : undefined,
          status: "SCHEDULED",
        }),
      });

      if (!response.ok) {
        throw new Error('Erro ao criar viagem');
      }

      toast.success("Viagem criada com sucesso!");
      setShowWizard(false);
      queryClient.invalidateQueries({ queryKey: ["trips"] });
    } catch (error: any) {
      toast.error(error.message || "Erro ao criar viagem");
      throw error;
    }
  };

  const columns: DataTableColumn<TripItem>[] = [
    {
      label: "Rota",
      accessor: (item) => formatShortId(item.route_id),
    },
    {
      label: "Ônibus",
      accessor: (item) => formatShortId(item.bus_id),
    },
    {
      label: "Saída",
      accessor: (item) => formatDateTime(item.departure_at),
    },
    {
      label: "Status",
      render: (item) => (
        <StatusBadge tone={statusTone(item.status)}>
          {tripStatusLabel[item.status] ?? item.status}
        </StatusBadge>
      ),
    },
    {
      label: "Operacional",
      render: (item) => (
        <StatusBadge tone={item.operational_status === "CLOSED" ? "success" : "info"}>
          {tripOperationalStatusLabel[item.operational_status] ?? item.operational_status}
        </StatusBadge>
      ),
    },
    {
      label: "Ações",
      render: (item) => (
        <Link className="button secondary sm" to={`/trips/${item.id}/operations`}>
          Operações
        </Link>
      ),
    },
  ];

  const trips = tripsQuery.data ?? [];

  return (
    <div className="page">
      <PageHeader
        eyebrow="OPERAÇÃO"
        title="Viagens"
        subtitle="Cadastro de viagens, datas e vínculo com rotas e ônibus."
        meta={<span className="badge">Core</span>}
        primaryAction={
          <button className="button" onClick={() => setShowWizard(true)}>
            Criar viagem
          </button>
        }
      />

      {showWizard && (
        <div style={{ marginBottom: "var(--space-6)" }}>
          <TripWizard
            onSubmit={handleCreateTrip}
            onCancel={() => setShowWizard(false)}
          />
        </div>
      )}

      {!showWizard && (
        <div className="section">
          <div className="section-header">
            <h3 className="section-title">Viagens cadastradas</h3>
          </div>

          <div style={{ display: "flex", gap: "var(--space-3)", marginBottom: "var(--space-4)" }}>
            <SearchToolbar
              placeholder="Buscar viagens..."
              value={searchQuery}
              onChange={setSearchQuery}
            />
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
          </div>

          <DataTable
            columns={columns}
            data={trips}
            loading={tripsQuery.isPending}
            getId={(item) => item.id}
            emptyState={{
              title: "Nenhuma viagem encontrada",
              description: "Tente ajustar a busca ou cadastre uma nova viagem.",
            }}
          />
        </div>
      )}
    </div>
  );
}
