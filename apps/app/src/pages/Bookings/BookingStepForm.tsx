import Stepper from "../../components/Stepper";
import ProgressBar from "../../components/feedback/ProgressBar";
import type { PassengerSuggestion, PreferredSeatOption } from "../../types/checkout";
import { useBookingForm, type BookingStep } from "./BookingFormContext";
import TripPassengerStep from "./steps/TripPassengerStep";
import BookingPaymentStep from "./steps/BookingPaymentStep";

type TripItem = { id: string };
type SeatItem = { id: string; seat_number: number; is_active: boolean; is_taken: boolean };
type TripStop = { id: string; city: string; stop_order: number };
type SummaryItem = {
  label: string;
  value: string;
};

type BookingStepFormProps = {
  trips: TripItem[];
  stops: TripStop[];
  alightStops: TripStop[];
  availableSeats: SeatItem[];
  passengerSuggestions: PassengerSuggestion[];
  preferredSeatOptions: PreferredSeatOption[];
  tripLabel: (tripId: string) => string;
  tripPassengerSummary: string;
  paymentSummary: string;
  stepTripPassengerComplete: boolean;
  canSubmitCheckout: boolean;
  minInitialAmount: number;
  paymentError: string | null;
  checkoutLoading: boolean;
  checkoutResult: {
    booking: { booking: { status: string } };
    payment: { id: string; status: string; amount: number; provider_ref?: string };
    provider_raw?: any;
    checkout_url?: string | null;
    pix_code?: string | null;
  } | null;
  syncLoading: boolean;
  onSubmitCheckout: () => void;
  onSync: () => void;
  onApplyPassengerSuggestion: (suggestion: PassengerSuggestion) => void;
  onSelectPreferredSeat: (option: PreferredSeatOption) => void;
  statusTone: (status: string) => "neutral" | "info" | "success" | "warning" | "danger";
  summaryItems: SummaryItem[];
  summaryTotals: SummaryItem[];
};

export default function BookingStepForm({
  trips,
  stops,
  alightStops,
  availableSeats,
  passengerSuggestions,
  preferredSeatOptions,
  tripLabel,
  tripPassengerSummary,
  paymentSummary,
  stepTripPassengerComplete,
  canSubmitCheckout,
  minInitialAmount,
  paymentError,
  checkoutLoading,
  checkoutResult,
  syncLoading,
  onSubmitCheckout,
  onSync,
  onApplyPassengerSuggestion,
  onSelectPreferredSeat,
  statusTone,
  summaryItems,
  summaryTotals,
}: BookingStepFormProps) {
  const { activeStep, setActiveStep } = useBookingForm();
  const progressValue = checkoutResult ? 100 : stepTripPassengerComplete ? 50 : 0;

  const statusFor = (step: BookingStep) => {
    if (activeStep === step) return "current";
    if (step === "trip_passenger") return stepTripPassengerComplete ? "complete" : "upcoming";
    return checkoutResult ? "complete" : "upcoming";
  };

  const steps = [
    {
      id: "trip_passenger",
      title: "1. Viagem e passageiro",
      summary: tripPassengerSummary,
      status: statusFor("trip_passenger"),
      disabled: checkoutLoading,
    },
    {
      id: "payment",
      title: "2. Pagamento inicial",
      summary: paymentSummary,
      status: statusFor("payment"),
      disabled: !stepTripPassengerComplete || checkoutLoading,
    },
  ];

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
      }}
      onKeyDown={(event) => {
        const target = event.target as HTMLElement;
        const isTextarea = target.tagName.toLowerCase() === "textarea";
        const isButton = target.tagName.toLowerCase() === "button";

        if (event.key === "Escape" && activeStep === "payment" && !checkoutLoading) {
          event.preventDefault();
          setActiveStep("trip_passenger");
          return;
        }

        if (event.key !== "Enter" || event.shiftKey || isTextarea || isButton) {
          return;
        }

        if (activeStep === "trip_passenger" && stepTripPassengerComplete && !checkoutLoading) {
          event.preventDefault();
          setActiveStep("payment");
          return;
        }

        if (activeStep === "payment" && canSubmitCheckout && !checkoutLoading) {
          event.preventDefault();
          onSubmitCheckout();
        }
      }}
    >
      <div className="booking-progress">
        <div className="booking-progress-meta">
          <span className="text-caption">Progresso da reserva</span>
          <strong>{progressValue}%</strong>
        </div>
        <ProgressBar value={progressValue} />
      </div>
      <Stepper steps={steps} onStepChange={(id) => setActiveStep(id as BookingStep)} />

      <div className="booking-layout">
        <div>
          {activeStep === "trip_passenger" ? (
            <div id="stepper-panel-trip_passenger" role="tabpanel" aria-label="Etapa 1">
              <TripPassengerStep
                trips={trips}
                stops={stops}
                alightStops={alightStops}
                availableSeats={availableSeats}
                passengerSuggestions={passengerSuggestions}
                preferredSeatOptions={preferredSeatOptions}
                tripLabel={tripLabel}
                stepTripPassengerComplete={stepTripPassengerComplete}
                onApplyPassengerSuggestion={onApplyPassengerSuggestion}
                onSelectPreferredSeat={onSelectPreferredSeat}
                onNext={() => setActiveStep("payment")}
              />
            </div>
          ) : null}
          {activeStep === "payment" ? (
            <div id="stepper-panel-payment" role="tabpanel" aria-label="Etapa 2">
              <BookingPaymentStep
                minInitialAmount={minInitialAmount}
                paymentError={paymentError}
                canSubmitCheckout={canSubmitCheckout}
                checkoutLoading={checkoutLoading}
                checkoutResult={checkoutResult}
                syncLoading={syncLoading}
                onBack={() => setActiveStep("trip_passenger")}
                onSubmitCheckout={onSubmitCheckout}
                onSync={onSync}
                statusTone={statusTone}
              />
            </div>
          ) : null}
        </div>

        <aside className="summary-panel">
          <div className="section-title">Resumo da reserva</div>
          {summaryItems.map((item) => (
            <div className="summary-item" key={item.label}>
              <span className="summary-label">{item.label}</span>
              <span className="summary-value">{item.value}</span>
            </div>
          ))}
          <div className="summary-highlight">
            {summaryTotals.map((item) => (
              <div className="summary-item" key={item.label}>
                <span className="summary-label">{item.label}</span>
                <span className="summary-value">{item.value}</span>
              </div>
            ))}
          </div>
        </aside>
      </div>
    </form>
  );
}
