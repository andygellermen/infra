import { useTransientControls } from "../hooks/useTransientControls";

const VIEW_LABELS = { clean: "Clean & Free", intense: "Intense", review: "Review" };

export default function TransientControlBar({
  book,
  chapter,
  workView = "clean",
  saveState = "",
  blocked = false,
  onWorkView,
  onAppearance,
  onSettings,
  onCommand,
  editorMode = "rich",
  onEditorMode,
  onSave,
  saveDisabled = false,
  onWritingTools,
  onHelp,
  onFullscreen,
  onKanban,
  isFullscreen = false,
}) {
  const controls = useTransientControls({ timeoutMs: 60000, blocked });
  return (
    <div className="transient-controls" onMouseMove={controls.reveal}>
      <div className="transient-controls__edge" onPointerEnter={controls.reveal} aria-hidden="true" />
      <button className="transient-controls__trigger" type="button" aria-label="Alle Bedienelemente anzeigen" onClick={controls.reveal}>•••</button>
      {controls.visible ? (
        <div
          className="transient-controls__bar"
          role="toolbar"
          aria-label="Schreibsteuerung"
          onPointerEnter={controls.onPointerEnter}
          onPointerLeave={controls.onPointerLeave}
          onFocus={controls.onFocusCapture}
          onBlur={controls.onBlurCapture}
          onClick={controls.resetTimer}
          onKeyDown={(event) => {
            if (event.key === "Escape") controls.dismiss();
          }}
        >
          <div className="transient-controls__location"><strong>{book?.title || "Buch"}</strong><span>{chapter?.title || "Kapitel"}</span></div>
          <button type="button" onClick={() => onWorkView?.(workView)}>{VIEW_LABELS[workView] || VIEW_LABELS.clean} · ▾</button>
          <div className="transient-controls__actions">
            {saveState ? <span aria-label={`Speicherstatus: ${saveState}`}>{saveState}</span> : null}
            <button type="button" aria-label="Kapitel speichern" onClick={onSave} disabled={saveDisabled}>💾</button>
            <button type="button" aria-label="Rich" aria-pressed={editorMode === "rich"} onClick={() => onEditorMode?.("rich")}>✍</button>
            <button type="button" aria-label="Markdown" aria-pressed={editorMode === "markdown"} onClick={() => onEditorMode?.("markdown")}>#</button>
            <button type="button" aria-label="Werkzeuge" onClick={onWritingTools}>✚</button>
            <button type="button" aria-label="Kanban öffnen" onClick={onKanban}>▥</button>
            <button type="button" aria-label="Hilfe" onClick={onHelp}>?</button>
            <button type="button" aria-label={isFullscreen ? "Vollbild verlassen" : "Vollbild"} aria-pressed={isFullscreen} onClick={onFullscreen}>⛶</button>
            <button type="button" aria-label="Befehlspalette öffnen" onClick={onCommand}>⌘</button>
            <button type="button" aria-label="Erscheinungsbild öffnen" onClick={onAppearance}>◐</button>
            <button type="button" aria-label="Einstellungen öffnen" onClick={onSettings}>⚙</button>
          </div>
        </div>
      ) : null}
    </div>
  );
}
