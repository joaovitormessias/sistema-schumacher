import FormField from "../../../components/FormField";
import InlineAlert from "../../../components/InlineAlert";
import LoadingState from "../../../components/LoadingState";
import { formatCurrency } from "../../../utils/format";
import { useBookingForm } from "../BookingFormContext";

type QuoteResult = {
  base_amount: number;
  calc_amount: number;
  final_amount: number;
  currency: string;
  fare_mode: string;
  occupancy_ratio: number;
};

type PaymentStepProps = {
  paymentError: string | null;
  quoteWarning: string | null;
  quoteLoading: boolean;
  quote: QuoteResult | null;
  stepPaymentReady: boolean;
  onBack: () => void;
};

export default function PaymentStep({
  paymentError,
  quoteWarning,
  quoteLoading,
  quote,
  stepPaymentReady,
  onBack,
}: PaymentStepProps) {
  const { form, setForm } = useBookingForm();

  return (
    <>
      <div className="form-step-grid">
        <FormField label="Método de cálculo" hint="Automático usa regras ativas">
          <select
            className="input"
            value={form.fare_mode}
            onChange={(e) => setForm({ ...form, fare_mode: e.target.value })}
          >
            <option value="AUTO">Automático</option>
            <option value="FIXED">Fixo</option>
            <option value="MANUAL">Manual</option>
          </select>
        </FormField>
        <FormField label="Total" required>
          <input
            className={`input${paymentError ? " error" : ""}`}
            type="number"
            placeholder="0,00"
            value={form.total_amount}
            min={0}
            step="0.01"
            onChange={(e) => setForm({ ...form, total_amount: Number(e.target.value) })}
            disabled={form.fare_mode !== "MANUAL"}
          />
        </FormField>
        <FormField label="Sinal" hint="Quanto j? foi pago">
          <input
            className={`input${paymentError ? " error" : ""}`}
            type="number"
            placeholder="0,00"
            value={form.deposit_amount}
            min={0}
            step="0.01"
            max={form.total_amount > 0 ? form.total_amount : undefined}
            onChange={(e) => setForm({ ...form, deposit_amount: Number(e.target.value) })}
          />
        </FormField>
        <FormField label="Restante" hint="Atualizado automaticamente">
          <input className="input" type="number" placeholder="0,00" value={form.remainder_amount} readOnly />
        </FormField>
      </div>

      {paymentError ? <div className="form-error">{paymentError}</div> : null}
      {quoteWarning ? (
        <InlineAlert tone="warning" title="Tarifa não encontrada">
          {quoteWarning}
          <div className="alert-actions">
            <button
              className="button secondary sm"
              type="button"
              onClick={() => setForm({ ...form, fare_mode: "MANUAL" })}
            >
              Definir manualmente
            </button>
          </div>
        </InlineAlert>
      ) : null}
      {quoteLoading ? <LoadingState label="Calculando tarifa..." /> : null}
      {quote ? (
        <div className="form-summary">
          <div>
            <strong>Base:</strong> {formatCurrency(quote.base_amount)}
          </div>
          <div>
            <strong>Calculado:</strong> {formatCurrency(quote.calc_amount)}
          </div>
          <div>
            <strong>Final:</strong> {formatCurrency(quote.final_amount)}
          </div>
        </div>
      ) : null}

      <div className="form-step-actions">
        <button className="button secondary" type="button" onClick={onBack}>
          Voltar
        </button>
        <button className="button" type="submit" disabled={!stepPaymentReady}>
          Criar reserva
        </button>
      </div>
    </>
  );
}
