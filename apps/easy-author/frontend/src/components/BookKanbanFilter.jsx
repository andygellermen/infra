import { useEffect, useMemo, useState } from "react";
import { allocateBookColor } from "../lib/bookColorAllocation";

export default function BookKanbanFilter({ books = [], selectedIds = [], counts = {}, colors, onToggle, onColorsChange }) {
  const [assignments, setAssignments] = useState(() => new Map(colors || []));
  const selectionKey = selectedIds.join("|");

  useEffect(() => {
    setAssignments((previous) => {
      let next = new Map([...previous].filter(([id]) => selectedIds.includes(id)));
      selectedIds.forEach((id) => { next = allocateBookColor(next, id); });
      return next;
    });
  }, [selectionKey]);

  useEffect(() => {
    onColorsChange?.(assignments);
  }, [assignments]);

  const selected = useMemo(() => new Set(selectedIds), [selectionKey]);
  return (
    <aside className="book-kanban-filter" aria-label="Bücher filtern">
      <div className="book-kanban-filter__heading"><strong>Bücher</strong><span>{selectedIds.length} gewählt</span></div>
      <div className="book-kanban-filter__list">
        {books.map((book) => {
          const active = selected.has(book.id);
          const value = counts[book.id] || {};
          return (
            <article className={`book-kanban-filter__book ${active ? "is-selected" : ""}`} key={book.id}>
              <button type="button" className="book-kanban-filter__toggle" aria-pressed={active}
                aria-label={`${book.title} ${active ? "aus Auswahl entfernen" : "zur Auswahl hinzufügen"}`}
                onClick={() => onToggle?.(book.id)}>
                <span className="book-kanban-filter__circle" style={active ? { backgroundColor: assignments.get(book.id) } : undefined} aria-hidden="true" />
                <strong>{book.title}</strong>
              </button>
              <div className="book-kanban-filter__counts">
                <span aria-label={`${book.title}: offen`}><b>{(value.backlog || 0) + (value.todo || 0)}</b> offen</span>
                <span aria-label={`${book.title}: in Arbeit`}><b>{(value.in_progress || 0) + (value.review || 0)}</b> aktiv</span>
                <span aria-label={`${book.title}: fertig`}><b>{value.done || 0}</b> fertig</span>
              </div>
            </article>
          );
        })}
      </div>
    </aside>
  );
}
