import { useEffect, useMemo, useState } from "react";
import EmptyState from "../../components/EmptyState";
import FormField from "../../components/FormField";
import InlineAlert from "../../components/InlineAlert";
import PageHeader from "../../components/PageHeader";
import StatusBadge, { type StatusTone } from "../../components/StatusBadge";
import { Skeleton } from "../../components/feedback/SkeletonLoader";
import DataTable, { type DataTableColumn } from "../../components/table/DataTable";
import VirtualDataTable from "../../components/data-display/VirtualDataTable";
import useMediaQuery from "../../hooks/useMediaQuery";
import { useTrips } from "../../hooks/useTrips";
import { apiBaseUrl, apiGet } from "../../services/api";
import { formatCurrency, formatDateTime } from "../../utils/format";

type TripItem = { id: string; departure_at: string };

type ReportRow = {
  passenger_id: string;
  name: string;
  document: string;
  phone: string;
  email: string;
  seat_number: number;
  booking_status: string;
  passenger_status: string;
  total_amount: number;
  deposit_amount: number;
  remainder_amount: number;
  amount_paid: number;
  payment_stage: string;
};

export default function Reports() {
  const tripsQuery = useTrips(200, 0);
  const trips = (tripsQuery.data as TripItem[] | undefined) ?? [];
  const isMobile = useMediaQuery("(max-width: 900px)");
  const [rows, setRows] = useState<ReportRow[]>([]);
  const [tripId, setTripId] = useState("");
  const [autoSelectLatest, setAutoSelectLatest] = useState(true);
  const [loading, setLoading] = useState(false);
  const [reportError, setReportError] = useState<string | null>(null);
  const tripsLoadError = (tripsQuery.error as Error | undefined)?.message || null;

  useEffect(() => {
    if (tripId || trips.length === 0 || !autoSelectLatest) return;
    const latest = [...trips].sort(
      (a, b) => new Date(b.departure_at).getTime() - new Date(a.departure_at).getTime()
    )[0];
    if (latest?.id) {
      setTripId(latest.id);
    }
  }, [autoSelectLatest, tripId, trips]);

  const loadReport = async (targetTripId: string) => {
    if (!targetTripId) return;
    try {
      setLoading(true);
      setReportError(null);
      const data = await apiGet<ReportRow[]>(`/reports/passengers?trip_id=${targetTripId}`);
      setRows(data);
    } catch (err: any) {
      setReportError(err.message || "Erro ao gerar relatorio");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!tripId) return;
    void loadReport(tripId);
  }, [tripId]);

  const totals = useMemo(() => {
    return rows.reduce(
      (acc, row) => {
        acc.total += row.total_amount;
        acc.paid += row.amount_paid;
        acc.pending += row.remainder_amount;
        return acc;
      },
      { total: 0, paid: 0, pending: 0 }
    );
  }, [rows]);

  const statusTone = (status: string): StatusTone => {
    if (status === "PAID" || status === "CONFIRMED") return "success";
    if (status === "PENDING") return "warning";
    if (status === "CANCELLED") return "danger";
    return "neutral";
  };

  const columns: DataTableColumn<ReportRow>[] = [
    { label: "Poltrona", accessor: (row) => row.seat_number, width: "90px" },
    { label: "Passageiro", accessor: (row) => row.name },
    {
      label: "Status pagamento",
      render: (row) => (
        <StatusBadge tone={statusTone(row.payment_stage)}>{row.payment_stage}</StatusBadge>
      ),
    },
    {
      label: "Pago",
      accessor: (row) => formatCurrency(row.amount_paid),
      align: "right",
      width: "130px",
    },
  ];

  const shouldVirtualize = !isMobile && rows.length > 100;

  return (
    <section className="page">
      <PageHeader
        title="Relatorios"
        subtitle="Manifesto de passageiros e status financeiro."
        meta={<span className="badge">MVP</span>}
      />

      <div className="section">
        <div className="section-header">
          <div className="section-title">Selecionar viagem</div>
        </div>
        <div className="form-grid">
          <FormField label="Viagem" required>
            <select
              className="input"
              value={tripId}
              onChange={(e) => {
                setAutoSelectLatest(true);
                setTripId(e.target.value);
              }}
            >
              <option value="">Selecione a viagem</option>
              {trips.map((trip) => (
                <option key={trip.id} value={trip.id}>
                  {formatDateTime(trip.departure_at)} - {trip.id.slice(0, 8)}
                </option>
              ))}
            </select>
          </FormField>
          <div className="form-actions full-span">
            <button className="button" type="button" onClick={() => loadReport(tripId)} disabled={!tripId}>
              Atualizar relatorio
            </button>
            <button
              className="button secondary"
              type="button"
              onClick={() => {
                setAutoSelectLatest(false);
                setTripId("");
                setRows([]);
                setReportError(null);
              }}
              disabled={!tripId && rows.length === 0}
            >
              Limpar selecao
            </button>
            {tripId ? (
              <a
                className="button secondary"
                href={`${apiBaseUrl()}/reports/passengers?trip_id=${tripId}&format=csv`}
                target="_blank"
                rel="noreferrer"
              >
                Baixar CSV
              </a>
            ) : null}
          </div>
        </div>
      </div>

      {tripsLoadError ? <InlineAlert tone="error">{tripsLoadError}</InlineAlert> : null}
      {reportError ? <InlineAlert tone="error">{reportError}</InlineAlert> : null}

      {loading ? (
        <Skeleton.Table rows={8} columns={4} />
      ) : rows.length === 0 ? (
        <EmptyState
          title="Nenhum relatorio gerado"
          description="Selecione uma viagem para visualizar o manifesto."
        />
      ) : (
        <div className="section">
          <div className="section-header">
            <div className="section-title">Resumo financeiro</div>
          </div>
          <div className="card-grid">
            <div className="card">
              <h3>Total esperado</h3>
              <p>{formatCurrency(totals.total)}</p>
            </div>
            <div className="card">
              <h3>Pago</h3>
              <p>{formatCurrency(totals.paid)}</p>
            </div>
            <div className="card">
              <h3>Saldo pendente</h3>
              <p>{formatCurrency(totals.pending)}</p>
            </div>
          </div>

          {shouldVirtualize ? (
            <VirtualDataTable
              columns={columns}
              rows={rows}
              rowKey={(row) => row.passenger_id}
              emptyState={
                <EmptyState
                  title="Nenhum passageiro encontrado"
                  description="A viagem selecionada ainda nao possui registros."
                />
              }
            />
          ) : (
            <DataTable
              columns={columns}
              rows={rows}
              rowKey={(row) => row.passenger_id}
              emptyState={
                <EmptyState
                  title="Nenhum passageiro encontrado"
                  description="A viagem selecionada ainda nao possui registros."
                />
              }
            />
          )}
        </div>
      )}
    </section>
  );
}
