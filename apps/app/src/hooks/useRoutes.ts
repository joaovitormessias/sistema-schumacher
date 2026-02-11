import { useQuery } from "@tanstack/react-query";
import { apiGet } from "../services/api";
import { buildListQuery } from "./buildListQuery";

export type RouteItem = {
  id: string;
  name: string;
  origin_city: string;
  destination_city: string;
  is_active: boolean;
  stop_count: number;
  configuration_status: "INCOMPLETE" | "READY" | "ACTIVE" | "SUSPENDED";
  missing_requirements: string[];
  has_linked_trips: boolean;
  duplicated_from_route_id?: string | null;
};

type UseRoutesOptions = {
  search?: string;
  status?: "active" | "inactive" | "all";
};

export function useRoutes(limit = 200, offset = 0, options: UseRoutesOptions = {}) {
  const search = options.search?.trim() ?? "";
  const status = options.status ?? "all";

  return useQuery({
    queryKey: ["routes", limit, offset, search, status],
    queryFn: () =>
      apiGet<RouteItem[]>(
        buildListQuery("/routes", {
          limit,
          offset,
          search,
          status,
        })
      ),
  });
}
