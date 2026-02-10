import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiGet, apiPost, apiPatch, apiDelete } from "../services/api";

export type Supplier = {
  id: string;
  name: string;
  document?: string;
  phone?: string;
  email?: string;
  payment_terms?: string;
  billing_day?: number;
  is_active: boolean;
  notes?: string;
  created_at: string;
  updated_at: string;
};

export type CreateSupplierInput = {
  name: string;
  document?: string;
  phone?: string;
  email?: string;
  payment_terms?: string;
  billing_day?: number;
  notes?: string;
};

export function useSuppliers(limit = 200, offset = 0, active?: boolean) {
  let url = `/suppliers?limit=${limit}&offset=${offset}`;
  if (active !== undefined) url += `&active=${active}`;
  
  return useQuery({
    queryKey: ["suppliers", limit, offset, active],
    queryFn: () => apiGet<Supplier[]>(url),
  });
}

export function useSupplier(id: string) {
  return useQuery({
    queryKey: ["supplier", id],
    queryFn: () => apiGet<Supplier>(`/suppliers/${id}`),
    enabled: !!id,
  });
}

export function useCreateSupplier() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSupplierInput) => apiPost<Supplier>("/suppliers", data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["suppliers"] }),
  });
}

export function useUpdateSupplier() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateSupplierInput> }) =>
      apiPatch<Supplier>(`/suppliers/${id}`, data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["suppliers"] }),
  });
}

export function useDeleteSupplier() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiDelete(`/suppliers/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["suppliers"] }),
  });
}
