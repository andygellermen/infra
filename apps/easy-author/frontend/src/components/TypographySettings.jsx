import { DEFAULT_TYPOGRAPHY, TYPOGRAPHY_RANGES, normalizeTypographyOverrides } from "../lib/typography";

const FIELDS = [
  ["bodyFont", "Fließtext-Schrift", "text"], ["headingFont", "Überschriften-Schrift", "text"],
  ["h1Size", "H1", "number"], ["h2Size", "H2", "number"], ["h3Size", "H3", "number"],
  ["h4Size", "H4", "number"], ["h5Size", "H5", "number"], ["h6Size", "H6", "number"],
  ["bodySize", "Fließtext", "number"], ["quoteSize", "Zitat", "number"], ["tableSize", "Tabelle", "number"],
  ["textWidth", "Textbreite", "number"], ["firstLineIndent", "Erstzeileneinzug", "number"],
  ["lineHeight", "Zeilenabstand", "number"], ["paragraphSpacing", "Absatzabstand", "number"],
];

export default function TypographySettings({
  globalDefaults = DEFAULT_TYPOGRAPHY,
  bookOverrides = {},
  onGlobalChange,
  onBookChange,
  onResetBook,
}) {
  const normalizedBook = normalizeTypographyOverrides(bookOverrides);
  const update = (scope, key, raw, type) => {
    if (scope === "book" && raw === "") {
      const next = { ...normalizedBook };
      delete next[key];
      onBookChange?.(next);
      return;
    }
    const value = type === "number" ? Number(raw) : raw;
    if (scope === "global") onGlobalChange?.({ ...globalDefaults, [key]: value });
    else onBookChange?.({ ...normalizedBook, [key]: value });
  };
  return (
    <section className="typography-settings" aria-label="Typografie">
      <div className="typography-settings__head">
        <div><strong>Schreibatmosphäre</strong><p>Gesamtstandard und Abweichungen für dieses Buch.</p></div>
        <button type="button" className="ghost-button" onClick={() => onResetBook?.({})}>Auf Gesamtstandard zurücksetzen</button>
      </div>
      <div className="typography-settings__grid">
        {FIELDS.map(([key, label, type]) => {
          const range = TYPOGRAPHY_RANGES[key];
          const overridden = Object.hasOwn(normalizedBook, key);
          return (
            <div className="typography-setting" key={key}>
              <label><span>{label} global</span><input type={type} aria-label={`${label} global`} min={range?.[0]} max={range?.[1]} step={key === "lineHeight" ? .05 : 1} value={globalDefaults[key]} onChange={(event) => update("global", key, event.target.value, type)} /></label>
              <label><span>{overridden ? `${label} · abweichend` : `${label} · geerbt (${globalDefaults[key]})`}</span><input type={type} aria-label={`${label} für dieses Buch`} min={range?.[0]} max={range?.[1]} step={key === "lineHeight" ? .05 : 1} value={overridden ? normalizedBook[key] : ""} placeholder={String(globalDefaults[key])} onChange={(event) => update("book", key, event.target.value, type)} /></label>
            </div>
          );
        })}
      </div>
    </section>
  );
}
