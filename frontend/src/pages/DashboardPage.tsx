import {
  AlertTriangle,
  ArrowRight,
  BadgeCheck,
  CircleDollarSign,
  ClipboardCheck,
  PackageSearch,
  ShieldCheck,
  Truck,
} from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { EWasteLotCard } from "../components/EWasteLotCard";
import { LoadingSkeleton } from "../components/LoadingSkeleton";
import { api } from "../services/api";
import { useAuth } from "../state/auth-context";
import type {
  HandoverRecord,
  LotView,
  OfferView,
  PriceBoard,
  Transaction,
} from "../types";

export default function DashboardPage() {
  const { user } = useAuth();
  const [lots, setLots] = useState<LotView[]>([]);
  const [lotTotal, setLotTotal] = useState(0);
  const [offers, setOffers] = useState<OfferView[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [handovers, setHandovers] = useState<HandoverRecord[]>([]);
  const [prices, setPrices] = useState<PriceBoard | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  useEffect(() => {
    Promise.all([
      api<{ items: LotView[]; pagination: { total: number } }>("/lots?limit=4"),
      api<{ items: OfferView[] }>("/offers?status=pending"),
      api<{ items: Transaction[] }>("/transactions"),
      api<{ items: HandoverRecord[] }>("/handovers"),
      api<PriceBoard>("/prices"),
    ])
      .then(([l, o, t, h, p]) => {
        setLots(l.items);
        setLotTotal(l.pagination.total);
        setOffers(o.items);
        setTransactions(t.items ?? []);
        setHandovers(h.items ?? []);
        setPrices(p);
      })
      .catch((e) =>
        setError(
          e instanceof Error ? e.message : "Unable to load recycler dashboard",
        ),
      )
      .finally(() => setLoading(false));
  }, []);
  const status = user?.effectiveAuthorizationStatus || "pending_verification";
  const demo = user?.authorization?.isDemo;
  const authGood = user?.participationAllowed === true;
  const matches = (lot: LotView) =>
    user?.acceptedMaterials?.some(
      (x) =>
        x.toLowerCase() === lot.lot.material.category.toLowerCase() ||
        x.toLowerCase() === (lot.lot.material.subCategory || "").toLowerCase(),
    ) || false;
  return (
    <div>
      <div className="page-heading">
        <div>
          <span className="eyebrow">Authorized recycler workspace</span>
          <h1>
            Welcome back,{" "}
            {user?.facilityName || user?.companyName || user?.name}
          </h1>
          <p>
            Discover e-waste, compare transparent development price records,
            send offers and maintain traceable handovers.
          </p>
        </div>
        <Link className="primary-button" to="/marketplace">
          <PackageSearch size={18} />
          Browse E-Waste Lots
        </Link>
      </div>
      <div className={`authorization-banner ${authGood ? "ok" : "warn"}`}>
        {authGood ? <BadgeCheck /> : <AlertTriangle />}
        <div>
          <strong>
            {demo
              ? `DEVELOPMENT AUTHORIZATION — ${status.replaceAll("_", " ").toUpperCase()}`
              : status.replaceAll("_", " ").toUpperCase()}
          </strong>
          <span>
            {demo
              ? "This local profile is allowed to test transactions. It is not a government authorization."
              : "Recycler participation depends on a valid authorization status."}
          </span>
        </div>
        <Link to="/profile">Review profile</Link>
      </div>
      <div className="summary-grid">
        <div className="summary-card">
          <span className="summary-icon">
            <PackageSearch />
          </span>
          <div>
            <small>Available E-Waste Lots</small>
            <strong>{loading ? "—" : lotTotal}</strong>
            <p>Collector-created lots</p>
          </div>
        </div>
        <div className="summary-card">
          <span className="summary-icon">
            <CircleDollarSign />
          </span>
          <div>
            <small>Active Offers</small>
            <strong>{loading ? "—" : offers.length}</strong>
            <p>Awaiting collector response</p>
          </div>
        </div>
        <div className="summary-card">
          <span className="summary-icon">
            <ClipboardCheck />
          </span>
          <div>
            <small>Active Handovers</small>
            <strong>
              {loading
                ? "—"
                : handovers.filter(
                    (h) => !["completed", "cancelled"].includes(h.status),
                  ).length}
            </strong>
            <p>Recycler confirmation queue</p>
          </div>
        </div>
        <div className="summary-card">
          <span className="summary-icon">
            <Truck />
          </span>
          <div>
            <small>Incoming Lots</small>
            <strong>
              {loading
                ? "—"
                : transactions.filter(
                    (t) => !["completed", "cancelled"].includes(t.status),
                  ).length}
            </strong>
            <p>Accepted transactions</p>
          </div>
        </div>
      </div>
      <section className="content-section">
        <div className="section-heading">
          <div>
            <h2>Available E-Waste Lots</h2>
            <p>
              Available lots that can be checked against your accepted-material
              profile.
            </p>
          </div>
          <Link className="text-link" to="/marketplace">
            View E-Waste Lots <ArrowRight size={16} />
          </Link>
        </div>
        {error && <div className="inline-error">{error}</div>}
        {loading ? (
          <LoadingSkeleton count={4} />
        ) : (
          <div className="listing-grid">
            {lots.map((x) => (
              <EWasteLotCard key={x.lot.id} item={x} compatible={matches(x)} />
            ))}
          </div>
        )}
      </section>
      <section className="content-section">
        <div className="section-heading">
          <div>
            <h2>Price Board snapshot</h2>
            <p>Historical/demo records only — not live Indian market rates.</p>
          </div>
          <Link className="text-link" to="/prices">
            Open Price Board
          </Link>
        </div>
        {prices && (
          <>
            <div className="price-board-grid">
              {prices.entries.slice(0, 4).map((e) => (
                <div
                  className="price-board-card"
                  key={`${e.materialCategory}-${e.materialSubCategory}-${e.city}-${e.state}-${e.unit}-${e.sourceType}`}
                >
                  <div>
                    <strong>
                      {e.materialSubCategory || e.materialCategory}
                    </strong>
                    <span>
                      {e.city}, {e.state} · {e.isDemo ? "DEMO" : ""}
                    </span>
                  </div>
                  <b>
                    ₹{e.currentBuyingRate}
                    <small>/{e.unit}</small>
                  </b>
                </div>
              ))}
            </div>
            <p className="price-board-notice">{prices.notice}</p>
          </>
        )}
      </section>
      <div className="dashboard-lower">
        <section className="compact-panel">
          <div className="section-heading">
            <div>
              <h2>Offers requiring attention</h2>
              <p>Offers currently waiting for collector action.</p>
            </div>
            <Link className="text-link" to="/offers">
              View all
            </Link>
          </div>
          {offers.length === 0 ? (
            <div className="compact-empty">No pending offers.</div>
          ) : (
            offers.slice(0, 3).map((o) => (
              <div className="offer-mini" key={o.offer.id}>
                <div>
                  <strong>{o.materialName}</strong>
                  <span>{o.collectorName}</span>
                </div>
                <div>
                  <strong>
                    ₹{o.offer.offeredPricePerUnit}/{o.offer.unit}
                  </strong>
                  <span>
                    {o.offer.approximateWeight} {o.offer.unit} · pending
                  </span>
                </div>
              </div>
            ))
          )}
        </section>
        <section className="compact-panel muted-panel">
          <div>
            <span className="eyebrow">
              <ShieldCheck size={14} /> Traceability
            </span>
            <h2>Digital handover records</h2>
            <p>
              Accepted transactions can progress to recycler-confirmed handovers
              with an auditable status history.
            </p>
          </div>
          <Link className="text-link" to="/handovers">
            Handover records →
          </Link>
        </section>
      </div>
    </div>
  );
}
