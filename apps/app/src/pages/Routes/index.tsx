import { useEffect, useMemo, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Copy, Edit3, PauseCircle, PlayCircle } from "lucide-react";
import DataTable, { type DataTableColumn } from "../../components/table/DataTable";
import PageHeader from "../../components/PageHeader";
import StatusBadge from "../../components/StatusBadge";
import SearchToolbar from "../../components/input/SearchToolbar";
import PaginationControls from "../../components/input/PaginationControls";
import InlineAlert from "../../components/InlineAlert";
import EmptyState from "../../components/EmptyState";
import Stepper, { type StepperStep } from "../../components/Stepper";
import Drawer from "../../components/overlay/Drawer";
import ConfirmDialog from "../../components/overlay/ConfirmDialog";
import { Skeleton } from "../../components/feedback/SkeletonLoader";
import useMediaQuery from "../../hooks/useMediaQuery";
import useDebouncedValue from "../../hooks/useDebouncedValue";
import useToast from "../../hooks/useToast";
import { apiDelete, apiGet, apiPatch, apiPost, APIRequestError } from "../../services/api";
import { buildListQuery } from "../../hooks/buildListQuery";
import type { RouteItem } from "../../hooks/useRoutes";
import { useRouteStops } from "../../hooks/useRouteStops";
import RouteStopsEditor, {
  type RouteStopFormItem,
} from "./components/RouteStopsEditor";
import RoutePublishChecklist from "./components/RoutePublishChecklist";

type EditorTarget =
  | { mode: "new" }
  | {
      mode: "route";
      routeId: string;
    };

type RouteFormState = {
  name: string;
  origin_city: string;
  destination_city: string;
};

const PAGE_SIZE = 20;

const initialForm: RouteFormState = {
  name: "",
  origin_city: "",
  destination_city: "",
};

const createClientId = () => {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `stop-${Date.now()}-${Math.random().toString(16).slice(2)}`;
};

const toStopForm = (item: {
  id: string;
  city: string;
  eta_offset_minutes?: number | null;
  notes?: string | null;
}): RouteStopFormItem => ({
  client_id: createClientId(),
  id: item.id,
  city: item.city,
  eta_offset_minutes:
    item.eta_offset_minutes === null || item.eta_offset_minutes === undefined
      ? ""
      : item.eta_offset_minutes,
  notes: item.notes ?? "",
  persisted: true,
});

const sameCity = (a: string, b: string) =>
  a.trim().toLowerCase() === b.trim().toLowerCase();

const computeMissingRequirements = (
  form: RouteFormState,
  stops: RouteStopFormItem[]
) => {
  const requirements = [] as string[];

  if (stops.length < 2) {
    requirements.push("at least two stops are required");
  }
  if (stops.some((stop) => stop.city.trim() === "")) {
    requirements.push("each stop must define a city");
  }
  if (stops.length > 0) {
    if (!sameCity(form.origin_city, stops[0].city)) {
      requirements.push("first stop city must match origin_city");
    }
    if (!sameCity(form.destination_city, stops[stops.length - 1].city)) {
      requirements.push("last stop city must match destination_city");
    }
  }

  let previousETA: number | null = null;
  for (const stop of stops) {
    if (stop.eta_offset_minutes === "") continue;
    if (stop.eta_offset_minutes < 0) {
      requirements.push("eta_offset_minutes must be >= 0");
    }
    if (previousETA !== null && stop.eta_offset_minutes < previousETA) {
      requirements.push("eta_offset_minutes must be non-decreasing");
    }
    previousETA = stop.eta_offset_minutes;
  }

  return Array.from(new Set(requirements));
};

const configStatusMeta = (status: RouteItem["configuration_status"]) => {
  if (status === "ACTIVE") return { label: "Ativa", tone: "success" as const };
  if (status === "READY") return { label: "Pronta", tone: "info" as const };
  if (status === "SUSPENDED") return { label: "Suspensa", tone: "neutral" as const };
  return { label: "Incompleta", tone: "warning" as const };
};

const mapErrorMessage = (error: unknown, fallback: string) => {
  if (error instanceof APIRequestError) {
    return error.message || fallback;
  }
  if (error instanceof Error) {
    return error.message || fallback;
  }
  return fallback;
};

export default function RoutesPage() {
  const toast = useToast();
  const queryClient = useQueryClient();
  const isMobile = useMediaQuery("(max-width: 900px)");

  const [page, setPage] = useState(0);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<"active" | "inactive" | "all">(
    "all"
  );
  const [editorTarget, setEditorTarget] = useState<EditorTarget | null>(null);
  const [pendingTarget, setPendingTarget] = useState<EditorTarget | null>(null);
  const [discardConfirmOpen, setDiscardConfirmOpen] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [form, setForm] = useState<RouteFormState>(initialForm);
  const [stops, setStops] = useState<RouteStopFormItem[]>([]);
  const [isDirty, setIsDirty] = useState(false);
  const [savingDraft, setSavingDraft] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [currentStep, setCurrentStep] = useState<"basic" | "stops" | "review">(
    "basic"
  );
  const [actionError, setActionError] = useState<string | null>(null);
  const [suspendTarget, setSuspendTarget] = useState<RouteItem | null>(null);

  const debouncedSearch = useDebouncedValue(search, 450);

  const routesQuery = useQuery({
    queryKey: ["routes-v2", PAGE_SIZE, page, debouncedSearch, statusFilter],
    queryFn: () =>
      apiGet<RouteItem[]>(
        buildListQuery("/routes", {
          limit: PAGE_SIZE,
          offset: page * PAGE_SIZE,
          search: debouncedSearch,
          status: statusFilter,
        })
      ),
  });

  const routes = routesQuery.data ?? [];
  const selectedRouteId =
    editorTarget?.mode === "route" ? editorTarget.routeId : null;
  const selectedRoute = useMemo(
    () => routes.find((route) => route.id === selectedRouteId) ?? null,
    [routes, selectedRouteId]
  );
  const stopsQuery = useRouteStops(selectedRouteId);

  useEffect(() => {
    if (!editorTarget) {
      if (routes.length > 0) {
        setEditorTarget({ mode: "route", routeId: routes[0].id });
      } else {
        setEditorTarget({ mode: "new" });
      }
    }
  }, [editorTarget, routes]);

  useEffect(() => {
    if (editorTarget?.mode !== "route" || !selectedRoute || isDirty) return;
    setForm({
      name: selectedRoute.name,
      origin_city: selectedRoute.origin_city,
      destination_city: selectedRoute.destination_city,
    });
  }, [editorTarget, selectedRoute, isDirty]);

  useEffect(() => {
    if (editorTarget?.mode !== "route" || isDirty) return;
    if (!stopsQuery.data) {
      setStops([]);
      return;
    }
    setStops(stopsQuery.data.map(toStopForm));
  }, [editorTarget, isDirty, stopsQuery.data]);

  useEffect(() => {
    if (!stopsQuery.error) return;
    setActionError(mapErrorMessage(stopsQuery.error, "Erro ao carregar paradas."));
  }, [stopsQuery.error]);

  const requestSwitchEditor = (target: EditorTarget, openOnMobile = false) => {
    if (isDirty) {
      setPendingTarget(target);
      setDiscardConfirmOpen(true);
      if (openOnMobile) {
        setDrawerOpen(true);
      }
      return;
    }
    setActionError(null);
    setCurrentStep("basic");
    setIsDirty(false);
    setEditorTarget(target);
    if (target.mode === "new") {
      setForm(initialForm);
      setStops([]);
    }
    if (openOnMobile) {
      setDrawerOpen(true);
    }
  };

  const confirmDiscardChanges = () => {
    if (!pendingTarget) {
      setDiscardConfirmOpen(false);
      return;
    }
    setActionError(null);
    setCurrentStep("basic");
    setIsDirty(false);
    setEditorTarget(pendingTarget);
    if (pendingTarget.mode === "new") {
      setForm(initialForm);
      setStops([]);
    }
    setPendingTarget(null);
    setDiscardConfirmOpen(false);
  };

  const missingRequirements = useMemo(
    () => computeMissingRequirements(form, stops),
    [form, stops]
  );

  const canSaveDraft =
    form.name.trim() !== "" &&
    form.origin_city.trim() !== "" &&
    form.destination_city.trim() !== "";

  const lockDestructiveStops = Boolean(selectedRoute?.has_linked_trips);

  const persistStops = async (routeId: string, lockStopOrder: boolean) => {
    const normalizedStops = stops.map((stop, index) => ({
      ...stop,
      city: stop.city.trim(),
      notes: stop.notes.trim(),
      stop_order: index + 1,
    }));

    if (normalizedStops.some((stop) => stop.city === "")) {
      throw new Error("Cada parada precisa ter cidade.");
    }

    if (!lockStopOrder) {
      for (let index = 0; index < normalizedStops.length; index += 1) {
        const stop = normalizedStops[index];
        if (!stop.persisted || !stop.id) continue;
        await apiPatch(`/routes/${routeId}/stops/${stop.id}`, {
          stop_order: 1000 + index + 1,
        });
      }
    }

    for (const stop of normalizedStops) {
      const payload = {
        city: stop.city,
        stop_order: stop.stop_order,
        eta_offset_minutes:
          stop.eta_offset_minutes === "" ? undefined : stop.eta_offset_minutes,
        notes: stop.notes === "" ? undefined : stop.notes,
      };

      if (stop.persisted && stop.id) {
        await apiPatch(`/routes/${routeId}/stops/${stop.id}`, {
          city: payload.city,
          stop_order: lockStopOrder ? undefined : payload.stop_order,
          eta_offset_minutes: payload.eta_offset_minutes,
          notes: payload.notes,
        });
      } else {
        await apiPost(`/routes/${routeId}/stops`, payload);
      }
    }
  };

  const saveDraft = async (showSuccess = true) => {
    if (!canSaveDraft) {
      setActionError("Preencha nome, origem e destino antes de salvar.");
      return null;
    }

    setSavingDraft(true);
    setActionError(null);
    try {
      let routeId = selectedRouteId;
      let createdNew = false;
      if (editorTarget?.mode === "new" || !routeId) {
        const created = await apiPost<RouteItem>("/routes", {
          ...form,
          is_active: false,
        });
        routeId = created.id;
        createdNew = true;
      } else {
        await apiPatch<RouteItem>(`/routes/${routeId}`, {
          ...form,
        });
      }

      await persistStops(routeId, lockDestructiveStops);

      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["routes-v2"] }),
        queryClient.invalidateQueries({ queryKey: ["route-stops", routeId] }),
      ]);

      if (createdNew) {
        setStatusFilter("all");
        setPage(0);
      }
      setEditorTarget({ mode: "route", routeId });
      setIsDirty(false);
      if (showSuccess) {
        toast.success("Rota salva como rascunho.");
      }
      return routeId;
    } catch (error) {
      const message = mapErrorMessage(error, "Erro ao salvar rascunho.");
      setActionError(message);
      toast.error(message);
      return null;
    } finally {
      setSavingDraft(false);
    }
  };

  const publishRoute = async () => {
    setPublishing(true);
    setActionError(null);
    try {
      let routeId = selectedRouteId;
      if (!routeId || isDirty || editorTarget?.mode === "new") {
        routeId = await saveDraft(false);
      }
      if (!routeId) return;

      await apiPost(`/routes/${routeId}/publish`, {});

      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["routes-v2"] }),
        queryClient.invalidateQueries({ queryKey: ["route-stops", routeId] }),
      ]);

      setIsDirty(false);
      setCurrentStep("review");
      toast.success("Rota publicada com sucesso.");
    } catch (error) {
      let message = mapErrorMessage(error, "Erro ao publicar rota.");
      if (error instanceof APIRequestError && error.code === "ROUTE_PUBLISH_BLOCKED") {
        const missing = error.requirementsMissing ?? [];
        if (missing.length > 0) {
          message = `Publicacao bloqueada: ${missing.join("; ")}.`;
        }
      }
      setActionError(message);
      toast.error(message);
      setCurrentStep("review");
    } finally {
      setPublishing(false);
    }
  };

  const suspendRoute = async () => {
    if (!suspendTarget) return;
    try {
      await apiPatch(`/routes/${suspendTarget.id}`, { is_active: false });
      await queryClient.invalidateQueries({ queryKey: ["routes-v2"] });
      toast.success("Rota suspensa.");
      setSuspendTarget(null);
    } catch (error) {
      const message = mapErrorMessage(error, "Erro ao suspender rota.");
      toast.error(message);
      setActionError(message);
      setSuspendTarget(null);
    }
  };

  const duplicateRoute = async (route: RouteItem) => {
    try {
      const duplicated = await apiPost<RouteItem>(`/routes/${route.id}/duplicate`, {});
      setStatusFilter("all");
      setPage(0);
      await queryClient.invalidateQueries({ queryKey: ["routes-v2"] });
      requestSwitchEditor({ mode: "route", routeId: duplicated.id }, isMobile);
      toast.success("Rota duplicada como rascunho.");
    } catch (error) {
      const message = mapErrorMessage(error, "Erro ao duplicar rota.");
      setActionError(message);
      toast.error(message);
    }
  };

  const addStop = () => {
    setStops((prev) => [
      ...prev,
      {
        client_id: createClientId(),
        city: "",
        eta_offset_minutes: "",
        notes: "",
        persisted: false,
      },
    ]);
    setIsDirty(true);
  };

  const moveStop = (index: number, direction: -1 | 1) => {
    if (lockDestructiveStops) {
      toast.warning("Reordenacao bloqueada para rota com viagens vinculadas.");
      return;
    }
    const targetIndex = index + direction;
    if (targetIndex < 0 || targetIndex >= stops.length) return;
    setStops((prev) => {
      const next = [...prev];
      const [item] = next.splice(index, 1);
      next.splice(targetIndex, 0, item);
      return next;
    });
    setIsDirty(true);
  };

  const removeStop = async (index: number) => {
    const stop = stops[index];
    if (!stop) return;
    if (lockDestructiveStops) {
      toast.warning("Remocao bloqueada para rota com viagens vinculadas.");
      return;
    }

    if (stop.persisted && stop.id && selectedRouteId) {
      try {
        await apiDelete(`/routes/${selectedRouteId}/stops/${stop.id}`);
      } catch (error) {
        const message = mapErrorMessage(error, "Erro ao remover parada.");
        setActionError(message);
        toast.error(message);
        return;
      }
    }

    setStops((prev) => prev.filter((_, itemIndex) => itemIndex !== index));
    setIsDirty(true);
  };

  const updateStop = (
    index: number,
    patch: Partial<RouteStopFormItem>
  ) => {
    setStops((prev) =>
      prev.map((item, itemIndex) =>
        itemIndex === index ? { ...item, ...patch } : item
      )
    );
    setIsDirty(true);
  };

  const rows = routes;

  const columns: DataTableColumn<RouteItem>[] = [
    { label: "Nome", accessor: (route) => route.name },
    {
      label: "Origem -> Destino",
      accessor: (route) => `${route.origin_city} -> ${route.destination_city}`,
    },
    { label: "Paradas", accessor: (route) => String(route.stop_count ?? 0), align: "center" },
    {
      label: "Status",
      render: (route) => {
        const meta = configStatusMeta(route.configuration_status);
        return <StatusBadge tone={meta.tone}>{meta.label}</StatusBadge>;
      },
    },
  ];

  const stepperSteps: StepperStep[] = [
    {
      id: "basic",
      title: "Dados basicos",
      status: currentStep === "basic" ? "current" : currentStep === "stops" || currentStep === "review" ? "complete" : "upcoming",
      summary: "Nome e cidades",
    },
    {
      id: "stops",
      title: "Paradas",
      status: currentStep === "stops" ? "current" : currentStep === "review" ? "complete" : "upcoming",
      summary: "Ordem operacional",
    },
    {
      id: "review",
      title: "Revisao",
      status: currentStep === "review" ? "current" : "upcoming",
      summary: "Checklist e publicacao",
    },
  ];

  const editorBody = (
    <div className="route-editor">
      <Stepper
        steps={stepperSteps}
        onStepChange={(id) => setCurrentStep(id as "basic" | "stops" | "review")}
      />

      {currentStep === "basic" ? (
        <div className="route-form-grid">
          <div className="form-field">
            <span className="form-label">Nome da rota</span>
            <input
              className="input"
              value={form.name}
              placeholder="Ex: Executivo Sul"
              onChange={(event) => {
                setForm((prev) => ({ ...prev, name: event.target.value }));
                setIsDirty(true);
              }}
            />
          </div>
          <div className="form-field">
            <span className="form-label">Cidade de origem</span>
            <input
              className="input"
              value={form.origin_city}
              placeholder="Ex: Curitiba"
              onChange={(event) => {
                setForm((prev) => ({ ...prev, origin_city: event.target.value }));
                setIsDirty(true);
              }}
            />
          </div>
          <div className="form-field">
            <span className="form-label">Cidade de destino</span>
            <input
              className="input"
              value={form.destination_city}
              placeholder="Ex: Florianopolis"
              onChange={(event) => {
                setForm((prev) => ({ ...prev, destination_city: event.target.value }));
                setIsDirty(true);
              }}
            />
          </div>
        </div>
      ) : null}

      {currentStep === "stops" ? (
        <>
          {lockDestructiveStops ? (
            <InlineAlert tone="warning">
              Esta rota possui viagens vinculadas. Reordenacao e exclusao de paradas estao bloqueadas.
            </InlineAlert>
          ) : null}
          {editorTarget?.mode === "route" && stopsQuery.isPending ? (
            <Skeleton.Table rows={4} columns={4} />
          ) : (
            <RouteStopsEditor
              stops={stops}
              onAdd={addStop}
              onMove={moveStop}
              onRemove={removeStop}
              onChange={updateStop}
              disableDelete={lockDestructiveStops}
              disableReorder={lockDestructiveStops}
              disabled={savingDraft || publishing}
            />
          )}
        </>
      ) : null}

      {currentStep === "review" ? (
        <div className="route-review-grid">
          <RoutePublishChecklist
            missingRequirements={missingRequirements}
            isActive={Boolean(selectedRoute?.is_active)}
          />
          <div className="route-review-summary">
            <div className="section-title">Resumo</div>
            <div className="route-summary-item">
              <span>Nome</span>
              <strong>{form.name || "-"}</strong>
            </div>
            <div className="route-summary-item">
              <span>Origem</span>
              <strong>{form.origin_city || "-"}</strong>
            </div>
            <div className="route-summary-item">
              <span>Destino</span>
              <strong>{form.destination_city || "-"}</strong>
            </div>
            <div className="route-summary-item">
              <span>Paradas</span>
              <strong>{stops.length}</strong>
            </div>
            <div className="route-summary-item">
              <span>Status atual</span>
              <strong>
                {selectedRoute
                  ? configStatusMeta(selectedRoute.configuration_status).label
                  : "Novo rascunho"}
              </strong>
            </div>
          </div>
        </div>
      ) : null}

      {actionError ? <InlineAlert tone="error">{actionError}</InlineAlert> : null}

      <div className="route-editor-footer">
        <div className="route-editor-nav">
          <button
            className="button secondary sm"
            type="button"
            onClick={() =>
              setCurrentStep((prev) =>
                prev === "review" ? "stops" : prev === "stops" ? "basic" : "basic"
              )
            }
            disabled={currentStep === "basic" || savingDraft || publishing}
          >
            Voltar
          </button>
          <button
            className="button secondary sm"
            type="button"
            onClick={() =>
              setCurrentStep((prev) =>
                prev === "basic" ? "stops" : prev === "stops" ? "review" : "review"
              )
            }
            disabled={currentStep === "review" || savingDraft || publishing}
          >
            Proximo
          </button>
        </div>
        <div className="route-editor-actions">
          <button
            className="button secondary"
            type="button"
            onClick={() => saveDraft(true)}
            disabled={savingDraft || publishing || !canSaveDraft}
          >
            {savingDraft ? "Salvando..." : "Salvar rascunho"}
          </button>
          <button
            className="button"
            type="button"
            onClick={publishRoute}
            disabled={savingDraft || publishing || !canSaveDraft}
          >
            {publishing ? "Publicando..." : "Publicar rota"}
          </button>
        </div>
      </div>
    </div>
  );

  return (
    <section className="page">
      <PageHeader
        title="Rotas"
        subtitle="Cadastro operacional de rotas, paradas e publicacao."
        meta={<span className="badge">Core</span>}
        primaryAction={
          <button
            className="button"
            type="button"
            onClick={() => requestSwitchEditor({ mode: "new" }, isMobile)}
          >
            Nova rota
          </button>
        }
      />

      <div className="route-workspace">
        <div className="section route-list-panel">
          <div className="section-header">
            <div className="section-title">Rotas cadastradas</div>
          </div>

          <SearchToolbar
            value={search}
            onChange={(next) => {
              setSearch(next);
              setPage(0);
            }}
            resultCount={rows.length}
            filters={
              <select
                className="input"
                value={statusFilter}
                onChange={(event) => {
                  setStatusFilter(event.target.value as "active" | "inactive" | "all");
                  setPage(0);
                }}
                aria-label="Filtrar status das rotas"
              >
                <option value="all">Todas</option>
                <option value="active">Ativas</option>
                <option value="inactive">Inativas</option>
              </select>
            }
          />

          <PaginationControls
            page={page}
            pageSize={PAGE_SIZE}
            itemCount={rows.length}
            onPageChange={setPage}
            disabled={routesQuery.isPending}
          />

          {routesQuery.error ? (
            <InlineAlert tone="error">
              {mapErrorMessage(routesQuery.error, "Erro ao listar rotas.")}
            </InlineAlert>
          ) : null}

          {routesQuery.isPending ? (
            <Skeleton.Table rows={6} columns={5} />
          ) : (
            <DataTable
              columns={columns}
              rows={rows}
              rowKey={(route) => route.id}
              actions={(route) => (
                <>
                  <button
                    className="icon-button"
                    type="button"
                    onClick={() => requestSwitchEditor({ mode: "route", routeId: route.id }, isMobile)}
                    aria-label="Editar rota"
                    title="Editar rota"
                  >
                    <Edit3 size={14} aria-hidden="true" />
                  </button>
                  <button
                    className="icon-button"
                    type="button"
                    onClick={() => duplicateRoute(route)}
                    aria-label="Duplicar rota"
                    title="Duplicar rota"
                  >
                    <Copy size={14} aria-hidden="true" />
                  </button>
                  {route.is_active ? (
                    <button
                      className="icon-button"
                      type="button"
                      onClick={() => setSuspendTarget(route)}
                      aria-label="Suspender rota"
                      title="Suspender rota"
                    >
                      <PauseCircle size={14} aria-hidden="true" />
                    </button>
                  ) : (
                    <button
                      className="icon-button"
                      type="button"
                      onClick={async () => {
                        try {
                          await apiPost(`/routes/${route.id}/publish`, {});
                          await queryClient.invalidateQueries({ queryKey: ["routes-v2"] });
                          toast.success("Rota publicada.");
                        } catch (error) {
                          const message = mapErrorMessage(error, "Erro ao publicar rota.");
                          setActionError(message);
                          toast.error(message);
                        }
                      }}
                      aria-label="Publicar rota"
                      title="Publicar rota"
                    >
                      <PlayCircle size={14} aria-hidden="true" />
                    </button>
                  )}
                </>
              )}
              emptyState={
                <EmptyState
                  title="Nenhuma rota encontrada"
                  description="Crie uma nova rota para iniciar o roteiro operacional."
                  action={
                    <button
                      className="button"
                      type="button"
                      onClick={() => requestSwitchEditor({ mode: "new" }, isMobile)}
                    >
                      Criar rota
                    </button>
                  }
                />
              }
            />
          )}
        </div>

        {!isMobile ? <div className="section route-editor-panel">{editorBody}</div> : null}
      </div>

      <Drawer
        open={isMobile && drawerOpen}
        title="Editor de rota"
        description="Fluxo guiado de configuracao e publicacao."
        onClose={() => setDrawerOpen(false)}
      >
        {editorBody}
      </Drawer>

      <ConfirmDialog
        open={discardConfirmOpen}
        title="Descartar alteracoes"
        description="Existem alteracoes nao salvas. Deseja descartar e continuar?"
        confirmLabel="Descartar"
        tone="danger"
        onCancel={() => {
          setDiscardConfirmOpen(false);
          setPendingTarget(null);
        }}
        onConfirm={confirmDiscardChanges}
      />

      <ConfirmDialog
        open={Boolean(suspendTarget)}
        title="Suspender rota"
        description="A rota sera removida da selecao para novas viagens."
        confirmLabel="Suspender"
        tone="danger"
        onCancel={() => setSuspendTarget(null)}
        onConfirm={suspendRoute}
      />
    </section>
  );
}
