import { useEffect, useMemo, useState, type FormEvent } from "react";
import InlineAlert from "../../components/InlineAlert";
import PageHeader from "../../components/PageHeader";
import useToast from "../../hooks/useToast";
import { formatCurrency, formatDateTime, formatShortId } from "../../utils/format";
import { apiGet, apiPost } from "../../services/api";
import BookingList from "./BookingList";
import { BookingFormProvider, type BookingFormState, type BookingStep } from "./BookingFormContext";
import BookingStepForm from "./BookingStepForm";

type TripItem = { id: string; route_id: string; bus_id: string; departure_at: string };

type RouteItem = { id: string; origin_city: string; destination_city: string };

type SeatItem = { id: string; seat_number: number; is_active: boolean; is_taken: boolean };

type TripStop = {
  id: string;
  route_stop_id: string;
  city: string;
  stop_order: number;
  arrive_at?: string | null;
  depart_at?: string | null;
};

type QuoteResult = {
  base_amount: number;
  calc_amount: number;
  final_amount: number;
  currency: string;
  fare_mode: string;
  occupancy_ratio: number;
};

type BookingItem = {
  id: string;
  trip_id: string;
  status: string;
  passenger_name: string;
  passenger_phone: string;
  passenger_email: string;
  seat_number: number;
  total_amount: number;
  deposit_amount: number;
  remainder_amount: number;
};

const initialForm: BookingFormState = {
  trip_id: "",
  seat_id: "",
  board_stop_id: "",
  alight_stop_id: "",
  fare_mode: "AUTO",
  name: "",
  document: "",
  phone: "",
  email: "",
  total_amount: 0,
  deposit_amount: 0,
  remainder_amount: 0,
};

export default function Bookings() {
  const toast = useToast();
  const [trips, setTrips] = useState<TripItem[]>([]);
  const [routes, setRoutes] = useState<RouteItem[]>([]);
  const [seats, setSeats] = useState<SeatItem[]>([]);
  const [stops, setStops] = useState<TripStop[]>([]);
  const [quote, setQuote] = useState<QuoteResult | null>(null);
  const [quoteLoading, setQuoteLoading] = useState(false);
  const [quoteWarning, setQuoteWarning] = useState<string | null>(null);
  const [bookings, setBookings] = useState<BookingItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(0);
  const [query, setQuery] = useState("");
  const pageSize = 50;

  const [form, setForm] = useState<BookingFormState>(initialForm);
  const [activeStep, setActiveStep] = useState<BookingStep>("trip");

  const load = async () => {
    try {
      setLoading(true);
      const [t, b, r] = await Promise.all([
        apiGet<TripItem[]>("/trips?limit=200&offset=0"),
        apiGet<BookingItem[]>(`/bookings?limit=${pageSize}&offset=${page * pageSize}`),
        apiGet<RouteItem[]>("/routes?limit=200&offset=0"),
      ]);
      setTrips(t);
      setBookings(b);
      setRoutes(r);
    } catch (err: any) {
      setError(err.message || "Erro ao carregar reservas");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [page]);

  useEffect(() => {
    if (!form.trip_id) {
      setSeats([]);
      setStops([]);
      setQuote(null);
      return;
    }
    apiGet<TripStop[]>(`/trips/${form.trip_id}/stops`)
      .then(setStops)
      .catch((err) => setError(err.message || "Erro ao carregar paradas"));
  }, [form.trip_id]);

  useEffect(() => {
    if (!form.trip_id) {
      setSeats([]);
      return;
    }
    const params =
      form.board_stop_id && form.alight_stop_id
        ? `?board_stop_id=${form.board_stop_id}&alight_stop_id=${form.alight_stop_id}`
        : "";
    apiGet<SeatItem[]>(`/trips/${form.trip_id}/seats${params}`)
      .then(setSeats)
      .catch((err) => setError(err.message || "Erro ao carregar poltronas"));
  }, [form.trip_id, form.board_stop_id, form.alight_stop_id]);

  useEffect(() => {
    if (!form.trip_id || !form.board_stop_id || !form.alight_stop_id) {
      setQuote(null);
      setQuoteWarning(null);
      return;
    }
    if (form.fare_mode === "MANUAL") {
      setQuote(null);
      setQuoteWarning(null);
      return;
    }
    setQuoteWarning(null);
    setQuoteLoading(true);
    apiPost<QuoteResult>("/pricing/quote", {
      trip_id: form.trip_id,
      board_stop_id: form.board_stop_id,
      alight_stop_id: form.alight_stop_id,
      fare_mode: form.fare_mode,
    })
      .then((data) => {
        setQuote(data);
        setForm((prev) => ({ ...prev, total_amount: data.final_amount }));
      })
      .catch((err) => {
        const message = err?.message || "Erro ao calcular tarifa";
        if (message === "segment fare not found") {
          setQuote(null);
          setQuoteWarning("Tarifa do trecho não encontrada. Informe o valor manualmente.");
          setError(null);
          return;
        }
        setError(message);
      })
      .finally(() => setQuoteLoading(false));
  }, [form.trip_id, form.board_stop_id, form.alight_stop_id, form.fare_mode]);

  useEffect(() => {
    const next = Math.max(form.total_amount - form.deposit_amount, 0);
    if (next !== form.remainder_amount) {
      setForm((prev) => ({ ...prev, remainder_amount: next }));
    }
  }, [form.total_amount, form.deposit_amount]);

  const availableSeats = seats.filter((s) => s.is_active && !s.is_taken);
  const boardStop = stops.find((stop) => stop.id === form.board_stop_id);
  const alightStops = boardStop
    ? stops.filter((stop) => stop.stop_order > boardStop.stop_order)
    : stops;
  const alightStop = stops.find((stop) => stop.id === form.alight_stop_id);
  const selectedSeat = seats.find((seat) => seat.id === form.seat_id);

  const paymentError =
    form.deposit_amount > form.total_amount ? "Sinal não pode ser maior que o total." : null;

  const stepTripComplete = Boolean(
    form.trip_id && form.board_stop_id && form.alight_stop_id && form.seat_id
  );
  const stepPassengerComplete = Boolean(form.name.trim());
  const stepPaymentReady =
    stepTripComplete &&
    stepPassengerComplete &&
    !paymentError &&
    (!quoteWarning || form.fare_mode === "MANUAL");

  useEffect(() => {
    if (activeStep === "payment" && !stepTripComplete) {
      setActiveStep("trip");
      return;
    }
    if (activeStep === "payment" && !stepPassengerComplete) {
      setActiveStep("passenger");
      return;
    }
    if (activeStep === "passenger" && !stepTripComplete) {
      setActiveStep("trip");
    }
  }, [activeStep, stepTripComplete, stepPassengerComplete]);

  const filtered = useMemo(() => {
    const term = query.trim().toLowerCase();
    if (!term) return bookings;
    return bookings.filter((b) =>
      [b.passenger_name, b.passenger_email, String(b.seat_number), b.status].some((value) =>
        value?.toLowerCase().includes(term)
      )
    );
  }, [bookings, query]);

  const routeLabel = (routeId: string) => {
    const route = routes.find((r) => r.id === routeId);
    return route ? `${route.origin_city} ? ${route.destination_city}` : formatShortId(routeId);
  };

  const tripLabel = (tripId: string) => {
    const trip = trips.find((t) => t.id === tripId);
    if (!trip) return formatShortId(tripId);
    return `${routeLabel(trip.route_id)} ? ${formatDateTime(trip.departure_at)}`;
  };

  const tripSummary = stepTripComplete ? tripLabel(form.trip_id) : "Selecione a viagem e as paradas.";
  const passengerSummary = stepPassengerComplete
    ? form.name.trim()
    : stepTripComplete
      ? "Dados de contato do passageiro."
      : "Complete a viagem para continuar.";
  const paymentSummary = !stepTripComplete
    ? "Complete a viagem para continuar."
    : !stepPassengerComplete
      ? "Complete o passageiro para continuar."
      : form.total_amount > 0
        ? `Total ${formatCurrency(form.total_amount)} ? Sinal ${formatCurrency(form.deposit_amount)}`
        : "Restante calculado automaticamente.";

  const summaryItems = [
    {
      label: "Viagem",
      value: form.trip_id ? tripLabel(form.trip_id) : "Selecione a viagem",
    },
    {
      label: "Trecho",
      value:
        boardStop && alightStop
          ? `${boardStop.city} ? ${alightStop.city}`
          : "Defina as paradas",
    },
    {
      label: "Poltrona",
      value: selectedSeat ? `Poltrona ${selectedSeat.seat_number}` : "Selecione a poltrona",
    },
    {
      label: "Passageiro",
      value: form.name ? form.name : "Informe o passageiro",
    },
  ];

  const summaryTotals = [
    { label: "Total", value: formatCurrency(form.total_amount || 0) },
    { label: "Sinal", value: formatCurrency(form.deposit_amount || 0) },
    { label: "Saldo", value: formatCurrency(form.remainder_amount || 0) },
  ];

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!stepPaymentReady) {
      return;
    }
    try {
      await apiPost("/bookings", {
        trip_id: form.trip_id,
        seat_id: form.seat_id,
        board_stop_id: form.board_stop_id,
        alight_stop_id: form.alight_stop_id,
        fare_mode: form.fare_mode,
        fare_amount_final: form.fare_mode === "MANUAL" ? Number(form.total_amount) : undefined,
        passenger: {
          name: form.name,
          document: form.document,
          phone: form.phone,
          email: form.email,
        },
        total_amount: Number(form.total_amount),
        deposit_amount: Number(form.deposit_amount),
        remainder_amount: Number(form.remainder_amount),
      });
      setForm(initialForm);
      setActiveStep("trip");
      setPage(0);
      await load();
      toast.success("Reserva criada com sucesso.");
    } catch (err: any) {
      setError(err.message || "Erro ao criar reserva");
    }
  };

  return (
    <section className="page">
      <PageHeader
        title="Reservas"
        subtitle="Passageiros, poltronas e status de pagamento."
        meta={<span className="badge">MVP</span>}
      />

      <BookingFormProvider value={{ form, setForm, activeStep, setActiveStep }}>
        <div className="section" id="booking-form">
          <div className="section-header">
            <div className="section-title">Nova reserva</div>
          </div>
          <BookingStepForm
            trips={trips}
            stops={stops}
            alightStops={alightStops}
            availableSeats={availableSeats}
            tripLabel={tripLabel}
            tripSummary={tripSummary}
            passengerSummary={passengerSummary}
            paymentSummary={paymentSummary}
            stepTripComplete={stepTripComplete}
            stepPassengerComplete={stepPassengerComplete}
            stepPaymentReady={stepPaymentReady}
            paymentError={paymentError}
            quoteWarning={quoteWarning}
            quoteLoading={quoteLoading}
            quote={quote}
            onSubmit={handleSubmit}
            summaryItems={summaryItems}
            summaryTotals={summaryTotals}
          />
        </div>
      </BookingFormProvider>

      {error ? <InlineAlert tone="error">{error}</InlineAlert> : null}

      <BookingList
        bookings={filtered}
        loading={loading}
        query={query}
        onQueryChange={setQuery}
        page={page}
        pageSize={pageSize}
        onPageChange={setPage}
        tripLabel={tripLabel}
      />
    </section>
  );
}
