import React from "react";
import "./success-animation.css";

interface SuccessAnimationProps {
  message?: string;
  size?: "sm" | "md" | "lg";
  autoHideDelay?: number | null;
  onDismiss?: () => void;
}

export default function SuccessAnimation({
  message = "Sucesso!",
  size = "md",
  autoHideDelay = 3000,
  onDismiss,
}: SuccessAnimationProps) {
  const [isVisible, setIsVisible] = React.useState(true);

  React.useEffect(() => {
    if (autoHideDelay === null) return;

    const timer = setTimeout(() => {
      setIsVisible(false);
      onDismiss?.();
    }, autoHideDelay);

    return () => clearTimeout(timer);
  }, [autoHideDelay, onDismiss]);

  if (!isVisible) return null;

  const sizeClasses = {
    sm: "success-animation-sm",
    md: "success-animation-md",
    lg: "success-animation-lg",
  };

  return (
    <div className={`success-animation ${sizeClasses[size]}`}>
      <div className="success-checkmark-wrapper">
        <svg
          className="success-checkmark"
          viewBox="0 0 100 100"
          xmlns="http://www.w3.org/2000/svg"
        >
          <circle
            cx="50"
            cy="50"
            r="45"
            fill="none"
            stroke="var(--success, #5B8C5A)"
            strokeWidth="3"
          />
          <path
            d="M 30 55 L 45 70 L 70 40"
            fill="none"
            stroke="var(--success, #5B8C5A)"
            strokeWidth="4"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </div>
      {message && <p className="success-message">{message}</p>}
    </div>
  );
}
