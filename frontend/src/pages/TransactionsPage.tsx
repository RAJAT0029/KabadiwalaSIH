import { ArrowRight, History } from "lucide-react";
import { useEffect, useState } from "react";
import { EmptyState } from "../components/EmptyState";
import { StatusBadge } from "../components/StatusBadge";
import { api } from "../services/api";
import type { Transaction } from "../types";
export default function TransactionsPage() {
  const [items, setItems] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  useEffect(() => {
    api<{ items: Transaction[] }>("/transactions")
      .then((d) => setItems(d.items ?? []))
      .catch((e) =>
        setError(
          e instanceof Error ? e.message : "Unable to load transactions",
        ),
      )
      .finally(() => setLoading(false));
  }, []);
  return (
    <div>
      <div className="page-heading">
        <div>
          <span className="eyebrow">Traceable commercial records</span>
          <h1>Transactions / Orders</h1>
          <p>
            Accepted offers become immutable transaction snapshots. Collector
            acceptance is an external shared-backend integration step and is not
            implemented in this recycler UI.
          </p>
        </div>
      </div>
      {error && <div className="inline-error">{error}</div>}
      {loading ? (
        <div className="table-loading">Loading transactions…</div>
      ) : items.length === 0 ? (
        <EmptyState
          title="No transactions yet"
          body="A transaction will appear here after a collector accepts a recycler offer and the shared backend creates the transaction snapshot."
        />
      ) : (
        <div className="record-list">
          {items.map((t) => (
            <article className="record-card" key={t.id}>
              <div>
                <small>{t.reference}</small>
                <h3>
                  {t.materialSnapshot.material.subCategory ||
                    t.materialSnapshot.material.category}
                </h3>
                <span>
                  {t.approximateQuantity} {t.unit} · ₹
                  {t.quotedTotal.toLocaleString("en-IN")}
                </span>
              </div>
              <div>
                <StatusBadge status={t.status} />
                <span>Payment: {t.paymentStatus}</span>
              </div>
              <div className="record-history">
                <History size={16} />
                {t.statusHistory.length} trace events
              </div>
              <ArrowRight />
            </article>
          ))}
        </div>
      )}
    </div>
  );
}
