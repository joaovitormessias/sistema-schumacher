import { createContext, useContext, useState, useCallback, type ReactNode } from "react";
import { X, CheckCircle, AlertTriangle, AlertCircle, Info } from "lucide-react";

type ToastVariant = "success" | "warning" | "danger" | "info";

interface ToastAction {
    label: string;
    onClick: () => void;
}

interface Toast {
    id: string;
    variant: ToastVariant;
    title: string;
    description?: string;
    actions?: ToastAction[];
    autoClose?: number;
}

interface ToastContextValue {
    toasts: Toast[];
    addToast: (toast: Omit<Toast, "id">) => void;
    removeToast: (id: string) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

export function useToastContext() {
    const context = useContext(ToastContext);
    if (!context) {
        throw new Error("useToastContext must be used within a ToastProvider");
    }
    return context;
}

interface ToastProviderProps {
    children: ReactNode;
}

export function ToastProvider({ children }: ToastProviderProps) {
    const [toasts, setToasts] = useState<Toast[]>([]);

    const addToast = useCallback((toast: Omit<Toast, "id">) => {
        const id = `toast-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
        const newToast: Toast = { ...toast, id };

        setToasts((prev) => [...prev, newToast]);

        if (toast.autoClose !== 0) {
            const duration = toast.autoClose ?? 5000;
            setTimeout(() => {
                setToasts((prev) => prev.filter((t) => t.id !== id));
            }, duration);
        }
    }, []);

    const removeToast = useCallback((id: string) => {
        setToasts((prev) => prev.filter((t) => t.id !== id));
    }, []);

    return (
        <ToastContext.Provider value={{ toasts, addToast, removeToast }}>
            {children}
            <ToastContainer />
        </ToastContext.Provider>
    );
}

function ToastContainer() {
    const { toasts, removeToast } = useToastContext();

    if (toasts.length === 0) return null;

    return (
        <div className="toast-container">
            {toasts.map((toast) => (
                <ToastItem key={toast.id} toast={toast} onClose={() => removeToast(toast.id)} />
            ))}
        </div>
    );
}

interface ToastItemProps {
    toast: Toast;
    onClose: () => void;
}

function ToastItem({ toast, onClose }: ToastItemProps) {
    const IconComponent = {
        success: CheckCircle,
        warning: AlertTriangle,
        danger: AlertCircle,
        info: Info,
    }[toast.variant];

    return (
        <div className={`toast ${toast.variant}`}>
            <IconComponent className={`toast-icon ${toast.variant}`} size={20} />
            <div className="toast-content">
                <div className="toast-title">{toast.title}</div>
                {toast.description && <div className="toast-description">{toast.description}</div>}
                {toast.actions && toast.actions.length > 0 && (
                    <div className="toast-actions">
                        {toast.actions.map((action, idx) => (
                            <button
                                key={idx}
                                className="button sm secondary"
                                onClick={() => {
                                    action.onClick();
                                    onClose();
                                }}
                            >
                                {action.label}
                            </button>
                        ))}
                    </div>
                )}
            </div>
            <button className="toast-close" onClick={onClose} aria-label="Fechar">
                <X size={16} />
            </button>
        </div>
    );
}

// Convenience hook for adding toasts
export function useToast() {
    const { addToast } = useToastContext();

    return {
        success: (title: string, description?: string, actions?: ToastAction[]) =>
            addToast({ variant: "success", title, description, actions }),
        warning: (title: string, description?: string, actions?: ToastAction[]) =>
            addToast({ variant: "warning", title, description, actions }),
        danger: (title: string, description?: string, actions?: ToastAction[]) =>
            addToast({ variant: "danger", title, description, actions }),
        error: (title: string, description?: string, actions?: ToastAction[]) =>
            addToast({ variant: "danger", title, description, actions }),
        info: (title: string, description?: string, actions?: ToastAction[]) =>
            addToast({ variant: "info", title, description, actions }),
        custom: addToast,
    };
}

export default ToastProvider;
