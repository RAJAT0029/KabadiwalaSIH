import { CheckCircle2, XCircle } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { useLocation } from "react-router-dom";
import { EmptyState } from "../components/EmptyState";
import { StatusBadge } from "../components/StatusBadge";
import { api } from "../services/api";
import type { OfferView } from "../types";
const tabs = ["all", "pending", "accepted", "rejected", "expired", "cancelled"];
export default function OffersPage() {
  const loc = useLocation();
  const [tab, setTab] = useState("all");
  const [items, setItems] = useState<OfferView[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(
    (loc.state as { created?: boolean } | null)?.created
      ? "Offer sent successfully. It is stored in the shared backend for collector-side review."
      : "",
  );
  const load = useCallback(async () => {
    setLoading(true);
    try {
      const d = await api<{ items: OfferView[] }>(`/offers?status=${tab}`);
      setItems(d.items);
      setError("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unable to load offers");
    } finally {
      setLoading(false);
    }
  }, [tab]);
  useEffect(() => {
    void load();
  }, [load]);
  const cancel = async (id: string) => {
    try {
      await api(`/offers/${id}/cancel`, { method: "PATCH" });
      setSuccess("Offer cancelled.");
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not cancel offer");
    }
  };
  return (
    <div>
      <div className="page-heading">
        <div>
          <span className="eyebrow">Recycler negotiations</span>
          <h1>My Offers</h1>
          <p>Track offers sent against collector-created e-waste lots.</p>
        </div>
      </div>
      {success && (
        <div className="success-banner">
          <CheckCircle2 />
          {success}
          <button
            aria-label="Dismiss success message"
            onClick={() => setSuccess("")}
          >
            <XCircle />
          </button>
        </div>
      )}
      {error && <div className="inline-error">{error}</div>}
      <div className="tabs" aria-label="Filter offers by status">
        {tabs.map((t) => (
          <button
            key={t}
            aria-pressed={tab === t}
            onClick={() => setTab(t)}
            className={tab === t ? "active" : ""}
          >
            {t}
          </button>
        ))}
      </div>
      <div className="offers-table-wrap">
        {loading ? (
          <div className="table-loading">Loading offers…</div>
        ) : items.length === 0 ? (
          <EmptyState
            title="No offers here"
            body={
              tab === "all"
                ? "You haven't sent any e-waste offers yet. Open a lot to start negotiating."
                : `You don't have any ${tab} offers.`
            }
          />
        ) : (
          <>
            <div className="offers-table header">
              <span>E-waste lot / Collector</span>
              <span>Approx. quantity</span>
              <span>Indicative rate</span>
              <span>Your offer</span>
              <span>Total</span>
              <span>Status</span>
              <span />
            </div>
            {items.map((v) => (
              <div className="offers-table row" key={v.offer.id}>
                <span className="offer-material">
                  <strong>{v.materialName}</strong>
                  <small>{v.collectorName}</small>
                </span>
                <span data-label="Quantity">
                  {v.offer.approximateWeight} {v.offer.unit}
                </span>
                <span data-label="Indicative rate">
                  ₹{v.estimatedPricePerUnit}/{v.unit}
                </span>
                <span data-label="Your offer">
                  <strong>
                    ₹{v.offer.offeredPricePerUnit}/{v.offer.unit}
                  </strong>
                </span>
                <span data-label="Total">
                  ₹{v.offer.offeredTotal.toLocaleString("en-IN")}
                </span>
                <span data-label="Status">
                  <StatusBadge status={v.offer.status} />
                </span>
                <span className="offer-actions">
                  {v.offer.status === "pending" ? (
                    <button
                      className="link-button danger"
                      onClick={() => cancel(v.offer.id)}
                    >
                      Cancel
                    </button>
                  ) : null}
                </span>
              </div>
            ))}
          </>
        )}
      </div>
    </div>
  );
}
