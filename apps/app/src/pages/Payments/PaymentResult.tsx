import { useState } from "react";
import { Copy, ExternalLink, QrCode } from "lucide-react";
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

function extractCheckoutUrl(providerRaw: any): string | null {
  if (!providerRaw || typeof providerRaw !== "object") return null;
  const direct = providerRaw.url;
  if (typeof direct === "string" && direct.trim() !== "") return direct;
  const nested = providerRaw.data?.url;
  if (typeof nested === "string" && nested.trim() !== "") return nested;
  return null;
}

function extractPixCode(providerRaw: any): string | null {
  if (!providerRaw || typeof providerRaw !== "object") return null;
  const code = providerRaw.data?.pixQrCode ?? providerRaw.pixQrCode;
  if (typeof code === "string" && code.trim() !== "") return code;
  return null;
}

export default function PaymentResult({ result, statusTone }: PaymentResultProps) {
  const checkoutUrl = extractCheckoutUrl(result.provider_raw);
  const pixCode = extractPixCode(result.provider_raw);
  const [copyFeedback, setCopyFeedback] = useState<string | null>(null);

  const handleCopyPix = async () => {
    if (!pixCode) return;
    try {
      await navigator.clipboard.writeText(pixCode);
      setCopyFeedback("Codigo PIX copiado.");
      window.setTimeout(() => setCopyFeedback(null), 2000);
    } catch {
      setCopyFeedback("Nao foi possivel copiar automaticamente.");
      window.setTimeout(() => setCopyFeedback(null), 2500);
    }
  };

  return (
    <div className="section">
      <div className="section-header">
        <div className="section-title">Confirmacao do pagamento</div>
      </div>

      <div className="card payment-result-grid">
        <div className="payment-result-main">
          <p>
            Status: <StatusBadge tone={statusTone(result.payment.status)}>{result.payment.status}</StatusBadge>
          </p>
          <p>Valor: {formatCurrency(result.payment.amount)}</p>
          {result.payment.provider_ref ? <p>Referencia: {result.payment.provider_ref}</p> : null}

          {checkoutUrl ? (
            <div className="payment-actions-inline">
              <a className="button secondary sm" href={checkoutUrl} target="_blank" rel="noreferrer">
                <ExternalLink size={14} />
                Abrir cobranca
              </a>
            </div>
          ) : null}
        </div>

        {pixCode ? (
          <div className="pix-code-card" aria-live="polite">
            <div className="pix-code-header">
              <QrCode size={16} />
              <span>PIX copia e cola</span>
            </div>
            <code className="pix-code-value">{pixCode}</code>
            <div className="pix-code-actions">
              <button className="button sm button-delightful" type="button" onClick={handleCopyPix}>
                <Copy size={14} />
                Copiar codigo PIX
              </button>
              {copyFeedback ? <span className="text-caption">{copyFeedback}</span> : null}
            </div>
          </div>
        ) : null}
      </div>
    </div>
  );
}
