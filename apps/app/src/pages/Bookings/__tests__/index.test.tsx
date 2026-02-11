import { QueryClientProvider } from "@tanstack/react-query";
import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useBookings } from "../../../hooks/useBookings";
import { useRoutes } from "../../../hooks/useRoutes";
import { useTrips } from "../../../hooks/useTrips";
import { apiGet, apiPost } from "../../../services/api";
import { createTestQueryClient } from "../../../test/queryTestUtils";
import Bookings from "../index";

vi.mock("../../../hooks/useTrips", () => ({
  useTrips: vi.fn(),
}));

vi.mock("../../../hooks/useRoutes", () => ({
  useRoutes: vi.fn(),
}));

vi.mock("../../../hooks/useBookings", () => ({
  useBookings: vi.fn(),
}));

vi.mock("../../../services/api", () => ({
  apiGet: vi.fn(),
  apiPost: vi.fn(),
  apiPatch: vi.fn(),
}));

vi.mock("../../../hooks/useToast", () => ({
  default: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}));

function renderBookings() {
  const client = createTestQueryClient();
  return render(
    <QueryClientProvider client={client}>
      <Bookings />
    </QueryClientProvider>
  );
}

describe("Bookings shortcuts", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();

    vi.mocked(useTrips).mockReturnValue({
      data: [
        {
          id: "trip-1",
          route_id: "route-1",
          bus_id: "bus-1",
          departure_at: "2026-02-06T10:00:00Z",
        },
      ],
      isLoading: false,
      error: null,
    } as any);

    vi.mocked(useRoutes).mockReturnValue({
      data: [{ id: "route-1", origin_city: "Chapeco", destination_city: "Lages" }],
      isLoading: false,
      error: null,
    } as any);

    vi.mocked(useBookings).mockImplementation((limit = 200) => {
      if (limit === 500) {
        return {
          data: [
            {
              id: "history-1",
              trip_id: "trip-1",
              status: "CONFIRMED",
              created_at: "2026-02-05T10:00:00Z",
              passenger_name: "Maria Oliveira",
              passenger_phone: "49999990000",
              passenger_email: "maria@example.com",
              seat_number: 12,
              total_amount: 180,
              deposit_amount: 60,
              remainder_amount: 120,
            },
            {
              id: "history-2",
              trip_id: "trip-1",
              status: "CONFIRMED",
              created_at: "2026-02-04T10:00:00Z",
              passenger_name: "Maria Oliveira",
              passenger_phone: "49999990000",
              passenger_email: "maria@example.com",
              seat_number: 9,
              total_amount: 180,
              deposit_amount: 60,
              remainder_amount: 120,
            },
          ],
          isLoading: false,
          error: null,
        } as any;
      }

      return {
        data: [],
        isLoading: false,
        error: null,
      } as any;
    });

    vi.mocked(apiGet).mockImplementation(async (path: string) => {
      if (path.startsWith("/trips/trip-1/stops")) {
        return [
          { id: "stop-1", route_stop_id: "rs-1", city: "Chapeco", stop_order: 1 },
          { id: "stop-2", route_stop_id: "rs-2", city: "Lages", stop_order: 2 },
        ];
      }
      if (path.startsWith("/trips/trip-1/seats")) {
        return [
          { id: "seat-1", seat_number: 12, is_active: true, is_taken: false },
          { id: "seat-2", seat_number: 9, is_active: true, is_taken: true },
        ];
      }
      return [];
    });

    vi.mocked(apiPost).mockImplementation(async (path: string) => {
      if (path === "/pricing/quote") {
        return {
          base_amount: 180,
          calc_amount: 180,
          final_amount: 180,
          currency: "BRL",
          fare_mode: "AUTO",
          occupancy_ratio: 0.3,
        };
      }
      return {};
    });
  });

  it("repete a ultima reserva salva no localStorage", async () => {
    window.localStorage.setItem(
      "booking.last_success",
      JSON.stringify({
        trip_id: "trip-1",
        seat_id: "seat-1",
        board_stop_id: "stop-1",
        alight_stop_id: "stop-2",
        fare_mode: "AUTO",
        name: "Cliente Snapshot",
        document: "123",
        phone: "49998887777",
        email: "snapshot@example.com",
        total_amount: 180,
        deposit_amount: 60,
        remainder_amount: 120,
        payment_method: "PIX",
        payment_description: "Passagem",
        payment_notes: "",
        saved_at: "2026-02-06T10:00:00Z",
      })
    );

    const user = userEvent.setup();
    renderBookings();

    await user.click(screen.getByRole("button", { name: /Repetir ultima reserva/i }));

    expect(screen.getByLabelText(/Nome do passageiro/i)).toHaveValue("Cliente Snapshot");
    expect(screen.getByLabelText(/Telefone/i)).toHaveValue("49998887777");
  });

  it("aplica auto-complete de passageiro e sugere apenas poltrona disponivel", async () => {
    const user = userEvent.setup();
    renderBookings();

    await user.selectOptions(screen.getByLabelText(/Viagem/i), "trip-1");
    const boardSelect = screen.getByLabelText(/Embarque/i, { selector: "select:not([disabled])" });
    await waitFor(() => {
      expect(within(boardSelect).getByRole("option", { name: /1 - Chapeco/i })).toBeInTheDocument();
    });
    await user.selectOptions(boardSelect, "stop-1");
    await user.selectOptions(screen.getByLabelText(/Desembarque/i), "stop-2");

    const passengerSelect = screen.getByLabelText(/Passageiro frequente/i);
    const mariaOption = within(passengerSelect).getByRole("option", { name: /Maria Oliveira/i });
    await user.selectOptions(passengerSelect, mariaOption);

    expect(screen.getByLabelText(/Nome do passageiro/i)).toHaveValue("Maria Oliveira");
    expect(screen.getByLabelText(/Telefone/i)).toHaveValue("49999990000");
    expect(screen.getByLabelText(/E-mail/i)).toHaveValue("maria@example.com");

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /Poltrona 12/i })).toBeInTheDocument();
    });
    expect(screen.queryByRole("button", { name: /Poltrona 9/i })).not.toBeInTheDocument();
  });

  it("preserva total automatico quando quote e digitacao ocorrem juntos", async () => {
    const user = userEvent.setup();
    let resolveQuote: ((value: any) => void) | null = null;

    vi.mocked(apiPost).mockImplementation((path: string) => {
      if (path === "/pricing/quote") {
        return new Promise((resolve) => {
          resolveQuote = resolve;
        });
      }
      return Promise.resolve({});
    });

    renderBookings();

    await user.selectOptions(screen.getByLabelText(/Viagem/i), "trip-1");
    const boardSelect = screen.getByLabelText(/Embarque/i, { selector: "select:not([disabled])" });
    await waitFor(() => {
      expect(within(boardSelect).getByRole("option", { name: /1 - Chapeco/i })).toBeInTheDocument();
    });
    await user.selectOptions(boardSelect, "stop-1");
    await user.selectOptions(screen.getByLabelText(/Desembarque/i), "stop-2");
    await user.selectOptions(screen.getByLabelText(/Poltrona/i), "seat-1");

    const passengerInput = screen.getByLabelText(/Nome do passageiro/i);
    await user.type(passengerInput, "M");

    await waitFor(() => {
      expect(resolveQuote).toBeTypeOf("function");
    });

    await act(async () => {
      resolveQuote?.({
        base_amount: 180,
        calc_amount: 180,
        final_amount: 180,
        currency: "BRL",
        fare_mode: "AUTO",
        occupancy_ratio: 0.3,
      });
      fireEvent.change(passengerInput, { target: { value: "Maria" } });
      await Promise.resolve();
    });

    await user.click(screen.getByRole("button", { name: /Continuar para pagamento/i }));
    await waitFor(() => {
      expect(screen.getByLabelText(/Total da reserva/i)).toHaveValue(180);
    });
  });

  it("mostra aviso quando o backend retorna FARE_NOT_FOUND", async () => {
    const user = userEvent.setup();

    vi.mocked(apiPost).mockImplementation((path: string) => {
      if (path === "/pricing/quote") {
        return Promise.reject({ code: "FARE_NOT_FOUND", message: "falha ao calcular" });
      }
      return Promise.resolve({});
    });

    renderBookings();

    await user.selectOptions(screen.getByLabelText(/Viagem/i), "trip-1");
    const boardSelect = screen.getByLabelText(/Embarque/i, { selector: "select:not([disabled])" });
    await waitFor(() => {
      expect(within(boardSelect).getByRole("option", { name: /1 - Chapeco/i })).toBeInTheDocument();
    });
    await user.selectOptions(boardSelect, "stop-1");
    await user.selectOptions(screen.getByLabelText(/Desembarque/i), "stop-2");

    await waitFor(() => {
      expect(screen.getByText(/Tarifa do trecho nao encontrada/i)).toBeInTheDocument();
    });
  });
});
