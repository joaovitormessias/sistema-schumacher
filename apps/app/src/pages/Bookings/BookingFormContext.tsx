import { createContext, useContext, type Dispatch, type ReactNode, type SetStateAction } from "react";

export type BookingFormState = {
  trip_id: string;
  seat_id: string;
  board_stop_id: string;
  alight_stop_id: string;
  fare_mode: string;
  name: string;
  document: string;
  phone: string;
  email: string;
  total_amount: number;
  deposit_amount: number;
  remainder_amount: number;
};

export type BookingStep = "trip" | "passenger" | "payment";

type BookingFormContextValue = {
  form: BookingFormState;
  setForm: Dispatch<SetStateAction<BookingFormState>>;
  activeStep: BookingStep;
  setActiveStep: Dispatch<SetStateAction<BookingStep>>;
};

const BookingFormContext = createContext<BookingFormContextValue | null>(null);

export function BookingFormProvider({
  value,
  children,
}: {
  value: BookingFormContextValue;
  children: ReactNode;
}) {
  return <BookingFormContext.Provider value={value}>{children}</BookingFormContext.Provider>;
}

export function useBookingForm() {
  const ctx = useContext(BookingFormContext);
  if (!ctx) {
    throw new Error("useBookingForm must be used within BookingFormProvider");
  }
  return ctx;
}
