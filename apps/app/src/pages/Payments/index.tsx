import { useEffect, useMemo, useState, type FormEvent } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useSearchParams } from "react-router-dom";
import DataTable, { type DataTableColumn } from "../../components/table/DataTable";
import EmptyState from "../../components/EmptyState";
import InlineAlert from "../../components/InlineAlert";
import PageHeader from "../../components/PageHeader";
import StatusBadge, { type StatusTone } from "../../components/StatusBadge";
import { useBookings } from "../../hooks/useBookings";
import { usePayments } from "../../hooks/usePayments";
import { apiPost } from "../../services/api";
import type { PaymentsSearchParams } from "../../types/checkout";
import useToast from "../../hooks/useToast";
import { formatCurrency, formatDateTime, formatShortId } from "../../utils/format";
import PaymentForm from "./PaymentForm";
import PaymentResult from "./PaymentResult";
import PaymentTabs from "./PaymentTabs";

type BookingItem = {
  id: string;
  passenger_name: string;
  total_amount: number;
  remainder_amount: number;
  status: string;
};

type PaymentResponse = {
  payment: { id: string; status: string; amount: number; provider_ref?: string };
  provider_raw?: any;
};

type PaymentItem = {
  id: string;
  booking_id: string;
  amount: number;
  method: string;
  status: string;
  provider?: string;
  provider_ref?: string;
  created_at: string;
  paid_at?: string | null;
};

type PaymentSyncResponse = {
  payment: PaymentItem;
  booking_status: string;
  synced: boolean;
};

type TabKey = "AUTOMATIC" | "MANUAL";

function isTabKey(value: string | null): value is TabKey {
  return value === "AUTOMATIC" || value === "MANUAL";
}

type AutomaticForm = {
  booking_id: string;
  amount: number;
  method: string;
  description: string;
  customer_name: string;
  customer_email: string;
  customer_phone: string;
  customer_document: string;
};

type ManualForm = {
  booking_id: string;
  amount: number;
  method: string;
  notes: string;
};

export default function Payments() {
  const toast = useToast();
  const queryClient = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();
  const bookingsQuery = useBookings(300, 0);
  const paymentsQuery = usePayments(300, 0);
  const bookings = (bookingsQuery.data as BookingItem[] | undefined) ?? [];
  const payments = (paymentsQuery.data as PaymentItem[] | undefined) ?? [];
  const loadingPayments = paymentsQuery.isLoading;
  const [syncingId, setSyncingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<PaymentResponse | null>(null);
  const parsedSearch = useMemo(() => {
    const mode = searchParams.get("mode");
    return {
      mode: isTabKey(mode) ? mode : "AUTOMATIC",
      booking_id: searchParams.get("booking_id") ?? "",
    } satisfies PaymentsSearchParams;
  }, [searchParams]);
  const [activeTab, setActiveTab] = useState<TabKey>(parsedSearch.mode);

  const [form, setForm] = useState<AutomaticForm>({
    booking_id: "",
    amount: 0,
    method: "PIX",
    description: "Passagem",
    customer_name: "",
    customer_email: "",
    customer_phone: "",
    customer_document: "",
  });

  const [manual, setManual] = useState<ManualForm>({
    booking_id: "",
    amount: 0,
    method: "CASH",
    notes: "",
  });

  useEffect(() => {
    if (!bookingsQuery.error && !paymentsQuery.error) return;
    setError(
      (bookingsQuery.error as Error | undefined)?.message ||
        (paymentsQuery.error as Error | undefined)?.message ||
        "Erro ao carregar pagamentos"
    );
  }, [bookingsQuery.error, paymentsQuery.error]);

  const bookingMap = useMemo(() => new Map(bookings.map((item) => [item.id, item])), [bookings]);
  const selectedBookingId =
    (activeTab === "AUTOMATIC" ? form.booking_id : manual.booking_id) || parsedSearch.booking_id;

  useEffect(() => {
    if (activeTab === parsedSearch.mode) return;
    setActiveTab(parsedSearch.mode);
  }, [activeTab, parsedSearch.mode]);

  useEffect(() => {
    if (!parsedSearch.booking_id) return;
    const booking = bookingMap.get(parsedSearch.booking_id);
    const nextAmount = booking && booking.remainder_amount > 0 ? booking.remainder_amount : undefined;

    setForm((prev) => {
      if (
        prev.booking_id === parsedSearch.booking_id &&
        (nextAmount === undefined || prev.amount === nextAmount)
      ) {
        return prev;
      }
      return {
        ...prev,
        booking_id: parsedSearch.booking_id,
        amount: nextAmount ?? prev.amount,
      };
    });

    setManual((prev) => {
      if (
        prev.booking_id === parsedSearch.booking_id &&
        (nextAmount === undefined || prev.amount === nextAmount)
      ) {
        return prev;
      }
      return {
        ...prev,
        booking_id: parsedSearch.booking_id,
        amount: nextAmount ?? prev.amount,
      };
    });
  }, [bookingMap, parsedSearch.booking_id]);

  useEffect(() => {
    const currentMode = searchParams.get("mode");
    const currentBooking = searchParams.get("booking_id") ?? "";
    if (currentMode === activeTab && currentBooking === selectedBookingId) return;

    const nextParams = new URLSearchParams(searchParams);
    nextParams.set("mode", activeTab);
    if (selectedBookingId) {
      nextParams.set("booking_id", selectedBookingId);
    } else {
      nextParams.delete("booking_id");
    }
    setSearchParams(nextParams, { replace: true });
  }, [activeTab, searchParams, selectedBookingId, setSearchParams]);

  const bookingOptions = useMemo(
    () =>
      bookings.map((b) => ({
        value: b.id,
        label: `${b.passenger_name} - ${formatCurrency(b.total_amount)} - saldo ${formatCurrency(b.remainder_amount)}`,
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
        customer:
          form.customer_name || form.customer_email || form.customer_phone || form.customer_document
            ? {
                name: form.customer_name || undefined,
                email: form.customer_email || undefined,
                phone: form.customer_phone || undefined,
                document: form.customer_document || undefined,
              }
            : undefined,
      });
      setResult(res);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["payments"] }),
        queryClient.invalidateQueries({ queryKey: ["bookings"] }),
        queryClient.invalidateQueries({ queryKey: ["reports"] }),
      ]);
      toast.success("Cobranca gerada com sucesso.");
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
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["payments"] }),
        queryClient.invalidateQueries({ queryKey: ["bookings"] }),
        queryClient.invalidateQueries({ queryKey: ["reports"] }),
      ]);
      toast.success("Pagamento manual registrado.");
    } catch (err: any) {
      setError(err.message || "Erro ao registrar pagamento manual");
    }
  };

  const handleSync = async (paymentId: string) => {
    try {
      setError(null);
      setSyncingId(paymentId);
      const response = await apiPost<PaymentSyncResponse>(`/payments/${paymentId}/sync`, {});
      if (result?.payment.id === paymentId) {
        setResult((prev) =>
          prev
            ? {
                ...prev,
                payment: {
                  ...prev.payment,
                  status: response.payment.status,
                },
              }
            : prev
        );
      }
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["payments"] }),
        queryClient.invalidateQueries({ queryKey: ["bookings"] }),
        queryClient.invalidateQueries({ queryKey: ["reports"] }),
      ]);
      toast.success(response.synced ? "Status sincronizado." : "Pagamento sem sincronizacao automatica.");
    } catch (err: any) {
      setError(err.message || "Erro ao sincronizar pagamento");
    } finally {
      setSyncingId(null);
    }
  };

  const statusTone = (status: string): StatusTone => {
    if (status === "PAID" || status === "CONFIRMED") return "success";
    if (status === "PENDING") return "warning";
    if (status === "FAILED") return "danger";
    return "neutral";
  };

  const columns: DataTableColumn<PaymentItem>[] = [
    {
      label: "Reserva",
      accessor: (payment) => {
        const booking = bookingMap.get(payment.booking_id);
        return booking
          ? `${booking.passenger_name} - ${formatCurrency(booking.total_amount)}`
          : formatShortId(payment.booking_id);
      },
    },
    { label: "Metodo", accessor: (payment) => payment.method, width: "100px" },
    { label: "Valor", accessor: (payment) => formatCurrency(payment.amount), width: "120px", align: "right" },
    {
      label: "Status",
      render: (payment) => <StatusBadge tone={statusTone(payment.status)}>{payment.status}</StatusBadge>,
      width: "120px",
    },
    {
      label: "Criado em",
      accessor: (payment) => formatDateTime(payment.created_at),
      hideOnMobile: true,
    },
  ];

  const handleAutomaticChange = (next: AutomaticForm | ManualForm) => {
    const candidate = next as AutomaticForm;
    setForm((prev) => {
      if (candidate.booking_id && candidate.booking_id !== prev.booking_id) {
        const booking = bookingMap.get(candidate.booking_id);
        if (booking && booking.remainder_amount > 0) {
          return { ...candidate, amount: booking.remainder_amount };
        }
      }
      return candidate;
    });
  };

  const handleManualChange = (next: AutomaticForm | ManualForm) => {
    const candidate = next as ManualForm;
    setManual((prev) => {
      if (candidate.booking_id && candidate.booking_id !== prev.booking_id) {
        const booking = bookingMap.get(candidate.booking_id);
        if (booking && booking.remainder_amount > 0) {
          return { ...candidate, amount: booking.remainder_amount };
        }
      }
      return candidate;
    });
  };

  return (
    <section className="page">
      <PageHeader
        title="Pagamentos"
        subtitle="Cobrancas de reservas existentes, sincronizacao e historico."
        meta={<span className="badge">Checkout support</span>}
      />

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
            title="Nenhuma reserva disponivel"
            description="Crie uma reserva antes de registrar pagamentos."
          />
        ) : activeTab === "AUTOMATIC" ? (
          <PaymentForm
            mode="AUTOMATIC"
            bookingOptions={bookingOptions}
            value={form}
            onChange={handleAutomaticChange}
            onSubmit={handlePayment}
          />
        ) : (
          <PaymentForm
            mode="MANUAL"
            bookingOptions={bookingOptions}
            value={manual}
            onChange={handleManualChange}
            onSubmit={handleManual}
          />
        )}
      </div>

      {error ? <InlineAlert tone="error">{error}</InlineAlert> : null}

      {result ? <PaymentResult result={result} statusTone={statusTone} /> : null}

      <div className="section">
        <div className="section-header">
          <div className="section-title">Historico e sincronizacao</div>
        </div>
        <DataTable
          columns={columns}
          rows={payments}
          rowKey={(payment) => payment.id}
          actions={(payment) =>
            payment.status === "PENDING" && payment.provider === "ABACATEPAY" ? (
              <button
                className="button secondary sm"
                type="button"
                onClick={() => handleSync(payment.id)}
                disabled={syncingId === payment.id}
              >
                {syncingId === payment.id ? "Atualizando..." : "Atualizar status"}
              </button>
            ) : null
          }
          emptyState={
            loadingPayments ? (
              <EmptyState title="Carregando pagamentos" description="Aguarde alguns segundos." />
            ) : (
              <EmptyState
                title="Nenhum pagamento registrado"
                description="Pagamentos criados aparecerao aqui."
              />
            )
          }
        />
      </div>
    </section>
  );
}
