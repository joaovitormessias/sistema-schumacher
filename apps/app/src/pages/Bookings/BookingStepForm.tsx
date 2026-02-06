import type { FormEvent } from "react";
import Stepper from "../../components/Stepper";
import { useBookingForm, type BookingStep } from "./BookingFormContext";
import TripStep from "./steps/TripStep";
import PassengerStep from "./steps/PassengerStep";
import PaymentStep from "./steps/PaymentStep";

type TripItem = { id: string };
type SeatItem = { id: string; seat_number: number; is_active: boolean; is_taken: boolean };
type TripStop = { id: string; city: string; stop_order: number };
type QuoteResult = {
  base_amount: number;
  calc_amount: number;
  final_amount: number;
  currency: string;
  fare_mode: string;
  occupancy_ratio: number;
};

type SummaryItem = {
  label: string;
  value: string;
};

type BookingStepFormProps = {
  trips: TripItem[];
  stops: TripStop[];
  alightStops: TripStop[];
  availableSeats: SeatItem[];
  tripLabel: (tripId: string) => string;
  tripSummary: string;
  passengerSummary: string;
  paymentSummary: string;
  stepTripComplete: boolean;
  stepPassengerComplete: boolean;
  stepPaymentReady: boolean;
  paymentError: string | null;
  quoteWarning: string | null;
  quoteLoading: boolean;
  quote: QuoteResult | null;
  onSubmit: (event: FormEvent) => void;
  summaryItems: SummaryItem[];
  summaryTotals: SummaryItem[];
};

export default function BookingStepForm({
  trips,
  stops,
  alightStops,
  availableSeats,
  tripLabel,
  tripSummary,
  passengerSummary,
  paymentSummary,
  stepTripComplete,
  stepPassengerComplete,
  stepPaymentReady,
  paymentError,
  quoteWarning,
  quoteLoading,
  quote,
  onSubmit,
  summaryItems,
  summaryTotals,
}: BookingStepFormProps) {
  const { activeStep, setActiveStep } = useBookingForm();

  const statusFor = (step: BookingStep) => {
    if (activeStep === step) return "current";
    if (step === "trip") return stepTripComplete ? "complete" : "upcoming";
    if (step === "passenger") return stepPassengerComplete ? "complete" : "upcoming";
    return stepPaymentReady ? "complete" : "upcoming";
  };

  const steps = [
    {
      id: "trip",
      title: "1. Escolha a viagem",
      summary: tripSummary,
      status: statusFor("trip"),
      disabled: false,
    },
    {
      id: "passenger",
      title: "2. Dados do passageiro",
      summary: passengerSummary,
      status: statusFor("passenger"),
      disabled: !stepTripComplete,
    },
    {
      id: "payment",
      title: "3. Pagamento",
      summary: paymentSummary,
      status: statusFor("payment"),
      disabled: !stepTripComplete || !stepPassengerComplete,
    },
  ];

  return (
    <form onSubmit={onSubmit}>
      <Stepper steps={steps} onStepChange={(id) => setActiveStep(id as BookingStep)} />

      <div className="booking-layout">
        <div>
          {activeStep === "trip" ? (
            <TripStep
              trips={trips}
              stops={stops}
              alightStops={alightStops}
              availableSeats={availableSeats}
              tripLabel={tripLabel}
              stepTripComplete={stepTripComplete}
              onNext={() => setActiveStep("passenger")}
            />
          ) : null}
          {activeStep === "passenger" ? (
            <PassengerStep
              stepPassengerComplete={stepPassengerComplete}
              onBack={() => setActiveStep("trip")}
              onNext={() => setActiveStep("payment")}
            />
          ) : null}
          {activeStep === "payment" ? (
            <PaymentStep
              paymentError={paymentError}
              quoteWarning={quoteWarning}
              quoteLoading={quoteLoading}
              quote={quote}
              stepPaymentReady={stepPaymentReady}
              onBack={() => setActiveStep("passenger")}
            />
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
