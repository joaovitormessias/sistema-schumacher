import { useQuery } from "@tanstack/react-query";
import { apiGet } from "../services/api";

export type RouteStopItem = {
  id: string;
  route_id: string;
  city: string;
  stop_order: number;
  eta_offset_minutes?: number | null;
  notes?: string | null;
  created_at: string;
};

export function useRouteStops(routeId: string | null) {
  return useQuery({
    queryKey: ["route-stops", routeId],
    enabled: Boolean(routeId),
    queryFn: () => apiGet<RouteStopItem[]>(`/routes/${routeId}/stops`),
  });
}
