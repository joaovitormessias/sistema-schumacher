import { type ButtonHTMLAttributes, type ReactNode } from "react";
import { Loader2 } from "lucide-react";

interface LoadingButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
    loading?: boolean;
    loadingText?: string;
    children: ReactNode;
    variant?: "primary" | "secondary" | "danger";
    size?: "sm" | "md" | "lg";
}

export function LoadingButton({
    loading = false,
    loadingText,
    children,
    variant = "primary",
    size = "md",
    className = "",
    disabled,
    ...props
}: LoadingButtonProps) {
    const variantClass = variant === "primary" ? "" : variant;
    const sizeClass = size === "md" ? "" : size;

    return (
        <button
            className={`button button-delightful ${variantClass} ${sizeClass} ${loading ? "loading" : ""} ${className}`.trim()}
            disabled={disabled || loading}
            {...props}
        >
            {loading ? (
                <>
                    <Loader2 size={16} className="spin" style={{ marginRight: 8 }} />
                    {loadingText || children}
                </>
            ) : (
                children
            )}
        </button>
    );
}

export default LoadingButton;
