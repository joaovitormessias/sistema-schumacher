import { useState, type FormEvent } from "react";
import InlineAlert from "../../components/InlineAlert";
import PageHeader from "../../components/PageHeader";
import { getSupabaseClient, isSupabaseConfigured } from "../../services/supabase";

export default function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!isSupabaseConfigured()) {
      setError("Supabase não configurado. Verifique VITE_SUPABASE_URL e VITE_SUPABASE_ANON_KEY.");
      return;
    }

    try {
      setLoading(true);
      const client = getSupabaseClient();
      if (!client) {
        setError("Supabase não configurado.");
        return;
      }
      const { error: signInError } = await client.auth.signInWithPassword({
        email,
        password,
      });
      if (signInError) {
        setError(signInError.message);
      }
    } catch (err: any) {
      setError(err.message || "Erro ao autenticar");
    } finally {
      setLoading(false);
    }
  };

  return (
    <section className="page">
      <div className="login-layout">
        <div className="login-card">
          <PageHeader
            title="Acesso ao Sistema"
            subtitle="Entre com seu e-mail e senha para continuar."
            eyebrow="Schumacher Turismo"
          />

          <form className="form-grid" onSubmit={handleSubmit}>
            <label className="form-field">
              <span className="form-label">E-mail</span>
              <input
                className="input"
                type="email"
                placeholder="seu@email.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </label>
            <label className="form-field">
              <span className="form-label">Senha</span>
              <input
                className="input"
                type="password"
                placeholder="Sua senha"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
              <span className="form-hint">Use as credenciais cadastradas pelo administrador.</span>
            </label>
            <div className="full-span">
              <button className="button" type="submit" disabled={loading}>
                {loading ? "Entrando..." : "Entrar"}
              </button>
            </div>
          </form>

          {error ? <InlineAlert tone="error">{error}</InlineAlert> : null}
        </div>
      </div>
    </section>
  );
}
