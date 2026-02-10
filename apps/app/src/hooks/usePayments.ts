import { useQuery } from "@tanstack/react-query";
import { apiGet } from "../services/api";
import { buildListQuery } from "./buildListQuery";

export type PaymentItem = {
  id: string;
  booking_id: string;
  amount: number;
  method: string;
  status: string;
  provider?: string;
  provider_ref?: string;
  paid_at?: string | null;
  created_at: string;
};

type UsePaymentsOptions = {
  booking_id?: string;
  status?: string;
  search?: string;
};

export function usePayments(limit = 200, offset = 0, options: UsePaymentsOptions = {}) {
  const booking_id = options.booking_id?.trim() ?? "";
  const status = options.status?.trim() ?? "";
  const search = options.search?.trim() ?? "";

  return useQuery({
    queryKey: ["payments", limit, offset, booking_id, status, search],
    queryFn: () =>
      apiGet<PaymentItem[]>(
        buildListQuery("/payments", {
          limit,
          offset,
          booking_id,
          status,
          search,
        })
      ),
  });
}
