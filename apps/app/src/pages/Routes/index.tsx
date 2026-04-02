import { useEffect, useMemo, useState } from "react";
import DataTable, { type DataTableColumn } from "../../components/table/DataTable";
import EmptyState from "../../components/EmptyState";
import InlineAlert from "../../components/InlineAlert";
import Modal from "../../components/overlay/Modal";
import PageHeader from "../../components/PageHeader";
import SearchToolbar from "../../components/input/SearchToolbar";
import { useRoutes, type RouteItem } from "../../hooks/useRoutes";
import { useRouteStops, type RouteStopItem } from "../../hooks/useRouteStops";
import { useCityCandidates, type CityCandidate } from "../../hooks/useCityCandidates";
import { apiGet, apiPatch, apiPost, apiPut, APIRequestError } from "../../services/api";

type SegmentPriceStop = {
  stop_id: string;
  display_name: string;
  stop_order: number;
};

type SegmentPriceItem = {
  origin_stop_id: string;
  origin_display_name: string;
  origin_stop_order: number;
  destination_stop_id: string;
  destination_display_name: string;
  destination_stop_order: number;
  price: number | null;
  status: string;
  configured: boolean;
};

type SegmentPriceMatrix = {
  route_id: string;
  stops: SegmentPriceStop[];
  items: SegmentPriceItem[];
};

type EditableSegmentPriceItem = SegmentPriceItem & {
  originalConfigured: boolean;
  enabled: boolean;
  priceInput: string;
  statusSelection: "ACTIVE" | "INACTIVE";
};

function toErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof APIRequestError) return error.message;
  if (error instanceof Error) return error.message;
  return fallback;
}

function hasSelectedCity(value: string, selected: CityCandidate | null): selected is CityCandidate {
  if (!selected) return false;
  return selected.city.trim().toLowerCase() === value.trim().toLowerCase();
}

function parseMoneyInput(value: string): number | null {
  const normalized = value.replace(/\./g, "").replace(",", ".").trim();
  if (!normalized) return null;
  const parsed = Number(normalized);
  if (Number.isNaN(parsed)) return null;
  return parsed;
}

function formatMoneyInput(value: number | null): string {
  if (value === null) return "";
  return value.toFixed(2).replace(".", ",");
}

function formatCompactId(id: string, prefix = 6, suffix = 6): string {
  if (!id) return "-";
  if (id.length <= prefix + suffix + 1) return id;
  return `${id.slice(0, prefix)}...${id.slice(-suffix)}`;
}

function toEditable(item: SegmentPriceItem): EditableSegmentPriceItem {
  return {
    ...item,
    originalConfigured: item.configured,
    enabled: item.configured,
    priceInput: formatMoneyInput(item.price),
    statusSelection: item.status === "INACTIVE" ? "INACTIVE" : "ACTIVE",
  };
}

export default function RoutesPage() {
  const [search, setSearch] = useState("");
  const [selectedRouteId, setSelectedRouteId] = useState<string | null>(null);

  const [createRouteOpen, setCreateRouteOpen] = useState(false);
  const [createRouteName, setCreateRouteName] = useState("");
  const [createRouteOrigin, setCreateRouteOrigin] = useState("");
  const [createRouteOriginCandidate, setCreateRouteOriginCandidate] = useState<CityCandidate | null>(null);
  const [createRouteDestination, setCreateRouteDestination] = useState("");
  const [createRouteDestinationCandidate, setCreateRouteDestinationCandidate] = useState<CityCandidate | null>(null);
  const [createRouteSaving, setCreateRouteSaving] = useState(false);
  const [createRouteError, setCreateRouteError] = useState<string | null>(null);

  const [editRouteOpen, setEditRouteOpen] = useState(false);
  const [editRouteName, setEditRouteName] = useState("");
  const [editRouteActive, setEditRouteActive] = useState(true);
  const [editRouteSaving, setEditRouteSaving] = useState(false);
  const [editRouteError, setEditRouteError] = useState<string | null>(null);

  const [addStopOpen, setAddStopOpen] = useState(false);
  const [newStopCity, setNewStopCity] = useState("");
  const [newStopCityCandidate, setNewStopCityCandidate] = useState<CityCandidate | null>(null);
  const [newStopOrder, setNewStopOrder] = useState("1");
  const [addStopSaving, setAddStopSaving] = useState(false);
  const [addStopError, setAddStopError] = useState<string | null>(null);

  const [pricesOpen, setPricesOpen] = useState(false);
  const [pricesLoading, setPricesLoading] = useState(false);
  const [pricesSaving, setPricesSaving] = useState(false);
  const [pricesError, setPricesError] = useState<string | null>(null);
  const [showAllPairs, setShowAllPairs] = useState(false);
  const [matrix, setMatrix] = useState<SegmentPriceMatrix | null>(null);
  const [editableItems, setEditableItems] = useState<EditableSegmentPriceItem[]>([]);

  const routesQuery = useRoutes(500, 0, { search, status: "all" });
  const routes = routesQuery.data ?? [];

  useEffect(() => {
    if (!selectedRouteId && routes.length > 0) {
      setSelectedRouteId(routes[0].id);
    }
  }, [routes, selectedRouteId]);

  const selectedRoute = useMemo(
    () => routes.find((item) => item.id === selectedRouteId) ?? null,
    [routes, selectedRouteId]
  );

  const stopsQuery = useRouteStops(selectedRouteId);
  const stops = stopsQuery.data ?? [];
  const originCandidatesQuery = useCityCandidates(createRouteOrigin, createRouteOpen);
  const destinationCandidatesQuery = useCityCandidates(createRouteDestination, createRouteOpen);
  const stopCandidatesQuery = useCityCandidates(newStopCity, addStopOpen);

  const loadError =
    (routesQuery.error as Error | undefined)?.message ||
    (stopsQuery.error as Error | undefined)?.message ||
    null;

  const routeColumns: DataTableColumn<RouteItem>[] = [
    { label: "Rota", accessor: (item) => item.name },
    { label: "Origem -> Destino", accessor: (item) => `${item.origin_city} -> ${item.destination_city}` },
    { label: "Paradas", accessor: (item) => String(item.stop_count ?? 0), width: "100px", align: "center" },
    { label: "Status", accessor: (item) => (item.is_active ? "ATIVA" : "INATIVA"), width: "130px" },
  ];

  const stopColumns: DataTableColumn<RouteStopItem>[] = [
    { label: "Ordem", accessor: (item) => String(item.stop_order), width: "100px", align: "center" },
    { label: "Cidade", accessor: (item) => item.city },
    {
      label: "ID",
      accessor: (item) => <span title={item.id}>{formatCompactId(item.id)}</span>,
      width: "180px",
    },
  ];

  const visibleEditableItems = showAllPairs
    ? editableItems
    : editableItems.filter((item) => item.enabled || item.originalConfigured);

  const openEditRouteModal = () => {
    if (!selectedRoute) return;
    setEditRouteName(selectedRoute.name);
    setEditRouteActive(selectedRoute.is_active);
    setEditRouteError(null);
    setEditRouteOpen(true);
  };

  const openCreateRouteModal = () => {
    setCreateRouteName("");
    setCreateRouteOrigin("");
    setCreateRouteOriginCandidate(null);
    setCreateRouteDestination("");
    setCreateRouteDestinationCandidate(null);
    setCreateRouteError(null);
    setCreateRouteOpen(true);
  };

  const saveCreateRoute = async () => {
    const name = createRouteName.trim();
    const origin = createRouteOrigin.trim();
    const destination = createRouteDestination.trim();
    if (!name || !origin || !destination) {
      setCreateRouteError("Nome, origem e destino sao obrigatorios.");
      return;
    }
    if (origin.toLowerCase() === destination.toLowerCase()) {
      setCreateRouteError("Origem e destino devem ser diferentes.");
      return;
    }
    if (!hasSelectedCity(origin, createRouteOriginCandidate)) {
      setCreateRouteError("Selecione a cidade de origem a partir das sugestoes.");
      return;
    }
    if (!hasSelectedCity(destination, createRouteDestinationCandidate)) {
      setCreateRouteError("Selecione a cidade de destino a partir das sugestoes.");
      return;
    }
    setCreateRouteSaving(true);
    setCreateRouteError(null);
    try {
      const created = await apiPost<RouteItem>("/routes", {
        name,
        origin_city: origin,
        origin_latitude: createRouteOriginCandidate.latitude,
        origin_longitude: createRouteOriginCandidate.longitude,
        destination_city: destination,
        destination_latitude: createRouteDestinationCandidate.latitude,
        destination_longitude: createRouteDestinationCandidate.longitude,
        is_active: false,
      });
      await routesQuery.refetch();
      setSelectedRouteId(created.id);
      setCreateRouteOpen(false);
    } catch (error) {
      setCreateRouteError(toErrorMessage(error, "Nao foi possivel criar a rota."));
    } finally {
      setCreateRouteSaving(false);
    }
  };

  const saveRoute = async () => {
    if (!selectedRoute) return;
    const trimmedName = editRouteName.trim();
    if (!trimmedName) {
      setEditRouteError("Nome da rota e obrigatorio.");
      return;
    }

    setEditRouteSaving(true);
    setEditRouteError(null);
    try {
      await apiPatch(`/routes/${selectedRoute.id}`, {
        name: trimmedName,
        is_active: editRouteActive,
      });
      await routesQuery.refetch();
      setEditRouteOpen(false);
    } catch (error) {
      setEditRouteError(toErrorMessage(error, "Nao foi possivel salvar a rota."));
    } finally {
      setEditRouteSaving(false);
    }
  };

  const openAddStopModal = () => {
    if (!selectedRoute) return;
    setNewStopCity("");
    setNewStopCityCandidate(null);
    setNewStopOrder(String((stops.length || 0) + 1));
    setAddStopError(null);
    setAddStopOpen(true);
  };

  const saveStop = async () => {
    if (!selectedRoute) return;
    const city = newStopCity.trim();
    const order = Number(newStopOrder);
    if (!city) {
      setAddStopError("Informe a cidade da parada.");
      return;
    }
    if (!Number.isFinite(order) || order <= 0) {
      setAddStopError("Ordem da parada invalida.");
      return;
    }
    if (!hasSelectedCity(city, newStopCityCandidate)) {
      setAddStopError("Selecione a cidade da parada a partir das sugestoes.");
      return;
    }
    setAddStopSaving(true);
    setAddStopError(null);
    try {
      await apiPost(`/routes/${selectedRoute.id}/stops`, {
        city,
        stop_order: order,
        latitude: newStopCityCandidate.latitude,
        longitude: newStopCityCandidate.longitude,
      });
      await Promise.all([stopsQuery.refetch(), routesQuery.refetch()]);
      setAddStopOpen(false);
    } catch (error) {
      setAddStopError(toErrorMessage(error, "Nao foi possivel adicionar a parada."));
    } finally {
      setAddStopSaving(false);
    }
  };

  const openPricesModal = async () => {
    if (!selectedRoute) return;
    setPricesOpen(true);
    setPricesLoading(true);
    setPricesSaving(false);
    setShowAllPairs(false);
    setPricesError(null);
    setMatrix(null);
    setEditableItems([]);
    try {
      const data = await apiGet<SegmentPriceMatrix>(`/routes/${selectedRoute.id}/segment-prices`);
      setMatrix(data);
      setEditableItems(data.items.map(toEditable));
    } catch (error) {
      setPricesError(toErrorMessage(error, "Nao foi possivel carregar os precos por trecho."));
    } finally {
      setPricesLoading(false);
    }
  };

  const savePrices = async () => {
    if (!selectedRoute) return;
    const updates: Array<{
      origin_stop_id: string;
      destination_stop_id: string;
      price: number | null;
      status?: "ACTIVE" | "INACTIVE";
    }> = [];

    for (const item of editableItems) {
      if (item.enabled) {
        const parsed = parseMoneyInput(item.priceInput);
        if (parsed === null || parsed < 0) {
          setPricesError(`Preco invalido para ${item.origin_display_name} -> ${item.destination_display_name}.`);
          return;
        }
        updates.push({
          origin_stop_id: item.origin_stop_id,
          destination_stop_id: item.destination_stop_id,
          price: parsed,
          status: item.statusSelection,
        });
      } else if (item.originalConfigured) {
        updates.push({
          origin_stop_id: item.origin_stop_id,
          destination_stop_id: item.destination_stop_id,
          price: null,
        });
      }
    }

    setPricesError(null);
    setPricesSaving(true);
    try {
      const data = await apiPut<SegmentPriceMatrix>(`/routes/${selectedRoute.id}/segment-prices`, { items: updates });
      setMatrix(data);
      setEditableItems(data.items.map(toEditable));
      await routesQuery.refetch();
    } catch (error) {
      setPricesError(toErrorMessage(error, "Nao foi possivel salvar os precos por trecho."));
    } finally {
      setPricesSaving(false);
    }
  };

  const renderCityCandidates = (
    query: string,
    selected: CityCandidate | null,
    candidates: CityCandidate[] | undefined,
    isLoading: boolean,
    error: unknown,
    onSelect: (candidate: CityCandidate) => void
  ) => {
    if (hasSelectedCity(query, selected)) {
      return (
        <div className="route-city-feedback route-city-selected">
          Selecionado: {selected.city} ({selected.display_name})
        </div>
      );
    }

    const trimmedQuery = query.trim();
    if (trimmedQuery.length < 2) {
      return <div className="route-city-feedback">Digite ao menos 2 caracteres para buscar sugestoes.</div>;
    }
    if (isLoading) {
      return <div className="route-city-feedback">Buscando cidades...</div>;
    }
    if (error) {
      return <InlineAlert tone="error">{toErrorMessage(error, "Nao foi possivel buscar cidades.")}</InlineAlert>;
    }
    if (!candidates || candidates.length === 0) {
      return <div className="route-city-feedback">Nenhuma cidade encontrada para essa busca.</div>;
    }

    return (
      <div className="route-city-candidate-list">
        {candidates.map((candidate) => (
          <button
            key={candidate.place_id}
            className="route-city-candidate-btn"
            type="button"
            onClick={() => onSelect(candidate)}
          >
            <span className="route-city-candidate-main">{candidate.city}</span>
            <span className="route-city-candidate-sub">{candidate.display_name}</span>
          </button>
        ))}
      </div>
    );
  };

  return (
    <section className="page">
      <PageHeader
        title="Rotas"
        subtitle="Edite rotas, adicione paradas e configure tarifas por trecho."
        meta={<span className="badge">Gestao</span>}
      />

      <div className="route-workspace">
        <div className="section route-list-panel">
          <div className="route-admin-header">
            <SearchToolbar value={search} onChange={setSearch} resultCount={routes.length} />
            <button className="button sm" type="button" onClick={openCreateRouteModal}>
              Nova rota
            </button>
          </div>
          {loadError ? <InlineAlert tone="error">{loadError}</InlineAlert> : null}
          <DataTable
            columns={routeColumns}
            rows={routes}
            rowKey={(item) => item.id}
            actions={(item) => (
              <button className="button secondary sm" type="button" onClick={() => setSelectedRouteId(item.id)}>
                Gerenciar
              </button>
            )}
            emptyState={
              routesQuery.isLoading ? (
                <EmptyState title="Carregando rotas" description="Aguarde alguns segundos." />
              ) : (
                <EmptyState title="Nenhuma rota encontrada" description="Tente ajustar os filtros." />
              )
            }
          />
        </div>

        <div className="section route-editor-panel">
          <div className="section-header route-admin-header">
            <div className="section-title">
              {selectedRoute ? `Paradas da rota: ${selectedRoute.name}` : "Paradas da rota"}
            </div>
            {selectedRoute ? (
              <div className="route-admin-actions">
                <button className="button secondary sm" type="button" onClick={openEditRouteModal}>
                  Editar rota
                </button>
                <button className="button secondary sm" type="button" onClick={openAddStopModal}>
                  Adicionar parada
                </button>
                <button className="button sm" type="button" onClick={() => void openPricesModal()}>
                  Precos por trecho
                </button>
              </div>
            ) : null}
          </div>

          <DataTable
            columns={stopColumns}
            rows={stops}
            rowKey={(item) => item.id}
            emptyState={
              selectedRouteId ? (
                stopsQuery.isLoading ? (
                  <EmptyState title="Carregando paradas" description="Aguarde alguns segundos." />
                ) : (
                  <EmptyState title="Sem paradas" description="Nao ha paradas cadastradas para esta rota." />
                )
              ) : (
                <EmptyState title="Selecione uma rota" description="Escolha uma rota para gerenciar." />
              )
            }
          />
        </div>
      </div>

      <Modal
        open={createRouteOpen}
        onClose={() => setCreateRouteOpen(false)}
        title="Nova rota"
        footer={
          <>
            <button className="button secondary" type="button" onClick={() => setCreateRouteOpen(false)} disabled={createRouteSaving}>
              Fechar
            </button>
            <button className="button" type="button" onClick={() => void saveCreateRoute()} disabled={createRouteSaving}>
              {createRouteSaving ? "Salvando..." : "Criar rota"}
            </button>
          </>
        }
      >
        {createRouteError ? <InlineAlert tone="error">{createRouteError}</InlineAlert> : null}
        <InlineAlert tone="info">
          Origem e destino sao geolocalizados automaticamente pelo nome da cidade.
        </InlineAlert>
        <label className="field">
          <span className="field-label">Nome da rota</span>
          <input className="input" type="text" value={createRouteName} onChange={(event) => setCreateRouteName(event.target.value)} />
        </label>
        <label className="field">
          <span className="field-label">Origem (ex.: Moncao/MA)</span>
          <input
            className="input"
            type="text"
            value={createRouteOrigin}
            onChange={(event) => {
              setCreateRouteOrigin(event.target.value);
              setCreateRouteOriginCandidate(null);
            }}
          />
        </label>
        {renderCityCandidates(
          createRouteOrigin,
          createRouteOriginCandidate,
          originCandidatesQuery.data,
          originCandidatesQuery.isLoading,
          originCandidatesQuery.error,
          (candidate) => {
            setCreateRouteOrigin(candidate.city);
            setCreateRouteOriginCandidate(candidate);
          }
        )}
        <label className="field">
          <span className="field-label">Destino (ex.: Fraiburgo/SC)</span>
          <input
            className="input"
            type="text"
            value={createRouteDestination}
            onChange={(event) => {
              setCreateRouteDestination(event.target.value);
              setCreateRouteDestinationCandidate(null);
            }}
          />
        </label>
        {renderCityCandidates(
          createRouteDestination,
          createRouteDestinationCandidate,
          destinationCandidatesQuery.data,
          destinationCandidatesQuery.isLoading,
          destinationCandidatesQuery.error,
          (candidate) => {
            setCreateRouteDestination(candidate.city);
            setCreateRouteDestinationCandidate(candidate);
          }
        )}
      </Modal>

      <Modal
        open={editRouteOpen}
        onClose={() => setEditRouteOpen(false)}
        title="Editar rota"
        description={selectedRoute ? `Rota ${selectedRoute.id}` : undefined}
        footer={
          <>
            <button className="button secondary" type="button" onClick={() => setEditRouteOpen(false)} disabled={editRouteSaving}>
              Fechar
            </button>
            <button className="button" type="button" onClick={() => void saveRoute()} disabled={editRouteSaving}>
              {editRouteSaving ? "Salvando..." : "Salvar"}
            </button>
          </>
        }
      >
        {editRouteError ? <InlineAlert tone="error">{editRouteError}</InlineAlert> : null}
        <label className="field">
          <span className="field-label">Nome da rota</span>
          <input className="input" type="text" value={editRouteName} onChange={(event) => setEditRouteName(event.target.value)} />
        </label>
        <label className="segment-prices-checkbox">
          <input type="checkbox" checked={editRouteActive} onChange={(event) => setEditRouteActive(event.target.checked)} />
          <span>Rota ativa</span>
        </label>
      </Modal>

      <Modal
        open={addStopOpen}
        onClose={() => setAddStopOpen(false)}
        title="Adicionar parada"
        description={selectedRoute ? `Rota ${selectedRoute.id}` : undefined}
        footer={
          <>
            <button className="button secondary" type="button" onClick={() => setAddStopOpen(false)} disabled={addStopSaving}>
              Fechar
            </button>
            <button className="button" type="button" onClick={() => void saveStop()} disabled={addStopSaving}>
              {addStopSaving ? "Salvando..." : "Adicionar"}
            </button>
          </>
        }
      >
        {addStopError ? <InlineAlert tone="error">{addStopError}</InlineAlert> : null}
        <label className="field">
          <span className="field-label">Cidade (ex.: Videira/SC)</span>
          <input
            className="input"
            type="text"
            value={newStopCity}
            onChange={(event) => {
              setNewStopCity(event.target.value);
              setNewStopCityCandidate(null);
            }}
          />
        </label>
        {renderCityCandidates(
          newStopCity,
          newStopCityCandidate,
          stopCandidatesQuery.data,
          stopCandidatesQuery.isLoading,
          stopCandidatesQuery.error,
          (candidate) => {
            setNewStopCity(candidate.city);
            setNewStopCityCandidate(candidate);
          }
        )}
        <label className="field">
          <span className="field-label">Ordem</span>
          <input
            className="input"
            type="number"
            min={1}
            value={newStopOrder}
            onChange={(event) => setNewStopOrder(event.target.value)}
          />
        </label>
        <InlineAlert tone="info">A localizacao da parada sera buscada automaticamente pelo nome da cidade.</InlineAlert>
      </Modal>

      <Modal
        open={pricesOpen}
        onClose={() => setPricesOpen(false)}
        title="Precos por trecho"
        size="lg"
        description={selectedRoute ? `As tarifas valem para a rota ${selectedRoute.name}.` : undefined}
        footer={
          <>
            <button className="button secondary" type="button" onClick={() => setPricesOpen(false)} disabled={pricesSaving}>
              Fechar
            </button>
            <button className="button" type="button" onClick={() => void savePrices()} disabled={pricesLoading || pricesSaving || !matrix}>
              {pricesSaving ? "Salvando..." : "Salvar matriz"}
            </button>
          </>
        }
      >
        <InlineAlert tone="info">
          Esta alteracao atualiza a matriz de precos da rota e impacta novas cotacoes.
        </InlineAlert>

        {pricesError ? <InlineAlert tone="error">{pricesError}</InlineAlert> : null}

        {pricesLoading ? (
          <EmptyState title="Carregando matriz de precos" description="Aguarde alguns segundos." />
        ) : null}

        {!pricesLoading && matrix && editableItems.length === 0 ? (
          <EmptyState title="Sem pares de parada" description="Cadastre pelo menos duas paradas na rota." />
        ) : null}

        {!pricesLoading && editableItems.length > 0 ? (
          <>
            <div className="segment-prices-toolbar">
              <label className="segment-prices-checkbox">
                <input type="checkbox" checked={showAllPairs} onChange={(event) => setShowAllPairs(event.target.checked)} />
                <span>Mostrar todos os pares ({editableItems.length})</span>
              </label>
              <span className="segment-prices-counter">
                Exibindo {visibleEditableItems.length} de {editableItems.length} pares
              </span>
            </div>

            <div className="segment-prices-stops">
              {matrix?.stops.map((stop) => (
                <span className="segment-prices-stop-chip" key={stop.stop_id}>
                  #{stop.stop_order} {stop.display_name}
                </span>
              ))}
            </div>

            <div className="segment-prices-table-wrap">
              <table className="segment-prices-table">
                <thead>
                  <tr>
                    <th>Configurar</th>
                    <th>Origem</th>
                    <th>Destino</th>
                    <th>Preco (R$)</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {visibleEditableItems.map((item) => {
                    const index = editableItems.findIndex(
                      (row) =>
                        row.origin_stop_id === item.origin_stop_id &&
                        row.destination_stop_id === item.destination_stop_id
                    );
                    return (
                      <tr key={`${item.origin_stop_id}-${item.destination_stop_id}`}>
                        <td>
                          <label className="segment-prices-checkbox">
                            <input
                              type="checkbox"
                              checked={item.enabled}
                              onChange={(event) => {
                                const enabled = event.target.checked;
                                setEditableItems((previous) =>
                                  previous.map((row, rowIndex) =>
                                    rowIndex === index
                                      ? {
                                          ...row,
                                          enabled,
                                          priceInput: enabled
                                            ? row.priceInput || formatMoneyInput(row.price ?? 0)
                                            : "",
                                        }
                                      : row
                                  )
                                );
                              }}
                            />
                            <span>{item.enabled ? "Ativo" : "Nao configurado"}</span>
                          </label>
                        </td>
                        <td>
                          {item.origin_display_name}
                          <span className="segment-prices-order">#{item.origin_stop_order}</span>
                        </td>
                        <td>
                          {item.destination_display_name}
                          <span className="segment-prices-order">#{item.destination_stop_order}</span>
                        </td>
                        <td>
                          <input
                            className="input"
                            type="text"
                            inputMode="decimal"
                            placeholder="0,00"
                            value={item.priceInput}
                            disabled={!item.enabled}
                            onChange={(event) => {
                              const priceInput = event.target.value;
                              setEditableItems((previous) =>
                                previous.map((row, rowIndex) =>
                                  rowIndex === index ? { ...row, priceInput } : row
                                )
                              );
                            }}
                          />
                        </td>
                        <td>
                          <select
                            className="input"
                            value={item.statusSelection}
                            disabled={!item.enabled}
                            onChange={(event) => {
                              const nextStatus = event.target.value === "INACTIVE" ? "INACTIVE" : "ACTIVE";
                              setEditableItems((previous) =>
                                previous.map((row, rowIndex) =>
                                  rowIndex === index
                                    ? {
                                        ...row,
                                        statusSelection: nextStatus,
                                      }
                                    : row
                                )
                              );
                            }}
                          >
                            <option value="ACTIVE">ACTIVE</option>
                            <option value="INACTIVE">INACTIVE</option>
                          </select>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </>
        ) : null}
      </Modal>
    </section>
  );
}
