import { CreditCard } from "lucide-react";
import { useEffect, useState } from "react";
import { EmptyState } from "../components/EmptyState";
import { StatusBadge } from "../components/StatusBadge";
import { api } from "../services/api";
import type { Payment } from "../types";
export default function PaymentsPage() {
  const [items, setItems] = useState<Payment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  useEffect(() => {
    api<{ items: Payment[] }>("/payments")
      .then((d) => setItems(d.items ?? []))
      .catch((e) =>
        setError(e instanceof Error ? e.message : "Unable to load payments"),
      )
      .finally(() => setLoading(false));
  }, []);
  return (
    <div>
      <div className="page-heading">
        <div>
          <span className="eyebrow">Transaction visibility</span>
          <h1>Payments</h1>
          <p>
            Recycler-visible payment records. Cash and digital methods are both
            supported by the data contract; no payment credentials are stored.
          </p>
        </div>
      </div>
      {error && <div className="inline-error">{error}</div>}
      {loading ? (
        <div className="table-loading">Loading payments…</div>
      ) : items.length === 0 ? (
        <EmptyState
          title="No payment records"
          body="Payment records will appear after transaction/payment workflow integration. Digital payment is not mandatory."
        />
      ) : (
        <div className="record-list">
          {items.map((p) => (
            <article className="record-card" key={p.id}>
              <CreditCard />
              <div>
                <small>{p.paymentId}</small>
                <h3>
                  ₹{p.amount.toLocaleString("en-IN")} {p.currency}
                </h3>
                <span>
                  {p.paymentMethod} ·{" "}
                  {new Date(p.paidAt || p.createdAt).toLocaleString("en-IN")}
                </span>
              </div>
              <StatusBadge status={p.status} />
            </article>
          ))}
        </div>
      )}
    </div>
  );
}
