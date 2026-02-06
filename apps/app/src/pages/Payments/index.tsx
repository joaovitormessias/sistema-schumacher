import { useEffect, useMemo, useState, type FormEvent } from "react";
import EmptyState from "../../components/EmptyState";
import InlineAlert from "../../components/InlineAlert";
import PageHeader from "../../components/PageHeader";
import useToast from "../../hooks/useToast";
import { formatCurrency } from "../../utils/format";
import { apiGet, apiPost } from "../../services/api";
import PaymentForm from "./PaymentForm";
import PaymentResult from "./PaymentResult";
import PaymentTabs from "./PaymentTabs";
import type { StatusTone } from "../../components/StatusBadge";

type BookingItem = { id: string; passenger_name: string; total_amount: number; status: string };

type PaymentResponse = {
  payment: { id: string; status: string; amount: number; provider_ref?: string };
  provider_raw?: any;
};

type TabKey = "AUTOMATIC" | "MANUAL";

type AutomaticForm = {
  booking_id: string;
  amount: number;
  method: string;
  description: string;
};

type ManualForm = {
  booking_id: string;
  amount: number;
  method: string;
  notes: string;
};

export default function Payments() {
  const toast = useToast();
  const [bookings, setBookings] = useState<BookingItem[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<PaymentResponse | null>(null);
  const [activeTab, setActiveTab] = useState<TabKey>("AUTOMATIC");

  const [form, setForm] = useState<AutomaticForm>({
    booking_id: "",
    amount: 0,
    method: "PIX",
    description: "Passagem",
  });

  const [manual, setManual] = useState<ManualForm>({
    booking_id: "",
    amount: 0,
    method: "CASH",
    notes: "",
  });

  useEffect(() => {
    apiGet<BookingItem[]>("/bookings?limit=200&offset=0")
      .then(setBookings)
      .catch((err) => setError(err.message || "Erro ao carregar reservas"));
  }, []);

  const bookingOptions = useMemo(
    () =>
      bookings.map((b) => ({
        value: b.id,
        label: `${b.passenger_name} • ${formatCurrency(b.total_amount)}`,
      })),
    [bookings]
  );

  const handlePayment = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    try {
      const res = await apiPost<PaymentResponse>("/payments", {
        booking_id: form.booking_id,
        amount: Number(form.amount),
        method: form.method,
        description: form.description,
      });
      setResult(res);
      toast.success("Cobrança gerada com sucesso.");
    } catch (err: any) {
      setError(err.message || "Erro ao criar pagamento");
    }
  };

  const handleManual = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    try {
      const res = await apiPost<PaymentResponse>("/payments/manual", {
        booking_id: manual.booking_id,
        amount: Number(manual.amount),
        method: manual.method,
        notes: manual.notes,
      });
      setResult(res);
      toast.success("Pagamento manual registrado.");
    } catch (err: any) {
      setError(err.message || "Erro ao registrar pagamento manual");
    }
  };

  const statusTone = (status: string): StatusTone => {
    if (status === "PAID" || status === "CONFIRMED") return "success";
    if (status === "PENDING") return "warning";
    if (status === "FAILED") return "danger";
    return "neutral";
  };

  return (
    <section className="page">
      <PageHeader title="Pagamentos" subtitle="Sinal, total e pagamentos manuais." meta={<span className="badge">MVP</span>} />

      <div className="section">
        <div className="section-header">
          <div className="section-title">Registrar pagamento</div>
          <PaymentTabs
            activeTab={activeTab}
            onChange={(tab) => {
              setActiveTab(tab);
              setResult(null);
              setError(null);
            }}
          />
        </div>

        {bookingOptions.length === 0 ? (
          <EmptyState
            title="Nenhuma reserva disponível"
            description="Crie uma reserva antes de registrar pagamentos."
          />
        ) : activeTab === "AUTOMATIC" ? (
          <PaymentForm
            mode="AUTOMATIC"
            bookingOptions={bookingOptions}
            value={form}
            onChange={setForm}
            onSubmit={handlePayment}
          />
        ) : (
          <PaymentForm
            mode="MANUAL"
            bookingOptions={bookingOptions}
            value={manual}
            onChange={setManual}
            onSubmit={handleManual}
          />
        )}
      </div>

      {error ? <InlineAlert tone="error">{error}</InlineAlert> : null}

      {result ? <PaymentResult result={result} statusTone={statusTone} /> : null}
    </section>
  );
}
