import { useQuery } from "@tanstack/react-query";
import { apiGet } from "../services/api";
import { buildListQuery } from "./buildListQuery";

export type DriverItem = {
  id: string;
  name: string;
};

type UseDriversOptions = {
  search?: string;
};

export function useDrivers(limit = 200, offset = 0, options: UseDriversOptions = {}) {
  const search = options.search?.trim() ?? "";

  return useQuery({
    queryKey: ["drivers", limit, offset, search],
    queryFn: () =>
      apiGet<DriverItem[]>(
        buildListQuery("/drivers", {
          limit,
          offset,
          search,
        })
      ),
  });
}
