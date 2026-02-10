type TimelineItem = {
  id: string;
  title: string;
  timestamp?: string;
  description?: string;
  tone?: "neutral" | "info" | "success" | "warning" | "danger";
};

type TimelineProps = {
  items: TimelineItem[];
  compact?: boolean;
};

export default function Timeline({ items, compact = false }: TimelineProps) {
  return (
    <ol className={`timeline${compact ? " compact" : ""}`}>
      {items.map((item, index) => (
        <li
          key={item.id}
          className={`timeline-item tone-${item.tone ?? "neutral"}`}
          style={{ animationDelay: `${index * 70}ms` }}
        >
          <span className="timeline-dot" aria-hidden="true" />
          <div className="timeline-content">
            <strong>{item.title}</strong>
            <span className="timeline-meta">{item.timestamp ?? "aguardando"}</span>
            {item.description ? <p>{item.description}</p> : null}
          </div>
        </li>
      ))}
    </ol>
  );
}
