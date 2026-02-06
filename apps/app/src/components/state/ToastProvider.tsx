import { createContext, useCallback, useMemo, useRef, useState, type ReactNode } from "react";

export type ToastTone = "success" | "error" | "info" | "warning";

export type Toast = {
  id: string;
  title?: string;
  message: string;
  tone: ToastTone;
  durationMs?: number;
};

export type ToastOptions = {
  title?: string;
  durationMs?: number;
};

type ToastContextValue = {
  toasts: Toast[];
  success: (message: string, options?: ToastOptions) => void;
  error: (message: string, options?: ToastOptions) => void;
  info: (message: string, options?: ToastOptions) => void;
  warning: (message: string, options?: ToastOptions) => void;
  dismiss: (id: string) => void;
};

export const ToastContext = createContext<ToastContextValue | null>(null);

const DEFAULT_DURATION = 4000;

const createId = () => {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `toast-${Date.now()}-${Math.random().toString(16).slice(2)}`;
};

const toneClass = (tone: ToastTone) => {
  if (tone === "error") return "danger";
  return tone;
};

const toneTitle = (tone: ToastTone) => {
  if (tone === "error") return "Erro";
  if (tone === "warning") return "Atenção";
  if (tone === "info") return "Info";
  return "Sucesso";
};

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const timers = useRef(new Map<string, number>());

  const dismiss = useCallback((id: string) => {
    const timer = timers.current.get(id);
    if (timer) {
      window.clearTimeout(timer);
      timers.current.delete(id);
    }
    setToasts((prev) => prev.filter((toast) => toast.id !== id));
  }, []);

  const enqueue = useCallback(
    (tone: ToastTone, message: string, options?: ToastOptions) => {
      const id = createId();
      const nextToast: Toast = {
        id,
        message,
        tone,
        title: options?.title,
        durationMs: options?.durationMs,
      };
      setToasts((prev) => [nextToast, ...prev]);
      const duration = options?.durationMs ?? DEFAULT_DURATION;
      const timer = window.setTimeout(() => dismiss(id), duration);
      timers.current.set(id, timer);
    },
    [dismiss]
  );

  const value = useMemo(
    () => ({
      toasts,
      success: (message: string, options?: ToastOptions) => enqueue("success", message, options),
      error: (message: string, options?: ToastOptions) => enqueue("error", message, options),
      info: (message: string, options?: ToastOptions) => enqueue("info", message, options),
      warning: (message: string, options?: ToastOptions) => enqueue("warning", message, options),
      dismiss,
    }),
    [dismiss, enqueue, toasts]
  );

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="toast-container" role="region" aria-live="polite" aria-label="Notificacoes">
        {toasts.map((toast) => (
          <div className={`toast ${toneClass(toast.tone)}`} key={toast.id}>
            <div className="toast-header">
            <div className="toast-title">{toast.title ?? toneTitle(toast.tone)}</div>
              <button className="toast-close" type="button" onClick={() => dismiss(toast.id)}>
                Fechar
              </button>
            </div>
            <div className="toast-message">{toast.message}</div>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}
