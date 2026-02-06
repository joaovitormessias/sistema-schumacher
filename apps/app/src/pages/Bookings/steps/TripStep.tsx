import FormField from "../../../components/FormField";
import { useBookingForm } from "../BookingFormContext";

type TripItem = { id: string };
type SeatItem = { id: string; seat_number: number; is_active: boolean; is_taken: boolean };
type TripStop = { id: string; city: string; stop_order: number };

type TripStepProps = {
  trips: TripItem[];
  stops: TripStop[];
  alightStops: TripStop[];
  availableSeats: SeatItem[];
  tripLabel: (tripId: string) => string;
  stepTripComplete: boolean;
  onNext: () => void;
};

export default function TripStep({
  trips,
  stops,
  alightStops,
  availableSeats,
  tripLabel,
  stepTripComplete,
  onNext,
}: TripStepProps) {
  const { form, setForm } = useBookingForm();

  return (
    <>
      <div className="form-step-grid">
        <FormField label="Viagem" required>
          <select
            className="input"
            value={form.trip_id}
            onChange={(e) =>
              setForm({
                ...form,
                trip_id: e.target.value,
                seat_id: "",
                board_stop_id: "",
                alight_stop_id: "",
              })
            }
            required
          >
            <option value="">Selecione a viagem</option>
            {trips.map((trip) => (
              <option key={trip.id} value={trip.id}>
                {tripLabel(trip.id)}
              </option>
            ))}
          </select>
        </FormField>

        <FormField label="Embarque" required>
          <select
            className="input"
            value={form.board_stop_id}
            onChange={(e) =>
              setForm({
                ...form,
                board_stop_id: e.target.value,
                alight_stop_id: "",
                seat_id: "",
              })
            }
            required
            disabled={!form.trip_id}
          >
            <option value="">Selecione o embarque</option>
            {stops.map((stop) => (
              <option key={stop.id} value={stop.id}>
                {stop.stop_order} - {stop.city}
              </option>
            ))}
          </select>
        </FormField>

        <FormField label="Desembarque" required>
          <select
            className="input"
            value={form.alight_stop_id}
            onChange={(e) => setForm({ ...form, alight_stop_id: e.target.value, seat_id: "" })}
            required
            disabled={!form.board_stop_id}
          >
            <option value="">Selecione o desembarque</option>
            {alightStops.map((stop) => (
              <option key={stop.id} value={stop.id}>
                {stop.stop_order} - {stop.city}
              </option>
            ))}
          </select>
        </FormField>

        <FormField label="Poltrona" required>
          <select
            className="input"
            value={form.seat_id}
            onChange={(e) => setForm({ ...form, seat_id: e.target.value })}
            required
            disabled={!form.board_stop_id || !form.alight_stop_id}
          >
            <option value="">Selecione a poltrona</option>
            {availableSeats.map((seat) => (
              <option key={seat.id} value={seat.id}>
                {seat.seat_number}
              </option>
            ))}
          </select>
        </FormField>
      </div>
      <div className="form-step-actions align-end">
        <button className="button" type="button" onClick={onNext} disabled={!stepTripComplete}>
          Continuar
        </button>
      </div>
    </>
  );
}
