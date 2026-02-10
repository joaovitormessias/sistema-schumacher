import { useQuery } from "@tanstack/react-query";
import { apiGet } from "../services/api";
import { buildListQuery } from "./buildListQuery";

export type BookingItem = {
  id: string;
  trip_id: string;
  status: string;
  passenger_name: string;
  passenger_phone?: string;
  passenger_email?: string;
  created_at?: string;
  seat_number: number;
  total_amount: number;
  deposit_amount: number;
  remainder_amount: number;
};

type UseBookingsOptions = {
  search?: string;
  status?: string;
  trip_id?: string;
};

export function useBookings(limit = 200, offset = 0, options: UseBookingsOptions = {}) {
  const search = options.search?.trim() ?? "";
  const status = options.status?.trim() ?? "";
  const trip_id = options.trip_id?.trim() ?? "";

  return useQuery({
    queryKey: ["bookings", limit, offset, search, status, trip_id],
    queryFn: () =>
      apiGet<BookingItem[]>(
        buildListQuery("/bookings", {
          limit,
          offset,
          search,
          status,
          trip_id,
        })
      ),
  });
}
