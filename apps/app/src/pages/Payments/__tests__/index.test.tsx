import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useBookings } from "../../../hooks/useBookings";
import { usePayments } from "../../../hooks/usePayments";
import { apiPost } from "../../../services/api";
import { createTestQueryClient } from "../../../test/queryTestUtils";
import Payments from "../index";

vi.mock("../../../hooks/useBookings", () => ({
  useBookings: vi.fn(),
}));

vi.mock("../../../hooks/usePayments", () => ({
  usePayments: vi.fn(),
}));

vi.mock("../../../services/api", () => ({
  apiPost: vi.fn(),
}));

vi.mock("../../../hooks/useToast", () => ({
  default: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}));

function renderPayments(initialPath = "/payments") {
  const client = createTestQueryClient();
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[initialPath]}>
        <Routes>
          <Route path="/payments" element={<Payments />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  );
}

describe("Payments page", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useBookings).mockReturnValue({
      data: [
        {
          id: "booking-1",
          passenger_name: "Maria",
          total_amount: 200,
          remainder_amount: 120,
          status: "PENDING",
        },
      ],
      isLoading: false,
      error: null,
    } as any);

    vi.mocked(usePayments).mockReturnValue({
      data: [],
      isLoading: false,
      error: null,
    } as any);
  });

  it("pre-seleciona modo e reserva via query string", async () => {
    renderPayments("/payments?booking_id=booking-1&mode=MANUAL");

    expect(screen.getByRole("tab", { name: /Manual/i })).toHaveAttribute("aria-selected", "true");

    await waitFor(() => {
      expect(screen.getByLabelText(/Reserva/i)).toHaveValue("booking-1");
    });
  });

  it("sincroniza pagamento pendente ao clicar em Atualizar status", async () => {
    const user = userEvent.setup();
    vi.mocked(usePayments).mockReturnValue({
      data: [
        {
          id: "payment-1",
          booking_id: "booking-1",
          amount: 120,
          method: "PIX",
          status: "PENDING",
          provider: "ABACATEPAY",
          created_at: "2026-02-06T10:00:00Z",
        },
      ],
      isLoading: false,
      error: null,
    } as any);

    vi.mocked(apiPost).mockResolvedValue({
      payment: {
        id: "payment-1",
        booking_id: "booking-1",
        amount: 120,
        method: "PIX",
        status: "PAID",
        provider: "ABACATEPAY",
        created_at: "2026-02-06T10:00:00Z",
      },
      booking_status: "CONFIRMED",
      synced: true,
    });

    renderPayments();

    await user.click(screen.getByRole("button", { name: /Atualizar status/i }));

    await waitFor(() => {
      expect(apiPost).toHaveBeenCalledWith("/payments/payment-1/sync", {});
    });
  });
});
