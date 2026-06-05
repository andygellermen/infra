import { useEffect, useMemo, useRef, useState } from "react";
import EditorPane from "./components/EditorPane";
import SidebarSection from "./components/SidebarSection";
import { api } from "./lib/api";
import { markdownToDoc, previewText } from "./lib/markdown";
import {
  extractWikiLinks,
  formatTagInput,
  knowledgeReference,
  knowledgeTypeLabel,
  normalizeKnowledgeKey,
  splitTagInput,
} from "./lib/knowledge";

const EMPTY_DRAFT = {
  title: "",
  markdown_content: "",
  editor_json: "",
};

const DEFAULT_EDITOR_APPEARANCE = {
  fontFamily: "serif",
  fontSize: 18,
  lineHeight: 1.8,
  contentWidth: 860,
  fullscreenContentWidth: 1040,
  fullscreenBackdrop: "linen",
  surfacePreset: "warm",
};

const WORKFLOW_TYPE_META = {
  notes: {
    label: "Notizen",
    hint: "Sammelt lose Gedanken, Formulierungen und Zwischenideen direkt am Kapitel.",
  },
  persons: {
    label: "Figuren",
    hint: "Bindet Passagen an Figurenentwicklung, Eigenschaften und offene Charakterfragen.",
  },
  events: {
    label: "Ereignisse",
    hint: "Markiert Schluesselmomente, Wendepunkte und Folgen im Kapitelverlauf.",
  },
  threads: {
    label: "Handlungsfaeden",
    hint: "Verknuepft Passagen mit roten Faeden, Konflikten und offenen Spannungen.",
  },
  reminders: {
    label: "Erinnerungen",
    hint: "Haelt spaetere To-dos, Rueckfragen und Ueberarbeitungsmarken fest.",
  },
  research: {
    label: "Recherche",
    hint: "Sammelt Stellen, die Faktencheck, Quellen oder Vertiefung benoetigen.",
  },
  clipboard: {
    label: "Clipboard",
    hint: "Fokussiert wiederverwendbare Textbausteine und schnelle Rueckgriffe im Schreibfluss.",
  },
  custom: {
    label: "Eigene Box",
    hint: "Freier Arbeitsraum fuer deinen individuellen Schreib- oder Review-Prozess.",
  },
};

function workflowTypeMeta(type) {
  return WORKFLOW_TYPE_META[type] || WORKFLOW_TYPE_META.custom;
}

function App() {
  const editorRef = useRef(null);
  const markdownTextareaRef = useRef(null);
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
  const [knowledgeItems, setKnowledgeItems] = useState([]);
  const [knowledgeQuery, setKnowledgeQuery] = useState("");
  const [selectedKnowledgeItemId, setSelectedKnowledgeItemId] = useState("");
  const [selectedKnowledgeRefKey, setSelectedKnowledgeRefKey] = useState("");
  const [selectedWorkflowBoxId, setSelectedWorkflowBoxId] = useState("");
  const [hasSelection, setHasSelection] = useState(false);
  const [editorMode, setEditorMode] = useState("rich");
  const [saveState, setSaveState] = useState("Synchron");
  const [errorMessage, setErrorMessage] = useState("");
  const [showEditorHelp, setShowEditorHelp] = useState(false);
  const [showEditorSettings, setShowEditorSettings] = useState(false);
  const [showClipboardPalette, setShowClipboardPalette] = useState(false);
  const [isEditorFullscreen, setIsEditorFullscreen] = useState(false);
  const [draggedChapterId, setDraggedChapterId] = useState("");
  const [chapterDropTargetId, setChapterDropTargetId] = useState("");
  const [editorAppearance, setEditorAppearance] = useState(DEFAULT_EDITOR_APPEARANCE);
  const slotNumbers = Array.from({ length: 9 }, (_, index) => index + 1);

  const currentChapter = useMemo(
    () => bookBundle?.chapters?.find((chapter) => chapter.id === selectedChapterId) || null,
    [bookBundle, selectedChapterId],
  );
  const currentChapterId = currentChapter?.id || "";
  const currentBook = useMemo(
    () => projectDetail?.books?.find((book) => book.id === selectedBookId) || bookBundle?.book || null,
    [projectDetail, selectedBookId, bookBundle],
  );
  const chaptersById = useMemo(
    () => new Map((bookBundle?.chapters || []).map((chapter) => [chapter.id, chapter])),
    [bookBundle?.chapters],
  );

  const pinnedSlots = useMemo(
    () =>
      clipboardItems
        .filter((item) => item.is_pinned && item.slot >= 1 && item.slot <= 9)
        .sort((left, right) => left.slot - right.slot),
    [clipboardItems],
  );
  const latestClipboardItems = useMemo(() => clipboardItems.slice(0, 3), [clipboardItems]);

  const chapterKnowledgeRefs = useMemo(
    () => extractWikiLinks(chapterDraft.markdown_content || currentChapter?.markdown_content || ""),
    [chapterDraft.markdown_content, currentChapter?.markdown_content],
  );

  const chapterKnowledgeMatches = useMemo(() => {
    const lookup = new Map(
      knowledgeItems.map((item) => [normalizeKnowledgeKey(item.type, item.name), item]),
    );
    return chapterKnowledgeRefs
      .map((reference) => ({
        reference,
        item: lookup.get(reference.key) || null,
      }))
      .filter(
        (entry, index, items) =>
          items.findIndex((candidate) => candidate.reference.key === entry.reference.key) === index,
      );
  }, [chapterKnowledgeRefs, knowledgeItems]);

  const unresolvedKnowledgeRefs = useMemo(
    () => chapterKnowledgeMatches.filter((entry) => !entry.item),
    [chapterKnowledgeMatches],
  );

  const filteredKnowledgeItems = useMemo(() => {
    const query = knowledgeQuery.trim().toLowerCase();
    if (!query) {
      return knowledgeItems;
    }
    return knowledgeItems.filter((item) =>
      [item.name, item.summary, item.type, ...(item.tags || [])].join(" ").toLowerCase().includes(query),
    );
  }, [knowledgeItems, knowledgeQuery]);

  const selectedKnowledgeItem = useMemo(
    () => filteredKnowledgeItems.find((item) => item.id === selectedKnowledgeItemId) || filteredKnowledgeItems[0] || null,
    [filteredKnowledgeItems, selectedKnowledgeItemId],
  );

  const selectedKnowledgeRef = useMemo(
    () =>
      chapterKnowledgeMatches.find((entry) => entry.reference.key === selectedKnowledgeRefKey) ||
      chapterKnowledgeMatches[0] ||
      null,
    [chapterKnowledgeMatches, selectedKnowledgeRefKey],
  );

  const activeWorkflowBox = useMemo(
    () => bookBundle?.workflow_boxes?.find((item) => item.id === selectedWorkflowBoxId) || null,
    [bookBundle, selectedWorkflowBoxId],
  );

  const anchorCountByWorkflowBox = useMemo(() => {
    const counts = new Map();
    anchors.forEach((anchor) => {
      counts.set(anchor.workflow_box_id, (counts.get(anchor.workflow_box_id) || 0) + 1);
    });
    return counts;
  }, [anchors]);
  const anchorsByWorkflowBox = useMemo(() => {
    const grouped = new Map();
    anchors.forEach((anchor) => {
      grouped.set(anchor.workflow_box_id, [...(grouped.get(anchor.workflow_box_id) || []), anchor]);
    });
    return grouped;
  }, [anchors]);
  const latestAnchorByWorkflowBox = useMemo(() => {
    const latest = new Map();
    anchors.forEach((anchor) => {
      latest.set(anchor.workflow_box_id, anchor);
    });
    return latest;
  }, [anchors]);
  const activeWorkflowAnchors = useMemo(
    () => anchorsByWorkflowBox.get(selectedWorkflowBoxId) || [],
    [anchorsByWorkflowBox, selectedWorkflowBoxId],
  );

  const editorSurfaceStyle = useMemo(
    () => ({
      "--editor-font-family":
        editorAppearance.fontFamily === "sans"
          ? '"Avenir Next", "Segoe UI", sans-serif'
          : editorAppearance.fontFamily === "mono"
            ? '"SFMono-Regular", "Menlo", "Monaco", monospace'
            : '"Iowan Old Style", "Palatino Linotype", "Book Antiqua", Palatino, serif',
      "--editor-font-size": `${editorAppearance.fontSize}px`,
      "--editor-line-height": String(editorAppearance.lineHeight),
      "--editor-max-width": `${editorAppearance.contentWidth}px`,
      "--editor-fullscreen-max-width": `${editorAppearance.fullscreenContentWidth}px`,
      "--editor-caret-color":
        editorAppearance.surfacePreset === "night" ? "rgba(243, 237, 229, 0.76)" : "rgba(109, 59, 16, 0.58)",
      "--editor-selection-bg":
        editorAppearance.surfacePreset === "night" ? "rgba(243, 237, 229, 0.16)" : "rgba(158, 91, 33, 0.14)",
      "--editor-selection-text":
        editorAppearance.surfacePreset === "night" ? "#f7f1ea" : "var(--editor-ink)",
    }),
    [editorAppearance],
  );

  useEffect(() => {
    loadProjects();
  }, []);

  useEffect(() => {
    if (!selectedProjectId) {
      return;
    }
    loadProject(selectedProjectId);
    loadKnowledgeItems(selectedProjectId);
  }, [selectedProjectId]);

  useEffect(() => {
    if (!filteredKnowledgeItems.some((item) => item.id === selectedKnowledgeItemId)) {
      setSelectedKnowledgeItemId(filteredKnowledgeItems[0]?.id || "");
    }
  }, [filteredKnowledgeItems, selectedKnowledgeItemId]);

  useEffect(() => {
    if (!chapterKnowledgeMatches.some((entry) => entry.reference.key === selectedKnowledgeRefKey)) {
      setSelectedKnowledgeRefKey(chapterKnowledgeMatches[0]?.reference.key || "");
    }
  }, [chapterKnowledgeMatches, selectedKnowledgeRefKey]);

  useEffect(() => {
    if (!selectedBookId) {
      setBookBundle(null);
      setSelectedChapterId("");
      setClipboardItems([]);
      setShowClipboardPalette(false);
      return;
    }
    loadBook(selectedBookId);
  }, [selectedBookId]);

  useEffect(() => {
    if (clipboardItems.length === 0) {
      setShowClipboardPalette(false);
    }
  }, [clipboardItems.length]);

  useEffect(() => {
    if (!selectedChapterId || !currentChapter) {
      setAnchors([]);
      setChapterDraft(EMPTY_DRAFT);
      setErrorMessage("");
      setSaveState("Synchron");
      return;
    }
    skipAutosaveRef.current = true;
    setErrorMessage("");
    setSaveState("Synchron");
    setChapterDraft({
      title: currentChapter.title,
      markdown_content: currentChapter.markdown_content || "",
      editor_json: currentChapter.editor_json || "",
    });
    loadAnchors(currentChapter.id);
  }, [selectedChapterId, currentChapterId]);

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

  useEffect(() => {
    function handleGlobalShortcuts(event) {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "s") {
        event.preventDefault();
        void saveChapter(true);
        return;
      }

      if (event.key === "Escape" && showClipboardPalette) {
        event.preventDefault();
        setShowClipboardPalette(false);
        return;
      }

      if (event.key === "Escape" && isEditorFullscreen) {
        event.preventDefault();
        setIsEditorFullscreen(false);
      }
    }

    window.addEventListener("keydown", handleGlobalShortcuts);
    return () => window.removeEventListener("keydown", handleGlobalShortcuts);
  }, [showClipboardPalette, isEditorFullscreen, currentChapter, chapterDraft, editorMode]);

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
      setErrorMessage("");
      const firstBook = response.books?.[0];
      if (!selectedBookId || !response.books?.some((book) => book.id === selectedBookId)) {
        setSelectedBookId(firstBook?.id || "");
      }
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function loadKnowledgeItems(projectId) {
    try {
      const response = await api.get(`/api/projects/${projectId}/knowledge-items`);
      setKnowledgeItems(response.knowledge_items || []);
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function loadBook(bookId) {
    try {
      const response = await api.get(`/api/books/${bookId}`);
      setBookBundle(response);
      setClipboardItems(response.clipboard || []);
      setErrorMessage("");
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
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function saveChapter(manual) {
    if (!currentChapter) {
      return;
    }
    try {
      const richSnapshot =
        editorMode === "rich"
          ? editorRef.current?.getDocumentSnapshot?.() || null
          : null;
      const liveDraft =
        richSnapshot && (richSnapshot.markdown_content || richSnapshot.editor_json)
          ? {
              ...chapterDraft,
              ...richSnapshot,
            }
          : chapterDraft;
      const payload =
        editorMode === "markdown"
          ? {
              ...liveDraft,
              editor_json: JSON.stringify(markdownToDoc(liveDraft.markdown_content || "")),
            }
          : liveDraft;
      if (payload !== chapterDraft) {
        setChapterDraft(payload);
      }
      setErrorMessage("");
      setSaveState(manual ? "Speichert ..." : "Autosave laeuft ...");
      const updated = await api.put(`/api/chapters/${currentChapter.id}`, payload);
      setBookBundle((previous) => ({
        ...previous,
        chapters: previous.chapters.map((chapter) => (chapter.id === updated.id ? updated : chapter)),
      }));
      const nextDraft = {
        title: updated.title,
        markdown_content: updated.markdown_content || "",
        editor_json: updated.editor_json || "",
      };
      setChapterDraft((previous) =>
        previous.title === nextDraft.title &&
        previous.markdown_content === nextDraft.markdown_content &&
        previous.editor_json === nextDraft.editor_json
          ? previous
          : nextDraft,
      );
      setErrorMessage("");
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
        subtitle: "",
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

  async function reorderChapters(sourceChapterId, targetChapterId) {
    if (!selectedBookId || !bookBundle?.chapters?.length || sourceChapterId === targetChapterId) {
      return;
    }

    const previousChapters = bookBundle.chapters;
    const sourceIndex = previousChapters.findIndex((chapter) => chapter.id === sourceChapterId);
    const targetIndex = previousChapters.findIndex((chapter) => chapter.id === targetChapterId);
    if (sourceIndex === -1 || targetIndex === -1) {
      return;
    }

    const nextChapters = [...previousChapters];
    const [movedChapter] = nextChapters.splice(sourceIndex, 1);
    nextChapters.splice(targetIndex, 0, movedChapter);
    const positionedChapters = nextChapters.map((chapter, index) => ({
      ...chapter,
      position: index + 1,
    }));

    setBookBundle((previous) => ({
      ...previous,
      chapters: positionedChapters,
    }));

    try {
      const response = await api.put(`/api/books/${selectedBookId}/chapters/reorder`, {
        chapter_ids: positionedChapters.map((chapter) => chapter.id),
      });
      setBookBundle((previous) => ({
        ...previous,
        chapters: response.chapters || positionedChapters,
      }));
      setErrorMessage("");
    } catch (error) {
      setBookBundle((previous) => ({
        ...previous,
        chapters: previousChapters,
      }));
      if (error.status === 404) {
        setErrorMessage("Kapitel konnten noch nicht neu sortiert werden. Bitte den easy-author-Backend-Prozess einmal neu starten.");
        return;
      }
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

  async function createKnowledgeItem() {
    if (!selectedProjectId) {
      return;
    }
    const name = window.prompt("Name des Wissenseintrags", "Neuer Eintrag");
    if (!name) {
      return;
    }
    try {
      const created = await api.post(`/api/projects/${selectedProjectId}/knowledge-items`, {
        type: "person",
        name,
        summary: "",
        body: "",
        tags: [],
      });
      setKnowledgeItems((previous) =>
        [...previous, created].sort((left, right) => left.name.localeCompare(right.name, "de")),
      );
      setSelectedKnowledgeItemId(created.id);
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function updateKnowledgeItem(item, nextFields) {
    try {
      const updated = await api.put(`/api/knowledge-items/${item.id}`, {
        type: nextFields.type ?? item.type,
        name: nextFields.name ?? item.name,
        summary: nextFields.summary ?? item.summary,
        body: nextFields.body ?? item.body,
        tags: nextFields.tags ?? item.tags,
      });
      setKnowledgeItems((previous) =>
        previous
          .map((entry) => (entry.id === updated.id ? updated : entry))
          .sort((left, right) => left.name.localeCompare(right.name, "de")),
      );
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  function patchLocalBook(bookId, nextFields) {
    setProjectDetail((previous) =>
      previous
        ? {
            ...previous,
            books: previous.books.map((entry) => (entry.id === bookId ? { ...entry, ...nextFields } : entry)),
          }
        : previous,
    );
    setBookBundle((previous) =>
      previous?.book?.id === bookId
        ? {
            ...previous,
            book: {
              ...previous.book,
              ...nextFields,
            },
          }
        : previous,
    );
  }

  async function updateBook(bookId, nextFields) {
    const book = currentBook?.id === bookId ? currentBook : projectDetail?.books?.find((entry) => entry.id === bookId);
    if (!book) {
      return;
    }
    try {
      const updated = await api.put(`/api/books/${bookId}`, {
        title: nextFields.title ?? book.title,
        subtitle: nextFields.subtitle ?? book.subtitle ?? "",
        author: nextFields.author ?? book.author ?? "",
        visibility: nextFields.visibility ?? book.visibility ?? "private",
      });
      patchLocalBook(bookId, updated);
      setErrorMessage("");
    } catch (error) {
      if (error.status === 405) {
        setErrorMessage("Buch-Details konnten noch nicht gespeichert werden. Bitte den easy-author-Backend-Prozess einmal neu starten.");
        return;
      }
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
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function createAnchor(workflowBoxId = selectedWorkflowBoxId, options = {}) {
    const { promptForNote = true } = options;
    if (!selectedChapterId || !workflowBoxId) {
      setErrorMessage("Bitte zuerst ein Kapitel und eine Workflow-Box waehlen.");
      return;
    }
    const payload =
      editorMode === "markdown" ? getMarkdownSelectionPayload() : editorRef.current?.getSelectionPayload();
    if (!payload?.selected_text) {
      setErrorMessage("Bitte zuerst eine Textpassage im Editor markieren.");
      return;
    }
    const note = promptForNote ? window.prompt("Optionale Notiz fuer diesen Anker", "") || "" : "";
    try {
      await api.post(`/api/chapters/${selectedChapterId}/anchors`, {
        ...payload,
        workflow_box_id: workflowBoxId,
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
    const payload =
      editorMode === "markdown" ? getMarkdownSelectionPayload() : editorRef.current?.getSelectionPayload();
    if (!payload?.selected_text) {
      setErrorMessage("Bitte zuerst eine Textpassage im Editor markieren.");
      return;
    }
    await captureClipboardPayload(payload);
  }

  async function captureClipboardPayload(payload) {
    if (!selectedBookId || !payload?.selected_text) {
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

  function handleEditorCopy(payload) {
    void captureClipboardPayload(payload);
  }

  async function updateClipboard(item, nextFields) {
    try {
      const updated = await api.put(`/api/clipboard/${item.id}`, {
        content: nextFields.content ?? item.content,
        is_pinned: nextFields.is_pinned ?? item.is_pinned,
        slot: nextFields.slot ?? item.slot,
      });
      setClipboardItems((previous) => previous.map((entry) => (entry.id === updated.id ? updated : entry)));
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function deleteClipboard(itemId) {
    try {
      await api.delete(`/api/clipboard/${itemId}`);
      setClipboardItems((previous) => previous.filter((entry) => entry.id !== itemId));
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  function nextFreeClipboardSlot(excludedItemId = "") {
    return (
      slotNumbers.find(
        (slot) => !clipboardItems.some((item) => item.id !== excludedItemId && item.is_pinned && item.slot === slot),
      ) || 1
    );
  }

  async function assignClipboardSlot(item, slot) {
    if (slot <= 0) {
      await updateClipboard(item, {
        is_pinned: false,
        slot: 0,
      });
      return;
    }

    await updateClipboard(item, {
      is_pinned: true,
      slot,
    });
  }

  async function toggleClipboardPinFromPalette(item) {
    if (item.is_pinned && item.slot >= 1) {
      await assignClipboardSlot(item, 0);
      return;
    }

    await assignClipboardSlot(item, item.slot >= 1 ? item.slot : nextFreeClipboardSlot(item.id));
  }

  function getMarkdownSelectionPayload() {
    const textarea = markdownTextareaRef.current;
    if (!textarea) {
      return null;
    }
    const start = textarea.selectionStart ?? 0;
    const end = textarea.selectionEnd ?? 0;
    if (start === end) {
      return null;
    }
    const content = chapterDraft.markdown_content || "";
    return {
      selected_text: content.slice(start, end),
      start_offset: start,
      end_offset: end,
      context_before: content.slice(Math.max(0, start - 60), start),
      context_after: content.slice(end, Math.min(content.length, end + 60)),
    };
  }

  function insertIntoMarkdown(content) {
    const textarea = markdownTextareaRef.current;
    if (!textarea) {
      setChapterDraft((previous) => ({
        ...previous,
        markdown_content: `${previous.markdown_content || ""}${content}`,
        editor_json: "",
      }));
      return;
    }
    const start = textarea.selectionStart ?? chapterDraft.markdown_content.length;
    const end = textarea.selectionEnd ?? chapterDraft.markdown_content.length;
    const current = chapterDraft.markdown_content || "";
    const nextValue = `${current.slice(0, start)}${content}${current.slice(end)}`;
    setChapterDraft((previous) => ({
      ...previous,
      markdown_content: nextValue,
      editor_json: "",
    }));
    window.requestAnimationFrame(() => {
      const cursor = start + content.length;
      textarea.focus();
      textarea.setSelectionRange(cursor, cursor);
      setHasSelection(false);
    });
  }

  function insertIntoActiveEditor(content) {
    if (editorMode === "markdown") {
      insertIntoMarkdown(content);
      return;
    }
    editorRef.current?.insertText(content);
  }

  function clipboardSourceLabel(item) {
    if (!item.chapter_id) {
      return "Allgemein";
    }
    return chaptersById.get(item.chapter_id)?.title || "Kapitel";
  }

  function nextFootnoteId(content) {
    const matches = Array.from(String(content || "").matchAll(/\[\^(\d+)\]/g)).map((match) => Number(match[1]));
    return String((matches.length ? Math.max(...matches) : 0) + 1);
  }

  function insertMarkdownFootnote() {
    const noteId = nextFootnoteId(chapterDraft.markdown_content);
    const reference = `[^${noteId}]`;
    const definition = `\n\n[^${noteId}]: `;
    const existing = String(chapterDraft.markdown_content || "");
    const hasDefinition = existing.includes(`[^${noteId}]:`);
    const textarea = markdownTextareaRef.current;
    const start = textarea?.selectionStart ?? existing.length;
    const end = textarea?.selectionEnd ?? existing.length;
    const nextValue = `${existing.slice(0, start)}${reference}${existing.slice(end)}${hasDefinition ? "" : definition}`;
    setChapterDraft((previous) => ({
      ...previous,
      markdown_content: nextValue,
      editor_json: "",
    }));
    window.requestAnimationFrame(() => {
      if (!textarea) {
        return;
      }
      const cursor = start + reference.length;
      textarea.focus();
      textarea.setSelectionRange(cursor, cursor);
    });
  }

  function toggleMarkdownBlockquote() {
    const textarea = markdownTextareaRef.current;
    if (!textarea) {
      insertIntoMarkdown("> ");
      return;
    }
    const content = chapterDraft.markdown_content || "";
    const start = textarea.selectionStart ?? 0;
    const end = textarea.selectionEnd ?? 0;
    const lineStart = content.lastIndexOf("\n", Math.max(0, start - 1)) + 1;
    const lineEndIndex = content.indexOf("\n", end);
    const lineEnd = lineEndIndex === -1 ? content.length : lineEndIndex;
    const segment = content.slice(lineStart, lineEnd);
    const lines = segment.split("\n");
    const alreadyQuoted = lines.every((line) => line.startsWith("> "));
    const nextSegment = lines.map((line) => (alreadyQuoted ? line.replace(/^> /, "") : `> ${line}`)).join("\n");
    const nextValue = `${content.slice(0, lineStart)}${nextSegment}${content.slice(lineEnd)}`;
    setChapterDraft((previous) => ({
      ...previous,
      markdown_content: nextValue,
      editor_json: "",
    }));
  }

  function updateEditorAppearance(field, value) {
    setEditorAppearance((previous) => ({
      ...previous,
      [field]: value,
    }));
  }

  function insertTable() {
    if (editorMode === "markdown") {
      insertIntoMarkdown(["| Kopf-Spalte 1 | Kopf-Spalte 2 |", "| --- | --- |", "|  |  |"].join("\n"));
      return;
    }
    editorRef.current?.insertTable?.();
  }

  function toggleQuote() {
    if (editorMode === "markdown") {
      toggleMarkdownBlockquote();
      return;
    }
    editorRef.current?.toggleBlockquote?.();
  }

  function insertFootnote() {
    if (editorMode === "markdown") {
      insertMarkdownFootnote();
      return;
    }
    editorRef.current?.insertFootnote?.();
  }

  function switchEditorMode(nextMode) {
    if (nextMode === editorMode) {
      return;
    }
    if (nextMode === "markdown") {
      const snapshot = editorRef.current?.getDocumentSnapshot?.();
      if (snapshot) {
        setChapterDraft((previous) => ({
          ...previous,
          ...snapshot,
        }));
      }
    }
    setHasSelection(false);
    setEditorMode(nextMode);
  }

  const bookTitle = bookBundle?.book?.title || "Kein Buch geladen";

  return (
    <div className={`app-shell ${isEditorFullscreen ? "editor-fullscreen-shell" : ""}`}>
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

      <main
        className={`workspace-grid ${isEditorFullscreen ? "editor-fullscreen" : ""}`}
        data-fullscreen-backdrop={editorAppearance.fullscreenBackdrop}
      >
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
                  <span>{book.subtitle || "Ohne Beschreibung"}</span>
                  <small>{book.visibility}</small>
                </button>
              ))}
            </div>
            {currentBook ? (
              <div className="book-meta-card">
                <div className="context-card-header">
                  <strong>Buchdetails</strong>
                  <span className="knowledge-chip">{currentBook.visibility}</span>
                </div>
                <label className="editor-setting">
                  <span>Titel</span>
                  <input
                    value={currentBook.title || ""}
                    onChange={(event) => patchLocalBook(currentBook.id, { title: event.target.value })}
                    onBlur={(event) => updateBook(currentBook.id, { title: event.target.value })}
                  />
                </label>
                <label className="editor-setting">
                  <span>Beschreibung</span>
                  <textarea
                    rows="3"
                    value={currentBook.subtitle || ""}
                    placeholder="Kurzbeschreibung oder Positionierung des Buchs"
                    onChange={(event) => patchLocalBook(currentBook.id, { subtitle: event.target.value })}
                    onBlur={(event) => updateBook(currentBook.id, { subtitle: event.target.value })}
                  />
                </label>
                <div className="book-meta-grid">
                  <label className="editor-setting">
                    <span>Autor</span>
                    <input
                      value={currentBook.author || ""}
                      placeholder="Autor oder Arbeitstitel"
                      onChange={(event) => patchLocalBook(currentBook.id, { author: event.target.value })}
                      onBlur={(event) => updateBook(currentBook.id, { author: event.target.value })}
                    />
                  </label>
                  <label className="editor-setting">
                    <span>Sichtbarkeit</span>
                    <select
                      value={currentBook.visibility || "private"}
                      onChange={(event) => {
                        patchLocalBook(currentBook.id, { visibility: event.target.value });
                        void updateBook(currentBook.id, { visibility: event.target.value });
                      }}
                    >
                      <option value="private">private</option>
                      <option value="shared">shared</option>
                      <option value="public">public</option>
                    </select>
                  </label>
                </div>
              </div>
            ) : (
              <p className="empty-note">Noch kein Buch ausgewaehlt.</p>
            )}
          </SidebarSection>

          <SidebarSection eyebrow="Kapitel" title={bookTitle} actionLabel="+ Kapitel" onAction={createChapter}>
            <div className="chapter-list">
              {(bookBundle?.chapters || []).map((chapter) => (
                <button
                  key={chapter.id}
                  type="button"
                  draggable
                  className={`chapter-row ${chapter.id === selectedChapterId ? "active" : ""} ${chapter.id === draggedChapterId ? "dragging" : ""} ${chapter.id === chapterDropTargetId ? "drop-target" : ""}`}
                  onClick={() => setSelectedChapterId(chapter.id)}
                  onDragStart={() => {
                    setDraggedChapterId(chapter.id);
                    setChapterDropTargetId(chapter.id);
                  }}
                  onDragOver={(event) => {
                    event.preventDefault();
                    if (chapterDropTargetId !== chapter.id) {
                      setChapterDropTargetId(chapter.id);
                    }
                  }}
                  onDrop={async (event) => {
                    event.preventDefault();
                    const sourceChapterId = draggedChapterId;
                    setDraggedChapterId("");
                    setChapterDropTargetId("");
                    await reorderChapters(sourceChapterId, chapter.id);
                  }}
                  onDragEnd={() => {
                    setDraggedChapterId("");
                    setChapterDropTargetId("");
                  }}
                >
                  <span className="chapter-drag-handle" aria-hidden="true">
                    ⋮⋮
                  </span>
                  <span className="chapter-index">{String(chapter.position).padStart(2, "0")}</span>
                  <span>{chapter.title}</span>
                </button>
              ))}
            </div>
          </SidebarSection>

          <SidebarSection eyebrow="Workflow" title="Workflow-Boxen" actionLabel="+ Box" onAction={createWorkflowBox}>
            {activeWorkflowBox ? (
              <div className="workflow-target-card">
                <div className="context-card-header">
                  <strong>Zielbox aktiv</strong>
                  <span className="knowledge-chip">{activeWorkflowBox.type}</span>
                </div>
                <p>{activeWorkflowBox.title}</p>
                <small>{workflowTypeMeta(activeWorkflowBox.type).hint}</small>
                <div className="workflow-target-meta">
                  <span className="knowledge-chip">
                    {anchorCountByWorkflowBox.get(activeWorkflowBox.id) || 0} Anker im aktuellen Kapitel
                  </span>
                  <span className="knowledge-chip">
                    {activeWorkflowAnchors.length > 0 ? "bereits verbunden" : "bereit fuer erste Passage"}
                  </span>
                </div>
                <div className="workflow-target-actions">
                  <button
                    type="button"
                    className="secondary-button"
                    onClick={() => createAnchor(activeWorkflowBox.id, { promptForNote: false })}
                    disabled={!hasSelection || !selectedChapterId}
                  >
                    Auswahl ankern
                  </button>
                  <button type="button" className="ghost-button" onClick={() => setShowEditorHelp(true)}>
                    Workflow-Hilfe
                  </button>
                </div>
                <div className="workflow-anchor-list">
                  {activeWorkflowAnchors.length === 0 ? (
                    <p className="empty-note">Noch keine Passagen in dieser Box. Markiere Text und verankere ihn direkt hier.</p>
                  ) : (
                    activeWorkflowAnchors.slice(-3).reverse().map((anchor) => (
                      <article key={anchor.id} className="workflow-anchor-card">
                        <strong>{anchor.title || "Passage"}</strong>
                        <p>{previewText(anchor.selected_text, 120)}</p>
                        <small>{anchor.note || "ohne Notiz"}</small>
                      </article>
                    ))
                  )}
                </div>
              </div>
            ) : null}
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
                  <div className="workflow-card-header">
                    <div>
                      <strong>{workflowTypeMeta(box.type).label}</strong>
                      <small>{anchorCountByWorkflowBox.get(box.id) || 0} Anker im Kapitel</small>
                    </div>
                    <div className="workflow-card-actions">
                      {selectedWorkflowBoxId === box.id ? <span className="knowledge-chip">Ziel</span> : null}
                      <button
                        type="button"
                        className="ghost-button"
                        onClick={(event) => {
                          event.stopPropagation();
                          setSelectedWorkflowBoxId(box.id);
                        }}
                      >
                        {selectedWorkflowBoxId === box.id ? "aktiv" : "als Ziel"}
                      </button>
                    </div>
                  </div>
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
                  {!box.is_collapsed ? (
                    <>
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
                        <span className="knowledge-chip">{anchorCountByWorkflowBox.get(box.id) || 0} Anker</span>
                      </div>
                      <p className="workflow-card-hint">{workflowTypeMeta(box.type).hint}</p>
                      <div className="workflow-row">
                        <button
                          type="button"
                          className="secondary-button"
                          onClick={(event) => {
                            event.stopPropagation();
                            setSelectedWorkflowBoxId(box.id);
                            void createAnchor(box.id, { promptForNote: false });
                          }}
                          disabled={!hasSelection || !selectedChapterId}
                        >
                          Auswahl ankern
                        </button>
                        {latestAnchorByWorkflowBox.get(box.id) ? (
                          <span className="knowledge-chip">
                            zuletzt: {previewText(latestAnchorByWorkflowBox.get(box.id).selected_text, 28)}
                          </span>
                        ) : (
                          <span className="knowledge-chip">noch leer</span>
                        )}
                      </div>
                    </>
                  ) : (
                    <div className="workflow-row">
                      <label className="checkbox-row">
                        <input
                          type="checkbox"
                          checked={box.is_collapsed}
                          onChange={(event) => updateWorkflowBox(box.id, { is_collapsed: event.target.checked })}
                        />
                        collapsed
                      </label>
                      <span className="knowledge-chip">{workflowTypeMeta(box.type).label}</span>
                      {latestAnchorByWorkflowBox.get(box.id) ? (
                        <span className="knowledge-chip">
                          {previewText(latestAnchorByWorkflowBox.get(box.id).selected_text, 24)}
                        </span>
                      ) : null}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </SidebarSection>

          <SidebarSection eyebrow="Wissen" title="Wissensbank" actionLabel="+ Eintrag" onAction={createKnowledgeItem}>
            <div className="knowledge-list compact">
              <input
                value={knowledgeQuery}
                placeholder="Begriffe filtern"
                onChange={(event) => setKnowledgeQuery(event.target.value)}
              />
              {knowledgeItems.length === 0 ? <p className="empty-note">Noch keine Wissenseintraege vorhanden.</p> : null}
              <div className="chip-cloud">
                {filteredKnowledgeItems.map((item) => {
                  const linkedInChapter = chapterKnowledgeMatches.some(
                    (entry) => entry.reference.key === normalizeKnowledgeKey(item.type, item.name),
                  );
                  return (
                    <button
                      key={item.id}
                      type="button"
                      className={`cloud-chip ${selectedKnowledgeItem?.id === item.id ? "active" : ""} ${linkedInChapter ? "linked" : ""}`}
                      onClick={() => setSelectedKnowledgeItemId(item.id)}
                    >
                      <span>{item.name}</span>
                      {linkedInChapter ? <small>im Kapitel</small> : null}
                    </button>
                  );
                })}
              </div>
              {selectedKnowledgeItem ? (
                <article className={`knowledge-card ${chapterKnowledgeMatches.some((entry) => entry.reference.key === normalizeKnowledgeKey(selectedKnowledgeItem.type, selectedKnowledgeItem.name)) ? "linked" : ""}`}>
                  <div className="context-card-header">
                    <strong>{knowledgeTypeLabel(selectedKnowledgeItem.type)}</strong>
                    <button type="button" className="ghost-button" onClick={() => insertIntoActiveEditor(knowledgeReference(selectedKnowledgeItem))}>
                      Link
                    </button>
                  </div>
                  <input
                    value={selectedKnowledgeItem.name}
                    onChange={(event) =>
                      setKnowledgeItems((previous) =>
                        previous.map((entry) =>
                          entry.id === selectedKnowledgeItem.id ? { ...entry, name: event.target.value } : entry,
                        ),
                      )
                    }
                    onBlur={(event) => updateKnowledgeItem(selectedKnowledgeItem, { name: event.target.value })}
                  />
                  <div className="workflow-row">
                    <select
                      value={selectedKnowledgeItem.type}
                      onChange={(event) => updateKnowledgeItem(selectedKnowledgeItem, { type: event.target.value })}
                    >
                      <option value="person">person</option>
                      <option value="location">location</option>
                      <option value="event">event</option>
                      <option value="thread">thread</option>
                      <option value="motif">motif</option>
                      <option value="term">term</option>
                      <option value="reminder">reminder</option>
                      <option value="research_note">research_note</option>
                      <option value="custom">custom</option>
                    </select>
                    <span className="knowledge-chip">{knowledgeReference(selectedKnowledgeItem)}</span>
                  </div>
                  <textarea
                    rows="2"
                    value={selectedKnowledgeItem.summary || ""}
                    placeholder="Kurze Zusammenfassung"
                    onChange={(event) =>
                      setKnowledgeItems((previous) =>
                        previous.map((entry) =>
                          entry.id === selectedKnowledgeItem.id ? { ...entry, summary: event.target.value } : entry,
                        ),
                      )
                    }
                    onBlur={(event) => updateKnowledgeItem(selectedKnowledgeItem, { summary: event.target.value })}
                  />
                  <input
                    value={formatTagInput(selectedKnowledgeItem.tags)}
                    placeholder="Tags, komma-getrennt"
                    onChange={(event) =>
                      setKnowledgeItems((previous) =>
                        previous.map((entry) =>
                          entry.id === selectedKnowledgeItem.id
                            ? { ...entry, tags: splitTagInput(event.target.value) }
                            : entry,
                        ),
                      )
                    }
                    onBlur={(event) => updateKnowledgeItem(selectedKnowledgeItem, { tags: splitTagInput(event.target.value) })}
                  />
                </article>
              ) : null}
            </div>
          </SidebarSection>
        </aside>

        <section className="editor-panel">
          {isEditorFullscreen ? (
            <button
              type="button"
              className="fullscreen-exit-button"
              aria-label="Fullscreen verlassen"
              onClick={() => setIsEditorFullscreen(false)}
            >
              ⎋
            </button>
          ) : null}
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
              <button
                type="button"
                className="ghost-button"
                aria-expanded={showEditorHelp}
                onClick={() => setShowEditorHelp((previous) => !previous)}
              >
                {showEditorHelp ? "Hilfe ausblenden" : "Hilfe"}
              </button>
              <button
                type="button"
                className="ghost-button"
                aria-expanded={showEditorSettings}
                onClick={() => setShowEditorSettings((previous) => !previous)}
              >
                {showEditorSettings ? "⚙ Einstellungen ausblenden" : "⚙ Einstellungen"}
              </button>
              <button
                type="button"
                className="ghost-button"
                aria-pressed={isEditorFullscreen}
                onClick={() => setIsEditorFullscreen((previous) => !previous)}
              >
                {isEditorFullscreen ? "Vollbild aus" : "Vollbild"}
              </button>
              <div className="mode-switch" role="tablist" aria-label="Editor-Modus">
                <button
                  type="button"
                  className={`mode-button ${editorMode === "rich" ? "active" : ""}`}
                  onClick={() => switchEditorMode("rich")}
                >
                  Rich
                </button>
                <button
                  type="button"
                  className={`mode-button ${editorMode === "markdown" ? "active" : ""}`}
                  onClick={() => switchEditorMode("markdown")}
                >
                  Markdown
                </button>
              </div>
              <button type="button" className="secondary-button" onClick={() => createAnchor()} disabled={!hasSelection}>
                Anker setzen
              </button>
              <button type="button" className="secondary-button" onClick={createClipboardItem} disabled={!hasSelection}>
                In Clipboard uebernehmen
              </button>
              <button type="button" className="secondary-button" onClick={insertTable} disabled={!currentChapter}>
                Tabelle
              </button>
              <button type="button" className="secondary-button" onClick={toggleQuote} disabled={!currentChapter}>
                Zitat
              </button>
              <button type="button" className="secondary-button" onClick={insertFootnote} disabled={!currentChapter}>
                Fussnote
              </button>
            </div>
          </div>

          <div className="editor-meta">
            <span>{currentChapter ? `Aktiv: ${currentChapter.title}` : "Noch kein Kapitel aktiv"}</span>
            <span>{selectedWorkflowBoxId ? `Zielbox: ${bookBundle?.workflow_boxes?.find((item) => item.id === selectedWorkflowBoxId)?.title || ""}` : "Keine Workflow-Box gewaehlt"}</span>
            <span>{editorMode === "markdown" ? "Markdown ist aktuell die Quelle" : "Tiptap-Editor mit Markdown-Snapshot"}</span>
          </div>

          {showEditorHelp ? (
            <section className="editor-help" role="dialog" aria-label="Editor-Hilfe">
              <div className="editor-help-header">
                <div>
                  <div className="panel-eyebrow">Editor-Hilfe</div>
                  <strong>MVP-Referenz fuer Schreiben, Workflow und Einfuegen</strong>
                </div>
                <button type="button" className="ghost-button" onClick={() => setShowEditorHelp(false)}>
                  Schliessen
                </button>
              </div>
              <div className="editor-help-grid">
                <article className="editor-help-card">
                  <strong>Markdown-Modus</strong>
                  <p>
                    Aktuell voll unterstuetzt sind Ueberschriften, Listen, verschachtelte Listen, Zitate, Code-Fences,
                    Trennlinien, harte Umbrueche, Escaping, Inline-Code, Fett, Kursiv, Durchgestrichen,
                    `[[...]]`-Wiki-Links und einfache Pipe-Tabellen.
                  </p>
                </article>
                <article className="editor-help-card">
                  <strong>Rich-Editor</strong>
                  <p>
                    Der Tiptap-Teil verarbeitet die gespeicherten Strukturen und erkennt typische Schreibmuster wie
                    Ueberschriften, Listen, Zitate, Code-Bloecke, Trennlinien und Tabellen. Abgetippte Pipe-Tabellen
                    wie `| Kopf | Kopf |` werden beim Weiterschreiben automatisch in eine Rich-Tabelle uebernommen.
                  </p>
                </article>
                <article className="editor-help-card">
                  <strong>Clipboard & Slots</strong>
                  <p>
                    Markiere Text und uebernimm ihn ueber `In Clipboard uebernehmen`. Rechts kannst du Eintraege
                    anpinnen, Slots `1-9` zuweisen und ueber `einfuegen` an der Cursorposition einsetzen. Im Rich-Editor
                    gehen gepinnte Slots auch per `Cmd/Ctrl + Shift + 1-9`. Die Floating-Liste sammelt alle Snippets
                    an einer Stelle und erlaubt schnelle Slot-Zuordnung ohne den Schreibfluss zu verlassen. Gefuellte
                    Slot-Karten lassen sich auch direkt per Klick wieder einfuegen.
                  </p>
                </article>
                <article className="editor-help-card">
                  <strong>Workflow-Anker</strong>
                  <p>
                    Waehle links zuerst eine Workflow-Box. Die aktive Box erscheint als `Zielbox`. Wenn du danach Text
                    markierst und `Anker setzen` klickst, wird die Passage genau dieser Workflow-Box zugeordnet. Im
                    Workflow-Cockpit kannst du markierte Passagen auch direkt ohne Zusatzdialog verankern.
                  </p>
                </article>
                <article className="editor-help-card">
                  <strong>Tabellen-Werkzeuge</strong>
                  <p>
                    `Tabelle` legt schnell eine Grundstruktur an. Sobald der Cursor in einer Tabelle steht, erscheint
                    direkt am Editor eine kontextuelle Table-Bar fuer Spalten, Zeilen, Kopfzeile und Loeschen.
                  </p>
                </article>
                <article className="editor-help-card">
                  <strong>Zitate & Fussnoten</strong>
                  <p>
                    `Zitat` schaltet Blockquotes um. `Fussnote` erzeugt im Rich-Editor eine Referenz samt Notizblock und
                    im Markdown-Modus eine `[^1]`-Referenz mit passender Definition am Kapitelende.
                  </p>
                </article>
              </div>
            </section>
          ) : null}

          {showEditorSettings ? (
            <section className="editor-settings" role="dialog" aria-label="Editor-Einstellungen">
              <div className="editor-help-header">
                <div>
                  <div className="panel-eyebrow">Einstellungen</div>
                  <strong>Look and Feel fuer konzentriertes Schreiben</strong>
                </div>
                <button type="button" className="ghost-button" onClick={() => setShowEditorSettings(false)}>
                  Schliessen
                </button>
              </div>
              <div className="editor-settings-grid">
                <label className="editor-setting">
                  <span>Schriftfamilie</span>
                  <select value={editorAppearance.fontFamily} onChange={(event) => updateEditorAppearance("fontFamily", event.target.value)}>
                    <option value="serif">Serif</option>
                    <option value="sans">Sans</option>
                    <option value="mono">Mono</option>
                  </select>
                </label>
                <label className="editor-setting">
                  <span>Schriftgroesse</span>
                  <input
                    type="range"
                    min="16"
                    max="24"
                    step="1"
                    value={editorAppearance.fontSize}
                    onChange={(event) => updateEditorAppearance("fontSize", Number(event.target.value))}
                  />
                </label>
                <label className="editor-setting">
                  <span>Zeilenhoehe</span>
                  <input
                    type="range"
                    min="1.5"
                    max="2.2"
                    step="0.05"
                    value={editorAppearance.lineHeight}
                    onChange={(event) => updateEditorAppearance("lineHeight", Number(event.target.value))}
                  />
                </label>
                <label className="editor-setting">
                  <span>Textbreite</span>
                  <select
                    value={editorAppearance.contentWidth}
                    onChange={(event) => updateEditorAppearance("contentWidth", Number(event.target.value))}
                  >
                    <option value="720">Schmal</option>
                    <option value="860">Standard</option>
                    <option value="1040">Breit</option>
                  </select>
                </label>
                <label className="editor-setting">
                  <span>Vollbild-Breite</span>
                  <select
                    value={editorAppearance.fullscreenContentWidth}
                    onChange={(event) => updateEditorAppearance("fullscreenContentWidth", Number(event.target.value))}
                  >
                    <option value="860">Kompakt</option>
                    <option value="1040">Standard</option>
                    <option value="1200">Breit</option>
                    <option value="1360">Sehr breit</option>
                  </select>
                </label>
                <label className="editor-setting">
                  <span>Vollbild-Hintergrund</span>
                  <select
                    value={editorAppearance.fullscreenBackdrop}
                    onChange={(event) => updateEditorAppearance("fullscreenBackdrop", event.target.value)}
                  >
                    <option value="linen">Leinen</option>
                    <option value="paper">Papier</option>
                    <option value="dusk">Dusk</option>
                    <option value="night">Nacht</option>
                  </select>
                </label>
                <label className="editor-setting">
                  <span>Farbprofil</span>
                  <select
                    value={editorAppearance.surfacePreset}
                    onChange={(event) => updateEditorAppearance("surfacePreset", event.target.value)}
                  >
                    <option value="warm">Warm</option>
                    <option value="paper">Papier</option>
                    <option value="night">Nacht</option>
                  </select>
                </label>
              </div>
            </section>
          ) : null}

          <div
            className={`editor-frame surface-${editorAppearance.surfacePreset}`}
            style={editorSurfaceStyle}
            data-surface={editorAppearance.surfacePreset}
          >
            <div className="editor-frame-inner">
              {editorMode === "markdown" ? (
                <textarea
                  ref={markdownTextareaRef}
                  className="markdown-textarea"
                  value={chapterDraft.markdown_content}
                  placeholder="Schreibe hier direkt in Markdown. Wiki-Links wie [[Mara]] oder [[Ort:Alter Garten]] bleiben erhalten."
                  onFocus={() => setHasSelection(false)}
                  onSelect={(event) => {
                    const target = event.target;
                    setHasSelection((target.selectionStart ?? 0) !== (target.selectionEnd ?? 0));
                  }}
                  onKeyUp={(event) => {
                    const target = event.target;
                    setHasSelection((target.selectionStart ?? 0) !== (target.selectionEnd ?? 0));
                  }}
                  onClick={(event) => {
                    const target = event.target;
                    setHasSelection((target.selectionStart ?? 0) !== (target.selectionEnd ?? 0));
                  }}
                  onCopy={() => handleEditorCopy(getMarkdownSelectionPayload())}
                  onCut={() => handleEditorCopy(getMarkdownSelectionPayload())}
                  onChange={(event) =>
                    setChapterDraft((previous) => ({
                      ...previous,
                      markdown_content: event.target.value,
                      editor_json: "",
                    }))
                  }
                />
              ) : (
                <EditorPane
                  ref={editorRef}
                  chapter={currentChapter ? { ...currentChapter, ...chapterDraft } : null}
                  pinnedSlots={pinnedSlots}
                  onSelectionChange={setHasSelection}
                  onClipboardCapture={handleEditorCopy}
                  onDocumentChange={(nextDocument) =>
                    setChapterDraft((previous) => ({
                      ...previous,
                      ...nextDocument,
                    }))
                  }
                />
              )}
            </div>
          </div>
        </section>

        <aside className="workspace-panel right-panel">
          <SidebarSection eyebrow="Kontext" title="Wiki-Links im Kapitel">
            <div className="knowledge-context-list compact">
              {chapterKnowledgeMatches.length === 0 ? (
                <p className="empty-note">Noch keine `[[...]]`-Referenzen im aktuellen Kapitel.</p>
              ) : null}
              <div className="chip-cloud">
                {chapterKnowledgeMatches.map((entry) => (
                  <button
                    key={entry.reference.key}
                    type="button"
                    className={`cloud-chip ${selectedKnowledgeRef?.reference.key === entry.reference.key ? "active" : ""} ${entry.item ? "linked" : "unresolved"}`}
                    onClick={() => setSelectedKnowledgeRefKey(entry.reference.key)}
                  >
                    <span>{entry.item ? entry.item.name : entry.reference.raw}</span>
                  </button>
                ))}
              </div>
              {selectedKnowledgeRef ? (
                <article className={`context-card ${selectedKnowledgeRef.item ? "" : "unresolved"}`}>
                  <div className="context-card-header">
                    <strong>{selectedKnowledgeRef.item ? selectedKnowledgeRef.item.name : selectedKnowledgeRef.reference.raw}</strong>
                    <span className="knowledge-chip">
                      {knowledgeTypeLabel(selectedKnowledgeRef.item?.type || selectedKnowledgeRef.reference.type)}
                    </span>
                  </div>
                  <p>{selectedKnowledgeRef.item?.summary || "Noch kein passender Wissenseintrag vorhanden."}</p>
                  <small>{selectedKnowledgeRef.item ? knowledgeReference(selectedKnowledgeRef.item) : `[[${selectedKnowledgeRef.reference.raw}]]`}</small>
                </article>
              ) : null}
              {unresolvedKnowledgeRefs.length > 0 ? (
                <p className="empty-note">Offene Referenzen koennen links als Wissenseintrag angelegt oder umbenannt werden.</p>
              ) : null}
            </div>
          </SidebarSection>

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
            <div className="clipboard-summary">
              <span>
                {clipboardItems.length} Eintraege · {pinnedSlots.length} Slots fixiert
              </span>
              <button
                type="button"
                className="ghost-button"
                onClick={() => setShowClipboardPalette(true)}
                disabled={clipboardItems.length === 0}
              >
                Floating-Liste
              </button>
            </div>

            <div className="slot-grid">
              {slotNumbers.map((slot) => {
                const item = pinnedSlots.find((entry) => entry.slot === slot);
                return (
                  <div key={slot} className={`slot-card ${item ? "filled" : ""}`}>
                    <strong>{slot}</strong>
                    <span>{item ? previewText(item.content, 36) : "leer"}</span>
                    {item ? (
                      <>
                        <small>{clipboardSourceLabel(item)}</small>
                        <button
                          type="button"
                          className="ghost-button"
                          aria-label={`Slot ${slot} einfuegen`}
                          onClick={() => insertIntoActiveEditor(item.content)}
                        >
                          einfuegen
                        </button>
                      </>
                    ) : (
                      <small>per Klick oder Shortcut belegbar</small>
                    )}
                  </div>
                );
              })}
            </div>

            {latestClipboardItems.length > 0 ? (
              <div className="clipboard-recent-row">
                {latestClipboardItems.map((item) => (
                  <button
                    key={item.id}
                    type="button"
                    className="cloud-chip"
                    onClick={() => insertIntoActiveEditor(item.content)}
                  >
                    <span>{previewText(item.content, 22)}</span>
                    <small>{clipboardSourceLabel(item)}</small>
                  </button>
                ))}
              </div>
            ) : null}

            <div className="clipboard-list">
              {clipboardItems.length === 0 ? <p className="empty-note">Noch keine Clipboard-Eintraege vorhanden.</p> : null}
              {clipboardItems.map((item) => (
                <article key={item.id} className="context-card">
                  <div className="context-card-header">
                    <strong>{previewText(item.content, 32)}</strong>
                    <div className="workflow-row">
                      <span className="knowledge-chip">{clipboardSourceLabel(item)}</span>
                      <button type="button" className="ghost-button" onClick={() => deleteClipboard(item.id)}>
                        loeschen
                      </button>
                    </div>
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
                    <button type="button" className="ghost-button" onClick={() => insertIntoActiveEditor(item.content)}>
                      einfuegen
                    </button>
                    <button
                      type="button"
                      className="ghost-button"
                      onClick={() => assignClipboardSlot(item, nextFreeClipboardSlot(item.id))}
                    >
                      naechster Slot
                    </button>
                  </div>
                </article>
              ))}
            </div>
          </SidebarSection>
        </aside>
      </main>

      {clipboardItems.length > 0 ? (
        <button
          type="button"
          className={`clipboard-fab ${showClipboardPalette ? "active" : ""}`}
          aria-expanded={showClipboardPalette}
          aria-controls="clipboard-palette"
          aria-label="Clipboard-Floating-Liste"
          onClick={() => setShowClipboardPalette((previous) => !previous)}
        >
          Clipboard · {clipboardItems.length}
        </button>
      ) : null}

      {showClipboardPalette ? (
        <section id="clipboard-palette" className="clipboard-palette" role="dialog" aria-label="Clipboard-Liste">
          <div className="clipboard-palette-header">
            <div>
              <div className="panel-eyebrow">Floating Clipboard</div>
              <strong>Gesammelte Ausschnitte und feste Slots</strong>
            </div>
            <button type="button" className="ghost-button" onClick={() => setShowClipboardPalette(false)}>
              Schliessen
            </button>
          </div>

          <div className="clipboard-palette-list">
            {clipboardItems.map((item) => (
              <article
                key={item.id}
                className={`context-card clipboard-palette-card ${item.is_pinned && item.slot >= 1 ? "pinned" : ""}`}
              >
                <div className="context-card-header">
                  <strong>{previewText(item.content, 36)}</strong>
                  <span className="knowledge-chip">{item.is_pinned && item.slot >= 1 ? `Slot ${item.slot}` : "frei"}</span>
                </div>
                <p>{previewText(item.content, 180)}</p>
                <div className="clipboard-controls">
                  <button type="button" className="secondary-button" onClick={() => insertIntoActiveEditor(item.content)}>
                    einfuegen
                  </button>
                  <button type="button" className="ghost-button" onClick={() => toggleClipboardPinFromPalette(item)}>
                    {item.is_pinned && item.slot >= 1 ? "Slot loesen" : "fixieren"}
                  </button>
                  <button type="button" className="ghost-button" onClick={() => deleteClipboard(item.id)}>
                    loeschen
                  </button>
                </div>
                <small>{clipboardSourceLabel(item)}</small>
                <div className="clipboard-slot-pills" aria-label={`Slot-Zuordnung fuer ${previewText(item.content, 24)}`}>
                  {slotNumbers.map((slot) => (
                    <button
                      key={slot}
                      type="button"
                      className={`slot-pill ${item.slot === slot ? "active" : ""}`}
                      aria-label={`Slot ${slot} zuweisen`}
                      onClick={() => assignClipboardSlot(item, slot)}
                    >
                      {slot}
                    </button>
                  ))}
                  <button
                    type="button"
                    className={`slot-pill clear ${item.slot === 0 ? "active" : ""}`}
                    aria-label="Slot freigeben"
                    onClick={() => assignClipboardSlot(item, 0)}
                  >
                    frei
                  </button>
                </div>
              </article>
            ))}
          </div>
        </section>
      ) : null}
    </div>
  );
}

export default App;
