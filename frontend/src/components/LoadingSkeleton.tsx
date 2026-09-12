export function LoadingSkeleton({ count = 3 }: { count?: number }) {
  return (
    <div className="skeleton-grid">
      {Array.from({ length: count }, (_, i) => (
        <div className="skeleton-card" key={i}>
          <div className="skeleton-media" />
          <div className="skeleton-line wide" />
          <div className="skeleton-line" />
          <div className="skeleton-line small" />
        </div>
      ))}
    </div>
  );
}
