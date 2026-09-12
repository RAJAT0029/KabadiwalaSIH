import { Check, MapPin, ShieldCheck } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { EmptyState } from "../components/EmptyState";
import { StatusBadge } from "../components/StatusBadge";
import { api } from "../services/api";
import type { HandoverRecord } from "../types";
export default function HandoverRecordsPage() {
  const [items, setItems] = useState<HandoverRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const load = useCallback(async () => {
    setLoading(true);
    try {
      const d = await api<{ items: HandoverRecord[] }>("/handovers");
      setItems(d.items ?? []);
      setError("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unable to load handovers");
    } finally {
      setLoading(false);
    }
  }, []);
  useEffect(() => {
    void load();
  }, [load]);
  const confirm = async (id: string) => {
    try {
      await api(`/handovers/${id}/confirm`, { method: "POST" });
      setSuccess(
        "Recycler handover confirmation recorded in the audit history.",
      );
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unable to confirm handover");
    }
  };
  return (
    <div>
      <div className="page-heading">
        <div>
          <span className="eyebrow">Digital traceability</span>
          <h1>Handover Records</h1>
          <p>
            Recycler-side confirmation is allowed only for handovers assigned to
            the authenticated recycler.
          </p>
        </div>
      </div>
      {success && (
        <div className="success-banner">
          <Check />
          {success}
        </div>
      )}
      {error && <div className="inline-error">{error}</div>}
      {loading ? (
        <div className="table-loading">Loading handover records…</div>
      ) : items.length === 0 ? (
        <EmptyState
          title="No handover records yet"
          body="Handover records are created after an accepted offer becomes a transaction. Collector-side confirmation remains owned by the collector app."
        />
      ) : (
        <div className="handover-grid">
          {items.map((h) => (
            <article className="handover-card" key={h.id}>
              <div className="handover-title">
                <div>
                  <small>{h.handoverReference}</small>
                  <h3>{h.materialLabel}</h3>
                </div>
                <StatusBadge status={h.status} />
              </div>
              <div className="handover-facts">
                <span>
                  <strong>
                    {h.confirmedWeight ?? h.approximateWeight} {h.weightUnit}
                  </strong>
                  weight
                </span>
                <span>
                  <MapPin />
                  {h.locationLabel || "Handover location recorded"}
                </span>
                <span>
                  <ShieldCheck />
                  Collector {h.collectorConfirmation
                    ? "confirmed"
                    : "pending"}{" "}
                  · Recycler {h.recyclerConfirmation ? "confirmed" : "pending"}
                </span>
              </div>
              {!h.recyclerConfirmation &&
                ["pending", "collector_confirmed"].includes(h.status) && (
                  <button
                    className="primary-button"
                    onClick={() => confirm(h.id)}
                  >
                    <Check size={17} />
                    Confirm Handover
                  </button>
                )}
            </article>
          ))}
        </div>
      )}
    </div>
  );
}
