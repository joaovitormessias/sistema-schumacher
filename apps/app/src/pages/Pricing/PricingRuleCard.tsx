import StatusBadge from "../../components/StatusBadge";
import RuleTypeIcon from "../../components/icons/RuleTypeIcon";
import { formatDate } from "../../utils/format";
import { pricingRuleTypeLabel, pricingScopeLabel } from "../../utils/labels";

type PricingRule = {
  id: string;
  name: string;
  scope: string;
  scope_id?: string | null;
  rule_type: string;
  priority: number;
  is_active: boolean;
  params: Record<string, unknown>;
};

type PricingRuleCardProps = {
  rule: PricingRule;
  onEdit: (rule: PricingRule) => void;
};

export default function PricingRuleCard({ rule, onEdit }: PricingRuleCardProps) {
  const typeLabel = pricingRuleTypeLabel[rule.rule_type] ?? rule.rule_type;
  const scopeLabel = pricingScopeLabel[rule.scope] ?? rule.scope;
  const params = rule.params ?? {};
  const paramKeys = typeof params === "object" && params ? Object.keys(params).length : 0;

  const windows = Array.isArray((params as any).windows) ? (params as any).windows : [];
  const showSeason = rule.rule_type === "SEASON" && windows.length > 0;

  return (
    <div className="card pricing-rule-card">
      <div className="pricing-rule-header">
        <div className="pricing-rule-title-group">
          <RuleTypeIcon ruleType={rule.rule_type as any} size={24} color="var(--accent)" />
          <div>
            <h3>{rule.name}</h3>
            <p className="text-caption">{typeLabel}</p>
          </div>
        </div>
        <StatusBadge
          tone={rule.is_active ? "success" : "warning"}
          label={rule.is_active ? "Ativa" : "Inativa"}
        />
      </div>

      <div className="pricing-rule-meta">
        <span className="meta-item">
          <span className="meta-label">Escopo:</span> {scopeLabel}
          {rule.scope_id ? ` • ${rule.scope_id}` : ""}
        </span>
        <span className="meta-item">
          <span className="meta-label">Prioridade:</span> {rule.priority}
        </span>
      </div>

      {showSeason ? (
        <div className="form-summary">
          {windows.map((window: any, index: number) => {
            const from = window?.from ? formatDate(window.from) : "-";
            const to = window?.to ? formatDate(window.to) : "-";
            const multiplier = typeof window?.multiplier === "number" ? window.multiplier : null;
            const percent = multiplier ? Math.round((multiplier - 1) * 100) : 0;
            const percentLabel = multiplier ? `${percent >= 0 ? "+" : ""}${percent}%` : "-";
            return (
              <div key={`${rule.id}-window-${index}`}>
                <strong>Período:</strong> {from} → {to}
                {" • "}
                <strong>Aumento:</strong> {percentLabel}
              </div>
            );
          })}
        </div>
      ) : (
        <div className="text-caption">
          {paramKeys > 0
            ? "Parâmetros avançados configurados."
            : "Parâmetros padrão para este tipo."}
        </div>
      )}

      <div className="section-spacing">
        <button className="button secondary sm" type="button" onClick={() => onEdit(rule)}>
          Editar
        </button>
      </div>
    </div>
  );
}
