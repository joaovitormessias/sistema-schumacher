import type { FormEvent } from "react";
import FormField from "../../components/FormField";
import StatusBadge from "../../components/StatusBadge";
import { pricingRuleTypeHelp, pricingRuleTypeLabel, pricingScopeLabel } from "../../utils/labels";

type PricingForm = {
  name: string;
  scope: string;
  scope_id: string;
  rule_type: string;
  priority: number;
  is_active: boolean;
  params: string;
};

type PricingRuleFormProps = {
  form: PricingForm;
  isEditing: boolean;
  showAdvanced: boolean;
  paramsPreview: Record<string, unknown> | null;
  paramsError: string | null;
  onSubmit: (event: FormEvent) => void;
  onCancel: () => void;
  onToggleAdvanced: () => void;
  onRuleTypeChange: (next: string) => void;
  onParamsChange: (next: string) => void;
  onResetParams: () => void;
  setForm: (value: PricingForm | ((prev: PricingForm) => PricingForm)) => void;
};

export default function PricingRuleForm({
  form,
  isEditing,
  showAdvanced,
  paramsPreview,
  paramsError,
  onSubmit,
  onCancel,
  onToggleAdvanced,
  onRuleTypeChange,
  onParamsChange,
  onResetParams,
  setForm,
}: PricingRuleFormProps) {
  const seasonParams = paramsPreview && typeof paramsPreview === "object" ? (paramsPreview as any) : null;
  const seasonWindow = Array.isArray(seasonParams?.windows) ? seasonParams.windows[0] : null;
  const seasonFrom = seasonWindow?.from ?? "";
  const seasonTo = seasonWindow?.to ?? "";
  const seasonMultiplier =
    typeof seasonWindow?.multiplier === "number" ? seasonWindow.multiplier : 1.1;

  const updateSeason = (next: { from: string; to: string; multiplier: number }) => {
    onParamsChange(
      JSON.stringify(
        {
          windows: [
            {
              from: next.from,
              to: next.to,
              multiplier: next.multiplier,
            },
          ],
        },
        null,
        2
      )
    );
  };

  return (
    <div className="section">
      <div className="section-header">
        <div className="section-title">Nova regra</div>
        {isEditing ? <StatusBadge tone="warning">Editando</StatusBadge> : null}
      </div>

      <form className="form-grid" onSubmit={onSubmit}>
        <FormField label="Nome da regra" required>
          <input
            className="input"
            placeholder="Ex: Alta temporada"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            required
          />
        </FormField>
        <FormField label="Tipo de regra">
          <select
            className="input"
            value={form.rule_type}
            onChange={(e) => onRuleTypeChange(e.target.value)}
          >
            {Object.keys(pricingRuleTypeLabel).map((key) => (
              <option key={key} value={key}>
                {pricingRuleTypeLabel[key]}
              </option>
            ))}
          </select>
          {pricingRuleTypeHelp[form.rule_type] ? (
            <span className="form-hint">{pricingRuleTypeHelp[form.rule_type]}</span>
          ) : null}
        </FormField>
        <FormField label="Escopo">
          <select
            className="input"
            value={form.scope}
            onChange={(e) => setForm({ ...form, scope: e.target.value })}
          >
            {Object.keys(pricingScopeLabel).map((key) => (
              <option key={key} value={key}>
                {pricingScopeLabel[key]}
              </option>
            ))}
          </select>
        </FormField>
        <FormField
          label="Selecione: aplicar em qual rota/viagem?"
          hint="Obrigatório para rota ou viagem. Ex: ID da rota."
        >
          <input
            className="input"
            placeholder="Ex: 9f3a1c..."
            value={form.scope_id}
            onChange={(e) => setForm({ ...form, scope_id: e.target.value })}
            disabled={form.scope === "GLOBAL"}
          />
        </FormField>
        <FormField label="Ordem de aplicação" hint="Menor número = maior prioridade">
          <input
            className="input"
            type="number"
            placeholder="100"
            value={form.priority}
            onChange={(e) => setForm({ ...form, priority: Number(e.target.value) })}
          />
        </FormField>
        <div className="form-field">
          <span className="form-label">Ativa</span>
          <label className="checkbox">
            <input
              type="checkbox"
              checked={form.is_active}
              onChange={(e) => setForm({ ...form, is_active: e.target.checked })}
            />
            Regra aplicada na cotação
          </label>
        </div>

        <div className="full-span">
          <div className="toolbar">
            <div className="toolbar-group">
              <button className="button secondary" type="button" onClick={onToggleAdvanced}>
                {showAdvanced ? "Ocultar parâmetros avançados" : "Mostrar parâmetros avançados"}
              </button>
              <button className="button ghost" type="button" onClick={onResetParams}>
                Restaurar padrão
              </button>
            </div>
            {!showAdvanced && (
              <span className="text-caption">Parâmetros padrão (você pode customizar).</span>
            )}
          </div>

          {!showAdvanced && form.rule_type === "SEASON" ? (
            <div className="form-grid">
              <FormField label="Data inicial" required>
                <input
                  className="input"
                  type="date"
                  value={seasonFrom}
                  onChange={(e) =>
                    updateSeason({ from: e.target.value, to: seasonTo, multiplier: seasonMultiplier })
                  }
                />
              </FormField>
              <FormField label="Data final" required>
                <input
                  className="input"
                  type="date"
                  value={seasonTo}
                  onChange={(e) =>
                    updateSeason({ from: seasonFrom, to: e.target.value, multiplier: seasonMultiplier })
                  }
                />
              </FormField>
              <FormField label="Multiplicador de preço (%)" hint="Ex: 1.15 = +15%">
                <input
                  className="input"
                  type="number"
                  step="0.01"
                  value={seasonMultiplier}
                  onChange={(e) =>
                    updateSeason({
                      from: seasonFrom,
                      to: seasonTo,
                      multiplier: Number(e.target.value),
                    })
                  }
                />
              </FormField>
            </div>
          ) : null}

          {!showAdvanced && form.rule_type !== "SEASON" ? (
            <div className="form-summary">
              <strong>Parâmetros padrão (você pode customizar).</strong>
              <span>Ative o modo avançado para editar o JSON.</span>
            </div>
          ) : null}

          {showAdvanced ? (
            <textarea
              className="input mono space-top-2"
              rows={7}
              placeholder="Parâmetros (JSON)"
              value={form.params}
              onChange={(e) => onParamsChange(e.target.value)}
            />
          ) : null}
        </div>

        <div className="form-actions full-span">
          <button className="button" type="submit">
            {isEditing ? "Salvar alterações" : "Criar regra"}
          </button>
          {isEditing ? (
            <button className="button secondary" type="button" onClick={onCancel}>
              Cancelar edição
            </button>
          ) : null}
        </div>
      </form>

      {paramsError && showAdvanced ? <div className="alert error">{paramsError}</div> : null}
    </div>
  );
}
