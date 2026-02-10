import { renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useTrips } from "../useTrips";
import { createTestQueryClient, withQueryClient } from "../../test/queryTestUtils";
import { apiGet } from "../../services/api";

vi.mock("../../services/api", () => ({
  apiGet: vi.fn(),
}));

describe("useTrips", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("carrega viagens com os parametros corretos", async () => {
    vi.mocked(apiGet).mockResolvedValueOnce([
      {
        id: "trip-1",
        route_id: "route-1",
        bus_id: "bus-1",
        departure_at: "2026-02-06T10:00:00Z",
        status: "SCHEDULED",
      },
    ]);

    const client = createTestQueryClient();
    const { result } = renderHook(() => useTrips(10, 20), {
      wrapper: withQueryClient(client),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(apiGet).toHaveBeenCalledWith("/trips?limit=10&offset=20");
    expect(result.current.data?.[0]?.id).toBe("trip-1");
  });
});
