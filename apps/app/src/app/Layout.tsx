import { useEffect, useMemo, useState, type ReactNode } from "react";
import { Link, NavLink, useLocation } from "react-router-dom";
import {
  BadgeCheck,
  BarChart3,
  Box,
  Bus,
  ChevronDown,
  CreditCard,
  FileText,
  IdCard,
  LayoutDashboard,
  MapPin,
  Package,
  Receipt,
  Route,
  ShoppingCart,
  Tag,
  Ticket,
  Truck,
  Moon,
  Sun,
  Wallet,
  Wrench,
} from "lucide-react";
import { ToastProvider } from "../components/state/ToastProvider";
import Drawer from "../components/overlay/Drawer";
import ConfirmDialog from "../components/overlay/ConfirmDialog";
import { getSupabaseClient } from "../services/supabase";

const navItems = [
  { path: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { path: "/trips", label: "Viagens", icon: Route },
  { path: "/routes", label: "Rotas", icon: MapPin },
  { path: "/buses", label: "Onibus", icon: Bus },
  { path: "/drivers", label: "Motoristas", icon: IdCard },
  { path: "/bookings", label: "Reservas", icon: Ticket },
  { path: "/payments", label: "Pagamentos", icon: CreditCard },
  { path: "/pricing", label: "Tarifas", icon: Tag },
  { path: "/reports", label: "Relatorios", icon: BarChart3 },
];

const financeItems = [
  { tab: "advances", label: "Adiantamentos", icon: Wallet },
  { tab: "expenses", label: "Despesas", icon: Receipt },
  { tab: "settlements", label: "Acertos", icon: BadgeCheck },
  { tab: "cards", label: "Cartoes", icon: CreditCard },
  { tab: "validations", label: "Validacoes", icon: BadgeCheck },
  { tab: "documents", label: "Documentos Fiscais", icon: FileText },
];

const warehouseItems = [
  { tab: "suppliers", label: "Fornecedores", icon: Truck },
  { tab: "products", label: "Produtos", icon: Package },
  { tab: "service-orders", label: "Ordens de Servico", icon: Wrench },
  { tab: "purchase-orders", label: "Pedidos de Compra", icon: ShoppingCart },
  { tab: "invoices", label: "Notas Fiscais", icon: FileText },
  { tab: "stock", label: "Estoque", icon: Box },
];

const THEME_STORAGE_KEY = "app-theme";
type AppTheme = "light" | "dark";

const SIDEBAR_STORAGE_KEY = "sidebar-collapsed";

export default function Layout({ children }: { children: ReactNode }) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(() => {
    if (typeof window === "undefined") return false;
    const saved = window.localStorage.getItem(SIDEBAR_STORAGE_KEY);
    return saved === "true";
  });
  const [shortcutsOpen, setShortcutsOpen] = useState(false);
  const [signOutOpen, setSignOutOpen] = useState(false);
  const [theme, setTheme] = useState<AppTheme>(() => {
    if (typeof window === "undefined") return "light";
    const saved = window.localStorage.getItem(THEME_STORAGE_KEY);
    return saved === "dark" ? "dark" : "light";
  });
  const [openSections, setOpenSections] = useState({
    financial: true,
    warehouse: true,
  });
  const location = useLocation();

  const activeFinanceTab = useMemo(() => {
    if (location.pathname !== "/financial") return null;
    const tab = new URLSearchParams(location.search).get("tab") ?? "advances";
    return financeItems.find((item) => item.tab === tab) ?? financeItems[0];
  }, [location.pathname, location.search]);

  const activeWarehouseTab = useMemo(() => {
    if (location.pathname !== "/warehouse") return null;
    return new URLSearchParams(location.search).get("tab") ?? "suppliers";
  }, [location.pathname, location.search]);

  const currentItem = useMemo(() => {
    if (activeFinanceTab) {
      return { path: "/financial", label: `Financeiro - ${activeFinanceTab.label}` };
    }
    return navItems.find((item) => location.pathname.startsWith(item.path));
  }, [activeFinanceTab, location.pathname]);

  const quickAction = useMemo(() => {
    const path = currentItem?.path ?? "";
    const map: Record<string, { label: string; to: string }> = {
      "/bookings": { label: "Nova reserva", to: "/bookings#booking-form" },
      "/trips": { label: "Criar viagem", to: "/trips#crud-form" },
      "/routes": { label: "Criar rota", to: "/routes#crud-form" },
      "/buses": { label: "Criar onibus", to: "/buses#crud-form" },
      "/drivers": { label: "Criar motorista", to: "/drivers#crud-form" },
      "/payments": { label: "Registrar pagamento", to: "/payments" },
      "/pricing": { label: "Nova tarifa", to: "/pricing" },
      "/financial": { label: "Abrir financeiro", to: "/financial?tab=advances" },
    };
    return map[path] ?? null;
  }, [currentItem]);

  useEffect(() => {
    document.documentElement.setAttribute("data-theme", theme);
    window.localStorage.setItem(THEME_STORAGE_KEY, theme);
  }, [theme]);

  useEffect(() => {
    window.localStorage.setItem(SIDEBAR_STORAGE_KEY, String(sidebarCollapsed));
  }, [sidebarCollapsed]);

  useEffect(() => {
    if (location.pathname === "/financial") {
      setOpenSections((prev) => (prev.financial ? prev : { ...prev, financial: true }));
    }
    if (location.pathname === "/warehouse") {
      setOpenSections((prev) => (prev.warehouse ? prev : { ...prev, warehouse: true }));
    }
  }, [location.pathname]);

  const handleSignOut = () => {
    setSignOutOpen(true);
  };

  const confirmSignOut = () => {
    const client = getSupabaseClient();
    client?.auth.signOut();
    setSignOutOpen(false);
  };

  const toggleSection = (section: "financial" | "warehouse") => {
    setOpenSections((prev) => ({ ...prev, [section]: !prev[section] }));
  };

  const toggleSidebar = () => {
    setSidebarCollapsed((prev) => !prev);
  };

  return (
    <ToastProvider>
      <a href="#main-content" className="skip-link">
        Pular para o conteudo principal
      </a>
      <div className={`app-shell ${sidebarCollapsed ? "sidebar-collapsed" : ""}`}>
        <div
          className={`sidebar-overlay ${menuOpen ? "open" : ""}`}
          onClick={() => setMenuOpen(false)}
          aria-hidden={!menuOpen}
        />
        <aside id="app-sidebar" className={`sidebar ${menuOpen ? "open" : ""}`} aria-label="Menu principal">
          <div className="sidebar-header">
            <div className="sidebar-brand">
              <div className="brand">Schumacher Turismo</div>
              <div className="page-subtitle">Sistema Interno</div>
            </div>
            <div className="sidebar-actions">
              <button
                className="icon-button"
                type="button"
                aria-label={theme === "dark" ? "Ativar tema claro" : "Ativar tema escuro"}
                onClick={() => setTheme((prev) => (prev === "light" ? "dark" : "light"))}
                title={theme === "dark" ? "Ativar tema claro" : "Ativar tema escuro"}
              >
                {theme === "dark" ? <Sun size={18} /> : <Moon size={18} />}
              </button>
              <button
                className="icon-button"
                type="button"
                onClick={toggleSidebar}
                aria-label={sidebarCollapsed ? "Expandir menu" : "Recolher menu"}
                title={sidebarCollapsed ? "Expandir menu" : "Recolher menu"}
              >
                <ChevronDown
                  size={18}
                  style={{
                    transform: sidebarCollapsed ? "rotate(-90deg)" : "rotate(90deg)",
                    transition: "transform 0.2s ease"
                  }}
                />
              </button>
            </div>
          </div>
          <nav className="nav">
            {navItems.map((item) => (
              <NavLink
                key={item.path}
                to={item.path}
                className={({ isActive }) => (isActive ? "nav-link active" : "nav-link")}
                onClick={() => setMenuOpen(false)}
              >
                <item.icon className="nav-icon" size={18} aria-hidden="true" />
                <span>{item.label}</span>
              </NavLink>
            ))}
            <div className="nav-section">
              <button
                type="button"
                className={`nav-section-toggle ${openSections.financial ? "open" : ""}`}
                onClick={() => toggleSection("financial")}
                aria-expanded={openSections.financial}
                aria-controls="financial-nav-section"
              >
                <span className="nav-section-title">Financeiro</span>
                <ChevronDown size={16} />
              </button>
              <div
                id="financial-nav-section"
                className={`nav-section-content ${openSections.financial ? "open" : ""}`}
              >
                {financeItems.map((item) => (
                  <Link
                    key={item.tab}
                    to={`/financial?tab=${item.tab}`}
                    className={
                      location.pathname === "/financial" && activeFinanceTab?.tab === item.tab
                        ? "nav-link active"
                        : "nav-link"
                    }
                    onClick={() => setMenuOpen(false)}
                  >
                    <item.icon className="nav-icon" size={18} aria-hidden="true" />
                    <span>{item.label}</span>
                  </Link>
                ))}
              </div>
            </div>
            <div className="nav-section">
              <button
                type="button"
                className={`nav-section-toggle ${openSections.warehouse ? "open" : ""}`}
                onClick={() => toggleSection("warehouse")}
                aria-expanded={openSections.warehouse}
                aria-controls="warehouse-nav-section"
              >
                <span className="nav-section-title">Almoxarifado</span>
                <ChevronDown size={16} />
              </button>
              <div
                id="warehouse-nav-section"
                className={`nav-section-content ${openSections.warehouse ? "open" : ""}`}
              >
                {warehouseItems.map((item) => (
                  <Link
                    key={item.tab}
                    to={`/warehouse?tab=${item.tab}`}
                    className={
                      location.pathname === "/warehouse" && activeWarehouseTab === item.tab
                        ? "nav-link active"
                        : "nav-link"
                    }
                    onClick={() => setMenuOpen(false)}
                  >
                    <item.icon className="nav-icon" size={18} aria-hidden="true" />
                    <span>{item.label}</span>
                  </Link>
                ))}
              </div>
            </div>
          </nav>
          <button className="button button-outline" onClick={handleSignOut}>
            Sair
          </button>
        </aside>
        <main id="main-content" className="main" tabIndex={-1}>
          {children}
        </main>
        <button
          className="menu-button-mobile"
          type="button"
          onClick={() => setMenuOpen(true)}
          aria-label="Abrir menu"
        >
          <LayoutDashboard size={24} />
        </button>
      </div>
      <Drawer
        open={shortcutsOpen}
        title="Atalhos rapidos"
        description="Acesse as areas mais usadas do sistema."
        onClose={() => setShortcutsOpen(false)}
      >
        <div className="shortcut-list">
          {navItems.slice(0, 6).map((item) => (
            <Link
              key={item.path}
              className="shortcut-item"
              to={item.path}
              onClick={() => setShortcutsOpen(false)}
            >
              <item.icon size={16} aria-hidden="true" />
              <span>{item.label}</span>
            </Link>
          ))}
          <Link
            className="shortcut-item"
            to="/financial?tab=advances"
            onClick={() => setShortcutsOpen(false)}
          >
            <Wallet size={16} aria-hidden="true" />
            <span>Financeiro</span>
          </Link>
          <Link
            className="shortcut-item"
            to="/warehouse?tab=suppliers"
            onClick={() => setShortcutsOpen(false)}
          >
            <Box size={16} aria-hidden="true" />
            <span>Almoxarifado</span>
          </Link>
        </div>
      </Drawer>
      <ConfirmDialog
        open={signOutOpen}
        title="Encerrar sessao"
        description="Deseja realmente sair do sistema?"
        confirmLabel="Sair"
        tone="danger"
        onCancel={() => setSignOutOpen(false)}
        onConfirm={confirmSignOut}
      />
    </ToastProvider>
  );
}
