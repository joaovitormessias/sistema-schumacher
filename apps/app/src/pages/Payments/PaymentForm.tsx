import type { FormEvent } from "react";
import FormField from "../../components/FormField";

type BookingOption = { value: string; label: string };

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
      <FormField label="Valor" required>
        <input
          className="input"
          type="number"
          placeholder="0,00"
          value={value.amount}
          onChange={(e) => onChange({ ...value, amount: Number(e.target.value) } as any)}
          min={0}
          step={0.01}
          required
        />
      </FormField>
      <FormField label="Metodo" required>
        <select
          className="input"
          value={value.method}
          onChange={(e) => onChange({ ...value, method: e.target.value } as any)}
          required
        >
          {isAutomatic ? (
            <>
              <option value="PIX">PIX</option>
              <option value="CARD">Cartao</option>
            </>
          ) : (
            <>
              <option value="CASH">Dinheiro</option>
              <option value="TRANSFER">Transferencia</option>
              <option value="OTHER">Outro</option>
            </>
          )}
        </select>
      </FormField>
      {isAutomatic ? (
        <>
          <FormField label="Descricao">
            <input
              className="input"
              placeholder="Ex: Passagem"
              value={(value as AutomaticForm).description}
              onChange={(e) =>
                onChange({ ...(value as AutomaticForm), description: e.target.value } as any)
              }
            />
          </FormField>
          <FormField label="Cliente" hint="Opcional">
            <input
              className="input"
              placeholder="Nome completo"
              value={(value as AutomaticForm).customer_name}
              onChange={(e) =>
                onChange({ ...(value as AutomaticForm), customer_name: e.target.value } as any)
              }
            />
          </FormField>
          <FormField label="Email" hint="Opcional">
            <input
              className="input"
              type="email"
              placeholder="cliente@exemplo.com"
              value={(value as AutomaticForm).customer_email}
              onChange={(e) =>
                onChange({ ...(value as AutomaticForm), customer_email: e.target.value } as any)
              }
            />
          </FormField>
          <FormField label="Telefone" hint="Opcional">
            <input
              className="input"
              placeholder="(49) 99999-9999"
              value={(value as AutomaticForm).customer_phone}
              onChange={(e) =>
                onChange({ ...(value as AutomaticForm), customer_phone: e.target.value } as any)
              }
            />
          </FormField>
          <FormField label="CPF/CNPJ" hint="Opcional">
            <input
              className="input"
              placeholder="Somente numeros"
              value={(value as AutomaticForm).customer_document}
              onChange={(e) =>
                onChange({ ...(value as AutomaticForm), customer_document: e.target.value } as any)
              }
            />
          </FormField>
        </>
      ) : (
        <FormField label="Observacoes" hint="Opcional">
          <input
            className="input"
            placeholder="Ex: Recebido no balcao"
            value={(value as ManualForm).notes}
            onChange={(e) =>
              onChange({ ...(value as ManualForm), notes: e.target.value } as any)
            }
          />
        </FormField>
      )}
      <div className="form-actions full-width-mobile full-span">
        <button className="button" type="submit">
          {isAutomatic ? "Gerar cobranca" : "Registrar pagamento"}
        </button>
      </div>
    </form>
  );
}

