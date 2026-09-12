import {
  AlertTriangle,
  ArrowLeft,
  BadgeCheck,
  BatteryCharging,
  CalendarDays,
  Check,
  Cpu,
  Database,
  IndianRupee,
  MapPin,
  PackageCheck,
  Scale,
  Send,
  ShieldCheck,
  X,
} from "lucide-react";
import {
  useEffect,
  useMemo,
  useState,
  type FormEvent,
  type ReactNode,
} from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useModalFocus } from "../components/useModalFocus";
import { api } from "../services/api";
import { useAuth } from "../state/auth-context";
import type { LotView } from "../types";

function attributeText(value: unknown) {
  if (typeof value === "boolean") return value ? "Yes" : "No";
  if (typeof value === "string" || typeof value === "number")
    return String(value);
  return "";
}
function Attribute({
  icon,
  label,
  value,
}: {
  icon: ReactNode;
  label: string;
  value: unknown;
}) {
  const text = attributeText(value);
  if (!text) return null;
  return (
    <div className="ewaste-attribute">
      <span>{icon}</span>
      <div>
        <small>{label}</small>
        <strong>{text}</strong>
      </div>
    </div>
  );
}
export default function ListingDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [item, setItem] = useState<LotView | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [modal, setModal] = useState(false);
  const [quantity, setQuantity] = useState("");
  const [price, setPrice] = useState("");
  const [message, setMessage] = useState("");
  const [pickup, setPickup] = useState(user?.pickup?.available || false);
  const [sending, setSending] = useState(false);
  const [offerError, setOfferError] = useState("");
  const total = useMemo(
    () => Number(quantity || 0) * Number(price || 0),
    [quantity, price],
  );
  const modalRef = useModalFocus<HTMLFormElement>(modal, () => setModal(false));
  useEffect(() => {
    if (!id) return;
    const controller = new AbortController();
    setLoading(true);
    setError("");
    setModal(false);
    api<LotView>(`/lots/${id}`, { signal: controller.signal })
      .then((d) => {
        setItem(d);
        setPrice(String(d.lot.valuation.estimatedPricePerUnit));
        setQuantity(String(Math.min(50, d.lot.quantity.availableWeight)));
      })
      .catch(
        (e) =>
          !controller.signal.aborted &&
          setError(e instanceof Error ? e.message : "Unable to load lot"),
      )
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    return () => controller.abort();
  }, [id]);
  const send = async (e: FormEvent) => {
    e.preventDefault();
    if (!id) return;
    setSending(true);
    setOfferError("");
    try {
      await api(`/lots/${id}/offers`, {
        method: "POST",
        body: JSON.stringify({
          approximateWeight: Number(quantity),
          offeredPricePerUnit: Number(price),
          pickupAvailable: pickup,
          message,
        }),
      });
      setModal(false);
      navigate("/offers", { state: { created: true } });
    } catch (e) {
      setOfferError(e instanceof Error ? e.message : "Could not send offer");
    } finally {
      setSending(false);
    }
  };
  if (loading)
    return <div className="detail-loading">Loading e-waste lot…</div>;
  if (error || !item)
    return (
      <div className="empty-state">
        <h3>Lot unavailable</h3>
        <p>{error || "This e-waste lot could not be found."}</p>
        <Link to="/marketplace" className="primary-button">
          Back to E-Waste Lots
        </Link>
      </div>
    );
  const { lot: l, collector } = item;
  const a = l.material.attributes || {};
  const authStatus =
    user?.effectiveAuthorizationStatus || "pending_verification";
  const authorized = user?.participationAllowed === true;
  const materialCompatible =
    user?.acceptedMaterials?.some(
      (x) =>
        x.toLowerCase() === l.material.category.toLowerCase() ||
        x.toLowerCase() === (l.material.subCategory || "").toLowerCase(),
    ) || false;
  const canOffer =
    authorized &&
    materialCompatible &&
    ["available", "offer_received"].includes(l.status) &&
    l.quantity.availableWeight > 0;
  const hazardous =
    l.material.category === "Battery" ||
    l.material.category === "CRT" ||
    a.hazardousHandling === true;
  const image = l.imageReferences?.[0]?.reference || l.images?.[0];
  const unit = l.quantity.unit;
  return (
    <div>
      <Link className="back-link" to="/marketplace">
        <ArrowLeft size={17} />
        Back to E-Waste Lots
      </Link>
      <div className="detail-grid">
        <section className="gallery-panel">
          <div className="main-image">
            {image ? (
              <img
                src={image}
                alt={`${l.material.subCategory || l.material.category} e-waste lot`}
              />
            ) : (
              <div className="detail-placeholder">
                <Cpu size={46} />
                <strong>Collector lot photograph</strong>
                <span>No image in development seed data</span>
              </div>
            )}
          </div>
          <div className="trust-strip">
            <span>
              <ShieldCheck />
              Authorization checked on offers
            </span>
            <span>
              <PackageCheck />
              Lot availability tracked
            </span>
            <span>
              <IndianRupee />
              Server-calculated totals
            </span>
          </div>
        </section>
        <section className="detail-info">
          <div className="detail-kicker">
            <span>{l.material.category}</span>
            <span className="negotiable">{l.status.replaceAll("_", " ")}</span>
            {hazardous && (
              <span className="hazard-inline">
                <AlertTriangle size={14} />
                Hazard-aware handling
              </span>
            )}
          </div>
          <h1>{l.material.subCategory || l.material.category}</h1>
          <p className="detail-description">
            {l.material.description || "E-waste lot"} ·{" "}
            {l.material.condition || "Condition not specified"}
          </p>
          <div className="reference-row">
            <span>Reference ID</span>
            <strong>{l.referenceId}</strong>
          </div>
          <div className="demo-warning">
            <strong>DEMO / INDICATIVE DEVELOPMENT DATA</strong>
            <span>Not live market prices.</span>
          </div>
          <div className="price-block">
            <strong>₹{l.valuation.estimatedPricePerUnit}</strong>
            <span>
              indicative / {unit} · range ₹{l.valuation.marketRangeMin}–₹
              {l.valuation.marketRangeMax}
            </span>
          </div>
          <div className="detail-facts">
            <div>
              <Scale />
              <span>
                Approx. available
                <strong>
                  {l.quantity.availableWeight} {unit.toUpperCase()}
                </strong>
              </span>
            </div>
            <div>
              <IndianRupee />
              <span>
                Estimated lot value
                <strong>
                  ₹{l.valuation.estimatedValue.toLocaleString("en-IN")}
                </strong>
              </span>
            </div>
            <div>
              <MapPin />
              <span>
                Collection location
                <strong>
                  {l.collection.label ||
                    collector.operatingLocation ||
                    "Provided with lot"}
                </strong>
              </span>
            </div>
            <div>
              <CalendarDays />
              <span>
                Collected / published
                <strong>
                  {new Date(
                    l.collection.collectedAt || l.createdAt,
                  ).toLocaleDateString("en-IN")}
                </strong>
              </span>
            </div>
          </div>
          <div className="ewaste-attributes">
            <h3>Material details</h3>
            <div className="ewaste-attribute-grid">
              <Attribute
                icon={<Cpu size={17} />}
                label="Source type"
                value={l.material.sourceType}
              />
              <Attribute
                icon={<PackageCheck size={17} />}
                label="Condition"
                value={l.material.condition}
              />
              <Attribute
                icon={<BatteryCharging size={17} />}
                label="Battery / chemistry"
                value={a.chemistry || a.batteryType}
              />
              <Attribute
                icon={<Database size={17} />}
                label="Panel / board / cable"
                value={a.boardType || a.panelType || a.cableType}
              />
              <Attribute
                icon={<Cpu size={17} />}
                label="Assembly / motor"
                value={a.assemblyType || a.motorType}
              />
              {Object.entries(a).map(([key, value]) => (
                <Attribute
                  key={key}
                  icon={<Database size={17} />}
                  label={key}
                  value={
                    typeof value === "object" ? JSON.stringify(value) : value
                  }
                />
              ))}
            </div>
            <p>
              Material characteristics are collector-provided/shared data. No
              automatic image classification is claimed in this version.
            </p>
          </div>
          {hazardous && (
            <div className="safety-note">
              <AlertTriangle />
              <div>
                <strong>Formal handling recommended</strong>
                <span>
                  Do not burn cables, open batteries/CRTs, acid-leach PCBs, or
                  perform unsafe backyard dismantling. Transfer through capable
                  recycling channels.
                </span>
              </div>
            </div>
          )}
          <div className="supplier-card">
            <div className="supplier-avatar">
              {collector.businessName.slice(0, 2).toUpperCase()}
            </div>
            <div>
              <span>Collector / aggregator</span>
              <h3>{collector.businessName}</h3>
              <p>
                {collector.operatingLocation || "Operating location available"}{" "}
                · profile: {collector.profileStatus}
              </p>
            </div>
          </div>
          <div className="compatibility-panel">
            <h3>Recycler compatibility</h3>
            <div>
              <span>
                {materialCompatible ? <Check /> : <X />}Accepts{" "}
                {l.material.category}
              </span>
              <span>
                {authorized ? <BadgeCheck /> : <AlertTriangle />}
                {authStatus.replaceAll("_", " ")}
              </span>
              <span>
                {user?.pickup?.available ? <Check /> : <X />}Pickup{" "}
                {user?.pickup?.available ? "available" : "not enabled"}
              </span>
            </div>
            {user?.authorization?.isDemo && (
              <p>
                Development verification is mock data and is not a real
                government authorization.
              </p>
            )}
          </div>
          <div className="detail-actions">
            <button
              className="primary-button"
              onClick={() => setModal(true)}
              disabled={!canOffer}
            >
              <Send size={17} />
              Send Offer
            </button>
          </div>
          {!canOffer && (
            <p className="purchase-note">
              Offer action is disabled because your authorization or
              accepted-material profile is not compatible with this lot. Update
              your profile where appropriate; the backend enforces the same
              rule.
            </p>
          )}
        </section>
      </div>
      {modal && (
        <div className="modal-root">
          <button
            className="modal-backdrop"
            onClick={() => setModal(false)}
            aria-label="Close offer form"
          />
          <form
            className="offer-modal"
            ref={modalRef}
            role="dialog"
            aria-modal="true"
            aria-label="Send recycler offer"
            onSubmit={send}
          >
            <div className="modal-title">
              <div>
                <span className="eyebrow">Recycler purchase offer</span>
                <h2>
                  Offer for {l.material.subCategory || l.material.category}
                </h2>
                <p>{collector.businessName}</p>
              </div>
              <button
                type="button"
                className="icon-button"
                aria-label="Close offer form"
                onClick={() => setModal(false)}
              >
                <X />
              </button>
            </div>
            {offerError && <div className="form-error">{offerError}</div>}
            <div className="form-grid">
              <label>
                Approx. quantity ({unit})
                <input
                  type="number"
                  min="0.01"
                  max={l.quantity.availableWeight}
                  step={unit === "unit" ? 1 : 0.01}
                  value={quantity}
                  onChange={(e) => setQuantity(e.target.value)}
                  required
                />
                <small>
                  {l.quantity.availableWeight} {unit} available
                </small>
              </label>
              <label>
                Offered rate / {unit}
                <input
                  type="number"
                  min="0.01"
                  step="0.01"
                  value={price}
                  onChange={(e) => setPrice(e.target.value)}
                  required
                />
                <small>
                  Indicative reference ₹{l.valuation.estimatedPricePerUnit}/
                  {unit}
                </small>
              </label>
              <label className="span-2 terms-check">
                <input
                  type="checkbox"
                  checked={pickup}
                  onChange={(e) => setPickup(e.target.checked)}
                />
                <span>Recycler can arrange pickup for this offer</span>
              </label>
              <label className="span-2">
                Message <small>optional</small>
                <textarea
                  maxLength={2000}
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  placeholder="Can arrange pickup after material inspection."
                  rows={3}
                />
              </label>
            </div>
            <div className="offer-total">
              <span>Estimated offered total</span>
              <strong>
                ₹{total.toLocaleString("en-IN", { maximumFractionDigits: 2 })}
              </strong>
              <small>Recalculated by server from quantity × offered rate</small>
            </div>
            <div className="modal-actions">
              <button
                type="button"
                className="secondary-button"
                onClick={() => setModal(false)}
              >
                Cancel
              </button>
              <button className="primary-button" disabled={sending}>
                {sending ? (
                  "Sending…"
                ) : (
                  <>
                    <Check size={17} />
                    Send Offer
                  </>
                )}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}
