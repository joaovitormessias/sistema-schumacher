import { renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { usePayments } from "../usePayments";
import { createTestQueryClient, withQueryClient } from "../../test/queryTestUtils";
import { apiGet } from "../../services/api";

vi.mock("../../services/api", () => ({
  apiGet: vi.fn(),
}));

describe("usePayments", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("carrega pagamentos e respeita limite/offset", async () => {
    vi.mocked(apiGet).mockResolvedValueOnce([
      {
        id: "payment-1",
        booking_id: "booking-1",
        amount: 40,
        method: "PIX",
        status: "PENDING",
        created_at: "2026-02-06T10:00:00Z",
      },
    ]);

    const client = createTestQueryClient();
    const { result } = renderHook(() => usePayments(15, 30), {
      wrapper: withQueryClient(client),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(apiGet).toHaveBeenCalledWith("/payments?limit=15&offset=30");
    expect(result.current.data?.[0]?.method).toBe("PIX");
  });
});
