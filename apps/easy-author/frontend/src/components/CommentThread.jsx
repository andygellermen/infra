import { useState } from "react";

export default function CommentThread({ thread, onReply, onStatusChange, onDelete, onClose }) {
  const [reply, setReply] = useState("");
  const resolved = thread?.status === "resolved";
  const submitReply = (event) => {
    event.preventDefault();
    const body = reply.trim();
    if (!body) return;
    onReply?.(body);
    setReply("");
  };
  return (
    <section className="comment-thread" role="dialog" aria-label="Kommentar-Thread">
      <header className="comment-thread__header">
        <div><strong>Kommentar</strong><span>{resolved ? "Erledigt" : "Offen"}</span></div>
        {onClose ? <button type="button" className="ghost-button" aria-label="Thread schließen" onClick={onClose}>×</button> : null}
      </header>
      <div className="comment-thread__messages" aria-live="polite">
        {(thread?.messages || []).map((message) => (
          <article key={message.id} className="comment-thread__message">
            <strong>{message.author || "Autor"}</strong>
            <p>{message.body}</p>
          </article>
        ))}
      </div>
      <form className="comment-thread__reply" onSubmit={submitReply}>
        <label><span>Antwort</span><textarea aria-label="Antwort" rows="3" value={reply} onChange={(event) => setReply(event.target.value)} /></label>
        <button type="submit" className="secondary-button">Antwort senden</button>
      </form>
      <footer className="comment-thread__actions">
        <button type="button" className="ghost-button" onClick={() => onStatusChange?.(resolved ? "open" : "resolved")}>{resolved ? "Wieder öffnen" : "Als erledigt markieren"}</button>
        <button type="button" className="ghost-button danger" onClick={onDelete}>Kommentar löschen</button>
      </footer>
    </section>
  );
}
