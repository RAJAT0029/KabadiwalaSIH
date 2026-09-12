import { PackageOpen } from "lucide-react";
export function EmptyState({ title, body }: { title: string; body: string }) {
  return (
    <div className="empty-state">
      <PackageOpen size={30} />
      <h3>{title}</h3>
      <p>{body}</p>
    </div>
  );
}
