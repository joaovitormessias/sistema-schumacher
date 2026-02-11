import type { ReactNode } from "react";

export type StatusTone = "neutral" | "info" | "success" | "warning" | "danger";

type StatusBadgeProps = {
  tone?: StatusTone;
  icon?: ReactNode;
  label?: string;
  children?: ReactNode;
};

export default function StatusBadge({
  tone = "neutral",
  icon,
  label,
  children
}: StatusBadgeProps) {
  const content = label ?? children;

  return (
    <span className={`status-badge ${tone}`}>
      {icon && <span className="status-badge-icon">{icon}</span>}
      {content}
    </span>
  );
}
