import { useQuery } from "@tanstack/react-query";
import { apiGet } from "../services/api";
import { buildListQuery } from "./buildListQuery";

export type CityCandidate = {
  place_id: string;
  city: string;
  state_code?: string | null;
  display_name: string;
  addresstype: string;
  latitude: number;
  longitude: number;
};

export function useCityCandidates(query: string, enabled = true, limit = 8) {
  const normalizedQuery = query.trim();

  return useQuery({
    queryKey: ["city-candidates", normalizedQuery, limit],
    enabled: enabled && normalizedQuery.length >= 2,
    queryFn: () =>
      apiGet<CityCandidate[]>(
        buildListQuery("/routes/cities/candidates", {
          query: normalizedQuery,
          limit,
        })
      ),
    staleTime: 5 * 60 * 1000,
  });
}
