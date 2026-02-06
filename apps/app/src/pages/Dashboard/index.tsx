import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import InlineAlert from "../../components/InlineAlert";
import LoadingState from "../../components/LoadingState";
import PageHeader from "../../components/PageHeader";
import StatusBadge from "../../components/StatusBadge";
import StatCard from "../../components/StatCard";
import { formatDateTime } from "../../utils/format";
import { apiGet } from "../../services/api";

type TripItem = { status: string };
type BookingItem = { status: string; remainder_amount: number };
type PaymentItem = { status: string; created_at: string };

type Summary = {
  activeTrips: number;
  pendingBookings: number;
  paymentsToday: number;
  alerts: number;
};

const emptySummary: Summary = {
  activeTrips: 0,
  pendingBookings: 0,
  paymentsToday: 0,
  alerts: 0,
};

function getTodayStartISO() {
  const now = new Date();
  const start = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0, 0);
  return start.toISOString();
}

export default function Dashboard() {
  const [summary, setSummary] = useState<Summary>(emptySummary);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [updatedAt, setUpdatedAt] = useState<Date | null>(null);

  const load = async () => {
    try {
      setLoading(true);
      setError(null);
      const today = getTodayStartISO();
      const [trips, bookings, payments] = await Promise.all([
        apiGet<TripItem[]>("/trips?limit=500&offset=0"),
        apiGet<BookingItem[]>("/bookings?limit=500&offset=0"),
        apiGet<PaymentItem[]>(`/payments?status=PAID&paid_since=${encodeURIComponent(today)}`),
      ]);

      const activeTrips = trips.filter((trip) => ["SCHEDULED", "IN_PROGRESS"].includes(trip.status)).length;
      const pendingBookings = bookings.filter((booking) => booking.status === "PENDING").length;
      const paymentsToday = payments.filter((payment) => payment.status === "PAID").length;
      const alerts = bookings.filter((booking) => booking.remainder_amount > 0).length;

      setSummary({ activeTrips, pendingBookings, paymentsToday, alerts });
      setUpdatedAt(new Date());
    } catch (err: any) {
      setError(err.message || "Erro ao carregar dashboard");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  return (
    <section className="page">
      <PageHeader
        title="Dashboard"
        subtitle="Visão geral da operação e alertas rápidos."
        meta={
          <>
            <span className="badge">MVP</span>
            {updatedAt ? (
              <StatusBadge tone="info">Atualizado {formatDateTime(updatedAt)}</StatusBadge>
            ) : null}
          </>
        }
      />

      {error ? <InlineAlert tone="error">{error}</InlineAlert> : null}

      {loading ? (
        <LoadingState label="Atualizando indicadores..." />
      ) : (
        <div className="stat-grid">
          <StatCard
            title="Viagens ativas"
            value={summary.activeTrips}
            helper="Programadas ou em andamento."
          />
          <StatCard
            title="Reservas pendentes"
            value={summary.pendingBookings}
            helper="Aguardando confirmação."
          />
          <StatCard
            title="Pagamentos do dia"
            value={summary.paymentsToday}
            helper="Confirmados hoje."
          />
          <StatCard
            title="Alertas financeiros"
            value={summary.alerts}
            helper="Reservas com saldo pendente."
          />
        </div>
      )}

      <div className="section section-spacing">
        <div className="section-header">
          <div className="section-title">Ações rápidas</div>
        </div>
        <div className="toolbar">
          <Link className="button" to="/bookings">Nova reserva</Link>
          <Link className="button secondary" to="/payments">Registrar pagamento</Link>
          <Link className="button secondary" to="/trips">Criar viagem</Link>
          <Link className="button ghost" to="/reports">Gerar relatório</Link>
        </div>
      </div>
    </section>
  );
}
