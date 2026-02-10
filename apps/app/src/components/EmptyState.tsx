import type { ReactNode } from "react";
import { Package } from "lucide-react";

type EmptyStateProps = {
  title: string;
  description?: string;
  action?: ReactNode;
  secondaryAction?: ReactNode;
  illustration?: ReactNode;
};

export default function EmptyState({
  title,
  description,
  action,
  secondaryAction,
  illustration,
}: EmptyStateProps) {
  return (
    <div className="empty-state">
      {illustration ? (
        <div className="empty-state-illustration">{illustration}</div>
      ) : (
        <div className="empty-state-illustration">
          <Package size={64} strokeWidth={1.5} />
        </div>
      )}
      <div className="empty-state-title">{title}</div>
      {description ? <div className="empty-state-description">{description}</div> : null}
      {(action || secondaryAction) && (
        <div className="empty-state-actions">
          {action}
          {secondaryAction}
        </div>
      )}
    </div>
  );
}
