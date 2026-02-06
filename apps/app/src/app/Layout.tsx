import { useMemo, useState, type ReactNode } from "react";
import { Link, NavLink, useLocation } from "react-router-dom";
import {
  BadgeCheck,
  BarChart3,
  Bus,
  CreditCard,
  FileText,
  IdCard,
  LayoutDashboard,
  MapPin,
  Route,
  Tag,
  Ticket,
  Wallet,
  Receipt,
} from "lucide-react";
import { getSupabaseClient } from "../services/supabase";
import { ToastProvider } from "../components/state/ToastProvider";

const navItems = [
  { path: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { path: "/trips", label: "Viagens", icon: Route },
  { path: "/routes", label: "Rotas", icon: MapPin },
  { path: "/buses", label: "Ônibus", icon: Bus },
  { path: "/drivers", label: "Motoristas", icon: IdCard },
  { path: "/bookings", label: "Reservas", icon: Ticket },
  { path: "/payments", label: "Pagamentos", icon: CreditCard },
  { path: "/pricing", label: "Tarifas", icon: Tag },
  { path: "/reports", label: "Relatórios", icon: BarChart3 },
];

const financeItems = [
  { path: "/trip-advances", label: "Adiantamentos", icon: Wallet },
  { path: "/trip-expenses", label: "Despesas", icon: Receipt },
  { path: "/trip-settlements", label: "Acertos", icon: BadgeCheck },
  { path: "/driver-cards", label: "Cartões", icon: CreditCard },
  { path: "/trip-validations", label: "Validações", icon: BadgeCheck },
  { path: "/fiscal-documents", label: "Documentos Fiscais", icon: FileText },
];

export default function Layout({ children }: { children: ReactNode }) {
  const [menuOpen, setMenuOpen] = useState(false);
  const location = useLocation();
  const allItems = useMemo(() => [...navItems, ...financeItems], []);
  const currentItem = allItems.find((item) => location.pathname.startsWith(item.path));
  const quickAction = useMemo(() => {
    const path = currentItem?.path ?? "";
    const map: Record<string, { label: string; to: string }> = {
      "/bookings": { label: "Nova reserva", to: "/bookings#booking-form" },
      "/trips": { label: "Criar viagem", to: "/trips#crud-form" },
      "/routes": { label: "Criar rota", to: "/routes#crud-form" },
      "/buses": { label: "Criar Ônibus", to: "/buses#crud-form" },
      "/drivers": { label: "Criar motorista", to: "/drivers#crud-form" },
      "/payments": { label: "Registrar pagamento", to: "/payments" },
      "/pricing": { label: "Nova tarifa", to: "/pricing" },
    };
    return map[path] ?? null;
  }, [currentItem]);

  const handleSignOut = () => {
    const client = getSupabaseClient();
    client?.auth.signOut();
  };

  return (
    <ToastProvider>
      <div className="app-shell">
        <header className="topbar">
          <div className="topbar-left">
            <button
              className="menu-button"
              type="button"
              onClick={() => setMenuOpen(true)}
              aria-label="Abrir menu"
            >
              <span />
              <span />
              <span />
            </button>
            <div className="topbar-title">
              <span className="topbar-eyebrow">Operação</span>
              <span className="topbar-module">{currentItem?.label ?? "Painel"}</span>
            </div>
          </div>
          <div className="topbar-actions">
            {quickAction ? (
              <Link className="button sm" to={quickAction.to}>
                {quickAction.label}
              </Link>
            ) : null}
          </div>
        </header>
        <div
          className={`sidebar-overlay ${menuOpen ? "open" : ""}`}
          onClick={() => setMenuOpen(false)}
          aria-hidden={!menuOpen}
        />
        <aside className={`sidebar ${menuOpen ? "open" : ""}`} aria-label="Menu principal">
          <div>
            <div className="brand">Schumacher Turismo</div>
            <div className="page-subtitle">Sistema Interno</div>
          </div>
          <nav className="nav">
            {navItems.map((item) => (
              <NavLink
                key={item.path}
                to={item.path}
                className={({ isActive }) =>
                  isActive ? "nav-link active" : "nav-link"
                }
                onClick={() => setMenuOpen(false)}
              >
                <item.icon className="nav-icon" size={18} aria-hidden="true" />
                <span>{item.label}</span>
              </NavLink>
            ))}
            <div className="nav-section">
              <div className="nav-section-title">Financeiro</div>
              {financeItems.map((item) => (
                <NavLink
                  key={item.path}
                  to={item.path}
                  className={({ isActive }) =>
                    isActive ? "nav-link active" : "nav-link"
                  }
                  onClick={() => setMenuOpen(false)}
                >
                  <item.icon className="nav-icon" size={18} aria-hidden="true" />
                  <span>{item.label}</span>
                </NavLink>
              ))}
            </div>
          </nav>
          <button className="button button-outline" onClick={handleSignOut}>
            Sair
          </button>
        </aside>
        <main className="main">{children}</main>
      </div>
    </ToastProvider>
  );
}
