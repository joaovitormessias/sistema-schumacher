import type { CSSProperties, ReactNode } from "react";
import { List, type RowComponentProps } from "react-window";
import type { DataTableColumn } from "../table/DataTable";

type VirtualDataTableProps<T> = {
  columns: DataTableColumn<T>[];
  rows: T[];
  rowKey: (item: T) => string;
  actions?: (item: T) => ReactNode;
  density?: "comfortable" | "compact";
  emptyState?: ReactNode;
  className?: string;
  maxHeight?: number;
};

const ROW_GAP = 8;

export default function VirtualDataTable<T>({
  columns,
  rows,
  rowKey,
  actions,
  density = "comfortable",
  emptyState,
  className,
  maxHeight = 520,
}: VirtualDataTableProps<T>) {
  if (rows.length === 0 && emptyState) {
    return <>{emptyState}</>;
  }

  const rowHeight = density === "compact" ? 48 : 64;
  const listRowHeight = rowHeight + ROW_GAP;
  const listHeight = Math.min(maxHeight, Math.max(listRowHeight, rows.length * listRowHeight));
  const columnTemplate = columns.map((column) => column.width ?? "minmax(0, 1fr)");
  const template = actions ? [...columnTemplate, "minmax(0, 0.6fr)"] : columnTemplate;
  const style = { "--table-columns": template.join(" ") } as CSSProperties;

  const Row = ({ index, style: itemStyle, ariaAttributes }: RowComponentProps<Record<string, never>>) => {
    const item = rows[index];
    const id = rowKey(item);

    return (
      <div
        {...ariaAttributes}
        className="virtual-data-table-item"
        style={itemStyle}
      >
        <div className="data-table-row virtual-data-table-row" style={{ height: rowHeight }}>
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
                key={`${id}-${column.label}`}
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
      </div>
    );
  };

  return (
    <div
      className={`data-table virtual-data-table ${density}${className ? ` ${className}` : ""}`}
      style={style}
    >
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
      <List
        className="virtual-data-table-list"
        style={{ height: listHeight, width: "100%" }}
        rowComponent={Row}
        rowCount={rows.length}
        rowHeight={listRowHeight}
        rowProps={{}}
        overscanCount={8}
      />
    </div>
  );
}
