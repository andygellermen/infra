export const KANBAN_PHASES = [
  { id: "backlog", label: "Backlog" },
  { id: "todo", label: "To-do" },
  { id: "in_progress", label: "In Arbeit" },
  { id: "review", label: "Prüfung" },
  { id: "done", label: "Fertig" },
];

function Card({ item, phaseIndex, phases, showRibbon, bookTitle, color, onMove, onOpenSource }) {
  const move = (direction) => {
    const target = phases[phaseIndex + direction];
    if (target) onMove?.(item.id, target.id);
  };
  return (
    <article className="kanban-card" role="article" aria-label={item.title} tabIndex="0" draggable
      onDragStart={(event) => { event.dataTransfer?.setData("text/plain", item.id); if (event.dataTransfer) event.dataTransfer.effectAllowed = "move"; }}
      onKeyDown={(event) => {
        if (!event.altKey || !["ArrowLeft", "ArrowRight"].includes(event.key)) return;
        event.preventDefault(); move(event.key === "ArrowRight" ? 1 : -1);
      }}>
      {showRibbon ? <span className="kanban-card__book-ribbon" style={{ borderColor: color }}>{bookTitle}</span> : null}
      <div className="kanban-card__body"><strong>{item.title}</strong>{item.priority ? <small>Priorität: {item.priority}</small> : null}</div>
      <div className="kanban-card__actions">
        <button type="button" aria-label={`${item.title} eine Phase zurück`} disabled={phaseIndex === 0} onClick={() => move(-1)}>←</button>
        <button type="button" aria-label={`Quelle von ${item.title} öffnen`} onClick={() => onOpenSource?.(item)}>Quelle</button>
        <button type="button" aria-label={`${item.title} eine Phase weiter`} disabled={phaseIndex === phases.length - 1} onClick={() => move(1)}>→</button>
      </div>
    </article>
  );
}

export default function KanbanBoard({ phases = KANBAN_PHASES, items = {}, totals = {}, limit = 12, onMove, onLoadMore, onOpenSource,
  selectedBookCount = 1, bookTitles = {}, colors = new Map() }) {
  return (
    <div className="kanban-board" aria-label="Kanban-Board">
      {phases.map((phase, phaseIndex) => {
        const cards = (items[phase.id] || []).slice(0, 20);
        const total = totals[phase.id] || 0;
        return (
          <section key={phase.id} className={`kanban-column kanban-column--${phase.id}`} role="region" aria-label={`${phase.label}, ${total} Vorgänge`}
            onDragOver={(event) => event.preventDefault()}
            onDrop={(event) => { event.preventDefault(); const id = event.dataTransfer?.getData("text/plain"); if (id) onMove?.(id, phase.id); }}>
            <header><h2>{phase.label} <span>{total}</span></h2></header>
            <div className="kanban-column__cards">
              {cards.map((item) => <Card key={item.id} item={item} phaseIndex={phaseIndex} phases={phases}
                showRibbon={selectedBookCount > 1} bookTitle={bookTitles[item.book_id]} color={colors.get?.(item.book_id) || colors[item.book_id]}
                onMove={onMove} onOpenSource={onOpenSource} />)}
              {cards.length === 0 ? <p className="kanban-column__empty">Keine Vorgänge</p> : null}
            </div>
            {cards.length < total && limit < 20 ? <button type="button" className="kanban-column__more"
              aria-label={`Weitere ${phase.label}-Karten anzeigen`} onClick={() => onLoadMore?.(phase.id)}>Weitere anzeigen</button> : null}
          </section>
        );
      })}
    </div>
  );
}
