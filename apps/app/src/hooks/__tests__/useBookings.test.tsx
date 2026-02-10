import { renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useBookings } from "../useBookings";
import { createTestQueryClient, withQueryClient } from "../../test/queryTestUtils";
import { apiGet } from "../../services/api";

vi.mock("../../services/api", () => ({
  apiGet: vi.fn(),
}));

describe("useBookings", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("carrega reservas com paginação", async () => {
    vi.mocked(apiGet).mockResolvedValueOnce([
      {
        id: "booking-1",
        trip_id: "trip-1",
        status: "PENDING",
        passenger_name: "Joao Silva",
        seat_number: 4,
        total_amount: 120,
        deposit_amount: 40,
        remainder_amount: 80,
      },
    ]);

    const client = createTestQueryClient();
    const { result } = renderHook(() => useBookings(25, 50), {
      wrapper: withQueryClient(client),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(apiGet).toHaveBeenCalledWith("/bookings?limit=25&offset=50");
    expect(result.current.data?.[0]?.passenger_name).toBe("Joao Silva");
  });
});
