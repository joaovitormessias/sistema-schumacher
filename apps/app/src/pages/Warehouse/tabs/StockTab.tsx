import { useMemo, useState } from "react";
import StatusBadge from "../../../components/StatusBadge";
import { apiGet } from "../../../services/api";
import { formatCurrency } from "../../../utils/financialLabels";
import type { Product } from "../../../hooks/useProducts";

export default function StockTab() {
    const [products, setProducts] = useState<Product[]>([]);
    const [loading, setLoading] = useState(true);
    const [filter, setFilter] = useState<"all" | "low">("all");

    useMemo(() => {
        const load = async () => {
            setLoading(true);
            try {
                const data = await apiGet<Product[]>("/products?limit=500&offset=0&active=true");
                setProducts(data);
            } finally {
                setLoading(false);
            }
        };
        load();
    }, []);

    const filteredProducts = useMemo(() => {
        if (filter === "low") {
            return products.filter((p) => p.current_stock < p.min_stock);
        }
        return products;
    }, [products, filter]);

    const lowStockCount = products.filter((p) => p.current_stock < p.min_stock).length;

    return (
        <div className="stock-tab">
            <div className="section-header" style={{ marginBottom: "1rem" }}>
                <div className="section-title">Posição de Estoque</div>
                <div className="section-actions" style={{ display: "flex", gap: "0.5rem" }}>
                    <button
                        className={`button ${filter === "all" ? "primary" : "ghost"} sm`}
                        onClick={() => setFilter("all")}
                    >
                        Todos ({products.length})
                    </button>
                    <button
                        className={`button ${filter === "low" ? "primary" : "ghost"} sm`}
                        onClick={() => setFilter("low")}
                    >
                        Abaixo do Mínimo ({lowStockCount})
                    </button>
                </div>
            </div>

            {loading ? (
                <div className="loading-state">Carregando...</div>
            ) : filteredProducts.length === 0 ? (
                <div className="empty-state">
                    <h3>Nenhum produto encontrado</h3>
                    <p>{filter === "low" ? "Nenhum produto abaixo do estoque mínimo." : "Cadastre produtos na aba Produtos."}</p>
                </div>
            ) : (
                <table className="table">
                    <thead>
                        <tr>
                            <th>Código</th>
                            <th>Produto</th>
                            <th>Categoria</th>
                            <th>Estoque Atual</th>
                            <th>Estoque Mínimo</th>
                            <th>Último Custo</th>
                            <th>Valor em Estoque</th>
                            <th>Status</th>
                        </tr>
                    </thead>
                    <tbody>
                        {filteredProducts.map((product) => {
                            const isLow = product.current_stock < product.min_stock;
                            const stockValue = product.current_stock * (product.last_cost ?? 0);
                            return (
                                <tr key={product.id}>
                                    <td>{product.code}</td>
                                    <td>{product.name}</td>
                                    <td>{product.category ?? "-"}</td>
                                    <td>
                                        <strong>{product.current_stock}</strong> {product.unit}
                                    </td>
                                    <td>
                                        {product.min_stock} {product.unit}
                                    </td>
                                    <td>{product.last_cost ? formatCurrency(product.last_cost) : "-"}</td>
                                    <td>{stockValue > 0 ? formatCurrency(stockValue) : "-"}</td>
                                    <td>
                                        <StatusBadge tone={isLow ? "danger" : "success"}>
                                            {isLow ? "Baixo" : "OK"}
                                        </StatusBadge>
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            )}

            <style>{`
        .stock-tab .table {
          width: 100%;
          border-collapse: collapse;
        }
        .stock-tab .table th,
        .stock-tab .table td {
          padding: 0.75rem;
          text-align: left;
          border-bottom: 1px solid var(--border-color, #e5e7eb);
        }
        .stock-tab .table th {
          font-weight: 600;
          color: var(--text-secondary, #6b7280);
          font-size: 0.875rem;
        }
        .stock-tab .empty-state {
          text-align: center;
          padding: 3rem;
          color: var(--text-secondary, #6b7280);
        }
        .stock-tab .loading-state {
          text-align: center;
          padding: 3rem;
          color: var(--text-secondary, #6b7280);
        }
      `}</style>
        </div>
    );
}
