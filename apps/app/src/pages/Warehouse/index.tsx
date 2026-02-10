import { Box, FileText, Package, ShoppingCart, Truck, Wrench } from "lucide-react";
import { Tabs } from "antd";
import { useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import PageHeader from "../../components/PageHeader";
import SuppliersTab from "./tabs/SuppliersTab";
import ProductsTab from "./tabs/ProductsTab";
import ServiceOrdersTab from "./tabs/ServiceOrdersTab";
import PurchaseOrdersTab from "./tabs/PurchaseOrdersTab";
import InvoicesTab from "./tabs/InvoicesTab";
import StockTab from "./tabs/StockTab";

type WarehouseTab = "suppliers" | "products" | "service-orders" | "purchase-orders" | "invoices" | "stock";

const tabOrder: WarehouseTab[] = ["suppliers", "products", "service-orders", "purchase-orders", "invoices", "stock"];
const defaultTab: WarehouseTab = "suppliers";

function isWarehouseTab(value: string | null): value is WarehouseTab {
    if (!value) return false;
    return tabOrder.includes(value as WarehouseTab);
}

export default function Warehouse() {
    const [searchParams, setSearchParams] = useSearchParams();
    const rawTab = searchParams.get("tab");
    const activeTab: WarehouseTab = isWarehouseTab(rawTab) ? rawTab : defaultTab;

    const items = useMemo(
        () => [
            {
                key: "suppliers",
                label: (
                    <span className="ant-tab-label">
                        <Truck size={16} />
                        <span>Fornecedores</span>
                    </span>
                ),
                children: <SuppliersTab />,
            },
            {
                key: "products",
                label: (
                    <span className="ant-tab-label">
                        <Package size={16} />
                        <span>Produtos</span>
                    </span>
                ),
                children: <ProductsTab />,
            },
            {
                key: "service-orders",
                label: (
                    <span className="ant-tab-label">
                        <Wrench size={16} />
                        <span>Ordens de Serviço</span>
                    </span>
                ),
                children: <ServiceOrdersTab />,
            },
            {
                key: "purchase-orders",
                label: (
                    <span className="ant-tab-label">
                        <ShoppingCart size={16} />
                        <span>Pedidos de Compra</span>
                    </span>
                ),
                children: <PurchaseOrdersTab />,
            },
            {
                key: "invoices",
                label: (
                    <span className="ant-tab-label">
                        <FileText size={16} />
                        <span>Notas Fiscais</span>
                    </span>
                ),
                children: <InvoicesTab />,
            },
            {
                key: "stock",
                label: (
                    <span className="ant-tab-label">
                        <Box size={16} />
                        <span>Estoque</span>
                    </span>
                ),
                children: <StockTab />,
            },
        ],
        []
    );

    return (
        <section className="page">
            <PageHeader
                title="Almoxarifado"
                subtitle="Gestão de compras, estoque e manutenção."
                meta={<span className="badge">Consolidado</span>}
            />

            <div className="section">
                <div className="section-header">
                    <div className="section-title">Módulos de almoxarifado</div>
                </div>

                <Tabs
                    activeKey={activeTab}
                    items={items}
                    onChange={(nextKey) => {
                        if (!isWarehouseTab(nextKey)) return;
                        setSearchParams({ tab: nextKey });
                    }}
                />
            </div>
        </section>
    );
}
