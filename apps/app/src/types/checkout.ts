export type PassengerSuggestion = {
  id: string;
  name: string;
  document: string;
  phone: string;
  email: string;
  lastBookingAt: string | null;
  preferredSeats: number[];
};

export type PreferredSeatOption = {
  seat_id: string;
  seat_number: number;
  score: number;
  source: "history";
};

export type PaymentsSearchParams = {
  booking_id: string;
  mode: "AUTOMATIC" | "MANUAL";
};
