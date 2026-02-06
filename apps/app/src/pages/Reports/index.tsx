import { useEffect, useMemo, useState, type CSSProperties } from "react";
import EmptyState from "../../components/EmptyState";
import FormField from "../../components/FormField";
import InlineAlert from "../../components/InlineAlert";
import LoadingState from "../../components/LoadingState";
import PageHeader from "../../components/PageHeader";
import StatusBadge, { type StatusTone } from "../../components/StatusBadge";
import { formatCurrency, formatDateTime } from "../../utils/format";
import { apiBaseUrl, apiGet } from "../../services/api";

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
  const [trips, setTrips] = useState<TripItem[]>([]);
  const [rows, setRows] = useState<ReportRow[]>([]);
  const [tripId, setTripId] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    apiGet<TripItem[]>("/trips?limit=200&offset=0")
      .then(setTrips)
      .catch((err) => setError(err.message || "Erro ao carregar viagens"));
  }, []);

  const loadReport = async () => {
    if (!tripId) return;
    try {
      setLoading(true);
      setError(null);
      const data = await apiGet<ReportRow[]>(`/reports/passengers?trip_id=${tripId}`);
      setRows(data);
    } catch (err: any) {
      setError(err.message || "Erro ao gerar relatório");
    } finally {
      setLoading(false);
    }
  };

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

  return (
    <section className="page">
      <PageHeader
        title="Relatórios"
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
              onChange={(e) => setTripId(e.target.value)}
            >
              <option value="">Selecione a viagem</option>
              {trips.map((trip) => (
                <option key={trip.id} value={trip.id}>
                  {formatDateTime(trip.departure_at)} • {trip.id.slice(0, 8)}
                </option>
              ))}
            </select>
          </FormField>
          <div className="form-actions full-span">
            <button className="button" type="button" onClick={loadReport} disabled={!tripId}>
              Gerar relatório
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

      {error ? <InlineAlert tone="error">{error}</InlineAlert> : null}

      {loading ? (
        <LoadingState label="Gerando relatório..." />
      ) : rows.length === 0 ? (
        <EmptyState
          title="Nenhum relatório gerado"
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

          <div
            className="table"
            style={{ "--table-columns": "repeat(4, minmax(0, 1fr))" } as CSSProperties}
          >
            <div className="table-row table-head">
              <div className="table-cell">Poltrona</div>
              <div className="table-cell">Passageiro</div>
              <div className="table-cell">Status pagamento</div>
              <div className="table-cell">Pago</div>
            </div>
            {rows.map((row) => (
              <div className="table-row" key={row.passenger_id}>
                <div className="table-cell" data-label="Poltrona">{row.seat_number}</div>
                <div className="table-cell" data-label="Passageiro">{row.name}</div>
                <div className="table-cell" data-label="Status">
                  <StatusBadge tone={statusTone(row.payment_stage)}>{row.payment_stage}</StatusBadge>
                </div>
                <div className="table-cell" data-label="Pago">{formatCurrency(row.amount_paid)}</div>
              </div>
            ))}
          </div>
        </div>
      )}
    </section>
  );
}
