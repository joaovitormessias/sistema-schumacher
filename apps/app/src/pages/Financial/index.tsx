import { BadgeCheck, CreditCard, FileText, Receipt, ShieldCheck, Wallet } from "lucide-react";
import { Tabs } from "antd";
import { useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import PageHeader from "../../components/PageHeader";
import { useRoutes } from "../../hooks/useRoutes";
import { useTrips } from "../../hooks/useTrips";
import { formatDateTime, formatShortId } from "../../utils/format";
import {
  FinancialFiltersProvider,
  useFinancialFiltersOptional,
} from "./FinancialContext";
import DriverCardsTab from "./tabs/DriverCardsTab";
import FiscalDocumentsTab from "./tabs/FiscalDocumentsTab";
import TripAdvancesTab from "./tabs/TripAdvancesTab";
import TripExpensesTab from "./tabs/TripExpensesTab";
import TripSettlementsTab from "./tabs/TripSettlementsTab";
import TripValidationsTab from "./tabs/TripValidationsTab";

type FinancialTab =
  | "advances"
  | "expenses"
  | "settlements"
  | "cards"
  | "validations"
  | "documents";

type RouteItem = { id: string; origin_city: string; destination_city: string };
type TripItem = { id: string; route_id: string; departure_at: string };

const tabOrder: FinancialTab[] = [
  "advances",
  "expenses",
  "settlements",
  "cards",
  "validations",
  "documents",
];
const defaultTab: FinancialTab = "advances";

function isFinancialTab(value: string | null): value is FinancialTab {
  if (!value) return false;
  return tabOrder.includes(value as FinancialTab);
}

function FinancialContent() {
  const [searchParams, setSearchParams] = useSearchParams();
  const financialFilters = useFinancialFiltersOptional();
  const rawTab = searchParams.get("tab");
  const activeTab: FinancialTab = isFinancialTab(rawTab) ? rawTab : defaultTab;
  const trips = (useTrips(300, 0).data as TripItem[] | undefined) ?? [];
  const routes = (useRoutes(300, 0).data as RouteItem[] | undefined) ?? [];
  const routeMap = useMemo(() => new Map(routes.map((route) => [route.id, route])), [routes]);

  const items = useMemo(
    () => [
      {
        key: "advances",
        label: (
          <span className="ant-tab-label">
            <Wallet size={16} />
            <span>Adiantamentos</span>
          </span>
        ),
        children: <TripAdvancesTab />,
      },
      {
        key: "expenses",
        label: (
          <span className="ant-tab-label">
            <Receipt size={16} />
            <span>Despesas</span>
          </span>
        ),
        children: <TripExpensesTab />,
      },
      {
        key: "settlements",
        label: (
          <span className="ant-tab-label">
            <BadgeCheck size={16} />
            <span>Acertos</span>
          </span>
        ),
        children: <TripSettlementsTab />,
      },
      {
        key: "cards",
        label: (
          <span className="ant-tab-label">
            <CreditCard size={16} />
            <span>Cartoes</span>
          </span>
        ),
        children: <DriverCardsTab />,
      },
      {
        key: "validations",
        label: (
          <span className="ant-tab-label">
            <ShieldCheck size={16} />
            <span>Validacoes</span>
          </span>
        ),
        children: <TripValidationsTab />,
      },
      {
        key: "documents",
        label: (
          <span className="ant-tab-label">
            <FileText size={16} />
            <span>Documentos</span>
          </span>
        ),
        children: <FiscalDocumentsTab />,
      },
    ],
    []
  );

  const tabCards = useMemo(
    () => [
      {
        key: "advances",
        label: "Adiantamentos",
        description: "Lancamento e entrega ao motorista.",
        icon: Wallet,
      },
      {
        key: "expenses",
        label: "Despesas",
        description: "Conferencia e aprovacao de gastos.",
        icon: Receipt,
      },
      {
        key: "settlements",
        label: "Acertos",
        description: "Fechamento final por viagem.",
        icon: BadgeCheck,
      },
      {
        key: "cards",
        label: "Cartoes",
        description: "Saldo e ajustes dos cartoes.",
        icon: CreditCard,
      },
      {
        key: "validations",
        label: "Validacoes",
        description: "Conferencias operacionais da viagem.",
        icon: ShieldCheck,
      },
      {
        key: "documents",
        label: "Documentos",
        description: "Notas e comprovantes fiscais.",
        icon: FileText,
      },
    ],
    []
  );

  const tripOptions = useMemo(() => {
    return trips
      .slice()
      .sort((a, b) => new Date(b.departure_at).getTime() - new Date(a.departure_at).getTime())
      .map((trip) => {
        const route = routeMap.get(trip.route_id);
        const routeLabel = route
          ? `${route.origin_city} -> ${route.destination_city}`
          : formatShortId(trip.route_id);
        return {
          value: trip.id,
          label: `${routeLabel} - ${formatDateTime(trip.departure_at)}`,
        };
      });
  }, [routeMap, trips]);

  return (
    <section className="page">
      <PageHeader
        title="Financeiro"
        subtitle="Operacoes financeiras consolidadas em abas."
        meta={<span className="badge">Consolidado</span>}
      />

      <div className="card-grid financial-cards-grid">
        {tabCards.map((card) => {
          const Icon = card.icon;
          const selected = activeTab === card.key;
          return (
            <button
              key={card.key}
              type="button"
              className={`financial-tab-card ${selected ? "is-active" : ""}`}
              onClick={() => setSearchParams({ tab: card.key })}
            >
              <span className="financial-tab-card-title">
                <Icon size={16} />
                {card.label}
              </span>
              <span className="financial-tab-card-description">{card.description}</span>
            </button>
          );
        })}
      </div>

      <div className="section">
        <div className="section-header">
          <div className="section-title">Modulos financeiros</div>
        </div>

        <div className="toolbar">
          <div className="toolbar-left">
            <div className="toolbar-group">
              <select
                className="input"
                value={financialFilters?.tripFilter ?? ""}
                onChange={(e) => financialFilters?.setTripFilter(e.target.value)}
                aria-label="Filtrar por viagem"
              >
                <option value="">Todas as viagens</option>
                {tripOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>
          </div>
          {financialFilters?.tripFilter ? (
            <div className="toolbar-right">
              <button className="button ghost sm" type="button" onClick={financialFilters.clearFilters}>
                Limpar filtro
              </button>
            </div>
          ) : null}
        </div>

        <div className="financial-tabs-shell">
          <Tabs
            activeKey={activeTab}
            items={items}
            animated={{ inkBar: true, tabPane: true }}
            onChange={(nextKey) => {
              if (!isFinancialTab(nextKey)) return;
              setSearchParams({ tab: nextKey });
            }}
          />
        </div>
      </div>
    </section>
  );
}

export default function Financial() {
  return (
    <FinancialFiltersProvider>
      <FinancialContent />
    </FinancialFiltersProvider>
  );
}
