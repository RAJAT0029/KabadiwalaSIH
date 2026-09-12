import {
  Bell,
  ChevronDown,
  CircleHelp,
  ClipboardList,
  CreditCard,
  FileBox,
  Handshake,
  LayoutDashboard,
  LogOut,
  Menu,
  PackageSearch,
  ReceiptIndianRupee,
  Search,
  Settings,
  TrendingUp,
  UsersRound,
  X,
} from "lucide-react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useState } from "react";
import { useModalFocus } from "../components/useModalFocus";
import { useAuth } from "../state/auth-context";

const nav = [
  ["Dashboard", "/dashboard", LayoutDashboard],
  ["E-Waste Lots", "/marketplace", PackageSearch],
  ["My Offers", "/offers", ReceiptIndianRupee],
  ["Transactions / Orders", "/transactions", Handshake],
  ["Material Requirements", "/requirements", FileBox],
  ["Collectors", "/collectors", UsersRound],
  ["Price Board", "/prices", TrendingUp],
  ["Handover Records", "/handovers", ClipboardList],
  ["Payments", "/payments", CreditCard],
] as const;

export default function AppLayout() {
  const [open, setOpen] = useState(false);
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const doLogout = async () => {
    try {
      await logout();
    } catch {
      /* Local session is cleared even when offline. */
    }
    navigate("/login", { replace: true });
  };
  const sidebarRef = useModalFocus<HTMLElement>(open, () => setOpen(false));
  return (
    <div className="app-shell">
      {open && (
        <button
          className="drawer-backdrop"
          onClick={() => setOpen(false)}
          aria-label="Close navigation"
        />
      )}
      <aside
        ref={sidebarRef}
        role={open ? "dialog" : undefined}
        aria-modal={open || undefined}
        aria-label="Recycler navigation"
        className={`sidebar ${open ? "open" : ""}`}
      >
        <div className="brand-row">
          <div className="brand-mark">
            <PackageSearch size={22} />
          </div>
          <div>
            <strong>KabadiConnect</strong>
            <span>E-Waste Formalization</span>
          </div>
          <button
            className="icon-button mobile-only"
            onClick={() => setOpen(false)}
            aria-label="Close menu"
          >
            <X />
          </button>
        </div>
        <nav className="main-nav">
          {nav.map(([label, to, Icon]) => (
            <NavLink
              key={to}
              to={to}
              onClick={() => setOpen(false)}
              className={({ isActive }) => (isActive ? "active" : "")}
            >
              <Icon size={19} />
              <span>{label}</span>
            </NavLink>
          ))}
        </nav>
        <div className="sidebar-spacer" />
        <nav className="main-nav secondary">
          <NavLink to="/profile" onClick={() => setOpen(false)}>
            <Settings size={19} />
            <span>Profile</span>
          </NavLink>
          <NavLink to="/help" onClick={() => setOpen(false)}>
            <CircleHelp size={19} />
            <span>Help</span>
          </NavLink>
        </nav>
        <button className="logout-button" onClick={doLogout}>
          <LogOut size={19} />
          Logout
        </button>
      </aside>
      <div className="app-content">
        <header className="topbar">
          <button
            className="icon-button mobile-only"
            onClick={() => setOpen(true)}
            aria-label="Open navigation"
            aria-expanded={open}
          >
            <Menu />
          </button>
          <div className="global-search">
            <Search size={18} />
            <input
              aria-label="Global search"
              placeholder="Search is available in E-Waste Lots"
              disabled
            />
          </div>
          <div className="topbar-actions">
            <button
              className="icon-button notification"
              aria-label="Notifications (planned)"
              title="Notifications are planned"
              disabled
            >
              <Bell />
            </button>
            <button
              className="profile-button"
              onClick={() => navigate("/profile")}
            >
              <div className="avatar">
                {(user?.facilityName || user?.companyName || user?.name || "R")
                  .slice(0, 2)
                  .toUpperCase()}
              </div>
              <div className="profile-copy">
                <strong>
                  {user?.facilityName || user?.companyName || "Recycler"}
                </strong>
                <span>
                  {user?.authorization?.isDemo
                    ? "Development authorization"
                    : user?.authorization?.status}
                </span>
              </div>
              <ChevronDown size={16} />
            </button>
          </div>
        </header>
        <main className="page-content">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
