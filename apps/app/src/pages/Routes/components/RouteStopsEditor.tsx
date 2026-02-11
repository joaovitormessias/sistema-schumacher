import { ArrowDown, ArrowUp, Plus, Trash2 } from "lucide-react";

export type RouteStopFormItem = {
  client_id: string;
  id?: string;
  city: string;
  eta_offset_minutes: number | "";
  notes: string;
  persisted: boolean;
};

type RouteStopsEditorProps = {
  stops: RouteStopFormItem[];
  disabled?: boolean;
  disableReorder?: boolean;
  disableDelete?: boolean;
  onAdd: () => void;
  onMove: (index: number, direction: -1 | 1) => void;
  onRemove: (index: number) => void;
  onChange: (index: number, patch: Partial<RouteStopFormItem>) => void;
};

export default function RouteStopsEditor({
  stops,
  disabled,
  disableReorder,
  disableDelete,
  onAdd,
  onMove,
  onRemove,
  onChange,
}: RouteStopsEditorProps) {
  return (
    <div className="route-stops-editor">
      <div className="route-stops-header">
        <div>
          <div className="section-title">Paradas</div>
          <div className="form-hint">
            Defina a ordem operacional do itinerario. O campo ETA (min) representa
            minutos apos a partida.
          </div>
        </div>
        <button className="button secondary sm" type="button" onClick={onAdd} disabled={disabled}>
          <Plus size={14} aria-hidden="true" />
          Adicionar parada
        </button>
      </div>

      {stops.length > 0 ? (
        <div className="route-stop-legend" aria-hidden="true">
          <span>Cidade</span>
          <span>ETA (min apos saida)</span>
          <span>Ponto/observacoes</span>
        </div>
      ) : null}

      {stops.length === 0 ? (
        <div className="route-stop-empty">Nenhuma parada adicionada.</div>
      ) : null}

      {stops.map((stop, index) => (
        <div className="route-stop-row" key={stop.client_id}>
          <div className="route-stop-order">{index + 1}</div>
          <div className="route-stop-fields">
            <input
              className="input"
              placeholder="Cidade da parada"
              value={stop.city}
              onChange={(event) => onChange(index, { city: event.target.value })}
              disabled={disabled}
            />
            <input
              className="input"
              type="number"
              min={0}
              placeholder="ETA (min)"
              value={stop.eta_offset_minutes}
              onChange={(event) =>
                onChange(index, {
                  eta_offset_minutes:
                    event.target.value === "" ? "" : Number(event.target.value),
                })
              }
              disabled={disabled}
            />
            <input
              className="input"
              placeholder="Observacoes (opcional)"
              value={stop.notes}
              onChange={(event) => onChange(index, { notes: event.target.value })}
              disabled={disabled}
            />
          </div>
          <div className="route-stop-actions">
            <button
              className="icon-button"
              type="button"
              onClick={() => onMove(index, -1)}
              disabled={disabled || disableReorder || index === 0}
              aria-label="Mover parada para cima"
              title="Mover parada para cima"
            >
              <ArrowUp size={14} aria-hidden="true" />
            </button>
            <button
              className="icon-button"
              type="button"
              onClick={() => onMove(index, 1)}
              disabled={disabled || disableReorder || index === stops.length - 1}
              aria-label="Mover parada para baixo"
              title="Mover parada para baixo"
            >
              <ArrowDown size={14} aria-hidden="true" />
            </button>
            <button
              className="icon-button"
              type="button"
              onClick={() => onRemove(index)}
              disabled={disabled || disableDelete}
              aria-label="Remover parada"
              title="Remover parada"
            >
              <Trash2 size={14} aria-hidden="true" />
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}
