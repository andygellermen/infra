const CONTEXT_META = {
  comment: { icon: "◌", label: "Kommentar" },
  work_item: { icon: "✓", label: "Aufgabe" },
  link: { icon: "↗", label: "Verknüpfung" },
  clipboard_insert: { icon: "↧", label: "Clipboard-Einfügung" },
  clipboard_source: { icon: "▣", label: "Clipboard-Quelle" },
};

export function groupContextsByAnchor(contexts = []) {
  const groups = new Map();
  contexts.forEach((context) => {
    const anchorId = context.anchor_id || context.anchor?.id;
    if (!anchorId) return;
    if (!groups.has(anchorId)) {
      groups.set(anchorId, { anchorId, anchor: context.anchor || {}, contexts: [] });
    }
    groups.get(anchorId).contexts.push(context);
  });
  return [...groups.values()].sort((left, right) =>
    (left.anchor.start_offset || 0) - (right.anchor.start_offset || 0) || left.anchorId.localeCompare(right.anchorId));
}

export default function ContextRail({ contexts = [], onActivate }) {
  return (
    <aside className="context-rail" aria-label="Textkontexte">
      {groupContextsByAnchor(contexts).map((group) => {
        const labels = group.contexts.map((item) => CONTEXT_META[item.context_type]?.label || item.context_type);
        const resolved = group.contexts.every((item) => item.status === "resolved");
        const countLabel = group.contexts.length === 1 ? "1 Kontext" : `${group.contexts.length} Kontexte`;
        const accessibleLabel = `${countLabel}: ${labels.join(", ")}${resolved ? " (erledigt)" : ""}`;
        const activate = () => onActivate?.(group);
        return (
          <button
            type="button"
            key={group.anchorId}
            className={`context-rail__group ${resolved ? "is-resolved" : ""}`}
            aria-label={accessibleLabel}
            title={labels.join(" · ")}
            onClick={activate}
            onKeyDown={(event) => {
              if (event.key === "Enter" || event.key === " ") {
                event.preventDefault();
                activate();
              }
            }}
          >
            <span className="context-rail__icons" aria-hidden="true">
              {group.contexts.slice(0, 3).map((item) => <span key={item.id}>{CONTEXT_META[item.context_type]?.icon || "•"}</span>)}
            </span>
            {group.contexts.length > 1 ? <span className="context-rail__count">{group.contexts.length}</span> : null}
          </button>
        );
      })}
    </aside>
  );
}
