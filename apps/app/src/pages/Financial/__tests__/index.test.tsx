import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import Financial from "../index";

vi.mock("../tabs/TripAdvancesTab", () => ({ default: () => <div>tab-advances</div> }));
vi.mock("../tabs/TripExpensesTab", () => ({ default: () => <div>tab-expenses</div> }));
vi.mock("../tabs/TripSettlementsTab", () => ({ default: () => <div>tab-settlements</div> }));
vi.mock("../tabs/DriverCardsTab", () => ({ default: () => <div>tab-cards</div> }));
vi.mock("../tabs/TripValidationsTab", () => ({ default: () => <div>tab-validations</div> }));
vi.mock("../tabs/FiscalDocumentsTab", () => ({ default: () => <div>tab-documents</div> }));
vi.mock("../../../hooks/useTrips", () => ({ useTrips: () => ({ data: [] }) }));
vi.mock("../../../hooks/useRoutes", () => ({ useRoutes: () => ({ data: [] }) }));

function FinancialWithLocation() {
  const location = useLocation();
  return (
    <>
      <Financial />
      <div data-testid="location-search">{location.search}</div>
    </>
  );
}

describe("Financial tabs", () => {
  it("abre a tab correta por deep-link", () => {
    render(
      <MemoryRouter initialEntries={["/financial?tab=documents"]}>
        <Routes>
          <Route path="/financial" element={<FinancialWithLocation />} />
        </Routes>
      </MemoryRouter>
    );

    expect(screen.getByText("tab-documents")).toBeInTheDocument();
  });

  it("sincroniza query param ao trocar de tab", async () => {
    const user = userEvent.setup();

    render(
      <MemoryRouter initialEntries={["/financial?tab=advances"]}>
        <Routes>
          <Route path="/financial" element={<FinancialWithLocation />} />
        </Routes>
      </MemoryRouter>
    );

    await user.click(screen.getByRole("tab", { name: /Despesas/i }));

    expect(screen.getByText("tab-expenses")).toBeInTheDocument();
    expect(screen.getByTestId("location-search").textContent).toContain("tab=expenses");
  });
});
