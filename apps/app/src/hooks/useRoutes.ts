import { useQuery } from "@tanstack/react-query";
import { apiGet } from "../services/api";
import { buildListQuery } from "./buildListQuery";

export type RouteItem = {
  id: string;
  name?: string;
  origin_city: string;
  destination_city: string;
};

type UseRoutesOptions = {
  search?: string;
};

export function useRoutes(limit = 200, offset = 0, options: UseRoutesOptions = {}) {
  const search = options.search?.trim() ?? "";

  return useQuery({
    queryKey: ["routes", limit, offset, search],
    queryFn: () =>
      apiGet<RouteItem[]>(
        buildListQuery("/routes", {
          limit,
          offset,
          search,
        })
      ),
  });
}
