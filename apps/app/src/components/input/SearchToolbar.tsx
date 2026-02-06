import type { ReactNode } from "react";

type SearchToolbarProps = {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  filters?: ReactNode;
  actions?: ReactNode;
  rightActions?: ReactNode;
  inputLabel?: string;
  resultCount?: number;
};

export default function SearchToolbar({
  value,
  onChange,
  placeholder = "Buscar",
  filters,
  actions,
  rightActions,
  inputLabel = "Buscar",
  resultCount,
}: SearchToolbarProps) {
  return (
    <div className="toolbar">
      <div className="toolbar-left">
        <div className="toolbar-group">
          <input
            className="input"
            placeholder={placeholder}
            value={value}
            onChange={(e) => onChange(e.target.value)}
            aria-label={inputLabel}
          />
          {filters}
        </div>
        {typeof resultCount === "number" ? (
          <span className="toolbar-count">{resultCount} itens</span>
        ) : null}
      </div>
      {rightActions ?? actions ? (
        <div className="toolbar-right">{rightActions ?? actions}</div>
      ) : null}
    </div>
  );
}
