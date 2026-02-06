import type { CSSProperties, ReactNode } from "react";

export type DataTableColumn<T> = {
  label: string;
  accessor?: (item: T) => ReactNode;
  render?: (item: T) => ReactNode;
  align?: "left" | "center" | "right";
  width?: string;
  hideOnMobile?: boolean;
};

type DataTableProps<T> = {
  columns: DataTableColumn<T>[];
  rows: T[];
  rowKey: (item: T) => string;
  actions?: (item: T) => ReactNode;
  density?: "comfortable" | "compact";
  emptyState?: ReactNode;
  className?: string;
};

export default function DataTable<T>({
  columns,
  rows,
  rowKey,
  actions,
  density = "comfortable",
  emptyState,
  className,
}: DataTableProps<T>) {
  if (rows.length === 0 && emptyState) {
    return <>{emptyState}</>;
  }

  const columnTemplate = columns.map((column) => column.width ?? "minmax(0, 1fr)");
  const template = actions ? [...columnTemplate, "minmax(0, 0.6fr)"] : columnTemplate;
  const style = { "--table-columns": template.join(" ") } as CSSProperties;

  return (
    <div className={`data-table ${density}${className ? ` ${className}` : ""}`} style={style}>
      <div className="data-table-row data-table-head">
        {columns.map((column) => (
          <div
            className={`data-table-cell${column.hideOnMobile ? " hide-mobile" : ""}`}
            key={column.label}
          >
            {column.label}
          </div>
        ))}
        {actions ? <div className="data-table-cell">Ações</div> : null}
      </div>
      {rows.map((item) => (
        <div className="data-table-row" key={rowKey(item)}>
          {columns.map((column) => {
            const content = column.render?.(item) ?? column.accessor?.(item) ?? "-";
            const justify =
              column.align === "center"
                ? "center"
                : column.align === "right"
                  ? "flex-end"
                  : "flex-start";
            return (
              <div
                className={`data-table-cell${column.hideOnMobile ? " hide-mobile" : ""}`}
                data-label={column.label}
                key={`${rowKey(item)}-${column.label}`}
                style={column.align ? { justifyContent: justify } : undefined}
              >
                {content}
              </div>
            );
          })}
          {actions ? (
            <div className="data-table-cell" data-label="Ações">
              <div className="data-table-actions">{actions(item)}</div>
            </div>
          ) : null}
        </div>
      ))}
    </div>
  );
}
