import type { ReactNode } from "react";

type PageHeaderProps = {
  title: string;
  subtitle?: string;
  eyebrow?: string;
  meta?: ReactNode;
  primaryAction?: ReactNode;
  secondaryActions?: ReactNode;
};

export default function PageHeader({
  title,
  subtitle,
  eyebrow,
  meta,
  primaryAction,
  secondaryActions,
}: PageHeaderProps) {
  return (
    <div className="page-header">
      <div className="page-title-group">
        {eyebrow ? <div className="page-eyebrow">{eyebrow}</div> : null}
        <div className="page-title">{title}</div>
        {subtitle ? <div className="page-subtitle">{subtitle}</div> : null}
      </div>
      <div className="page-header-actions">
        {meta ? <div className="page-header-meta">{meta}</div> : null}
        {primaryAction || secondaryActions ? (
          <div className="page-actions">
            {secondaryActions}
            {primaryAction}
          </div>
        ) : null}
      </div>
    </div>
  );
}
