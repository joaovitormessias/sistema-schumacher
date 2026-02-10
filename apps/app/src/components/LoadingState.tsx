import { SkeletonCircle } from "./feedback/SkeletonLoader";

type LoadingStateProps = {
  label?: string;
};

export default function LoadingState({ label = "Carregando..." }: LoadingStateProps) {
  return (
    <div className="loading-state" role="status" aria-live="polite">
      <SkeletonCircle size={16} />
      <span>{label}</span>
    </div>
  );
}
