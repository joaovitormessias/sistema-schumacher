type LoadingStateProps = {
  label?: string;
};

export default function LoadingState({ label = "Carregando..." }: LoadingStateProps) {
  return (
    <div className="loading-state" role="status" aria-live="polite">
      <div className="loading-spinner" />
      <span>{label}</span>
    </div>
  );
}
