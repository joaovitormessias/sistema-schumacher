import { useEffect, useMemo, useState, type FormEvent } from "react";
import EmptyState from "../../components/EmptyState";
import InlineAlert from "../../components/InlineAlert";
import { Skeleton } from "../../components/feedback/SkeletonLoader";
import PageHeader from "../../components/PageHeader";
import SearchToolbar from "../../components/input/SearchToolbar";
import useToast from "../../hooks/useToast";
import { apiGet, apiPatch, apiPost } from "../../services/api";
import { pricingRuleTypeLabel, pricingScopeLabel } from "../../utils/labels";
import PricingRuleCard from "./PricingRuleCard";
import PricingRuleForm from "./PricingRuleForm";

type PricingRule = {
  id: string;
  name: string;
  scope: string;
  scope_id?: string | null;
  rule_type: string;
  priority: number;
  is_active: boolean;
  params: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

const defaultParamsByType: Record<string, Record<string, unknown>> = {
  OCCUPANCY: {
    apply: "max",
    bands: [
      { min: 0, max: 0.4, multiplier: 0.9 },
      { min: 0.41, max: 0.7, multiplier: 1.0 },
      { min: 0.71, max: 0.9, multiplier: 1.1 },
      { min: 0.91, max: 1.0, multiplier: 1.25 },
    ],
  },
  LEAD_TIME: {
    bands: [
      { min_hours: 168, max_hours: 9999, multiplier: 0.95 },
      { min_hours: 48, max_hours: 167.99, multiplier: 1.0 },
      { min_hours: 0, max_hours: 47.99, multiplier: 1.1 },
    ],
  },
  DOW: {
    days: [
      { days: [1, 2, 3, 4], multiplier: 1.0 },
      { days: [5, 6, 0], multiplier: 1.1 },
    ],
  },
  SEASON: {
    windows: [
      { from: "2026-12-15", to: "2027-01-10", multiplier: 1.2 },
    ],
  },
};

export default function Pricing() {
  const toast = useToast();
  const [rules, setRules] = useState<PricingRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [query, setQuery] = useState("");
  const [form, setForm] = useState({
    name: "",
    scope: "GLOBAL",
    scope_id: "",
    rule_type: "OCCUPANCY",
    priority: 100,
    is_active: true,
    params: JSON.stringify(defaultParamsByType.OCCUPANCY, null, 2),
  });

  const isEditing = Boolean(editingId);

  const paramsPreview = useMemo(() => {
    try {
      return JSON.parse(form.params);
    } catch {
      return null;
    }
  }, [form.params]);

  const paramsError = useMemo(() => {
    if (!form.params.trim()) return null;
    return paramsPreview === null ? "Parâmetros precisam ser um JSON válido (objeto)." : null;
  }, [form.params, paramsPreview]);

  const filtered = useMemo(() => {
    const term = query.trim().toLowerCase();
    if (!term) return rules;
    return rules.filter((rule) => {
      const typeLabel = pricingRuleTypeLabel[rule.rule_type] ?? rule.rule_type;
      const scopeLabel = pricingScopeLabel[rule.scope] ?? rule.scope;
      return [rule.name, typeLabel, scopeLabel, rule.scope_id ?? ""].some((value) =>
        value?.toLowerCase().includes(term)
      );
    });
  }, [rules, query]);

  const load = async () => {
    try {
      setLoading(true);
      const data = await apiGet<PricingRule[]>("/pricing/rules?limit=200");
      setRules(data);
    } catch (err: any) {
      setError(err.message || "Erro ao carregar regras de tarifa");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const resetForm = () => {
    setError(null);
    setEditingId(null);
    setShowAdvanced(false);
    setForm({
      name: "",
      scope: "GLOBAL",
      scope_id: "",
      rule_type: "OCCUPANCY",
      priority: 100,
      is_active: true,
      params: JSON.stringify(defaultParamsByType.OCCUPANCY, null, 2),
    });
  };

  const handleRuleTypeChange = (next: string) => {
    const nextParams = defaultParamsByType[next] ?? {};
    setForm((prev) => ({
      ...prev,
      rule_type: next,
      params: JSON.stringify(nextParams, null, 2),
    }));
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);

    let parsedParams: Record<string, unknown> = {};
    try {
      parsedParams = form.params.trim() ? JSON.parse(form.params) : {};
      if (Array.isArray(parsedParams) || typeof parsedParams !== "object") {
        throw new Error();
      }
    } catch {
      setError("Parâmetros precisam ser um JSON válido (objeto).");
      return;
    }

    if (form.scope !== "GLOBAL" && !form.scope_id.trim()) {
      setError("Selecione uma rota ou viagem para aplicar a regra.");
      return;
    }

    const payload = {
      name: form.name.trim(),
      scope: form.scope,
      scope_id: form.scope === "GLOBAL" ? undefined : form.scope_id.trim(),
      rule_type: form.rule_type,
      priority: Number(form.priority),
      is_active: form.is_active,
      params: parsedParams,
    };

    try {
      if (editingId) {
        await apiPatch(`/pricing/rules/${editingId}`, payload);
        toast.success("Regra atualizada com sucesso.");
      } else {
        await apiPost("/pricing/rules", payload);
        toast.success("Regra criada com sucesso.");
      }
      resetForm();
      await load();
    } catch (err: any) {
      setError(err.message || "Erro ao salvar regra");
    }
  };

  const startEdit = (rule: PricingRule) => {
    setEditingId(rule.id);
    setShowAdvanced(false);
    setForm({
      name: rule.name,
      scope: rule.scope,
      scope_id: rule.scope_id ?? "",
      rule_type: rule.rule_type,
      priority: rule.priority,
      is_active: rule.is_active,
      params: JSON.stringify(rule.params ?? {}, null, 2),
    });
  };

  return (
    <section className="page">
      <PageHeader
        title="Configuração de Tarifas"
        subtitle="Regras de multiplicadores por ocupação, antecedência e sazonalidade."
        meta={<span className="badge">Config</span>}
      />

      <PricingRuleForm
        form={form}
        isEditing={isEditing}
        showAdvanced={showAdvanced}
        paramsPreview={paramsPreview}
        paramsError={paramsError}
        onSubmit={handleSubmit}
        onCancel={resetForm}
        onToggleAdvanced={() => setShowAdvanced((prev) => !prev)}
        onRuleTypeChange={handleRuleTypeChange}
        onParamsChange={(params) => setForm((prev) => ({ ...prev, params }))}
        onResetParams={() =>
          setForm((prev) => ({
            ...prev,
            params: JSON.stringify(defaultParamsByType[prev.rule_type] ?? {}, null, 2),
          }))
        }
        setForm={setForm}
      />

      {error ? <InlineAlert tone="error">{error}</InlineAlert> : null}

      <div className="section">
        <div className="section-header">
          <div className="section-title">Regras ativas</div>
        </div>
        <SearchToolbar
          value={query}
          onChange={setQuery}
          placeholder="Buscar por nome, tipo ou escopo"
          inputLabel="Buscar regras"
        />

        {loading ? (
          <div className="card-grid">
            {Array.from({ length: 4 }).map((_, index) => (
              <Skeleton.Card key={index}>
                <Skeleton.Text lines={2} />
                <Skeleton.Text lines={3} className="section-spacing" />
                <Skeleton.Button width={110} className="section-spacing" />
              </Skeleton.Card>
            ))}
          </div>
        ) : filtered.length === 0 ? (
          <EmptyState
            title="Nenhuma regra encontrada"
            description="Crie uma regra para automatizar a tarifa."
          />
        ) : (
          <div className="card-grid">
            {filtered.map((rule) => (
              <PricingRuleCard key={rule.id} rule={rule} onEdit={startEdit} />
            ))}
          </div>
        )}
      </div>
    </section>
  );
}
