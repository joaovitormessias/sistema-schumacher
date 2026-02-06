import type { ReactNode } from "react";

type StatCardProps = {
  title: string;
  value: ReactNode;
  helper?: string;
  trend?: ReactNode;
};

export default function StatCard({ title, value, helper, trend }: StatCardProps) {
  return (
    <div className="stat-card">
      <div className="stat-card-header">
        <span className="stat-card-title">{title}</span>
        {trend ? <span className="stat-card-trend">{trend}</span> : null}
      </div>
      <div className="stat-card-value">{value}</div>
      {helper ? <div className="stat-card-helper">{helper}</div> : null}
    </div>
  );
}
