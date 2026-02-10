import { useMemo, useState, type CSSProperties, type FormEvent, type ReactNode } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate } from "react-router-dom";
import { Bus, Calendar, CreditCard, AlertTriangle } from "lucide-react";
import InlineAlert from "../../components/InlineAlert";
import PageHeader from "../../components/PageHeader";
import StatusBadge from "../../components/StatusBadge";
import EmptyState from "../../components/EmptyState";
import { Skeleton } from "../../components/feedback/SkeletonLoader";
import Modal from "../../components/overlay/Modal";
import { useBookings } from "../../hooks/useBookings";
import { useRoutes } from "../../hooks/useRoutes";
import { usePayments } from "../../hooks/usePayments";
import { useTrips } from "../../hooks/useTrips";
import useToast from "../../hooks/useToast";
import { apiPost } from "../../services/api";
import { formatCurrency, formatDateTime, formatShortId } from "../../utils/format";

type Summary = {
  activeTrips: number;
  pendingBookings: number;
  paymentsToday: number;
  alerts: number;
};

type CreatePaymentResponse = {
  payment: { id: string; status: string; amount: number; provider_ref?: string };
  provider_raw?: unknown;
};

function isTodayInLocalTime(isoDate: string) {
  const date = new Date(isoDate);
  const now = new Date();
  return (
    date.getFullYear() === now.getFullYear() &&
    date.getMonth() === now.getMonth() &&
    date.getDate() === now.getDate()
  );
}

function toDateKey(isoDate: string) {
  const date = new Date(isoDate);
  return [date.getFullYear(), String(date.getMonth() + 1).padStart(2, "0"), String(date.getDate()).padStart(2, "0")].join("-");
}

interface InteractiveStatCardProps {
  title: string;
  value: number;
  helper: string;
  icon: ReactNode;
  to?: string;
  onClick?: () => void;
}

function InteractiveStatCard({ title, value, helper, icon, to, onClick }: InteractiveStatCardProps) {
  const navigate = useNavigate();

  const handleClick = () => {
    if (onClick) {
      onClick();
    } else if (to) {
      navigate(to);
    }
  };

  const isClickable = Boolean(to || onClick);

  return (
    <div
      className={`stat-card ${isClickable ? "stat-card-interactive" : ""}`}
      onClick={isClickable ? handleClick : undefined}
      role={isClickable ? "button" : undefined}
      tabIndex={isClickable ? 0 : undefined}
      onKeyDown={(e) => {
        if (isClickable && (e.key === "Enter" || e.key === " ")) {
          e.preventDefault();
          handleClick();
        }
      }}
    >
      <div className="stat-card-header">
        <div className="stat-card-title">{title}</div>
        <div style={{ color: "var(--accent)", opacity: 0.7 }}>{icon}</div>
      </div>
      <div className="stat-card-value">{value}</div>
      <div className="stat-card-helper">{helper}</div>
    </div>
  );
}

function StatCardSkeleton() {
  return (
    <div className="stat-grid">
      {[1, 2, 3, 4].map((i) => (
        <Skeleton.StatCard key={i} />
      ))}
    </div>
  );
}

function TableSkeleton({ columns }: { columns: number }) {
  return <Skeleton.Table rows={4} columns={columns} />;
}

export default function Dashboard() {
  const toast = useToast();
  const queryClient = useQueryClient();
  const tripsQuery = useTrips(500, 0);
  const routesQuery = useRoutes(500, 0);
  const bookingsQuery = useBookings(500, 0);
  const paymentsQuery = usePayments(500, 0);

  const trips = tripsQuery.data ?? [];
  const routes = routesQuery.data ?? [];
  const bookings = bookingsQuery.data ?? [];
  const payments = paymentsQuery.data ?? [];

  const [chargeTargetId, setChargeTargetId] = useState<string | null>(null);
  const [chargeAmount, setChargeAmount] = useState<number>(0);
  const [chargeMethod, setChargeMethod] = useState("PIX");
  const [chargeLoading, setChargeLoading] = useState(false);
  const [chargeError, setChargeError] = useState<string | null>(null);

  const loading =
    tripsQuery.isLoading || routesQuery.isLoading || bookingsQuery.isLoading || paymentsQuery.isLoading;

  const loadError =
    (tripsQuery.error as Error | undefined)?.message ||
    (routesQuery.error as Error | undefined)?.message ||
    (bookingsQuery.error as Error | undefined)?.message ||
    (paymentsQuery.error as Error | undefined)?.message ||
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

  const paidToday = payments.filter(
    (payment) => payment.status === "PAID" && payment.paid_at && isTodayInLocalTime(payment.paid_at)
  );

  const summary: Summary = {
    activeTrips: activeTrips.length,
    pendingBookings: bookings.filter((booking) => booking.status === "PENDING").length,
    paymentsToday: paidToday.length,
    alerts: bookings.filter((booking) => booking.remainder_amount > 0).length,
  };

  const revenueLast7Days = useMemo(() => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const days = Array.from({ length: 7 }, (_, index) => {
      const day = new Date(today);
      day.setDate(today.getDate() - (6 - index));
      return {
        key: toDateKey(day.toISOString()),
        label: new Intl.DateTimeFormat("pt-BR", { weekday: "short" }).format(day).replace(".", ""),
        total: 0,
      };
    });

    const byKey = new Map(days.map((day) => [day.key, day]));

    for (const payment of payments) {
      if (payment.status !== "PAID") continue;
      const sourceDate = payment.paid_at ?? payment.created_at;
      if (!sourceDate) continue;
      const key = toDateKey(sourceDate);
      const target = byKey.get(key);
      if (!target) continue;
      target.total += Number(payment.amount) || 0;
    }

    return days;
  }, [payments]);

  const maxRevenueValue = useMemo(
    () => Math.max(...revenueLast7Days.map((item) => item.total), 1),
    [revenueLast7Days]
  );

  const chargeTarget = useMemo(
    () => bookings.find((booking) => booking.id === chargeTargetId) ?? null,
    [bookings, chargeTargetId]
  );

  const openChargeModal = (bookingId: string) => {
    const booking = bookings.find((item) => item.id === bookingId);
    if (!booking) return;
    setChargeTargetId(booking.id);
    setChargeAmount(Number(booking.remainder_amount) || 0);
    setChargeMethod("PIX");
    setChargeError(null);
  };

  const closeChargeModal = () => {
    if (chargeLoading) return;
    setChargeTargetId(null);
    setChargeError(null);
  };

  const handleInlineCharge = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!chargeTarget) return;
    if (!chargeAmount || chargeAmount <= 0) {
      setChargeError("Informe um valor maior que zero para gerar a cobranca.");
      return;
    }

    try {
      setChargeLoading(true);
      setChargeError(null);
      await apiPost<CreatePaymentResponse>("/payments", {
        booking_id: chargeTarget.id,
        amount: Number(chargeAmount),
        method: chargeMethod,
        description: "Saldo pendente da reserva",
      });
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["payments"] }),
        queryClient.invalidateQueries({ queryKey: ["bookings"] }),
        queryClient.invalidateQueries({ queryKey: ["reports"] }),
      ]);
      toast.success("Cobranca de saldo gerada com sucesso.");
      setChargeTargetId(null);
    } catch (err: any) {
      setChargeError(err?.message || "Erro ao gerar cobranca de saldo.");
    } finally {
      setChargeLoading(false);
    }
  };

  return (
    <section className="page">
      <PageHeader
        title="Dashboard"
        subtitle="Visao geral da operacao com foco em viagens de hoje e pendencias financeiras."
        meta={<span className="badge">MVP</span>}
      />

      {loadError ? <InlineAlert tone="error">{loadError}</InlineAlert> : null}

      {loading ? (
        <StatCardSkeleton />
      ) : (
        <div className="stat-grid">
          <InteractiveStatCard
            title="Viagens ativas"
            value={summary.activeTrips}
            helper="Programadas ou em andamento."
            icon={<Bus size={24} />}
            to="/trips?status=SCHEDULED,IN_PROGRESS"
          />
          <InteractiveStatCard
            title="Reservas pendentes"
            value={summary.pendingBookings}
            helper="Aguardando confirmacao."
            icon={<Calendar size={24} />}
            to="/bookings?status=PENDING"
          />
          <InteractiveStatCard
            title="Pagamentos do dia"
            value={summary.paymentsToday}
            helper="Confirmados hoje."
            icon={<CreditCard size={24} />}
            to="/payments"
          />
          <InteractiveStatCard
            title="Alertas financeiros"
            value={summary.alerts}
            helper="Reservas com saldo pendente."
            icon={<AlertTriangle size={24} />}
            to="/bookings?has_balance=true"
          />
        </div>
      )}

      <div className="section section-spacing">
        <div className="section-header">
          <div className="section-title">Receita dos ultimos 7 dias</div>
          <Link className="button secondary sm" to="/reports">
            Abrir relatorios
          </Link>
        </div>
        {loading ? (
          <Skeleton.Table rows={1} columns={7} />
        ) : (
          <div className="revenue-chart">
            {revenueLast7Days.map((item) => {
              const heightPercent = Math.max((item.total / maxRevenueValue) * 100, item.total > 0 ? 8 : 0);
              return (
                <div className="revenue-chart-item" key={item.key}>
                  <div className="revenue-chart-track">
                    <div className="revenue-chart-fill" style={{ height: `${heightPercent}%` }} />
                  </div>
                  <div className="revenue-chart-label">{item.label}</div>
                  <div className="revenue-chart-value">{formatCurrency(item.total)}</div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      <div className="section section-spacing">
        <div className="section-header">
          <div className="section-title">Viagens de hoje</div>
          <Link className="button secondary sm" to="/trips">
            Ver todas
          </Link>
        </div>

        {loading ? (
          <TableSkeleton columns={2} />
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
              <Link
                to={`/trips?id=${trip.id}`}
                key={trip.id}
                className="table-row card-interactive"
                style={{ textDecoration: "none", color: "inherit" }}
              >
                <div className="table-cell" data-label="Rota">
                  {routeLabel(trip.route_id)}
                </div>
                <div className="table-cell" data-label="Saida">
                  {formatDateTime(trip.departure_at)}
                </div>
              </Link>
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
          <TableSkeleton columns={4} />
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
          <div className="table" style={{ "--table-columns": "2fr 1fr 1fr 1fr" } as CSSProperties}>
            <div className="table-row table-head">
              <div className="table-cell">Passageiro</div>
              <div className="table-cell">Status</div>
              <div className="table-cell">Saldo</div>
              <div className="table-cell">Acao</div>
            </div>
            {latestBookings.map((booking) => (
              <div className="table-row" key={booking.id}>
                <div className="table-cell" data-label="Passageiro">
                  {booking.passenger_name}
                </div>
                <div className="table-cell" data-label="Status">
                  <StatusBadge tone={booking.status === "CONFIRMED" ? "success" : "warning"}>
                    {booking.status}
                  </StatusBadge>
                </div>
                <div className="table-cell" data-label="Saldo">
                  {formatCurrency(booking.remainder_amount)}
                </div>
                <div className="table-cell" data-label="Acao">
                  {booking.remainder_amount > 0 ? (
                    <div className="toolbar-group">
                      <button
                        className="button sm button-delightful"
                        type="button"
                        onClick={() => openChargeModal(booking.id)}
                      >
                        Cobrar saldo
                      </button>
                      <Link className="button secondary sm" to={`/payments?booking_id=${booking.id}&mode=AUTOMATIC`}>
                        Ver no modulo
                      </Link>
                    </div>
                  ) : (
                    <span className="text-muted">Sem saldo</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <Modal
        open={Boolean(chargeTarget)}
        title="Cobrar saldo pendente"
        description="Gere uma cobranca rapida para a reserva selecionada."
        onClose={closeChargeModal}
        footer={
          <>
            <button className="button secondary" type="button" onClick={closeChargeModal} disabled={chargeLoading}>
              Cancelar
            </button>
            <button
              className={`button button-delightful ${chargeLoading ? "loading" : ""}`}
              type="submit"
              form="charge-modal-form"
              disabled={chargeLoading}
            >
              {chargeLoading ? "Gerando cobranca..." : "Gerar cobranca"}
            </button>
          </>
        }
      >
        {chargeTarget ? (
          <form id="charge-modal-form" className="form-grid" onSubmit={handleInlineCharge}>
            <label className="form-field">
              <span className="form-label">Reserva</span>
              <input
                className="input"
                readOnly
                value={`${chargeTarget.passenger_name} (${formatShortId(chargeTarget.id)})`}
              />
            </label>
            <label className="form-field">
              <span className="form-label">Metodo</span>
              <select
                className="input"
                value={chargeMethod}
                onChange={(event) => setChargeMethod(event.target.value)}
                disabled={chargeLoading}
              >
                <option value="PIX">PIX</option>
                <option value="CARD">Cartao</option>
                <option value="TRANSFER">Transferencia</option>
                <option value="CASH">Dinheiro</option>
                <option value="OTHER">Outro</option>
              </select>
            </label>
            <label className="form-field">
              <span className="form-label">Valor da cobranca</span>
              <input
                className="input"
                type="number"
                min={0}
                step="0.01"
                value={chargeAmount}
                onChange={(event) => setChargeAmount(Number(event.target.value))}
                disabled={chargeLoading}
              />
              <span className="form-hint">Saldo atual: {formatCurrency(chargeTarget.remainder_amount)}</span>
            </label>
            {chargeError ? <InlineAlert tone="error">{chargeError}</InlineAlert> : null}
          </form>
        ) : null}
      </Modal>
    </section>
  );
}
