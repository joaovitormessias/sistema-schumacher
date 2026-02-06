import { useEffect, useState, type ReactNode } from "react";
import type { Session } from "@supabase/supabase-js";
import InlineAlert from "../components/InlineAlert";
import LoadingState from "../components/LoadingState";
import Login from "../pages/Login";
import { getSupabaseClient, isSupabaseConfigured } from "../services/supabase";

export default function AuthGate({ children }: { children: ReactNode }) {
  const devMode = import.meta.env.VITE_DEV_MODE === "true";
  const [session, setSession] = useState<Session | null>(null);
  const [loading, setLoading] = useState(true);

  if (devMode) {
    return <>{children}</>;
  }

  useEffect(() => {
    let active = true;
    const client = getSupabaseClient();
    if (!client) {
      setLoading(false);
      return () => {};
    }
    client.auth.getSession().then(({ data }) => {
      if (active) {
        setSession(data.session);
        setLoading(false);
      }
    });
    const { data } = client.auth.onAuthStateChange((_event, newSession) => {
      setSession(newSession);
      setLoading(false);
    });
    return () => {
      active = false;
      data.subscription.unsubscribe();
    };
  }, []);

  if (!isSupabaseConfigured()) {
    return (
      <section className="page">
        <div className="page-header">
          <div>
            <div className="page-title">Supabase não configurado</div>
            <div className="page-subtitle">
              Defina VITE_SUPABASE_URL e VITE_SUPABASE_ANON_KEY no .env.
            </div>
          </div>
        </div>
        <InlineAlert tone="warning">Configure o ambiente para liberar o login.</InlineAlert>
      </section>
    );
  }

  if (loading) {
    return (
      <section className="page">
        <LoadingState label="Validando sessão..." />
      </section>
    );
  }

  if (!session) {
    return <Login />;
  }

  return <>{children}</>;
}
