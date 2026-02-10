import { useMemo, useState } from "react";
import FormField from "../../../components/FormField";
import Autocomplete from "../../../components/form/Autocomplete";
import SeatMap, { type Seat } from "../../../components/domain/SeatMap";
import type { PassengerSuggestion, PreferredSeatOption } from "../../../types/checkout";
import { useBookingForm } from "../BookingFormContext";
import { formatCurrency } from "../../../utils/format";

type TripItem = { id: string };
type SeatItem = { id: string; seat_number: number; is_active: boolean; is_taken: boolean };
type TripStop = { id: string; city: string; stop_order: number };

type TripPassengerStepProps = {
  trips: TripItem[];
  stops: TripStop[];
  alightStops: TripStop[];
  availableSeats: SeatItem[];
  allSeats?: SeatItem[];
  passengerSuggestions: PassengerSuggestion[];
  preferredSeatOptions: PreferredSeatOption[];
  tripLabel: (tripId: string) => string;
  stepTripPassengerComplete: boolean;
  calculatedFare?: number;
  onApplyPassengerSuggestion: (suggestion: PassengerSuggestion) => void;
  onSelectPreferredSeat: (option: PreferredSeatOption) => void;
  onNext: () => void;
};

export default function TripPassengerStep({
  trips,
  stops,
  alightStops,
  availableSeats,
  allSeats,
  passengerSuggestions,
  preferredSeatOptions,
  tripLabel,
  stepTripPassengerComplete,
  calculatedFare,
  onApplyPassengerSuggestion,
  onSelectPreferredSeat,
  onNext,
}: TripPassengerStepProps) {
  const { form, setForm } = useBookingForm();
  const [selectedSuggestionId, setSelectedSuggestionId] = useState("");
  const [showSeatMap, setShowSeatMap] = useState(false);

  const seatMapSeats: Seat[] = useMemo(() => {
    const seatsToUse = allSeats && allSeats.length > 0 ? allSeats : availableSeats;
    return seatsToUse.map((seat) => ({
      id: seat.id,
      number: seat.seat_number,
      status: seat.is_taken ? "occupied" : form.seat_id === seat.id ? "selected" : "available",
      isPreferred: preferredSeatOptions.some((p) => p.seat_id === seat.id),
    }));
  }, [allSeats, availableSeats, form.seat_id, preferredSeatOptions]);

  const preferredSeatIds = useMemo(
    () => preferredSeatOptions.map((p) => p.seat_id),
    [preferredSeatOptions]
  );

  const handleSeatClick = (seat: Seat) => {
    setForm({ ...form, seat_id: seat.id });
  };

  const selectedSeatNumber = useMemo(() => {
    const seat = seatMapSeats.find((s) => s.id === form.seat_id);
    return seat?.number;
  }, [seatMapSeats, form.seat_id]);

  const boardStopName = useMemo(() => {
    const stop = stops.find((s) => s.id === form.board_stop_id);
    return stop?.city;
  }, [stops, form.board_stop_id]);

  const alightStopName = useMemo(() => {
    const stop = alightStops.find((s) => s.id === form.alight_stop_id);
    return stop?.city;
  }, [alightStops, form.alight_stop_id]);

  return (
    <>
      <div className="form-step-grid">
        <FormField label="Viagem" required>
          <select
            className="input input-delightful"
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
            className="input input-delightful"
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
            className="input input-delightful"
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
          <div style={{ display: "flex", gap: "8px" }}>
            <select
              className="input input-delightful"
              value={form.seat_id}
              onChange={(e) => setForm({ ...form, seat_id: e.target.value })}
              required
              disabled={!form.board_stop_id || !form.alight_stop_id}
              style={{ flex: 1 }}
            >
              <option value="">Selecione a poltrona</option>
              {availableSeats.map((seat) => (
                <option key={seat.id} value={seat.id}>
                  {seat.seat_number}
                </option>
              ))}
            </select>
            <button
              type="button"
              className="button secondary sm"
              onClick={() => setShowSeatMap(!showSeatMap)}
              disabled={!form.board_stop_id || !form.alight_stop_id}
            >
              {showSeatMap ? "Ocultar mapa" : "Ver mapa"}
            </button>
          </div>
        </FormField>

        {showSeatMap && seatMapSeats.length > 0 ? (
          <div className="full-span">
            <SeatMap
              seats={seatMapSeats}
              selectedSeatId={form.seat_id}
              preferredSeatIds={preferredSeatIds}
              onSeatClick={handleSeatClick}
              seatsPerRow={4}
            />
          </div>
        ) : null}

        {preferredSeatOptions.length > 0 && !showSeatMap ? (
          <div className="form-field full-span">
            <label className="form-label">Poltronas sugeridas</label>
            <span className="form-hint">Com base no historico deste passageiro.</span>
            <div className="chip-list">
              {preferredSeatOptions.map((option) => (
                <button
                  key={option.seat_id}
                  type="button"
                  className={`chip-button${form.seat_id === option.seat_id ? " active" : ""}`}
                  onClick={() => onSelectPreferredSeat(option)}
                >
                  Poltrona {option.seat_number}
                </button>
              ))}
            </div>
          </div>
        ) : null}

        <FormField label="Passageiro frequente" hint="Opcional">
          <select
            className="input input-delightful"
            value={selectedSuggestionId}
            onChange={(event) => {
              const suggestionId = event.target.value;
              setSelectedSuggestionId(suggestionId);
              if (!suggestionId) return;
              const suggestion = passengerSuggestions.find((item) => item.id === suggestionId);
              if (suggestion) {
                onApplyPassengerSuggestion(suggestion);
              }
            }}
            disabled={passengerSuggestions.length === 0}
          >
            <option value="">
              {passengerSuggestions.length === 0
                ? "Sem historico para auto-complete"
                : "Selecionar passageiro do historico"}
            </option>
            {passengerSuggestions.map((suggestion) => (
              <option key={suggestion.id} value={suggestion.id}>
                {suggestion.name}
                {suggestion.phone ? ` - ${suggestion.phone}` : ""}
                {suggestion.email ? ` - ${suggestion.email}` : ""}
              </option>
            ))}
          </select>
        </FormField>

        <FormField label="Nome do passageiro" required>
          <Autocomplete
            value={form.name}
            ariaLabel="Buscar passageiro"
            placeholder={
              passengerSuggestions.length > 0
                ? "Digite para buscar no historico"
                : "Ex: Maria Oliveira"
            }
            options={passengerSuggestions.map((suggestion) => ({
              id: suggestion.id,
              label: suggestion.name,
              meta: [suggestion.phone, suggestion.email].filter(Boolean).join(" - "),
            }))}
            onInputChange={(value) => setForm({ ...form, name: value })}
            onOptionSelect={(option) => {
              const suggestion = passengerSuggestions.find((item) => item.id === option.id);
              if (suggestion) {
                onApplyPassengerSuggestion(suggestion);
              }
            }}
          />
        </FormField>

        <FormField label="Documento" hint="Opcional">
          <input
            className="input input-delightful"
            placeholder="CPF ou documento"
            value={form.document}
            onChange={(e) => setForm({ ...form, document: e.target.value })}
          />
        </FormField>

        <FormField label="Telefone" hint="Opcional">
          <input
            className="input input-delightful"
            placeholder="(00) 00000-0000"
            value={form.phone}
            onChange={(e) => setForm({ ...form, phone: e.target.value })}
          />
        </FormField>

        <FormField label="E-mail" hint="Opcional">
          <input
            className="input input-delightful"
            placeholder="email@exemplo.com"
            value={form.email}
            onChange={(e) => setForm({ ...form, email: e.target.value })}
          />
        </FormField>
      </div>

      {form.name || form.seat_id || form.board_stop_id ? (
        <div className="live-preview" style={{ marginTop: "16px" }}>
          <div className="live-preview-header">Resumo da reserva</div>
          {form.name ? (
            <div className="live-preview-item">
              <span className="live-preview-label">Passageiro</span>
              <span className="live-preview-value">{form.name}</span>
            </div>
          ) : null}
          {form.trip_id ? (
            <div className="live-preview-item">
              <span className="live-preview-label">Viagem</span>
              <span className="live-preview-value">{tripLabel(form.trip_id)}</span>
            </div>
          ) : null}
          {boardStopName && alightStopName ? (
            <div className="live-preview-item">
              <span className="live-preview-label">Trecho</span>
              <span className="live-preview-value">
                {boardStopName} {"->"} {alightStopName}
              </span>
            </div>
          ) : null}
          {selectedSeatNumber ? (
            <div className="live-preview-item">
              <span className="live-preview-label">Poltrona</span>
              <span className="live-preview-value">#{selectedSeatNumber}</span>
            </div>
          ) : null}
          {calculatedFare && calculatedFare > 0 ? (
            <div className="live-preview-item">
              <span className="live-preview-label">Valor estimado</span>
              <span className="live-preview-value highlight">{formatCurrency(calculatedFare)}</span>
            </div>
          ) : null}
        </div>
      ) : null}

      <div className="form-step-actions align-end">
        <button
          className="button button-delightful"
          type="button"
          onClick={onNext}
          disabled={!stepTripPassengerComplete}
        >
          Continuar para pagamento
        </button>
      </div>
    </>
  );
}
