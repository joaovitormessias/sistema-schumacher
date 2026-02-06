import type { ReactNode } from "react";

type FormFieldGroupProps = {
  legend: string;
  hint?: string;
  layout?: "grid" | "stack";
  children: ReactNode;
};

export default function FormFieldGroup({
  legend,
  hint,
  layout = "grid",
  children,
}: FormFieldGroupProps) {
  const layoutClass = layout === "grid" ? "form-grid" : "form-stack";

  return (
    <fieldset className="form-fieldset">
      <legend className="form-legend">{legend}</legend>
      {hint ? <div className="form-hint">{hint}</div> : null}
      <div className={layoutClass}>{children}</div>
    </fieldset>
  );
}
