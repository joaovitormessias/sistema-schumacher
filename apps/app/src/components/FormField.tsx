import type { ReactNode } from "react";

type FormFieldProps = {
  label: string;
  hint?: string;
  required?: boolean;
  children: ReactNode;
};

export default function FormField({ label, hint, required, children }: FormFieldProps) {
  return (
    <label className="form-field">
      <span className="form-label">
        {label}
        {required ? <span className="form-required">*</span> : null}
      </span>
      {children}
      {hint ? <span className="form-hint">{hint}</span> : null}
    </label>
  );
}
