import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiGet, apiPost, apiPatch, apiDelete } from "../services/api";

export type PurchaseOrderStatus = "DRAFT" | "SENT" | "PARTIAL" | "RECEIVED" | "CANCELLED";

export type PurchaseOrderItem = {
    id: string;
    purchase_order_id: string;
    product_id: string;
    product_name?: string;
    product_code?: string;
    quantity: number;
    unit_price: number;
    discount: number;
    total: number;
    received_quantity: number;
    created_at: string;
};

export type PurchaseOrder = {
    id: string;
    order_number: number;
    service_order_id?: string;
    supplier_id: string;
    supplier_name?: string;
    status: PurchaseOrderStatus;
    order_date: string;
    expected_delivery?: string;
    own_delivery: boolean;
    subtotal: number;
    discount: number;
    freight: number;
    total: number;
    notes?: string;
    created_at: string;
    updated_at: string;
    items?: PurchaseOrderItem[];
};

export type CreatePurchaseOrderInput = {
    service_order_id?: string;
    supplier_id: string;
    expected_delivery?: string;
    own_delivery?: boolean;
    discount?: number;
    freight?: number;
    notes?: string;
    items: { product_id: string; quantity: number; unit_price: number; discount?: number }[];
};

export function usePurchaseOrders(limit = 200, offset = 0, status?: PurchaseOrderStatus, supplier_id?: string, service_order_id?: string) {
    let url = `/purchase-orders?limit=${limit}&offset=${offset}`;
    if (status) url += `&status=${status}`;
    if (supplier_id) url += `&supplier_id=${supplier_id}`;
    if (service_order_id) url += `&service_order_id=${service_order_id}`;

    return useQuery({
        queryKey: ["purchase-orders", limit, offset, status, supplier_id, service_order_id],
        queryFn: () => apiGet<PurchaseOrder[]>(url),
    });
}

export function usePurchaseOrder(id: string) {
    return useQuery({
        queryKey: ["purchase-order", id],
        queryFn: () => apiGet<PurchaseOrder>(`/purchase-orders/${id}`),
        enabled: !!id,
    });
}

export function useCreatePurchaseOrder() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: CreatePurchaseOrderInput) => apiPost<PurchaseOrder>("/purchase-orders", data),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["purchase-orders"] }),
    });
}

export function useUpdatePurchaseOrder() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ id, data }: { id: string; data: { expected_delivery?: string; own_delivery?: boolean; discount?: number; freight?: number; notes?: string } }) =>
            apiPatch<PurchaseOrder>(`/purchase-orders/${id}`, data),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["purchase-orders"] }),
    });
}

export function useSendPurchaseOrder() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => apiPost<PurchaseOrder>(`/purchase-orders/${id}/send`, {}),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["purchase-orders"] }),
    });
}

export function useReceivePurchaseOrder() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => apiPost<PurchaseOrder>(`/purchase-orders/${id}/receive`, {}),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["purchase-orders"] }),
    });
}

export function useCancelPurchaseOrder() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => apiPost<PurchaseOrder>(`/purchase-orders/${id}/cancel`, {}),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["purchase-orders"] }),
    });
}

export function useDeletePurchaseOrder() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => apiDelete(`/purchase-orders/${id}`),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["purchase-orders"] }),
    });
}
