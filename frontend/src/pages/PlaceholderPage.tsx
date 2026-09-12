import { Construction } from "lucide-react";
import { useLocation } from "react-router-dom";
const copy: Record<string, { title: string; body: string }> = {
  requirements: {
    title: "Material Requirements",
    body: "Planned recycler module for publishing e-waste categories, approximate quantities, preferred rates, pickup/service area and validity. The shared contract will be finalized before collector-app matching is implemented.",
  },
  collectors: {
    title: "Collectors",
    body: "Planned recycler-visible relationship view for collectors/aggregators previously involved in transactions. It will expose only transaction-relevant business/location information, not unnecessary personal data.",
  },
  help: {
    title: "Help",
    body: "Recycler help content is planned. Collector vernacular, offline and safety UX remains owned by the collector mobile application.",
  },
};
export default function PlaceholderPage() {
  const name = useLocation().pathname.split("/")[1] || "help";
  const c = copy[name] || {
    title: "Planned module",
    body: "This module is intentionally not implemented in the current verified vertical slice.",
  };
  return (
    <div className="placeholder-page">
      <Construction />
      <span className="eyebrow">Planned V1 module</span>
      <h1>{c.title}</h1>
      <p>{c.body}</p>
    </div>
  );
}
