import { createContext, useContext, useMemo, useState, type ReactNode } from "react";

type FinancialFiltersContextValue = {
  tripFilter: string;
  setTripFilter: (value: string) => void;
  clearFilters: () => void;
};

const FinancialFiltersContext = createContext<FinancialFiltersContextValue | null>(null);

export function FinancialFiltersProvider({ children }: { children: ReactNode }) {
  const [tripFilter, setTripFilter] = useState("");

  const value = useMemo(
    () => ({
      tripFilter,
      setTripFilter,
      clearFilters: () => setTripFilter(""),
    }),
    [tripFilter]
  );

  return (
    <FinancialFiltersContext.Provider value={value}>{children}</FinancialFiltersContext.Provider>
  );
}

export function useFinancialFiltersOptional() {
  return useContext(FinancialFiltersContext);
}
