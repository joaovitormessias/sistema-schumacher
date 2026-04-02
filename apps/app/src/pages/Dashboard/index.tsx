import { useMemo, type CSSProperties } from "react";
import { Link } from "react-router-dom";
import { Bus, Calendar, Ticket } from "lucide-react";
import InlineAlert from "../../components/InlineAlert";
import PageHeader from "../../components/PageHeader";
import EmptyState from "../../components/EmptyState";
import { Skeleton } from "../../components/feedback/SkeletonLoader";
import { useBookings } from "../../hooks/useBookings";
import { useRoutes } from "../../hooks/useRoutes";
import { useTrips } from "../../hooks/useTrips";
import { formatCurrency, formatDateTime, formatShortId } from "../../utils/format";

function isTodayInLocalTime(isoDate: string) {
  const date = new Date(isoDate);
  const now = new Date();
  return (
    date.getFullYear() === now.getFullYear() &&
    date.getMonth() === now.getMonth() &&
    date.getDate() === now.getDate()
  );
}

interface StatCardProps {
  title: string;
  value: number;
  helper: string;
  icon: React.ReactNode;
}

function StatCard({ title, value, helper, icon }: StatCardProps) {
  return (
    <div className="stat-card">
      <div className="stat-card-header">
        <div className="stat-card-title">{title}</div>
        <div style={{ color: "var(--accent)", opacity: 0.7 }}>{icon}</div>
      </div>
      <div className="stat-card-value">{value}</div>
      <div className="stat-card-helper">{helper}</div>
    </div>
  );
}

export default function Dashboard() {
  const tripsQuery = useTrips(500, 0);
  const routesQuery = useRoutes(500, 0);
  const bookingsQuery = useBookings(500, 0);

  const trips = tripsQuery.data ?? [];
  const routes = routesQuery.data ?? [];
  const bookings = bookingsQuery.data ?? [];

  const loading = tripsQuery.isLoading || routesQuery.isLoading || bookingsQuery.isLoading;

  const loadError =
    (tripsQuery.error as Error | undefined)?.message ||
    (routesQuery.error as Error | undefined)?.message ||
    (bookingsQuery.error as Error | undefined)?.message ||
    null;

  const activeTrips = trips.filter((trip) => ["SCHEDULED", "IN_PROGRESS"].includes(trip.status));
  const tripsToday = activeTrips
    .filter((trip) => isTodayInLocalTime(trip.departure_at))
    .sort((a, b) => new Date(a.departure_at).getTime() - new Date(b.departure_at).getTime())
    .slice(0, 6);

  const routeMap = new Map(routes.map((route) => [route.id, route]));
  const routeLabel = (routeId: string) => {
    const route = routeMap.get(routeId);
    if (!route) return formatShortId(routeId);
    return `${route.origin_city} -> ${route.destination_city}`;
  };

  const latestBookings = [...bookings]
    .sort((a, b) => {
      const aDate = a.created_at ? new Date(a.created_at).getTime() : 0;
      const bDate = b.created_at ? new Date(b.created_at).getTime() : 0;
      return bDate - aDate;
    })
    .slice(0, 6);

  const pendingBookings = bookings.filter((booking) => booking.status === "PENDING").length;
  const bookingsWithBalance = bookings.filter((booking) => booking.remainder_amount > 0).length;
  const todayRevenue = bookings
    .filter((booking) => booking.created_at && isTodayInLocalTime(booking.created_at))
    .reduce((acc, booking) => acc + Number(booking.deposit_amount || 0), 0);

  return (
    <section className="page">
      <PageHeader
        title="Dashboard"
        subtitle="Visao geral da operacao de passagens."
        meta={<span className="badge">Ticketing</span>}
      />

      {loadError ? <InlineAlert tone="error">{loadError}</InlineAlert> : null}

      {loading ? (
        <div className="stat-grid">
          {[1, 2, 3, 4].map((i) => (
            <Skeleton.StatCard key={i} />
          ))}
        </div>
      ) : (
        <div className="stat-grid">
          <StatCard
            title="Viagens ativas"
            value={activeTrips.length}
            helper="Programadas ou em andamento."
            icon={<Bus size={24} />}
          />
          <StatCard
            title="Reservas pendentes"
            value={pendingBookings}
            helper="Aguardando confirmacao."
            icon={<Ticket size={24} />}
          />
          <StatCard
            title="Com saldo pendente"
            value={bookingsWithBalance}
            helper="Reservas com valor em aberto."
            icon={<Calendar size={24} />}
          />
          <StatCard
            title="Entrada de hoje"
            value={Math.round(todayRevenue)}
            helper={formatCurrency(todayRevenue)}
            icon={<Ticket size={24} />}
          />
        </div>
      )}

      <div className="section section-spacing">
        <div className="section-header">
          <div className="section-title">Viagens de hoje</div>
          <Link className="button secondary sm" to="/trips">
            Ver todas
          </Link>
        </div>

        {loading ? (
          <Skeleton.Table rows={4} columns={2} />
        ) : tripsToday.length === 0 ? (
          <EmptyState
            title="Nenhuma viagem hoje"
            description="Nao ha viagens programadas para hoje."
            action={
              <Link className="button sm" to="/trips">
                Ver todas as viagens
              </Link>
            }
          />
        ) : (
          <div className="table" style={{ "--table-columns": "2fr 1fr" } as CSSProperties}>
            <div className="table-row table-head">
              <div className="table-cell">Rota</div>
              <div className="table-cell">Saida</div>
            </div>
            {tripsToday.map((trip) => (
              <div key={trip.id} className="table-row">
                <div className="table-cell" data-label="Rota">
                  {routeLabel(trip.route_id)}
                </div>
                <div className="table-cell" data-label="Saida">
                  {formatDateTime(trip.departure_at)}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="section section-spacing">
        <div className="section-header">
          <div className="section-title">Ultimas reservas</div>
          <Link className="button secondary sm" to="/bookings">
            Abrir reservas
          </Link>
        </div>

        {loading ? (
          <Skeleton.Table rows={4} columns={3} />
        ) : latestBookings.length === 0 ? (
          <EmptyState
            title="Nenhuma reserva recente"
            description="Comece criando a primeira reserva."
            action={
              <Link className="button" to="/bookings">
                Criar reserva
              </Link>
            }
          />
        ) : (
          <div className="table" style={{ "--table-columns": "2fr 1fr 1fr" } as CSSProperties}>
            <div className="table-row table-head">
              <div className="table-cell">Passageiro</div>
              <div className="table-cell">Status</div>
              <div className="table-cell">Saldo</div>
            </div>
            {latestBookings.map((booking) => (
              <div className="table-row" key={booking.id}>
                <div className="table-cell" data-label="Passageiro">
                  {booking.passenger_name}
                </div>
                <div className="table-cell" data-label="Status">
                  {booking.status}
                </div>
                <div className="table-cell" data-label="Saldo">
                  {formatCurrency(booking.remainder_amount)}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}

