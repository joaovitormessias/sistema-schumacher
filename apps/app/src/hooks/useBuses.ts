import { useQuery } from "@tanstack/react-query";
import { apiGet } from "../services/api";
import { buildListQuery } from "./buildListQuery";

export type BusItem = {
  id: string;
  name: string;
};

type UseBusesOptions = {
  search?: string;
};

export function useBuses(limit = 200, offset = 0, options: UseBusesOptions = {}) {
  const search = options.search?.trim() ?? "";

  return useQuery({
    queryKey: ["buses", limit, offset, search],
    queryFn: () =>
      apiGet<BusItem[]>(
        buildListQuery("/buses", {
          limit,
          offset,
          search,
        })
      ),
  });
}
