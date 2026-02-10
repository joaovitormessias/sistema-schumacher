type TabKey = "AUTOMATIC" | "MANUAL";

type PaymentTabsProps = {
  activeTab: TabKey;
  onChange: (tab: TabKey) => void;
};

export default function PaymentTabs({ activeTab, onChange }: PaymentTabsProps) {
  return (
    <div className="segmented" role="tablist" aria-label="Tipo de pagamento">
      <button
        type="button"
        className={activeTab === "AUTOMATIC" ? "active" : ""}
        onClick={() => onChange("AUTOMATIC")}
        role="tab"
        aria-selected={activeTab === "AUTOMATIC"}
      >
        Automatico (PIX/CARTAO)
      </button>
      <button
        type="button"
        className={activeTab === "MANUAL" ? "active" : ""}
        onClick={() => onChange("MANUAL")}
        role="tab"
        aria-selected={activeTab === "MANUAL"}
      >
        Manual (Dinheiro)
      </button>
    </div>
  );
}
