import { useEffect, useMemo, useRef, useState } from "react";
import EditorPane from "./components/EditorPane";
import SidebarSection from "./components/SidebarSection";
import { api } from "./lib/api";
import { previewText } from "./lib/markdown";

const EMPTY_DRAFT = {
  title: "",
  markdown_content: "",
  editor_json: "",
};

function App() {
  const editorRef = useRef(null);
  const autosaveRef = useRef(null);
  const skipAutosaveRef = useRef(true);

  const [projects, setProjects] = useState([]);
  const [projectDetail, setProjectDetail] = useState(null);
  const [selectedProjectId, setSelectedProjectId] = useState("");
  const [selectedBookId, setSelectedBookId] = useState("");
  const [bookBundle, setBookBundle] = useState(null);
  const [selectedChapterId, setSelectedChapterId] = useState("");
  const [chapterDraft, setChapterDraft] = useState(EMPTY_DRAFT);
  const [anchors, setAnchors] = useState([]);
  const [clipboardItems, setClipboardItems] = useState([]);
  const [selectedWorkflowBoxId, setSelectedWorkflowBoxId] = useState("");
  const [hasSelection, setHasSelection] = useState(false);
  const [saveState, setSaveState] = useState("Synchron");
  const [errorMessage, setErrorMessage] = useState("");

  const currentChapter = useMemo(
    () => bookBundle?.chapters?.find((chapter) => chapter.id === selectedChapterId) || null,
    [bookBundle, selectedChapterId],
  );

  const pinnedSlots = useMemo(
    () =>
      clipboardItems
        .filter((item) => item.is_pinned && item.slot >= 1 && item.slot <= 9)
        .sort((left, right) => left.slot - right.slot),
    [clipboardItems],
  );

  useEffect(() => {
    loadProjects();
  }, []);

  useEffect(() => {
    if (!selectedProjectId) {
      return;
    }
    loadProject(selectedProjectId);
  }, [selectedProjectId]);

  useEffect(() => {
    if (!selectedBookId) {
      setBookBundle(null);
      setSelectedChapterId("");
      setClipboardItems([]);
      return;
    }
    loadBook(selectedBookId);
  }, [selectedBookId]);

  useEffect(() => {
    if (!selectedChapterId) {
      setAnchors([]);
      setChapterDraft(EMPTY_DRAFT);
      return;
    }
    const chapter = bookBundle?.chapters?.find((entry) => entry.id === selectedChapterId);
    if (chapter) {
      skipAutosaveRef.current = true;
      setChapterDraft({
        title: chapter.title,
        markdown_content: chapter.markdown_content || "",
        editor_json: chapter.editor_json || "",
      });
      loadAnchors(chapter.id);
    }
  }, [selectedChapterId, bookBundle]);

  useEffect(() => {
    if (!currentChapter) {
      return undefined;
    }
    if (skipAutosaveRef.current) {
      skipAutosaveRef.current = false;
      return undefined;
    }

    const changed =
      chapterDraft.title !== currentChapter.title ||
      chapterDraft.markdown_content !== (currentChapter.markdown_content || "") ||
      chapterDraft.editor_json !== (currentChapter.editor_json || "");

    if (!changed) {
      setSaveState("Synchron");
      return undefined;
    }

    setSaveState("Autosave ausstehend");
    clearTimeout(autosaveRef.current);
    autosaveRef.current = window.setTimeout(() => {
      saveChapter(false);
    }, 1500);

    return () => clearTimeout(autosaveRef.current);
  }, [chapterDraft, currentChapter]);

  async function loadProjects() {
    try {
      setErrorMessage("");
      const response = await api.get("/api/projects");
      const items = response.projects || [];
      setProjects(items);
      if (!selectedProjectId && items.length > 0) {
        setSelectedProjectId(items[0].id);
      }
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function loadProject(projectId) {
    try {
      const response = await api.get(`/api/projects/${projectId}`);
      setProjectDetail(response);
      const firstBook = response.books?.[0];
      if (!selectedBookId || !response.books?.some((book) => book.id === selectedBookId)) {
        setSelectedBookId(firstBook?.id || "");
      }
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function loadBook(bookId) {
    try {
      const response = await api.get(`/api/books/${bookId}`);
      setBookBundle(response);
      setClipboardItems(response.clipboard || []);
      if (!selectedChapterId || !response.chapters?.some((chapter) => chapter.id === selectedChapterId)) {
        setSelectedChapterId(response.chapters?.[0]?.id || "");
      }
      if (!selectedWorkflowBoxId || !response.workflow_boxes?.some((item) => item.id === selectedWorkflowBoxId)) {
        setSelectedWorkflowBoxId(response.workflow_boxes?.[0]?.id || "");
      }
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function loadAnchors(chapterId) {
    try {
      const response = await api.get(`/api/chapters/${chapterId}/anchors`);
      setAnchors(response.anchors || []);
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function saveChapter(manual) {
    if (!currentChapter) {
      return;
    }
    try {
      setSaveState(manual ? "Speichert ..." : "Autosave laeuft ...");
      const updated = await api.put(`/api/chapters/${currentChapter.id}`, chapterDraft);
      setBookBundle((previous) => ({
        ...previous,
        chapters: previous.chapters.map((chapter) => (chapter.id === updated.id ? updated : chapter)),
      }));
      setSaveState(manual ? "Gespeichert" : "Autosave gespeichert");
      window.setTimeout(() => setSaveState("Synchron"), 1200);
    } catch (error) {
      setSaveState("Fehler beim Speichern");
      setErrorMessage(error.message);
    }
  }

  async function createProject() {
    const title = window.prompt("Titel des neuen Projekts", "Neues Projekt");
    if (!title) {
      return;
    }
    try {
      const project = await api.post("/api/projects", { title, description: "" });
      await loadProjects();
      setSelectedProjectId(project.id);
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function createBook() {
    if (!selectedProjectId) {
      return;
    }
    const title = window.prompt("Titel des neuen Buchs", "Neues Buch");
    if (!title) {
      return;
    }
    try {
      const book = await api.post(`/api/projects/${selectedProjectId}/books`, {
        title,
        author: "",
        visibility: "private",
      });
      await loadProject(selectedProjectId);
      setSelectedBookId(book.id);
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function createChapter() {
    if (!selectedBookId) {
      return;
    }
    const title = window.prompt("Titel des neuen Kapitels", `Kapitel ${(bookBundle?.chapters?.length || 0) + 1}`);
    if (!title) {
      return;
    }
    try {
      const chapter = await api.post(`/api/books/${selectedBookId}/chapters`, {
        title,
        markdown_content: `# ${title}\n`,
        editor_json: "",
      });
      await loadBook(selectedBookId);
      setSelectedChapterId(chapter.id);
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function createWorkflowBox() {
    if (!selectedBookId) {
      return;
    }
    const title = window.prompt("Name der Workflow-Box", "Neue Box");
    if (!title) {
      return;
    }
    try {
      const box = await api.post(`/api/books/${selectedBookId}/workflow-boxes`, {
        title,
        type: "custom",
        is_collapsed: false,
      });
      await loadBook(selectedBookId);
      setSelectedWorkflowBoxId(box.id);
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function updateWorkflowBox(id, nextFields) {
    const box = bookBundle?.workflow_boxes?.find((entry) => entry.id === id);
    if (!box) {
      return;
    }
    try {
      const updated = await api.put(`/api/workflow-boxes/${id}`, {
        title: nextFields.title ?? box.title,
        type: nextFields.type ?? box.type,
        is_collapsed: nextFields.is_collapsed ?? box.is_collapsed,
      });
      setBookBundle((previous) => ({
        ...previous,
        workflow_boxes: previous.workflow_boxes.map((entry) => (entry.id === updated.id ? updated : entry)),
      }));
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function createAnchor() {
    if (!selectedChapterId || !selectedWorkflowBoxId) {
      setErrorMessage("Bitte zuerst ein Kapitel und eine Workflow-Box waehlen.");
      return;
    }
    const payload = editorRef.current?.getSelectionPayload();
    if (!payload?.selected_text) {
      setErrorMessage("Bitte zuerst eine Textpassage im Editor markieren.");
      return;
    }
    const note = window.prompt("Optionale Notiz fuer diesen Anker", "") || "";
    try {
      await api.post(`/api/chapters/${selectedChapterId}/anchors`, {
        ...payload,
        workflow_box_id: selectedWorkflowBoxId,
        anchor_type: "passage",
        title: previewText(payload.selected_text, 40),
        note,
      });
      await loadAnchors(selectedChapterId);
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function deleteAnchor(anchorId) {
    try {
      await api.delete(`/api/anchors/${anchorId}`);
      await loadAnchors(selectedChapterId);
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function createClipboardItem() {
    if (!selectedBookId) {
      return;
    }
    const payload = editorRef.current?.getSelectionPayload();
    if (!payload?.selected_text) {
      setErrorMessage("Bitte zuerst eine Textpassage im Editor markieren.");
      return;
    }
    try {
      const created = await api.post(`/api/books/${selectedBookId}/clipboard`, {
        chapter_id: selectedChapterId,
        content: payload.selected_text,
        content_type: "text/markdown",
        source_anchor_id: "",
        is_pinned: false,
        slot: 0,
      });
      setClipboardItems((previous) => [created, ...previous]);
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function updateClipboard(item, nextFields) {
    try {
      const updated = await api.put(`/api/clipboard/${item.id}`, {
        content: nextFields.content ?? item.content,
        is_pinned: nextFields.is_pinned ?? item.is_pinned,
        slot: nextFields.slot ?? item.slot,
      });
      setClipboardItems((previous) => previous.map((entry) => (entry.id === updated.id ? updated : entry)));
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function deleteClipboard(itemId) {
    try {
      await api.delete(`/api/clipboard/${itemId}`);
      setClipboardItems((previous) => previous.filter((entry) => entry.id !== itemId));
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  const bookTitle = bookBundle?.book?.title || "Kein Buch geladen";

  return (
    <div className="app-shell">
      <header className="topbar">
        <div>
          <div className="brand-kicker">Markdown-first Author Studio</div>
          <h1>easy-author</h1>
        </div>
        <div className="topbar-actions">
          <span className="status-pill">{saveState}</span>
          <button className="primary-button" type="button" onClick={() => saveChapter(true)} disabled={!currentChapter}>
            Kapitel speichern
          </button>
        </div>
      </header>

      {errorMessage ? <div className="error-banner">{errorMessage}</div> : null}

      <main className="workspace-grid">
        <aside className="workspace-panel left-panel">
          <SidebarSection eyebrow="Projekt" title="Arbeitsraum" actionLabel="+ Projekt" onAction={createProject}>
            <div className="pill-list">
              {projects.map((project) => (
                <button
                  key={project.id}
                  type="button"
                  className={`pill-button ${project.id === selectedProjectId ? "active" : ""}`}
                  onClick={() => setSelectedProjectId(project.id)}
                >
                  <strong>{project.title}</strong>
                  <span>{project.description || "Ohne Beschreibung"}</span>
                </button>
              ))}
            </div>
          </SidebarSection>

          <SidebarSection eyebrow="Buch" title={projectDetail?.project?.title || "Noch kein Projekt"} actionLabel="+ Buch" onAction={createBook}>
            <div className="book-stack">
              {(projectDetail?.books || []).map((book) => (
                <button
                  key={book.id}
                  type="button"
                  className={`book-card ${book.id === selectedBookId ? "active" : ""}`}
                  onClick={() => setSelectedBookId(book.id)}
                >
                  <strong>{book.title}</strong>
                  <span>{book.visibility}</span>
                </button>
              ))}
            </div>
          </SidebarSection>

          <SidebarSection eyebrow="Kapitel" title={bookTitle} actionLabel="+ Kapitel" onAction={createChapter}>
            <div className="chapter-list">
              {(bookBundle?.chapters || []).map((chapter) => (
                <button
                  key={chapter.id}
                  type="button"
                  className={`chapter-row ${chapter.id === selectedChapterId ? "active" : ""}`}
                  onClick={() => setSelectedChapterId(chapter.id)}
                >
                  <span className="chapter-index">{String(chapter.position).padStart(2, "0")}</span>
                  <span>{chapter.title}</span>
                </button>
              ))}
            </div>
          </SidebarSection>

          <SidebarSection eyebrow="Workflow" title="Workflow-Boxen" actionLabel="+ Box" onAction={createWorkflowBox}>
            <div className="workflow-list">
              {(bookBundle?.workflow_boxes || []).map((box) => (
                <div
                  key={box.id}
                  className={`workflow-card ${box.id === selectedWorkflowBoxId ? "active" : ""}`}
                  onClick={() => setSelectedWorkflowBoxId(box.id)}
                  role="button"
                  tabIndex={0}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                      setSelectedWorkflowBoxId(box.id);
                    }
                  }}
                >
                  <input
                    value={box.title}
                    onChange={(event) =>
                      setBookBundle((previous) => ({
                        ...previous,
                        workflow_boxes: previous.workflow_boxes.map((entry) =>
                          entry.id === box.id ? { ...entry, title: event.target.value } : entry,
                        ),
                      }))
                    }
                    onBlur={(event) => updateWorkflowBox(box.id, { title: event.target.value })}
                  />
                  <div className="workflow-row">
                    <select value={box.type} onChange={(event) => updateWorkflowBox(box.id, { type: event.target.value })}>
                      <option value="notes">notes</option>
                      <option value="persons">persons</option>
                      <option value="events">events</option>
                      <option value="threads">threads</option>
                      <option value="reminders">reminders</option>
                      <option value="research">research</option>
                      <option value="clipboard">clipboard</option>
                      <option value="custom">custom</option>
                    </select>
                    <label className="checkbox-row">
                      <input
                        type="checkbox"
                        checked={box.is_collapsed}
                        onChange={(event) => updateWorkflowBox(box.id, { is_collapsed: event.target.checked })}
                      />
                      collapsed
                    </label>
                  </div>
                </div>
              ))}
            </div>
          </SidebarSection>
        </aside>

        <section className="editor-panel">
          <div className="editor-header">
            <div>
              <div className="panel-eyebrow">Editor</div>
              <input
                className="chapter-title-input"
                value={chapterDraft.title}
                onChange={(event) => setChapterDraft((previous) => ({ ...previous, title: event.target.value }))}
                placeholder="Kapitelueberschrift"
                disabled={!currentChapter}
              />
            </div>
            <div className="editor-actions">
              <button type="button" className="secondary-button" onClick={createAnchor} disabled={!hasSelection}>
                Anker setzen
              </button>
              <button type="button" className="secondary-button" onClick={createClipboardItem} disabled={!hasSelection}>
                In Clipboard uebernehmen
              </button>
            </div>
          </div>

          <div className="editor-meta">
            <span>{currentChapter ? `Aktiv: ${currentChapter.title}` : "Noch kein Kapitel aktiv"}</span>
            <span>{selectedWorkflowBoxId ? `Zielbox: ${bookBundle?.workflow_boxes?.find((item) => item.id === selectedWorkflowBoxId)?.title || ""}` : "Keine Workflow-Box gewaehlt"}</span>
          </div>

          <div className="editor-frame">
            <EditorPane
              ref={editorRef}
              chapter={currentChapter ? { ...currentChapter, ...chapterDraft } : null}
              pinnedSlots={pinnedSlots}
              onSelectionChange={setHasSelection}
              onDocumentChange={(nextDocument) =>
                setChapterDraft((previous) => ({
                  ...previous,
                  ...nextDocument,
                }))
              }
            />
          </div>
        </section>

        <aside className="workspace-panel right-panel">
          <SidebarSection eyebrow="Anker" title="Aktuelle Textstelle">
            <div className="anchor-list">
              {anchors.length === 0 ? <p className="empty-note">Noch keine Anker fuer dieses Kapitel.</p> : null}
              {anchors.map((anchor) => (
                <article key={anchor.id} className="context-card">
                  <div className="context-card-header">
                    <strong>{anchor.title || "Anker"}</strong>
                    <button type="button" className="ghost-button" onClick={() => deleteAnchor(anchor.id)}>
                      loeschen
                    </button>
                  </div>
                  <p>{anchor.selected_text}</p>
                  <small>
                    {anchor.anchor_type} | {anchor.note || "ohne Notiz"}
                  </small>
                </article>
              ))}
            </div>
          </SidebarSection>

          <SidebarSection eyebrow="Workflow" title="Clipboard & Slots">
            <div className="slot-grid">
              {Array.from({ length: 9 }, (_, index) => {
                const slot = index + 1;
                const item = pinnedSlots.find((entry) => entry.slot === slot);
                return (
                  <div key={slot} className={`slot-card ${item ? "filled" : ""}`}>
                    <strong>{slot}</strong>
                    <span>{item ? previewText(item.content, 36) : "leer"}</span>
                  </div>
                );
              })}
            </div>

            <div className="clipboard-list">
              {clipboardItems.length === 0 ? <p className="empty-note">Noch keine Clipboard-Eintraege vorhanden.</p> : null}
              {clipboardItems.map((item) => (
                <article key={item.id} className="context-card">
                  <div className="context-card-header">
                    <strong>{previewText(item.content, 32)}</strong>
                    <button type="button" className="ghost-button" onClick={() => deleteClipboard(item.id)}>
                      loeschen
                    </button>
                  </div>
                  <p>{previewText(item.content, 120)}</p>
                  <div className="clipboard-controls">
                    <label className="checkbox-row">
                      <input
                        type="checkbox"
                        checked={item.is_pinned}
                        onChange={(event) => updateClipboard(item, { is_pinned: event.target.checked })}
                      />
                      anpinnen
                    </label>
                    <label className="slot-picker">
                      Slot
                      <input
                        type="number"
                        min="0"
                        max="9"
                        value={item.slot}
                        onChange={(event) => updateClipboard(item, { slot: Number(event.target.value) })}
                      />
                    </label>
                    <button type="button" className="ghost-button" onClick={() => editorRef.current?.insertClipboardContent(item.content)}>
                      einfuegen
                    </button>
                  </div>
                </article>
              ))}
            </div>
          </SidebarSection>
        </aside>
      </main>
    </div>
  );
}

export default App;
