import type { PropsWithChildren } from "react";

type FadeTransitionProps = PropsWithChildren<{
  show: boolean;
  className?: string;
}>;

export default function FadeTransition({ show, className, children }: FadeTransitionProps) {
  return (
    <div className={`fade-transition ${show ? "show" : "hide"}${className ? ` ${className}` : ""}`}>
      {children}
    </div>
  );
}
