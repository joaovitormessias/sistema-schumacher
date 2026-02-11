import { useQuery } from "@tanstack/react-query";
import { apiGet } from "../services/api";
import { buildListQuery } from "./buildListQuery";

export type TripItem = {
  id: string;
  route_id: string;
  bus_id: string;
  driver_id?: string;
  request_id?: string;
  departure_at: string;
  status: string;
  operational_status: string;
  estimated_km: number;
  dispatch_validated_at?: string;
};

type UseTripsOptions = {
  search?: string;
  status?: string;
};

export function useTrips(limit = 200, offset = 0, options: UseTripsOptions = {}) {
  const search = options.search?.trim() ?? "";
  const status = options.status?.trim() ?? "";

  return useQuery({
    queryKey: ["trips", limit, offset, search, status],
    queryFn: () =>
      apiGet<TripItem[]>(
        buildListQuery("/trips", {
          limit,
          offset,
          search,
          status,
        })
      ),
  });
}
