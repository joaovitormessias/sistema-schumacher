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
  const enabledSteps = steps.filter((step) => !step.disabled);

  const focusStep = (stepId: string) => {
    const element = document.getElementById(`stepper-tab-${stepId}`);
    element?.focus();
  };

  const moveFocus = (currentId: string, direction: 1 | -1) => {
    if (enabledSteps.length === 0) return;
    const index = enabledSteps.findIndex((step) => step.id === currentId);
    if (index < 0) return;
    const nextIndex = (index + direction + enabledSteps.length) % enabledSteps.length;
    const nextStep = enabledSteps[nextIndex];
    onStepChange?.(nextStep.id);
    focusStep(nextStep.id);
  };

  return (
    <div className="stepper" role="tablist" aria-label="Etapas">
      {steps.map((step, index) => {
        const isCurrent = step.status === "current";
        return (
          <button
            key={step.id}
            type="button"
            className={`stepper-item ${step.status}${step.disabled ? " is-disabled" : ""}`}
            id={`stepper-tab-${step.id}`}
            role="tab"
            aria-selected={isCurrent}
            aria-controls={`stepper-panel-${step.id}`}
            onClick={() => {
              if (!step.disabled) {
                onStepChange?.(step.id);
              }
            }}
            onKeyDown={(event) => {
              if (step.disabled) return;
              if (event.key === "ArrowRight") {
                event.preventDefault();
                moveFocus(step.id, 1);
              } else if (event.key === "ArrowLeft") {
                event.preventDefault();
                moveFocus(step.id, -1);
              } else if (event.key === "Home") {
                event.preventDefault();
                const first = enabledSteps[0];
                if (!first) return;
                onStepChange?.(first.id);
                focusStep(first.id);
              } else if (event.key === "End") {
                event.preventDefault();
                const last = enabledSteps[enabledSteps.length - 1];
                if (!last) return;
                onStepChange?.(last.id);
                focusStep(last.id);
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
