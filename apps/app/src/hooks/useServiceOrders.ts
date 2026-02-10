import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiGet, apiPost, apiPatch, apiDelete } from "../services/api";

export type ServiceOrderType = "PREVENTIVE" | "CORRECTIVE";
export type ServiceOrderStatus = "OPEN" | "IN_PROGRESS" | "CLOSED" | "CANCELLED";

export type ServiceOrder = {
    id: string;
    order_number: number;
    bus_id: string;
    bus_plate?: string;
    driver_id?: string;
    driver_name?: string;
    order_type: ServiceOrderType;
    status: ServiceOrderStatus;
    description: string;
    odometer_km?: number;
    scheduled_date?: string;
    location: string;
    opened_at: string;
    closed_at?: string;
    closed_odometer_km?: number;
    next_preventive_km?: number;
    notes?: string;
    created_at: string;
    updated_at: string;
};

export type CreateServiceOrderInput = {
    bus_id: string;
    driver_id?: string;
    order_type: ServiceOrderType;
    description: string;
    odometer_km?: number;
    scheduled_date?: string;
    location?: string;
    notes?: string;
};

export function useServiceOrders(limit = 200, offset = 0, status?: ServiceOrderStatus, order_type?: ServiceOrderType, bus_id?: string) {
    let url = `/service-orders?limit=${limit}&offset=${offset}`;
    if (status) url += `&status=${status}`;
    if (order_type) url += `&order_type=${order_type}`;
    if (bus_id) url += `&bus_id=${bus_id}`;

    return useQuery({
        queryKey: ["service-orders", limit, offset, status, order_type, bus_id],
        queryFn: () => apiGet<ServiceOrder[]>(url),
    });
}

export function useServiceOrder(id: string) {
    return useQuery({
        queryKey: ["service-order", id],
        queryFn: () => apiGet<ServiceOrder>(`/service-orders/${id}`),
        enabled: !!id,
    });
}

export function useCreateServiceOrder() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: CreateServiceOrderInput) => apiPost<ServiceOrder>("/service-orders", data),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["service-orders"] }),
    });
}

export function useUpdateServiceOrder() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ id, data }: { id: string; data: Partial<CreateServiceOrderInput> }) =>
            apiPatch<ServiceOrder>(`/service-orders/${id}`, data),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["service-orders"] }),
    });
}

export function useStartServiceOrder() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => apiPost<ServiceOrder>(`/service-orders/${id}/start`, {}),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["service-orders"] }),
    });
}

export function useCloseServiceOrder() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ id, data }: { id: string; data: { closed_odometer_km?: number; next_preventive_km?: number } }) =>
            apiPost<ServiceOrder>(`/service-orders/${id}/close`, data),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["service-orders"] }),
    });
}

export function useCancelServiceOrder() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => apiPost<ServiceOrder>(`/service-orders/${id}/cancel`, {}),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["service-orders"] }),
    });
}

export function useDeleteServiceOrder() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => apiDelete(`/service-orders/${id}`),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["service-orders"] }),
    });
}
