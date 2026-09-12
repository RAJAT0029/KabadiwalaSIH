import {
  ArrowDownRight,
  ArrowRight,
  ArrowUpRight,
  Minus,
  RefreshCw,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { api } from "../services/api";
import type { PriceBoard } from "../types";
function Trend({ value }: { value: string }) {
  if (value === "unknown")
    return <span className="trend">Not enough history</span>;
  if (value === "up")
    return (
      <span className="trend up">
        <ArrowUpRight />
        Up
      </span>
    );
  if (value === "down")
    return (
      <span className="trend down">
        <ArrowDownRight />
        Down
      </span>
    );
  return (
    <span className="trend">
      <Minus />
      Stable
    </span>
  );
}
export default function PriceBoardPage() {
  const [data, setData] = useState<PriceBoard | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const load = useCallback(async () => {
    setLoading(true);
    try {
      setData(await api<PriceBoard>("/prices"));
      setError("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unable to load price board");
    } finally {
      setLoading(false);
    }
  }, []);
  useEffect(() => {
    void load();
  }, [load]);
  return (
    <div>
      <div className="page-heading">
        <div>
          <span className="eyebrow">Transparent price dataset</span>
          <h1>Price Board</h1>
          <p>
            Indicative recycler-side buying records with recent ranges and
            simple historical direction.
          </p>
        </div>
        <button className="secondary-button" onClick={() => void load()}>
          <RefreshCw size={17} />
          Refresh
        </button>
      </div>
      <div className="demo-warning">
        <strong>DEVELOPMENT / DEMO PRICE DATA</strong>
        <span>
          These values are not live Indian market prices and must not be
          represented as production quotations.
        </span>
      </div>
      {error && <div className="inline-error">{error}</div>}
      {loading ? (
        <div className="table-loading">Loading price records…</div>
      ) : (
        <div className="price-table">
          {data?.entries.length === 0 && <p>No price records yet.</p>}
          {data?.entries.map((e) => (
            <article
              key={`${e.materialCategory}-${e.materialSubCategory}-${e.city}-${e.state}-${e.unit}-${e.sourceType}`}
              className="price-row"
            >
              <div>
                <span className="category-chip">{e.materialCategory}</span>
                <strong>{e.materialSubCategory || e.materialCategory}</strong>
                <small>
                  {e.city}, {e.state}
                </small>
              </div>
              <div>
                <small>Indicative buying rate</small>
                <strong>
                  ₹{e.currentBuyingRate}/{e.unit}
                </strong>
              </div>
              <div>
                <small>Recent demo range</small>
                <strong>
                  ₹{e.marketRangeMin}–₹{e.marketRangeMax}/{e.unit}
                </strong>
              </div>
              <div>
                <small>Trend</small>
                <Trend value={e.trend} />
              </div>
              <div>
                <small>Last updated</small>
                <strong>{new Date(e.updatedAt).toLocaleString("en-IN")}</strong>
              </div>
              <ArrowRight className="muted-icon" />
            </article>
          ))}
        </div>
      )}
      {data && <p className="price-board-notice">{data.notice}</p>}
    </div>
  );
}
