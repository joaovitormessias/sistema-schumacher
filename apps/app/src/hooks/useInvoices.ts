import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiGet, apiPost, apiPatch, apiDelete } from "../services/api";

export type InvoiceStatus = "PENDING" | "PROCESSED" | "CANCELLED";

export type InvoiceItem = {
    id: string;
    invoice_id: string;
    product_id: string;
    product_name?: string;
    product_code?: string;
    quantity: number;
    unit_price: number;
    discount: number;
    total: number;
    created_at: string;
};

export type Invoice = {
    id: string;
    invoice_number: string;
    barcode?: string;
    supplier_id: string;
    supplier_name?: string;
    purchase_order_id?: string;
    service_order_id?: string;
    bus_id?: string;
    bus_plate?: string;
    issue_date: string;
    issue_time?: string;
    entry_date: string;
    entry_time?: string;
    cfop: string;
    payment_type?: string;
    due_date?: string;
    subtotal: number;
    discount: number;
    freight: number;
    total: number;
    status: InvoiceStatus;
    notes?: string;
    driver_id?: string;
    driver_name?: string;
    odometer_km?: number;
    created_at: string;
    items?: InvoiceItem[];
};

export type CreateInvoiceInput = {
    invoice_number: string;
    barcode?: string;
    supplier_id: string;
    purchase_order_id?: string;
    service_order_id?: string;
    bus_id?: string;
    issue_date: string;
    issue_time?: string;
    cfop?: string;
    payment_type?: string;
    due_date?: string;
    discount?: number;
    freight?: number;
    notes?: string;
    driver_id?: string;
    odometer_km?: number;
    items: { product_id: string; quantity: number; unit_price: number; discount?: number }[];
};

export function useInvoices(limit = 200, offset = 0, status?: InvoiceStatus, supplier_id?: string, bus_id?: string) {
    let url = `/invoices?limit=${limit}&offset=${offset}`;
    if (status) url += `&status=${status}`;
    if (supplier_id) url += `&supplier_id=${supplier_id}`;
    if (bus_id) url += `&bus_id=${bus_id}`;

    return useQuery({
        queryKey: ["invoices", limit, offset, status, supplier_id, bus_id],
        queryFn: () => apiGet<Invoice[]>(url),
    });
}

export function useInvoice(id: string) {
    return useQuery({
        queryKey: ["invoice", id],
        queryFn: () => apiGet<Invoice>(`/invoices/${id}`),
        enabled: !!id,
    });
}

export function useCreateInvoice() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: CreateInvoiceInput) => apiPost<Invoice>("/invoices", data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["invoices"] });
            queryClient.invalidateQueries({ queryKey: ["products"] }); // Stock updated
        },
    });
}

export function useUpdateInvoice() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ id, data }: { id: string; data: { barcode?: string; payment_type?: string; due_date?: string; notes?: string; driver_id?: string; odometer_km?: number } }) =>
            apiPatch<Invoice>(`/invoices/${id}`, data),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["invoices"] }),
    });
}

export function useProcessInvoice() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => apiPost<Invoice>(`/invoices/${id}/process`, {}),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["invoices"] }),
    });
}

export function useCancelInvoice() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => apiPost<Invoice>(`/invoices/${id}/cancel`, {}),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["invoices"] }),
    });
}

export function useDeleteInvoice() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => apiDelete(`/invoices/${id}`),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["invoices"] }),
    });
}
