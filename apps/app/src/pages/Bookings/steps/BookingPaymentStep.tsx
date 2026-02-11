import { useState, useEffect } from "react";
import { CheckCircle } from "lucide-react";
import InlineAlert from "../../../components/InlineAlert";
import FormField from "../../../components/FormField";
import { formatCurrency } from "../../../utils/format";
import PaymentResult from "../../Payments/PaymentResult";
import { useBookingForm } from "../BookingFormContext";
import ConfettiAnimation from "../../../components/feedback/ConfettiAnimation";

type CheckoutResult = {
  booking: { booking: { status: string } };
  payment: { id: string; status: string; amount: number; provider_ref?: string };
  provider_raw?: any;
  checkout_url?: string | null;
  pix_code?: string | null;
};

type BookingPaymentStepProps = {
  minInitialAmount: number;
  paymentError: string | null;
  canSubmitCheckout: boolean;
  checkoutLoading: boolean;
  checkoutResult: CheckoutResult | null;
  syncLoading: boolean;
  onBack: () => void;
  onSubmitCheckout: () => void;
  onSync: () => void;
  statusTone: (status: string) => "neutral" | "info" | "success" | "warning" | "danger";
};

export default function BookingPaymentStep({
  minInitialAmount,
  paymentError,
  canSubmitCheckout,
  checkoutLoading,
  checkoutResult,
  syncLoading,
  onBack,
  onSubmitCheckout,
  onSync,
  statusTone,
}: BookingPaymentStepProps) {
  const { form, setForm } = useBookingForm();
  const isAutomatic = form.payment_method === "PIX" || form.payment_method === "CARD";
  const [showConfetti, setShowConfetti] = useState(false);

  // Trigger confetti when payment is successful
  const isPaid = checkoutResult?.payment?.status?.toUpperCase() === "PAID";
  useEffect(() => {
    if (isPaid) {
      setShowConfetti(true);
    }
  }, [isPaid]);

  return (
    <>
      {showConfetti && <ConfettiAnimation onComplete={() => setShowConfetti(false)} />}

      {isPaid && (
        <div className="success-animation">
          <CheckCircle className="success-icon" />
          <div className="success-title">Reserva confirmada!</div>
          <div className="success-description">O pagamento foi processado com sucesso.</div>
        </div>
      )}

      <div className="form-step-grid">
        <FormField label="Método de cálculo" hint="Automático usa regras ativas">
          <select
            className="input input-delightful"
            value={form.fare_mode}
            onChange={(e) => setForm((prev) => ({ ...prev, fare_mode: e.target.value }))}
          >
            <option value="AUTO">Automático</option>
            <option value="FIXED">Fixo</option>
            <option value="MANUAL">Manual</option>
          </select>
        </FormField>
        <FormField label="Total da reserva" required>
          <input
            className="input input-delightful"
            type="number"
            value={form.total_amount}
            min={0}
            step="0.01"
            onChange={(e) =>
              setForm((prev) => ({ ...prev, total_amount: Number(e.target.value) }))
            }
            disabled={form.fare_mode !== "MANUAL"}
          />
        </FormField>
        <FormField label="Método do pagamento inicial" required>
          <select
            className="input input-delightful"
            value={form.payment_method}
            onChange={(e) => setForm((prev) => ({ ...prev, payment_method: e.target.value }))}
            required
          >
            <option value="PIX">PIX</option>
            <option value="CARD">Cartão</option>
            <option value="CASH">Dinheiro</option>
            <option value="TRANSFER">Transferência</option>
            <option value="OTHER">Outro</option>
          </select>
        </FormField>
        <FormField label="Valor inicial" hint={`Mínimo de ${formatCurrency(minInitialAmount)}`} required>
          <input
            className={`input input-delightful${paymentError ? " error" : ""}`}
            type="number"
            value={form.deposit_amount}
            min={0}
            max={form.total_amount > 0 ? form.total_amount : undefined}
            step="0.01"
            onChange={(e) => {
              const amount = Number(e.target.value);
              setForm((prev) => {
                const remainder = Math.max((Number(prev.total_amount) || 0) - amount, 0);
                return {
                  ...prev,
                  deposit_amount: amount,
                  remainder_amount: remainder,
                };
              });
            }}
            required
          />
        </FormField>
        <FormField label="Saldo restante" hint="Calculado automaticamente">
          <input className="input" type="number" value={form.remainder_amount} readOnly />
        </FormField>
        <FormField label="Descrição da cobrança" hint="Opcional">
          <input
            className="input input-delightful"
            value={form.payment_description}
            onChange={(e) =>
              setForm((prev) => ({ ...prev, payment_description: e.target.value }))
            }
            placeholder="Ex: Passagem"
          />
        </FormField>
        {!isAutomatic ? (
          <FormField label="Observações do pagamento" hint="Opcional">
            <input
              className="input input-delightful"
              value={form.payment_notes}
              onChange={(e) => setForm((prev) => ({ ...prev, payment_notes: e.target.value }))}
              placeholder="Ex: Recebido no balcão"
            />
          </FormField>
        ) : null}
      </div>

      {/* Live Preview Summary */}
      {form.total_amount > 0 && (
        <div className="live-preview" style={{ marginTop: "16px" }}>
          <div className="live-preview-header">Resumo financeiro</div>
          <div className="live-preview-item">
            <span className="live-preview-label">Total da reserva</span>
            <span className="live-preview-value">{formatCurrency(form.total_amount)}</span>
          </div>
          <div className="live-preview-item">
            <span className="live-preview-label">Pagamento inicial</span>
            <span className="live-preview-value">{formatCurrency(form.deposit_amount)}</span>
          </div>
          <div className="live-preview-item">
            <span className="live-preview-label">Saldo restante</span>
            <span className={`live-preview-value ${form.remainder_amount > 0 ? "" : "highlight"}`}>
              {formatCurrency(form.remainder_amount)}
            </span>
          </div>
        </div>
      )}

      {paymentError ? <InlineAlert tone="error">{paymentError}</InlineAlert> : null}

      <div className="form-step-actions">
        <button
          className="button secondary"
          type="button"
          onClick={onBack}
          disabled={checkoutLoading}
        >
          Voltar
        </button>
        <button
          className={`button button-delightful ${checkoutLoading ? "loading" : ""}`}
          type="button"
          onClick={onSubmitCheckout}
          disabled={checkoutLoading || !canSubmitCheckout}
        >
          {checkoutLoading ? "Processando..." : "Concluir reserva"}
        </button>
      </div>

      {checkoutResult && !isPaid ? (
        <>
          <PaymentResult
            result={{
              payment: checkoutResult.payment,
              provider_raw: {
                ...(checkoutResult.provider_raw ?? {}),
                url: checkoutResult.checkout_url ?? undefined,
                pixQrCode: checkoutResult.pix_code ?? undefined,
              },
            }}
            statusTone={statusTone}
          />
          {(checkoutResult.payment.status ?? "").toUpperCase() === "PENDING" ? (
            <div className="form-step-actions">
              <button
                className="button secondary"
                type="button"
                onClick={onSync}
                disabled={syncLoading}
              >
                {syncLoading ? "Atualizando..." : "Atualizar status"}
              </button>
            </div>
          ) : null}
        </>
      ) : null}
    </>
  );
}
