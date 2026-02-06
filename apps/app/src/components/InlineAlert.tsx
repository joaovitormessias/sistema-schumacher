import type { ReactNode } from "react";

type AlertTone = "error" | "success" | "info" | "warning";

type InlineAlertProps = {
  tone?: AlertTone;
  title?: string;
  children: ReactNode;
};

export default function InlineAlert({ tone = "info", title, children }: InlineAlertProps) {
  return (
    <div className={`alert ${tone}`} role={tone === "error" ? "alert" : "status"}>
      {title ? <div className="alert-title">{title}</div> : null}
      <div className="alert-body">{children}</div>
    </div>
  );
}
