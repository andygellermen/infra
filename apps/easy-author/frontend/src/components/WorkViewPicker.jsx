const WORK_VIEWS = [
  { key: "clean", icon: "✦", label: "Clean & Free", description: "Nur Manuskript und minimale Navigation." },
  { key: "intense", icon: "◎", label: "Intense", description: "Recherche, Struktur und Story-Verknüpfungen." },
  { key: "review", icon: "✓", label: "Review", description: "Kommentare, Aufgaben und Prüfhinweise." },
];

export default function WorkViewPicker({ open, currentView = "clean", persistDefault = false, onSelect, onClose }) {
  if (!open) return null;
  return (
    <div className="work-view-picker__backdrop">
      <section className="work-view-picker" role="dialog" aria-modal="true" aria-labelledby="work-view-picker-title">
        <header>
          <div>
            <span className="eyebrow">Arbeitskontext</span>
            <h2 id="work-view-picker-title">Arbeitsansicht wählen</h2>
          </div>
          {!persistDefault && onClose ? <button type="button" aria-label="Auswahl schließen" onClick={onClose}>×</button> : null}
        </header>
        <div className="work-view-picker__cards">
          {WORK_VIEWS.map((view) => (
            <button
              key={view.key}
              type="button"
              className={`work-view-picker__card ${currentView === view.key ? "is-current" : ""}`}
              aria-pressed={currentView === view.key}
              onClick={() => onSelect?.(view.key, { persistDefault })}
            >
              <span className="work-view-picker__icon" aria-hidden="true">{view.icon}</span>
              <strong>{view.label}</strong>
              <span>{view.description}</span>
            </button>
          ))}
        </div>
      </section>
    </div>
  );
}
