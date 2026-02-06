import { Navigate, Route, Routes } from "react-router-dom";
import Layout from "./Layout";
import Dashboard from "../pages/Dashboard";
import Trips from "../pages/Trips";
import RoutesPage from "../pages/Routes";
import Buses from "../pages/Buses";
import Drivers from "../pages/Drivers";
import Bookings from "../pages/Bookings";
import Payments from "../pages/Payments";
import Reports from "../pages/Reports";
import Pricing from "../pages/Pricing";
import TripAdvances from "../pages/TripAdvances";
import TripExpenses from "../pages/TripExpenses";
import TripSettlements from "../pages/TripSettlements";
import DriverCards from "../pages/DriverCards";
import TripValidations from "../pages/TripValidations";
import FiscalDocuments from "../pages/FiscalDocuments";

export default function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/trips" element={<Trips />} />
        <Route path="/routes" element={<RoutesPage />} />
        <Route path="/buses" element={<Buses />} />
        <Route path="/drivers" element={<Drivers />} />
        <Route path="/bookings" element={<Bookings />} />
        <Route path="/payments" element={<Payments />} />
        <Route path="/reports" element={<Reports />} />
        <Route path="/pricing" element={<Pricing />} />
        <Route path="/trip-advances" element={<TripAdvances />} />
        <Route path="/trip-expenses" element={<TripExpenses />} />
        <Route path="/trip-settlements" element={<TripSettlements />} />
        <Route path="/driver-cards" element={<DriverCards />} />
        <Route path="/trip-validations" element={<TripValidations />} />
        <Route path="/fiscal-documents" element={<FiscalDocuments />} />
      </Routes>
    </Layout>
  );
}
