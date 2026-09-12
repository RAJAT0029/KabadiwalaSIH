import { Grid2X2, List, Search, SlidersHorizontal, X } from "lucide-react";
import { useEffect, useState } from "react";
import { EmptyState } from "../components/EmptyState";
import { EWasteLotCard } from "../components/EWasteLotCard";
import { LoadingSkeleton } from "../components/LoadingSkeleton";
import { useModalFocus } from "../components/useModalFocus";
import { api } from "../services/api";
import { useAuth } from "../state/auth-context";
import type { LotView } from "../types";

const categories = [
  "CRT",
  "LCD Panel",
  "PCB",
  "Cable",
  "Battery",
  "Motor",
  "Magnet-bearing Assembly",
  "Mixed Plastics from EEE",
  "Other E-Waste",
];
const conditions = [
  "Dismantled",
  "Damaged",
  "Mixed condition",
  "Loose / sorted",
  "For recovery",
  "Sorted fractions",
];

export default function MarketplacePage() {
  const { user } = useAuth();
  const [items, setItems] = useState<LotView[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState("");
  const [subCategory, setSubCategory] = useState("");
  const [condition, setCondition] = useState("");
  const [sort, setSort] = useState("newest");
  const [maxPrice, setMaxPrice] = useState("");
  const [mobileFilters, setMobileFilters] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const filterRef = useModalFocus<HTMLDivElement>(mobileFilters, () =>
    setMobileFilters(false),
  );
  useEffect(() => {
    setPage(1);
  }, [search, category, subCategory, condition, sort, maxPrice]);
  useEffect(() => {
    const controller = new AbortController();
    const t = setTimeout(() => {
      setLoading(true);
      const p = new URLSearchParams({ page: String(page), limit: "20", sort });
      if (search) p.set("search", search);
      if (category) p.set("category", category);
      if (subCategory) p.set("subCategory", subCategory);
      if (condition) p.set("condition", condition);
      if (maxPrice) p.set("maxPrice", maxPrice);
      api<{ items: LotView[]; pagination: { total: number } }>(`/lots?${p}`, {
        signal: controller.signal,
      })
        .then((d) => {
          setItems(d.items);
          setTotal(d.pagination.total);
          setError("");
        })
        .catch(
          (e) =>
            !controller.signal.aborted &&
            setError(
              e instanceof Error ? e.message : "Unable to load e-waste lots",
            ),
        )
        .finally(() => {
          if (!controller.signal.aborted) setLoading(false);
        });
    }, 250);
    return () => {
      clearTimeout(t);
      controller.abort();
    };
  }, [page, search, category, subCategory, condition, sort, maxPrice]);
  const clear = () => {
    setCategory("");
    setSubCategory("");
    setCondition("");
    setMaxPrice("");
  };
  const compatible = (x: LotView) =>
    user?.acceptedMaterials?.some(
      (m) =>
        m.toLowerCase() === x.lot.material.category.toLowerCase() ||
        m.toLowerCase() === (x.lot.material.subCategory || "").toLowerCase(),
    ) || false;
  const filters = (
    <div className="filter-stack">
      <div className="filter-title">
        <strong>Filters</strong>
        <button className="link-button" onClick={clear}>
          Clear all
        </button>
      </div>
      <label>
        Material category
        <select value={category} onChange={(e) => setCategory(e.target.value)}>
          <option value="">All e-waste categories</option>
          {categories.map((x) => (
            <option key={x}>{x}</option>
          ))}
        </select>
      </label>
      <label>
        Sub-category
        <input
          value={subCategory}
          onChange={(e) => setSubCategory(e.target.value)}
          placeholder="e.g. Mixed PCBs"
        />
      </label>
      <label>
        Condition
        <select
          value={condition}
          onChange={(e) => setCondition(e.target.value)}
        >
          <option value="">Any condition</option>
          {conditions.map((x) => (
            <option key={x}>{x}</option>
          ))}
        </select>
      </label>
      <label>
        Maximum indicative rate
        <input
          type="number"
          min="0"
          value={maxPrice}
          onChange={(e) => setMaxPrice(e.target.value)}
          placeholder="Any rate"
        />
      </label>
      <div className="filter-note">
        <strong>E-waste scope only</strong>
        <p>
          Ordinary PET bottles, newspaper, cardboard and municipal scrap are
          intentionally excluded from this recycler experience.
        </p>
      </div>
    </div>
  );
  return (
    <div>
      <div className="page-heading marketplace-heading">
        <div>
          <span className="eyebrow">E-waste procurement</span>
          <h1>E-Waste Lots</h1>
          <p>
            Discover collector-created lots from end-of-life electrical and
            electronic equipment.
          </p>
        </div>
      </div>
      <div className="market-toolbar">
        <div className="market-search">
          <Search size={19} />
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search PCB, cable, battery, LCD, location..."
            aria-label="Search e-waste lots"
          />
        </div>
        <button
          className="secondary-button mobile-filter-button"
          onClick={() => setMobileFilters(true)}
        >
          <SlidersHorizontal size={17} />
          Filters
        </button>
        <select
          className="sort-select"
          value={sort}
          onChange={(e) => setSort(e.target.value)}
          aria-label="Sort lots"
        >
          <option value="newest">Newest</option>
          <option value="lowest_price">Lowest indicative rate</option>
          <option value="highest_quantity">Highest quantity</option>
        </select>
        <div className="view-toggle">
          <button className="active" aria-label="Grid view">
            <Grid2X2 size={17} />
          </button>
          <button disabled aria-label="List view">
            <List size={17} />
          </button>
        </div>
      </div>
      <div className="market-layout">
        <aside className="desktop-filters">{filters}</aside>
        <section className="market-results">
          <div className="results-caption">
            <span>
              {loading
                ? "Loading e-waste lots…"
                : `${items.length} e-waste lots shown`}
            </span>
            <span>
              Values are indicative/demo unless field data says otherwise
            </span>
          </div>
          {error && <div className="inline-error">{error}</div>}
          {loading ? (
            <LoadingSkeleton count={6} />
          ) : items.length === 0 ? (
            <EmptyState
              title="No e-waste lots match these filters"
              body="Try widening the material, condition or rate filters."
            />
          ) : (
            <div className="listing-grid">
              {items.map((i) => (
                <EWasteLotCard
                  key={i.lot.id}
                  item={i}
                  compatible={compatible(i)}
                />
              ))}
            </div>
          )}
          {total > 20 && (
            <nav className="pagination" aria-label="Lot pages">
              <button
                className="secondary-button"
                disabled={loading || page === 1}
                onClick={() => setPage((p) => p - 1)}
              >
                Previous
              </button>
              <span>
                Page {page} of {Math.ceil(total / 20)}
              </span>
              <button
                className="secondary-button"
                disabled={loading || page * 20 >= total}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
              </button>
            </nav>
          )}
        </section>
      </div>
      {mobileFilters && (
        <div className="filter-drawer">
          <button
            className="drawer-backdrop"
            onClick={() => setMobileFilters(false)}
            aria-label="Close filters"
          />
          <div
            className="filter-drawer-panel"
            ref={filterRef}
            role="dialog"
            aria-modal="true"
            aria-label="E-waste filters"
          >
            <div className="drawer-title">
              <h2>E-waste filters</h2>
              <button
                className="icon-button"
                aria-label="Close filters"
                onClick={() => setMobileFilters(false)}
              >
                <X />
              </button>
            </div>
            {filters}
            <button
              className="primary-button full"
              onClick={() => setMobileFilters(false)}
            >
              Show results
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
