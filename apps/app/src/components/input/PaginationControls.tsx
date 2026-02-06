type PaginationControlsProps = {
  page: number;
  pageSize: number;
  itemCount: number;
  disabled?: boolean;
  onPageChange: (nextPage: number) => void;
};

export default function PaginationControls({
  page,
  pageSize,
  itemCount,
  disabled,
  onPageChange,
}: PaginationControlsProps) {
  const hasNext = itemCount >= pageSize;
  const canPrev = page > 0 && !disabled;
  const canNext = hasNext && !disabled;

  return (
    <div className="pagination" aria-label="Paginação">
      <button
        className="button secondary"
        type="button"
        onClick={() => onPageChange(Math.max(page - 1, 0))}
        disabled={!canPrev}
      >
        Anterior
      </button>
      <span>Página {page + 1}</span>
      <button
        className="button secondary"
        type="button"
        onClick={() => onPageChange(page + 1)}
        disabled={!canNext}
      >
        Próxima
      </button>
    </div>
  );
}
