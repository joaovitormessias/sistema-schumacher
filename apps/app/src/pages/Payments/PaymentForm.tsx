import type { FormEvent } from "react";
import FormField from "../../components/FormField";

type BookingOption = { value: string; label: string };

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

type PaymentFormProps = {
  mode: "AUTOMATIC" | "MANUAL";
  bookingOptions: BookingOption[];
  value: AutomaticForm | ManualForm;
  onChange: (next: AutomaticForm | ManualForm) => void;
  onSubmit: (event: FormEvent) => void;
};

export default function PaymentForm({
  mode,
  bookingOptions,
  value,
  onChange,
  onSubmit,
}: PaymentFormProps) {
  const isAutomatic = mode === "AUTOMATIC";

  return (
    <form className="form-grid" onSubmit={onSubmit}>
      <FormField label="Reserva" required>
        <select
          className="input"
          value={value.booking_id}
          onChange={(e) => onChange({ ...value, booking_id: e.target.value } as any)}
          required
        >
          <option value="">Selecione a reserva</option>
          {bookingOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </FormField>
      <FormField label="Valor">
        <input
          className="input"
          type="number"
          placeholder="0,00"
          value={value.amount}
          onChange={(e) => onChange({ ...value, amount: Number(e.target.value) } as any)}
        />
      </FormField>
      <FormField label="Método">
        <select
          className="input"
          value={value.method}
          onChange={(e) => onChange({ ...value, method: e.target.value } as any)}
        >
          {isAutomatic ? (
            <>
              <option value="PIX">PIX</option>
              <option value="CARD">Cartão</option>
            </>
          ) : (
            <>
              <option value="CASH">Dinheiro</option>
              <option value="TRANSFER">Transferência</option>
              <option value="OTHER">Outro</option>
            </>
          )}
        </select>
      </FormField>
      {isAutomatic ? (
        <FormField label="Descrição">
          <input
            className="input"
            placeholder="Ex: Passagem"
            value={(value as AutomaticForm).description}
            onChange={(e) =>
              onChange({ ...(value as AutomaticForm), description: e.target.value } as any)
            }
          />
        </FormField>
      ) : (
        <FormField label="Observações" hint="Opcional">
          <input
            className="input"
            placeholder="Ex: Recebido no balcão"
            value={(value as ManualForm).notes}
            onChange={(e) =>
              onChange({ ...(value as ManualForm), notes: e.target.value } as any)
            }
          />
        </FormField>
      )}
      <div className="form-actions full-width-mobile full-span">
        <button className="button" type="submit">
          {isAutomatic ? "Gerar cobrança" : "Registrar pagamento"}
        </button>
      </div>
    </form>
  );
}
