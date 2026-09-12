import { AlertTriangle, CheckCircle2, Clock3, Cpu, MapPin } from "lucide-react";
import { Link } from "react-router-dom";
import { StatusBadge } from "./StatusBadge";
import type { LotView } from "../types";

function age(iso: string) {
  const h = Math.max(
    1,
    Math.round((Date.now() - new Date(iso).getTime()) / 36e5),
  );
  return h < 24 ? `${h}h ago` : `${Math.round(h / 24)}d ago`;
}
function hazardous(item: LotView) {
  return (
    item.lot.material.category === "Battery" ||
    item.lot.material.category === "CRT" ||
    item.lot.material.attributes?.hazardousHandling === true
  );
}

export function EWasteLotCard({
  item,
  compatible = false,
}: {
  item: LotView;
  compatible?: boolean;
}) {
  const { lot, collector, distanceKm } = item;
  const title = lot.material.subCategory || lot.material.category;
  const image = lot.imageReferences?.[0]?.reference || lot.images?.[0];
  const location =
    lot.collection.label ||
    collector.operatingLocation ||
    "Location available on lot";
  return (
    <article className="listing-card">
      <div className="listing-image-wrap">
        {image ? (
          <img
            src={image}
            alt={`${title} e-waste lot`}
            className="listing-image"
          />
        ) : (
          <div className="listing-image placeholder ewaste-placeholder">
            <Cpu size={34} />
            <span>E-waste lot</span>
          </div>
        )}
        <span className="material-pill">{lot.material.category}</span>
        {hazardous(item) && (
          <span className="hazard-pill">
            <AlertTriangle size={12} />
            Hazard-aware handling
          </span>
        )}
      </div>
      <div className="listing-body">
        <div className="listing-title-row">
          <div>
            <h3>{title}</h3>
            <p>
              {lot.material.description ||
                lot.material.condition ||
                "End-of-life electrical/electronic material"}
            </p>
          </div>
          <strong className="listing-price">
            ₹{lot.valuation.estimatedPricePerUnit}
            <small>/{lot.quantity.unit}</small>
          </strong>
        </div>
        <div className="listing-meta">
          <span>
            <strong>{lot.quantity.availableWeight}</strong>{" "}
            {lot.quantity.unit.toUpperCase()} approx.
          </span>
          <span
            className="negotiable"
            title="Indicative market range (development data)"
          >
            ₹{lot.valuation.marketRangeMin}–₹{lot.valuation.marketRangeMax}/
            {lot.quantity.unit}
          </span>
        </div>
        <div className="supplier-line">
          <strong>{collector.businessName}</strong>
          <span>
            <MapPin size={14} />
            {distanceKm != null ? `${distanceKm} km` : location}
          </span>
        </div>
        {compatible && (
          <div className="compatibility-line">
            <CheckCircle2 size={14} />
            Matches your accepted material profile
          </div>
        )}
        <StatusBadge status={lot.status} />
        <div className="listing-footer">
          <span>
            <Clock3 size={14} />
            {age(lot.createdAt)}
          </span>
          <Link className="text-link" to={`/marketplace/${lot.id}`}>
            View lot →
          </Link>
        </div>
      </div>
    </article>
  );
}
