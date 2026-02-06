import StatusBadge from "../../components/StatusBadge";
import { formatCurrency } from "../../utils/format";

type PaymentResponse = {
  payment: { id: string; status: string; amount: number; provider_ref?: string };
  provider_raw?: any;
};

type PaymentResultProps = {
  result: PaymentResponse;
  statusTone: (status: string) => "neutral" | "info" | "success" | "warning" | "danger";
};

export default function PaymentResult({ result, statusTone }: PaymentResultProps) {
  return (
    <div className="section">
      <div className="section-header">
        <div className="section-title">Confirmação do pagamento</div>
      </div>
      <div className="card">
        <p>
          Status: <StatusBadge tone={statusTone(result.payment.status)}>{result.payment.status}</StatusBadge>
        </p>
        <p>Valor: {formatCurrency(result.payment.amount)}</p>
        {result.payment.provider_ref ? <p>Referência: {result.payment.provider_ref}</p> : null}
      </div>
    </div>
  );
}
