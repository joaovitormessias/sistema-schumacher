import { ArchiveX, Pencil, RotateCcw } from "lucide-react";

type TableActionButtonsProps = {
  isActive?: boolean;
  onEdit?: () => void;
  onDeactivate?: () => void;
  onRestore?: () => void;
  disableEdit?: boolean;
  labels?: {
    edit?: string;
    deactivate?: string;
    restore?: string;
  };
};

export default function TableActionButtons({
  isActive = true,
  onEdit,
  onDeactivate,
  onRestore,
  disableEdit,
  labels,
}: TableActionButtonsProps) {
  const editLabel = labels?.edit ?? "Editar";
  const deactivateLabel = labels?.deactivate ?? "Desativar";
  const restoreLabel = labels?.restore ?? "Reativar";

  return (
    <>
      {onEdit ? (
        <button
          className="icon-button"
          type="button"
          onClick={onEdit}
          disabled={disableEdit}
          aria-label={editLabel}
          title={editLabel}
        >
          <Pencil size={16} aria-hidden="true" />
        </button>
      ) : null}
      {isActive ? (
        onDeactivate ? (
          <button
            className="icon-button"
            type="button"
            onClick={() => {
              if (window.confirm("Deseja desativar este item?")) {
                onDeactivate();
              }
            }}
            aria-label={deactivateLabel}
            title={deactivateLabel}
          >
            <ArchiveX size={16} aria-hidden="true" />
          </button>
        ) : null
      ) : onRestore ? (
        <button
          className="icon-button"
          type="button"
          onClick={() => {
            if (window.confirm("Deseja reativar este item?")) {
              onRestore();
            }
          }}
          aria-label={restoreLabel}
          title={restoreLabel}
        >
          <RotateCcw size={16} aria-hidden="true" />
        </button>
      ) : null}
    </>
  );
}
