import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiGet, apiPost, apiPatch, apiDelete } from "../services/api";

export type Product = {
    id: string;
    code: string;
    name: string;
    category?: string;
    unit: string;
    min_stock: number;
    current_stock: number;
    last_cost?: number;
    is_active: boolean;
    created_at: string;
    updated_at: string;
};

export type CreateProductInput = {
    code: string;
    name: string;
    category?: string;
    unit?: string;
    min_stock?: number;
};

export function useProducts(limit = 200, offset = 0, active?: boolean, category?: string, search?: string) {
    let url = `/products?limit=${limit}&offset=${offset}`;
    if (active !== undefined) url += `&active=${active}`;
    if (category) url += `&category=${encodeURIComponent(category)}`;
    if (search) url += `&search=${encodeURIComponent(search)}`;

    return useQuery({
        queryKey: ["products", limit, offset, active, category, search],
        queryFn: () => apiGet<Product[]>(url),
    });
}

export function useProduct(id: string) {
    return useQuery({
        queryKey: ["product", id],
        queryFn: () => apiGet<Product>(`/products/${id}`),
        enabled: !!id,
    });
}

export function useCreateProduct() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: CreateProductInput) => apiPost<Product>("/products", data),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["products"] }),
    });
}

export function useUpdateProduct() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ id, data }: { id: string; data: Partial<CreateProductInput> }) =>
            apiPatch<Product>(`/products/${id}`, data),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["products"] }),
    });
}

export function useDeleteProduct() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => apiDelete(`/products/${id}`),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["products"] }),
    });
}
