import { Navigate, Route, Routes, useLocation } from "react-router-dom";
import AppLayout from "./layouts/AppLayout";
import DashboardPage from "./pages/DashboardPage";
import HandoverRecordsPage from "./pages/HandoverRecordsPage";
import ListingDetailPage from "./pages/ListingDetailPage";
import LoginPage from "./pages/LoginPage";
import MarketplacePage from "./pages/MarketplacePage";
import OffersPage from "./pages/OffersPage";
import PaymentsPage from "./pages/PaymentsPage";
import PlaceholderPage from "./pages/PlaceholderPage";
import PriceBoardPage from "./pages/PriceBoardPage";
import ProfilePage from "./pages/ProfilePage";
import RegisterPage from "./pages/RegisterPage";
import TransactionsPage from "./pages/TransactionsPage";
import { useAuth } from "./state/auth-context";

function Protected() {
  const { user, loading } = useAuth();
  const loc = useLocation();
  if (loading) return <div className="boot-screen">KabadiConnect</div>;
  if (!user)
    return <Navigate to="/login" replace state={{ from: loc.pathname }} />;
  return <AppLayout />;
}
export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route element={<Protected />}>
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/marketplace" element={<MarketplacePage />} />
        <Route path="/marketplace/:id" element={<ListingDetailPage />} />
        <Route path="/offers" element={<OffersPage />} />
        <Route path="/transactions" element={<TransactionsPage />} />
        <Route
          path="/orders"
          element={<Navigate to="/transactions" replace />}
        />
        <Route path="/prices" element={<PriceBoardPage />} />
        <Route path="/handovers" element={<HandoverRecordsPage />} />
        <Route path="/payments" element={<PaymentsPage />} />
        <Route path="/profile" element={<ProfilePage />} />
        {["requirements", "collectors", "help"].map((p) => (
          <Route key={p} path={`/${p}`} element={<PlaceholderPage />} />
        ))}
      </Route>
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}
