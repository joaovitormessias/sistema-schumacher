import type { ReactNode } from "react";

export type StepperStatus = "complete" | "current" | "upcoming";

export type StepperStep = {
  id: string;
  title: string;
  summary?: string;
  status: StepperStatus;
  disabled?: boolean;
  icon?: ReactNode;
};

type StepperProps = {
  steps: StepperStep[];
  onStepChange?: (id: string) => void;
};

export default function Stepper({ steps, onStepChange }: StepperProps) {
  return (
    <div className="stepper" role="tablist" aria-label="Etapas">
      {steps.map((step, index) => {
        const isCurrent = step.status === "current";
        return (
          <button
            key={step.id}
            type="button"
            className={`stepper-item ${step.status}${step.disabled ? " is-disabled" : ""}`}
            onClick={() => {
              if (!step.disabled) {
                onStepChange?.(step.id);
              }
            }}
            disabled={step.disabled}
            aria-current={isCurrent ? "step" : undefined}
          >
            <span className="stepper-index">{index + 1}</span>
            <span className="stepper-content">
              <span className="stepper-title">{step.title}</span>
              {step.summary ? <span className="stepper-summary">{step.summary}</span> : null}
            </span>
          </button>
        );
      })}
    </div>
  );
}
