import type { ReactNode } from "react";

export type StatusTone = "neutral" | "info" | "success" | "warning" | "danger";

type StatusBadgeProps = {
  tone?: StatusTone;
  children: ReactNode;
};

export default function StatusBadge({ tone = "neutral", children }: StatusBadgeProps) {
  return <span className={`status-badge ${tone}`}>{children}</span>;
}
