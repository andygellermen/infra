import { forwardRef, useEffect, useImperativeHandle } from "react";
import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";
import App, { splitMarkdownIntoChapterSections } from "./App";
import { api } from "./lib/api";
import { markdownToDoc } from "./lib/markdown";

const RICH_SNAPSHOT_MARKDOWN = "# Kapitel 1\n\nRich Snapshot aus dem Editor";
let mockRichSnapshotMarkdown = null;

vi.mock("./lib/api", () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock("./components/EditorPane", () => ({
  default: forwardRef(function MockEditorPane({ chapter, onSelectionChange }, ref) {
    useImperativeHandle(ref, () => ({
      getSelectionPayload: () => null,
      getDocumentSnapshot: () => {
        const markdown = mockRichSnapshotMarkdown ?? chapter?.markdown_content ?? "";
        return {
          markdown_content: markdown,
          editor_json: JSON.stringify(markdownToDoc(markdown)),
        };
      },
      insertClipboardContent: () => {},
      insertText: () => {},
      applyReviewCommentMark: () => {},
      removeReviewCommentMark: () => {},
      replaceReviewCommentText: () => {},
      insertTable: () => {},
      toggleBlockquote: () => {},
      insertFootnote: () => {},
      addColumnAfter: () => {},
      addRowAfter: () => {},
      deleteColumn: () => {},
      deleteRow: () => {},
      toggleHeaderRow: () => {},
      deleteTable: () => {},
      isTableActive: () => false,
      focus: () => {},
    }));

    useEffect(() => {
      onSelectionChange?.(false);
    }, [onSelectionChange]);

    return <div data-testid="editor-pane">Rich Editor: {chapter?.title || "leer"}</div>;
  }),
}));

const project = {
  id: "project-1",
  title: "Romanprojekt",
  description: "Demo",
};

const projectTwo = {
  id: "project-2",
  title: "Sachbuchprojekt",
  description: "Alternative Welt",
};

const book = {
  id: "book-1",
  title: "Buch Eins",
  subtitle: "Erste Buchbeschreibung",
  author: "A. Autor",
  visibility: "private",
};

const bookTwo = {
  id: "book-2",
  title: "Buch Zwei",
  subtitle: "",
  author: "",
  visibility: "private",
};

const bookThree = {
  id: "book-3",
  title: "Buch Drei",
  subtitle: "",
  author: "",
  visibility: "private",
};

const workflowBox = {
  id: "box-1",
  title: "Notizen",
  type: "notes",
  tags: ["Alter", "Text", "idee"],
  is_collapsed: false,
};

const workflowTimeBox = {
  id: "box-2",
  title: "Timeline",
  type: "research",
  tags: ["zeit", "datum", "jahr"],
  is_collapsed: false,
};

const knowledgeItem = {
  id: "knowledge-1",
  type: "person",
  name: "Mara",
  summary: "Protagonistin mit scharfem Blick",
  body: "",
  tags: ["figur"],
};

const chapter = {
  id: "chapter-1",
  position: 1,
  title: "Kapitel 1",
  markdown_content: "# Kapitel 1\n\nAlter Text",
  editor_json: "",
};

const chapterTwo = {
  id: "chapter-2",
  position: 2,
  title: "Kapitel 2",
  markdown_content: "# Kapitel 2\n\nZweiter Auftakt",
  editor_json: "",
};

const revisionOne = {
  id: "revision-1",
  chapter_id: chapter.id,
  revision_type: "manual",
  title: "Bewusster Speicherpunkt",
  description: "Manuelle Sicherung vor Strukturwechsel.",
  markdown_content: "# Kapitel 1\n\nFrueherer Stand",
  editor_json: JSON.stringify(markdownToDoc("# Kapitel 1\n\nFrueherer Stand")),
  word_count: 4,
  added_words: 4,
  removed_words: 0,
  change_summary: "Kapitelstand mit 4 Woertern gesichert.",
  session_id: "session-revision",
  created_by: "Tester",
  created_at: "2026-06-11T18:15:00Z",
};

const autosaveDraftOne = {
  id: "autosave-1",
  chapter_id: chapter.id,
  markdown_content: "# Kapitel 1\n\nAutosave Recovery",
  editor_json: JSON.stringify(markdownToDoc("# Kapitel 1\n\nAutosave Recovery")),
  reason: "idle_autosave",
  session_id: "session-autosave",
  word_count: 4,
  created_at: "2026-06-11T18:45:00Z",
  expires_at: "2026-06-14T18:45:00Z",
};

const reviewCommentOne = {
  id: "comment-1",
  chapter_id: chapter.id,
  revision_id: revisionOne.id,
  comment_type: "suggestion",
  author: "Lektorat",
  body: "Bitte den Satzbau vereinfachen.",
  suggested_text: "Der Satz ist klarer und direkter formuliert.",
  selected_text: "Der alte Satz ist unnötig kompliziert.",
  start_offset: 4,
  end_offset: 44,
  context_before: "",
  context_after: "",
  status: "open",
  is_todo_done: false,
  created_at: "2026-06-11T18:10:00Z",
};

const reviewCommentTwo = {
  id: "comment-2",
  chapter_id: chapter.id,
  revision_id: revisionOne.id,
  comment_type: "todo",
  author: "Review",
  body: "Zeitangabe gegen Timeline prüfen.",
  suggested_text: "",
  selected_text: "Im Sommer 1987",
  start_offset: 52,
  end_offset: 66,
  context_before: "",
  context_after: "",
  status: "open",
  is_todo_done: false,
  created_at: "2026-06-11T18:05:00Z",
};

const reviewCommentThree = {
  id: "comment-3",
  chapter_id: chapter.id,
  revision_id: "",
  comment_type: "comment",
  author: "Review",
  body: "Bereits geprüft.",
  suggested_text: "",
  selected_text: "Alter Abschnitt",
  start_offset: 70,
  end_offset: 84,
  context_before: "",
  context_after: "",
  status: "resolved",
  is_todo_done: true,
  created_at: "2026-06-11T17:55:00Z",
};

const MARKDOWN_PLACEHOLDER =
  "Schreibe hier direkt in Markdown. Wiki-Links wie [[Mara]] oder [[Ort:Alter Garten]] bleiben erhalten.";

async function openMarkdownEditor(user) {
  await user.click(screen.getByRole("button", { name: "Markdown" }));
  return screen.findByPlaceholderText(MARKDOWN_PLACEHOLDER);
}

function selectMarkdownText(textarea, selectedText) {
  const start = textarea.value.indexOf(selectedText);
  const end = start + selectedText.length;
  textarea.focus();
  textarea.setSelectionRange(start, end);
  fireEvent.select(textarea);
  fireEvent.keyUp(textarea);
  return { start, end };
}

function mockApi() {
  const state = {
    projects: [{ ...project }, { ...projectTwo }],
    booksByProject: {
      [project.id]: [{ ...book }, { ...bookTwo }],
      [projectTwo.id]: [{ ...bookThree }],
    },
    anchors: {
      [chapter.id]: [],
      [chapterTwo.id]: [],
    },
    comments: {
      [chapter.id]: [{ ...reviewCommentOne }, { ...reviewCommentTwo }, { ...reviewCommentThree }],
      [chapterTwo.id]: [],
    },
    clipboard: [],
    chaptersByBook: {
      [book.id]: [{ ...chapter }, { ...chapterTwo }],
      [bookTwo.id]: [{ ...chapter }, { ...chapterTwo }],
      [bookThree.id]: [{ ...chapter }, { ...chapterTwo }],
    },
    workflowBoxes: [{ ...workflowBox }, { ...workflowTimeBox }],
    knowledgeItems: [{ ...knowledgeItem }],
    revisionsByChapter: {
      [chapter.id]: [{ ...revisionOne }],
      [chapterTwo.id]: [],
    },
    autosavesByChapter: {
      [chapter.id]: [{ ...autosaveDraftOne }],
      [chapterTwo.id]: [],
    },
    revisionEventsByRevisionId: {
      [revisionOne.id]: [
        {
          id: "event-1",
          revision_id: revisionOne.id,
          chapter_id: chapter.id,
          event_type: "chapter_saved",
          title: "Kapitel gespeichert",
          description: "Manuelle Sicherung vor Strukturwechsel.",
          created_at: "2026-06-11T18:15:00Z",
        },
      ],
    },
    milestonesByBook: {
      [book.id]: [],
      [bookTwo.id]: [],
      [bookThree.id]: [],
    },
  };

  api.get.mockImplementation(async (path) => {
    if (path.startsWith("/api/projects/") && path.endsWith("/knowledge-items")) {
      const projectId = path.split("/")[3];
      return { knowledge_items: projectId === project.id ? state.knowledgeItems : [] };
    }

    if (path.startsWith("/api/projects/")) {
      const projectId = path.split("/")[3];
      const projectEntry = state.projects.find((entry) => entry.id === projectId);
      if (projectEntry) {
        return { project: projectEntry, books: state.booksByProject[projectId] || [] };
      }
    }

    if (path.startsWith("/api/books/")) {
      const bookId = path.split("/")[3];
      const resolvedBook = Object.values(state.booksByProject)
        .flat()
        .find((entry) => entry.id === bookId);
      if (resolvedBook) {
        return {
          book: resolvedBook,
          chapters: state.chaptersByBook[resolvedBook.id] || [],
          workflow_boxes: state.workflowBoxes,
          clipboard: state.clipboard,
        };
      }
    }

    if (path.startsWith("/api/revisions/") && path.endsWith("/events")) {
      const revisionId = path.split("/")[3];
      return { events: state.revisionEventsByRevisionId[revisionId] || [] };
    }

    switch (path) {
      case "/api/projects":
        return { projects: state.projects };
      case `/api/chapters/${chapter.id}/revisions`:
        return { revisions: state.revisionsByChapter[chapter.id] };
      case `/api/chapters/${chapterTwo.id}/revisions`:
        return { revisions: state.revisionsByChapter[chapterTwo.id] };
      case `/api/chapters/${chapter.id}/autosaves`:
        return { autosaves: state.autosavesByChapter[chapter.id] };
      case `/api/chapters/${chapterTwo.id}/autosaves`:
        return { autosaves: state.autosavesByChapter[chapterTwo.id] };
      case `/api/books/${book.id}/milestones`:
        return { milestones: state.milestonesByBook[book.id] };
      case `/api/books/${bookTwo.id}/milestones`:
        return { milestones: state.milestonesByBook[bookTwo.id] };
      case `/api/books/${bookThree.id}/milestones`:
        return { milestones: state.milestonesByBook[bookThree.id] };
      case `/api/chapters/${chapter.id}/anchors`:
        return { anchors: state.anchors[chapter.id] };
      case `/api/chapters/${chapterTwo.id}/anchors`:
        return { anchors: state.anchors[chapterTwo.id] };
      case `/api/chapters/${chapter.id}/comments`:
        return { comments: state.comments[chapter.id] };
      case `/api/chapters/${chapterTwo.id}/comments`:
        return { comments: state.comments[chapterTwo.id] };
      default:
        throw new Error(`Unexpected GET ${path}`);
    }
  });

  api.post.mockImplementation(async (path, payload) => {
    if (path === "/api/projects") {
      const created = {
        id: `project-${state.projects.length + 1}`,
        title: payload.title,
        description: payload.description || "",
      };
      state.projects = [...state.projects, created];
      state.booksByProject[created.id] = [];
      return created;
    }

    if (path.startsWith("/api/projects/") && path.endsWith("/books")) {
      const projectId = path.split("/")[3];
      const created = {
        id: `book-${Object.values(state.booksByProject).flat().length + 1}`,
        title: payload.title,
        subtitle: payload.subtitle || "",
        author: payload.author || "",
        visibility: payload.visibility || "private",
      };
      state.booksByProject[projectId] = [...(state.booksByProject[projectId] || []), created];
      state.chaptersByBook[created.id] = [];
      return created;
    }

    if (path.startsWith("/api/books/") && path.endsWith("/chapters")) {
      const bookId = path.split("/")[3];
      const created = {
        id: `chapter-${Object.values(state.chaptersByBook).flat().length + 1}`,
        position: (state.chaptersByBook[bookId]?.length || 0) + 1,
        title: payload.title,
        markdown_content: payload.markdown_content,
        editor_json: payload.editor_json || "",
      };
      state.chaptersByBook[bookId] = [...(state.chaptersByBook[bookId] || []), created];
      state.anchors[created.id] = [];
      return created;
    }

    if (path.startsWith("/api/books/") && path.endsWith("/workflow-boxes")) {
      const created = {
        id: `box-${state.workflowBoxes.length + 1}`,
        title: payload.title,
        type: payload.type || "custom",
        tags: payload.tags || [],
        is_collapsed: Boolean(payload.is_collapsed),
      };
      state.workflowBoxes = [...state.workflowBoxes, created];
      return created;
    }

    if (path.startsWith("/api/chapters/") && path.endsWith("/comments")) {
      const chapterId = path.split("/")[3];
      const created = {
        id: `comment-${state.comments[chapterId]?.length || 0}`,
        revision_id: payload.revision_id || "",
        comment_type: payload.comment_type || "comment",
        author: payload.author || "Review",
        body: payload.body || "",
        suggested_text: payload.suggested_text || "",
        selected_text: payload.selected_text || "",
        start_offset: payload.start_offset || 0,
        end_offset: payload.end_offset || 0,
        context_before: payload.context_before || "",
        context_after: payload.context_after || "",
        status: payload.status || "open",
        is_todo_done: Boolean(payload.is_todo_done),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      state.comments[chapterId] = [created, ...(state.comments[chapterId] || [])];
      return created;
    }

    if (path === `/api/chapters/${chapter.id}/anchors` || path === `/api/chapters/${chapterTwo.id}/anchors`) {
      const chapterId = path.split("/")[3];
      const created = {
        id: `anchor-${state.anchors[chapterId].length + 1}`,
        ...payload,
      };
      state.anchors[chapterId] = [...state.anchors[chapterId], created];
      return created;
    }

    if (path === `/api/books/${book.id}/clipboard`) {
      const created = {
        id: `clipboard-${state.clipboard.length + 1}`,
        ...payload,
      };
      state.clipboard = [created, ...state.clipboard];
      return created;
    }

    if (path.endsWith("/milestones")) {
      const bookId = path.split("/")[3];
      const created = {
        id: `milestone-${(state.milestonesByBook[bookId] || []).length + 1}`,
        book_id: bookId,
        chapter_id: chapter.id,
        revision_id: payload.revision_id,
        title: payload.title,
        description: payload.description || "",
        milestone_type: payload.milestone_type || "custom",
        locked: Boolean(payload.locked),
        created_by: payload.created_by || "",
        created_at: "2026-06-11T19:10:00Z",
      };
      state.milestonesByBook[bookId] = [created, ...(state.milestonesByBook[bookId] || [])];
      return created;
    }

    if (path.startsWith("/api/revisions/") && path.endsWith("/restore")) {
      const revisionId = path.split("/")[3];
      const targetRevision = Object.values(state.revisionsByChapter)
        .flat()
        .find((entry) => entry.id === revisionId);
      if (!targetRevision) {
        throw new Error(`Unexpected restore ${path}`);
      }
      const restoredChapter = {
        ...(Object.values(state.chaptersByBook).flat().find((entry) => entry.id === targetRevision.chapter_id) || chapter),
        markdown_content: targetRevision.markdown_content,
        editor_json: targetRevision.editor_json,
      };
      state.chaptersByBook = Object.fromEntries(
        Object.entries(state.chaptersByBook).map(([bookId, chapters]) => [
          bookId,
          chapters.map((entry) => (entry.id === restoredChapter.id ? restoredChapter : entry)),
        ]),
      );
      const protectionRevision = {
        ...revisionOne,
        id: "revision-protection-1",
        title: "Sicherungsstand vor Wiederherstellung",
        markdown_content: chapter.markdown_content,
        editor_json: JSON.stringify(markdownToDoc(chapter.markdown_content)),
        created_at: "2026-06-11T19:00:00Z",
      };
      const restoredRevision = {
        ...targetRevision,
        id: "revision-restored-1",
        revision_type: "restore",
        title: "Wiederhergestellter Stand",
        description: `Wiederhergestellt aus Revision ${targetRevision.id}.`,
        created_at: "2026-06-11T19:05:00Z",
      };
      state.revisionsByChapter[targetRevision.chapter_id] = [
        restoredRevision,
        protectionRevision,
        ...state.revisionsByChapter[targetRevision.chapter_id],
      ];
      state.revisionEventsByRevisionId[restoredRevision.id] = [
        {
          id: "event-restore-1",
          revision_id: restoredRevision.id,
          chapter_id: targetRevision.chapter_id,
          event_type: "restore_performed",
          title: "Wiederherstellung ausgefuehrt",
          description: `Wiederhergestellt aus Revision ${targetRevision.id}.`,
          created_at: "2026-06-11T19:05:00Z",
        },
      ];
      return {
        chapter: restoredChapter,
        protection_revision: protectionRevision,
        restored_revision: restoredRevision,
      };
    }

    throw new Error(`Unexpected POST ${path}`);
  });

  api.put.mockImplementation(async (path, payload) => {
    if (path.endsWith("/chapters/reorder")) {
      const bookId = path.split("/")[3];
      const orderedIds = payload.chapter_ids;
      const existing = state.chaptersByBook[bookId] || [];
      const lookup = new Map(existing.map((entry) => [entry.id, entry]));
      const reordered = orderedIds.map((id, index) => ({
        ...lookup.get(id),
        position: index + 1,
      }));
      state.chaptersByBook[bookId] = reordered;
      return { chapters: reordered };
    }

    if (path.startsWith("/api/books/")) {
      const bookId = path.split("/").pop();
      const existing = Object.values(state.booksByProject)
        .flat()
        .find((entry) => entry.id === bookId);
      if (!existing) {
        throw new Error(`Unexpected book ${path}`);
      }
      const updated = {
        ...existing,
        ...payload,
      };
      state.booksByProject = Object.fromEntries(
        Object.entries(state.booksByProject).map(([projectId, books]) => [
          projectId,
          books.map((entry) => (entry.id === bookId ? updated : entry)),
        ]),
      );
      return updated;
    }

    if (path.startsWith("/api/chapters/")) {
      const chapterId = path.split("/").pop();
      const existing = Object.values(state.chaptersByBook)
        .flat()
        .find((entry) => entry.id === chapterId);
      if (!existing) {
        throw new Error(`Unexpected chapter ${path}`);
      }
      const updated = {
        ...existing,
        ...payload,
      };
      state.chaptersByBook = Object.fromEntries(
        Object.entries(state.chaptersByBook).map(([bookId, chapters]) => [
          bookId,
          chapters.map((entry) => (entry.id === chapterId ? updated : entry)),
        ]),
      );
      return updated;
    }

    if (path.startsWith("/api/clipboard/")) {
      const itemId = path.split("/").pop();
      const existing = state.clipboard.find((item) => item.id === itemId);
      if (!existing) {
        throw new Error(`Unexpected clipboard item ${path}`);
      }
      const updated = {
        ...existing,
        ...payload,
      };
      state.clipboard = state.clipboard.map((item) => (item.id === itemId ? updated : item));
      return updated;
    }

    if (path.startsWith("/api/comments/")) {
      const commentId = path.split("/").pop();
      const chapterId = Object.keys(state.comments).find((key) => state.comments[key].some((entry) => entry.id === commentId));
      if (!chapterId) {
        throw new Error(`Unexpected comment ${path}`);
      }
      const existing = state.comments[chapterId].find((entry) => entry.id === commentId);
      const updated = {
        ...existing,
        ...payload,
      };
      state.comments[chapterId] = state.comments[chapterId].map((entry) => (entry.id === commentId ? updated : entry));
      return updated;
    }

    if (path.startsWith("/api/milestones/")) {
      const milestoneId = path.split("/").pop();
      const bookId = Object.keys(state.milestonesByBook).find((key) =>
        (state.milestonesByBook[key] || []).some((entry) => entry.id === milestoneId),
      );
      if (!bookId) {
        throw new Error(`Unexpected milestone ${path}`);
      }
      const existing = state.milestonesByBook[bookId].find((entry) => entry.id === milestoneId);
      const updated = {
        ...existing,
        ...payload,
      };
      state.milestonesByBook[bookId] = state.milestonesByBook[bookId].map((entry) => (entry.id === milestoneId ? updated : entry));
      return updated;
    }

    if (path.startsWith("/api/workflow-boxes/")) {
      const boxId = path.split("/").pop();
      const existing = state.workflowBoxes.find((entry) => entry.id === boxId);
      if (!existing) {
        throw new Error(`Unexpected workflow box ${path}`);
      }
      const updated = {
        ...existing,
        ...payload,
      };
      state.workflowBoxes = state.workflowBoxes.map((entry) => (entry.id === boxId ? updated : entry));
      return updated;
    }

    if (path.startsWith("/api/knowledge-items/")) {
      const itemId = path.split("/").pop();
      const existing = state.knowledgeItems.find((entry) => entry.id === itemId);
      if (!existing) {
        throw new Error(`Unexpected knowledge item ${path}`);
      }
      const updated = {
        ...existing,
        ...payload,
      };
      state.knowledgeItems = state.knowledgeItems.map((entry) => (entry.id === itemId ? updated : entry));
      return updated;
    }

    throw new Error(`Unexpected PUT ${path}`);
  });

  api.delete.mockResolvedValue(null);
  api.delete.mockImplementation(async (path) => {
    if (path.startsWith("/api/clipboard/")) {
      const itemId = path.split("/").pop();
      state.clipboard = state.clipboard.filter((item) => item.id !== itemId);
      return null;
    }
    if (path.startsWith("/api/comments/")) {
      const commentId = path.split("/").pop();
      state.comments = Object.fromEntries(
        Object.entries(state.comments).map(([chapterId, comments]) => [
          chapterId,
          comments.filter((entry) => entry.id !== commentId),
        ]),
      );
      return null;
    }
    if (path.startsWith("/api/milestones/")) {
      const milestoneId = path.split("/").pop();
      state.milestonesByBook = Object.fromEntries(
        Object.entries(state.milestonesByBook).map(([bookId, milestones]) => [
          bookId,
          milestones.filter((entry) => entry.id !== milestoneId),
        ]),
      );
      return null;
    }
    return null;
  });

  return state;
}

describe("App editor smoke test", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockRichSnapshotMarkdown = null;
    window.localStorage.clear();
    window.localStorage.setItem("easy-author.work-mode.v1", "structure");
    mockApi();
  });

  it("keeps footnote definitions with the earlier chapter when a new H1 starts below their references", () => {
    const sections = splitMarkdownIntoChapterSections(
      [
        "# Kapitel 1",
        "",
        "Der erste Gedanke bleibt wichtig[^1] und bekommt spaeter noch eine zweite Quelle[^2].",
        "",
        "# Kapitel 2",
        "",
        "Hier beginnt bereits das neue Kapitel.",
        "",
        "[^1]: Das ist die erste Fussnote.",
        "[^2]: Das ist die zweite Fussnote.",
      ].join("\n"),
      "Kapitel 1",
    );

    expect(sections).toHaveLength(2);
    expect(sections[0].content).toContain("[^1]: Das ist die erste Fussnote.");
    expect(sections[0].content).toContain("[^2]: Das ist die zweite Fussnote.");
    expect(sections[1].content).not.toContain("[^1]: Das ist die erste Fussnote.");
    expect(sections[1].content).not.toContain("[^2]: Das ist die zweite Fussnote.");
    expect(sections[1].content).toContain("# Kapitel 2");
  });

  it("loads a chapter, switches to markdown, saves, and returns to rich mode", async () => {
    const user = userEvent.setup();

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    expect(screen.getByTestId("editor-pane")).toHaveTextContent("Rich Editor: Kapitel 1");

    const textarea = await openMarkdownEditor(user);

    const nextMarkdown = "# Kapitel 1\n\n- Punkt A\n- Punkt B\n\n[[Person:Mara]]";
    fireEvent.change(textarea, {
      target: { value: nextMarkdown },
    });

    await user.click(screen.getByRole("button", { name: "Kapitel speichern" }));

    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith(`/api/chapters/${chapter.id}`, expect.objectContaining({
        title: "Kapitel 1",
        markdown_content: nextMarkdown,
        editor_json: JSON.stringify(markdownToDoc(nextMarkdown)),
        save_mode: "manual",
        create_revision: true,
        revision_type: "manual",
        created_by: "easy-author-editor",
        autosave_reason: "manual_save",
        session_id: expect.any(String),
      }));
    });

    expect(api.put).toHaveBeenCalledTimes(1);
    expect(screen.getByText("Modus · Markdown")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Rich" }));

    expect(await screen.findByTestId("editor-pane")).toHaveTextContent("Rich Editor: Kapitel 1");
  });

  it("hydrates markdown mode from the current rich editor snapshot", async () => {
    const user = userEvent.setup();
    mockRichSnapshotMarkdown = RICH_SNAPSHOT_MARKDOWN;

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    expect(screen.getByTestId("editor-pane")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Markdown" }));

    expect(await screen.findByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(RICH_SNAPSHOT_MARKDOWN);
    expect(screen.getByText("Modus · Markdown")).toBeInTheDocument();
  });

  it("saves markdown content via Cmd/Ctrl+S instead of the browser default", async () => {
    const user = userEvent.setup();

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);
    const nextMarkdown = "# Kapitel 1\n\nShortcut Save";

    fireEvent.change(textarea, {
      target: { value: nextMarkdown },
    });

    const keyboardEvent = new KeyboardEvent("keydown", {
      key: "s",
      ctrlKey: true,
      bubbles: true,
      cancelable: true,
    });
    window.dispatchEvent(keyboardEvent);

    expect(keyboardEvent.defaultPrevented).toBe(true);

    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith(`/api/chapters/${chapter.id}`, expect.objectContaining({
        title: "Kapitel 1",
        markdown_content: nextMarkdown,
        editor_json: JSON.stringify(markdownToDoc(nextMarkdown)),
        save_mode: "manual",
        create_revision: true,
        revision_type: "manual",
        created_by: "easy-author-editor",
        autosave_reason: "manual_save",
        session_id: expect.any(String),
      }));
    });
  });

  it("shows revision history, loads autosave drafts, and restores a selected revision", async () => {
    const user = userEvent.setup();
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(true);
    const { container } = render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    await user.click(screen.getByRole("button", { name: "Revisionen" }));

    expect(await screen.findByText("Bewusster Speicherpunkt")).toBeInTheDocument();
    expect(screen.getByText("Draft · Automatisch")).toBeInTheDocument();

    const detailCard = container.querySelector(".revision-detail-card");
    expect(detailCard).toBeTruthy();
    await user.click(within(detailCard).getByRole("button", { name: "Als Entwurf laden" }));

    await waitFor(() => {
      expect(textarea).toHaveValue(autosaveDraftOne.markdown_content);
    });

    const revisionCard = screen.getByText("Bewusster Speicherpunkt").closest(".revision-card");
    expect(revisionCard).toBeTruthy();
    await user.click(revisionCard);

    await waitFor(() => {
      expect(screen.getByText("Manuelle Sicherung vor Strukturwechsel.")).toBeInTheDocument();
    });

    const activeDetailCard = container.querySelector(".revision-detail-card");
    await user.click(within(activeDetailCard).getByRole("button", { name: "Wiederherstellen" }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/api/revisions/${revisionOne.id}/restore`, {
        created_by: "easy-author-editor",
      });
    });

    expect(confirmSpy).toHaveBeenCalled();

    await waitFor(() => {
      expect(textarea).toHaveValue(revisionOne.markdown_content);
      expect(screen.getByText(/Revision wiederhergestellt/)).toBeInTheDocument();
    });

    confirmSpy.mockRestore();
  });

  it("creates a bookmark milestone from a selected revision in the timeline", async () => {
    const user = userEvent.setup();

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Revisionen" }));

    const revisionCard = await screen.findByText("Bewusster Speicherpunkt");
    await user.click(revisionCard.closest(".revision-card"));
    await user.click(screen.getByRole("button", { name: "Vor Review" }));

    const bookmarkButton = await screen.findByRole("button", { name: "Bookmark setzen" });
    await user.click(bookmarkButton);

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/api/books/${book.id}/milestones`, {
        revision_id: revisionOne.id,
        title: revisionOne.title,
        description: revisionOne.change_summary,
        milestone_type: "before_review",
        locked: true,
        created_by: "easy-author-editor",
      });
    });

    expect(screen.getAllByText(revisionOne.title).length).toBeGreaterThan(0);
  });

  it("shows calm proofing summary and filters review plus timeline entries", async () => {
    const user = userEvent.setup();

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Revisionen" }));

    expect(await screen.findByText("Proofing-Uebersicht")).toBeInTheDocument();
    expect(screen.getByText("Ruhiger Blick auf offene Hinweise und Korrekturen")).toBeInTheDocument();
    await user.click(screen.getByText("Bewusster Speicherpunkt").closest(".revision-card"));
    expect(screen.getByRole("button", { name: /Zur Revision\s*2/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Vorschlaege\s*1/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /To-dos\s*1/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Abgeschlossen\s*1/i })).toBeInTheDocument();
    expect(screen.getAllByText("Bitte den Satzbau vereinfachen.").length).toBeGreaterThan(0);
    expect(screen.getAllByText(/Bewusster Speicherpunkt/).length).toBeGreaterThan(0);

    await user.click(screen.getByRole("button", { name: /Abgeschlossen\s*1/i }));
    expect(await screen.findByText("Bereits geprüft.")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Drafts\s*1/i }));
    expect(screen.getByText("Draft · Automatisch")).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.queryByText("Bewusster Speicherpunkt")).not.toBeInTheDocument();
    });
  });

  it("toggles editor fullscreen and exits it with Escape", async () => {
    const user = userEvent.setup();
    const { container } = render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Vollbild" }));

    expect(container.querySelector(".workspace-grid")?.className).toContain("editor-fullscreen");
    expect(screen.queryByRole("button", { name: "Vollbild" })).not.toBeInTheDocument();

    const keyboardEvent = new KeyboardEvent("keydown", {
      key: "Escape",
      bubbles: true,
      cancelable: true,
    });
    window.dispatchEvent(keyboardEvent);

    await waitFor(() => {
      expect(container.querySelector(".workspace-grid")?.className).not.toContain("editor-fullscreen");
    });
  });

  it("edits and persists book description metadata", async () => {
    const user = userEvent.setup();

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Buchdetails" }));
    await user.click(screen.getByRole("button", { name: "Buch ändern" }));
    const descriptionField = screen.getByRole("textbox", { name: "Beschreibung" });
    await user.clear(descriptionField);
    await user.type(descriptionField, "Neue kompakte Buchbeschreibung");
    fireEvent.blur(descriptionField);

    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith(`/api/books/${book.id}`, {
        title: "Buch Eins",
        subtitle: "Neue kompakte Buchbeschreibung",
        author: "A. Autor",
        visibility: "private",
      });
    });
  });

  it("reorders chapters via drag and drop and persists the new order", async () => {
    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();

    const chapterOneButton = screen.getByRole("button", { name: /Kapitel 1/ });
    const chapterTwoButton = screen.getByRole("button", { name: /Kapitel 2/ });

    fireEvent.dragStart(chapterOneButton);
    fireEvent.dragOver(chapterTwoButton);
    fireEvent.drop(chapterTwoButton);

    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith(`/api/books/${book.id}/chapters/reorder`, {
        chapter_ids: [chapterTwo.id, chapter.id],
      });
    });

    const chapterButtons = screen.getAllByRole("button", { name: /Kapitel [12]/ });
    expect(chapterButtons[0]).toHaveTextContent("Kapitel 2");
    expect(chapterButtons[1]).toHaveTextContent("Kapitel 1");
  });

  it("shows contextual editor help for markdown, workflow, and clipboard usage", async () => {
    const user = userEvent.setup();

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Hilfe" }));

    expect(screen.getByRole("dialog", { name: "Editor-Hilfe" })).toBeInTheDocument();
    expect(screen.getByText("MVP-Referenz fuer Schreiben, Workflow und Einfuegen")).toBeInTheDocument();
    expect(screen.getByText(/verschachtelte Listen, Zitate, Code-Fences/)).toBeInTheDocument();
    expect(screen.getByText(/einfache Pipe-Tabellen/)).toBeInTheDocument();
    expect(screen.getByText(/Cmd\/Ctrl \+ Shift \+ 1-9/)).toBeInTheDocument();
    expect(screen.getByText(/aktive Box erscheint als `Zielbox`/)).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Schliessen" }));

    await waitFor(() => {
      expect(screen.queryByRole("dialog", { name: "Editor-Hilfe" })).not.toBeInTheDocument();
    });
  });

  it("opens editor settings and updates the editor appearance controls", async () => {
    const user = userEvent.setup();
    const { container, unmount } = render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const editorFrame = container.querySelector(".editor-frame");
    const workspaceGrid = container.querySelector(".workspace-grid");
    expect(editorFrame).toBeTruthy();
    expect(workspaceGrid).toBeTruthy();

    await user.click(screen.getByRole("button", { name: /Einstellungen/ }));

    expect(screen.getByRole("dialog", { name: "Editor-Einstellungen" })).toBeInTheDocument();

    await user.selectOptions(screen.getByRole("combobox", { name: "Farbprofil" }), "night");
    await user.selectOptions(screen.getByRole("combobox", { name: "Textbreite" }), "1040");
    await user.selectOptions(screen.getByRole("combobox", { name: "Vollbild-Breite" }), "1200");
    await user.selectOptions(screen.getByRole("combobox", { name: "Vollbild-Hintergrund" }), "dusk");

    expect(editorFrame.className).toContain("surface-night");
    expect(editorFrame.style.getPropertyValue("--editor-max-width")).toBe("1040px");
    expect(editorFrame.style.getPropertyValue("--editor-fullscreen-max-width")).toBe("1200px");
    expect(workspaceGrid?.getAttribute("data-fullscreen-backdrop")).toBe("dusk");
    expect(editorFrame?.getAttribute("data-fullscreen-backdrop")).toBe("dusk");
    expect(JSON.parse(window.localStorage.getItem("easy-author.editor-appearance.v1"))).toEqual(
      expect.objectContaining({
        surfacePreset: "night",
        contentWidth: 1040,
        fullscreenContentWidth: 1200,
        fullscreenBackdrop: "dusk",
      }),
    );

    await user.click(screen.getByRole("button", { name: "Vollbild" }));
    expect(container.querySelector(".workspace-grid")?.className).toContain("editor-fullscreen");
    expect(editorFrame?.style.getPropertyValue("--fullscreen-backdrop-start")).toBe("#ede1d9");
    fireEvent.keyDown(window, {
      key: "Escape",
      bubbles: true,
      cancelable: true,
    });

    await waitFor(() => {
      expect(container.querySelector(".workspace-grid")?.className).not.toContain("editor-fullscreen");
      expect(screen.queryByRole("dialog", { name: "Editor-Einstellungen" })).not.toBeInTheDocument();
    });

    unmount();

    const rerendered = render(<App />);
    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    expect(rerendered.container.querySelector(".workspace-grid")?.getAttribute("data-fullscreen-backdrop")).toBe("dusk");
  });

  it("creates an anchor and clipboard item from markdown selection", async () => {
    const user = userEvent.setup();
    const promptSpy = vi.spyOn(window, "prompt").mockReturnValue("Wichtige Notiz");

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const selectedText = "Alter Text";
    const { start, end } = selectMarkdownText(textarea, selectedText);

    const anchorButton = screen.getByRole("button", { name: "Anker setzen" });
    const clipboardButton = screen.getByRole("button", { name: "In Clipboard uebernehmen" });

    await waitFor(() => {
      expect(anchorButton).toBeEnabled();
      expect(clipboardButton).toBeEnabled();
    });

    await user.click(anchorButton);

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(
        `/api/chapters/${chapter.id}/anchors`,
        expect.objectContaining({
          selected_text: selectedText,
          start_offset: start,
          end_offset: end,
          workflow_box_id: workflowBox.id,
          anchor_type: "passage",
          title: selectedText,
          note: "Wichtige Notiz",
        }),
      );
    });

    expect(await screen.findByText("passage | Wichtige Notiz")).toBeInTheDocument();

    selectMarkdownText(textarea, selectedText);

    await user.click(clipboardButton);

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/api/books/${book.id}/clipboard`, {
        chapter_id: chapter.id,
        content: selectedText,
        content_type: "text/markdown",
        source_anchor_id: "",
        is_pinned: false,
        slot: 0,
      });
    });

    expect(await screen.findByRole("button", { name: "einfuegen" })).toBeInTheDocument();
    expect(screen.queryByText("Noch keine Clipboard-Eintraege vorhanden.")).not.toBeInTheDocument();

    promptSpy.mockRestore();
  });

  it("shows workflow suggestions in the selection popup and anchors directly into the suggested box", async () => {
    const user = userEvent.setup();
    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);
    const { start, end } = selectMarkdownText(textarea, "Alter Text");

    expect(screen.queryByRole("dialog", { name: "Auswahl-Aktionen" })).not.toBeInTheDocument();

    const popup = await screen.findByRole("dialog", { name: "Auswahl-Aktionen" }, { timeout: 2600 });
    expect(within(popup).getByRole("button", { name: "Zu Notizen" })).toBeInTheDocument();

    await user.click(within(popup).getByRole("button", { name: "Zu Notizen" }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(
        `/api/chapters/${chapter.id}/anchors`,
        expect.objectContaining({
          selected_text: "Alter Text",
          start_offset: start,
          end_offset: end,
          workflow_box_id: workflowBox.id,
          note: "",
        }),
      );
    });
  });

  it("boosts workflow suggestions for explicit time cues", async () => {
    const user = userEvent.setup();
    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);
    fireEvent.change(textarea, {
      target: { value: "# Kapitel 1\n\nAm 1.4.2011 begann die Reise bis 2014." },
    });

    selectMarkdownText(textarea, "1.4.2011 begann die Reise bis 2014");

    const popup = await screen.findByRole("dialog", { name: "Auswahl-Aktionen" }, { timeout: 2600 });
    expect(within(popup).getByRole("button", { name: "Zu Timeline" })).toBeInTheDocument();
  });

  it("marks workflow boxes semantically as focused when context strongly matches", async () => {
    const user = userEvent.setup();
    const { container } = render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);
    fireEvent.change(textarea, {
      target: { value: "# Kapitel 1\n\nAm 1.4.2011 begann die Reise bis 2014." },
    });

    selectMarkdownText(textarea, "1.4.2011 begann die Reise bis 2014");

    await screen.findByRole("dialog", { name: "Auswahl-Aktionen" }, { timeout: 2600 });

    const workflowCards = () => Array.from(container.querySelectorAll(".workflow-card"));
    const timelineInput = screen.getByDisplayValue("Timeline");
    const timelineCard = workflowCards().find((card) => card.contains(timelineInput));

    expect(timelineCard).toBeTruthy();
    expect(within(timelineCard).getByText("Auto-Ziel")).toBeInTheDocument();
    expect(within(timelineCard).getByText(/Zeitbezug|Zeitraum|Faktenbezug/)).toBeInTheDocument();
  });

  it("uses the temporary auto target from the selection popup without overwriting the manual base target", async () => {
    const user = userEvent.setup();
    const promptSpy = vi.spyOn(window, "prompt").mockReturnValue("");

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);
    fireEvent.change(textarea, {
      target: { value: "# Kapitel 1\n\nAm 1.4.2011 begann die Reise bis 2014." },
    });

    selectMarkdownText(textarea, "1.4.2011 begann die Reise bis 2014");

    const popup = await screen.findByRole("dialog", { name: "Auswahl-Aktionen" }, { timeout: 2600 });
    expect(screen.getByText(/Auto-Ziel: Timeline · Basis: Notizen/)).toBeInTheDocument();

    await user.click(within(popup).getByRole("button", { name: "Anker · Timeline" }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(
        `/api/chapters/${chapter.id}/anchors`,
        expect.objectContaining({
          selected_text: "1.4.2011 begann die Reise bis 2014",
          workflow_box_id: "box-2",
          note: "",
        }),
      );
    });

    promptSpy.mockRestore();
  });

  it("activates multiple workflow boxes together for a scene combination", async () => {
    const user = userEvent.setup();
    const promptSpy = vi
      .spyOn(window, "prompt")
      .mockReturnValueOnce("Figurenboard")
      .mockReturnValueOnce("Ereignisboard");
    const { container } = render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "+ Box" }));
    await user.click(screen.getByRole("button", { name: "+ Box" }));

    await screen.findByDisplayValue("Figurenboard");
    await screen.findByDisplayValue("Ereignisboard");

    const workflowCards = () => Array.from(container.querySelectorAll(".workflow-card"));
    const figuresCard = workflowCards().find((card) => card.contains(screen.getByDisplayValue("Figurenboard")));
    const eventsCard = workflowCards().find((card) => card.contains(screen.getByDisplayValue("Ereignisboard")));

    fireEvent.change(within(figuresCard).getByDisplayValue("custom"), { target: { value: "persons" } });
    fireEvent.change(within(eventsCard).getByDisplayValue("custom"), { target: { value: "events" } });

    const notesCard = workflowCards().find((card) => card.contains(screen.getByDisplayValue("Notizen")));
    await user.click(notesCard);

    const textarea = await openMarkdownEditor(user);
    fireEvent.change(textarea, {
      target: { value: "# Kapitel 1\n\nAm 1.4.2011 traf Mara eine Entscheidung." },
    });

    selectMarkdownText(textarea, "1.4.2011 traf Mara eine Entscheidung");

    await screen.findByRole("dialog", { name: "Auswahl-Aktionen" }, { timeout: 2600 });

    const timelineCard = workflowCards().find((card) => card.contains(screen.getByDisplayValue("Timeline")));

    expect(within(timelineCard).getByText("Auto-Ziel")).toBeInTheDocument();
    expect(within(figuresCard).getByText("Kombi")).toBeInTheDocument();
    expect(within(eventsCard).getByText("Kombi")).toBeInTheDocument();
    expect(screen.getByText(/Kombi: Zeit · Figur · Ereignis/)).toBeInTheDocument();

    promptSpy.mockRestore();
  });

  it("captures a markdown copy event directly into the clipboard list", async () => {
    const user = userEvent.setup();
    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const selectedText = "Alter Text";
    selectMarkdownText(textarea, selectedText);
    fireEvent.copy(textarea);

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/api/books/${book.id}/clipboard`, {
        chapter_id: chapter.id,
        content: selectedText,
        content_type: "text/markdown",
        source_anchor_id: "",
        is_pinned: false,
        slot: 0,
      });
    });

    expect(await screen.findByText(selectedText, { selector: ".context-card strong" })).toBeInTheDocument();
  });

  it("inserts a knowledge link, reinserts clipboard content, and saves the markdown roundtrip", async () => {
    const user = userEvent.setup();

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    await act(async () => {
      textarea.focus();
      textarea.setSelectionRange(textarea.value.length, textarea.value.length);
      fireEvent.select(textarea);
    });

    await user.click(screen.getByRole("button", { name: "Link" }));

    await waitFor(() => {
      expect(textarea).toHaveValue(`${chapter.markdown_content}[[Mara]]`);
    });

    await act(async () => {
      await new Promise((resolve) => window.requestAnimationFrame(resolve));
    });

    expect((await screen.findAllByText("Protagonistin mit scharfem Blick")).length).toBeGreaterThan(0);

    const selectedText = "Alter Text";
    const { start, end } = selectMarkdownText(textarea, selectedText);

    const clipboardButton = screen.getByRole("button", { name: "In Clipboard uebernehmen" });

    await waitFor(() => {
      expect(clipboardButton).toBeEnabled();
    });

    await user.click(clipboardButton);

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/api/books/${book.id}/clipboard`, {
        chapter_id: chapter.id,
        content: selectedText,
        content_type: "text/markdown",
        source_anchor_id: "",
        is_pinned: false,
        slot: 0,
      });
    });

    const insertButtons = await screen.findAllByRole("button", { name: "einfuegen" });
    await act(async () => {
      textarea.focus();
      textarea.setSelectionRange(textarea.value.length, textarea.value.length);
      fireEvent.select(textarea);
    });
    await user.click(insertButtons[0]);

    const expectedMarkdown = `${chapter.markdown_content}[[Mara]]${selectedText}`;

    await waitFor(() => {
      expect(textarea).toHaveValue(expectedMarkdown);
    });

    await user.click(screen.getByRole("button", { name: "Kapitel speichern" }));

    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith(`/api/chapters/${chapter.id}`, expect.objectContaining({
        title: chapter.title,
        markdown_content: expectedMarkdown,
        editor_json: JSON.stringify(markdownToDoc(expectedMarkdown)),
        save_mode: "manual",
        create_revision: true,
        revision_type: "manual",
        created_by: "easy-author-editor",
        autosave_reason: "manual_save",
        session_id: expect.any(String),
      }));
    });

    expect(api.put).toHaveBeenCalledTimes(1);
    expect(screen.getByText("Modus · Markdown")).toBeInTheDocument();
  });

  it("pins clipboard content into a slot, survives rich-markdown switching, and shows unresolved wiki refs", async () => {
    const user = userEvent.setup();

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Markdown" }));

    const textarea = await screen.findByPlaceholderText(
      "Schreibe hier direkt in Markdown. Wiki-Links wie [[Mara]] oder [[Ort:Alter Garten]] bleiben erhalten.",
    );

    const selectedText = "Alter Text";
    const start = textarea.value.indexOf(selectedText);
    const end = start + selectedText.length;

    textarea.focus();
    textarea.setSelectionRange(start, end);
    fireEvent.select(textarea);
    fireEvent.keyUp(textarea);

    const clipboardButton = screen.getByRole("button", { name: "In Clipboard uebernehmen" });

    await waitFor(() => {
      expect(clipboardButton).toBeEnabled();
    });

    await user.click(clipboardButton);

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/api/books/${book.id}/clipboard`, {
        chapter_id: chapter.id,
        content: selectedText,
        content_type: "text/markdown",
        source_anchor_id: "",
        is_pinned: false,
        slot: 0,
      });
    });

    const pinCheckbox = await screen.findByRole("checkbox", { name: "anpinnen" });
    await user.click(pinCheckbox);

    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/clipboard/clipboard-1", {
        content: selectedText,
        is_pinned: true,
        slot: 0,
      });
    });

    const slotInput = screen.getByRole("spinbutton");
    await user.clear(slotInput);
    await user.type(slotInput, "3");

    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/clipboard/clipboard-1", {
        content: selectedText,
        is_pinned: true,
        slot: 3,
      });
    });

    expect(screen.getByText("Alter Text", { selector: ".slot-card span" })).toBeInTheDocument();

    const markdownWithUnresolvedRef = `${textarea.value}\n\n[[Ort:Verlassenes Haus]]`;
    await act(async () => {
      fireEvent.change(textarea, {
        target: { value: markdownWithUnresolvedRef },
      });
    });

    expect((await screen.findAllByText("Ort:Verlassenes Haus")).length).toBeGreaterThan(0);
    expect(screen.getByText("Noch kein passender Wissenseintrag vorhanden.")).toBeInTheDocument();
    expect(
      screen.getByText("Offene Referenzen koennen links als Wissenseintrag angelegt oder umbenannt werden."),
    ).toBeInTheDocument();

    const insertButton = screen.getByRole("button", { name: "einfuegen" });
    textarea.focus();
    textarea.setSelectionRange(textarea.value.length, textarea.value.length);
    fireEvent.select(textarea);
    await user.click(insertButton);

    const expectedMarkdown = `${markdownWithUnresolvedRef}${selectedText}`;

    await waitFor(() => {
      expect(textarea).toHaveValue(expectedMarkdown);
    });

    await user.click(screen.getByRole("button", { name: "Rich" }));
    expect(await screen.findByTestId("editor-pane")).toHaveTextContent("Rich Editor: Kapitel 1");

    await user.click(screen.getByRole("button", { name: "Markdown" }));
    const returnedTextarea = await screen.findByPlaceholderText(MARKDOWN_PLACEHOLDER);

    await waitFor(() => {
      expect(returnedTextarea).toHaveValue(expectedMarkdown);
    });

    expect(screen.getByText("Alter Text", { selector: ".slot-card span" })).toBeInTheDocument();
    expect(screen.getAllByText("Ort:Verlassenes Haus").length).toBeGreaterThan(0);
  });

  it("keeps multiple pinned clipboard slots stable across reassignment and editor switching", async () => {
    const user = userEvent.setup();
    const { container } = render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const firstSelection = "Alter Text";
    selectMarkdownText(textarea, firstSelection);

    const clipboardButton = screen.getByRole("button", { name: "In Clipboard uebernehmen" });
    await waitFor(() => {
      expect(clipboardButton).toBeEnabled();
    });
    await user.click(clipboardButton);

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/api/books/${book.id}/clipboard`, {
        chapter_id: chapter.id,
        content: firstSelection,
        content_type: "text/markdown",
        source_anchor_id: "",
        is_pinned: false,
        slot: 0,
      });
    });

    const extendedMarkdown = `${chapter.markdown_content}\n\nZusatz Szene`;
    await act(async () => {
      fireEvent.change(textarea, {
        target: { value: extendedMarkdown },
      });
    });

    const secondSelection = "Zusatz Szene";
    selectMarkdownText(textarea, secondSelection);

    await waitFor(() => {
      expect(clipboardButton).toBeEnabled();
    });
    await user.click(clipboardButton);

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/api/books/${book.id}/clipboard`, {
        chapter_id: chapter.id,
        content: secondSelection,
        content_type: "text/markdown",
        source_anchor_id: "",
        is_pinned: false,
        slot: 0,
      });
    });

    const clipboardCards = Array.from(container.querySelectorAll(".clipboard-list article"));
    const firstCard = clipboardCards.find((card) => card.textContent.includes(firstSelection));
    const secondCard = clipboardCards.find((card) => card.textContent.includes(secondSelection));

    expect(firstCard).toBeTruthy();
    expect(secondCard).toBeTruthy();

    await user.click(within(secondCard).getByRole("checkbox", { name: "anpinnen" }));
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/clipboard/clipboard-2", {
        content: secondSelection,
        is_pinned: true,
        slot: 0,
      });
    });
    await user.clear(within(secondCard).getByRole("spinbutton"));
    await user.type(within(secondCard).getByRole("spinbutton"), "2");
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/clipboard/clipboard-2", {
        content: secondSelection,
        is_pinned: true,
        slot: 2,
      });
    });

    await user.click(within(firstCard).getByRole("checkbox", { name: "anpinnen" }));
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/clipboard/clipboard-1", {
        content: firstSelection,
        is_pinned: true,
        slot: 0,
      });
    });
    await user.clear(within(firstCard).getByRole("spinbutton"));
    await user.type(within(firstCard).getByRole("spinbutton"), "7");
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/clipboard/clipboard-1", {
        content: firstSelection,
        is_pinned: true,
        slot: 7,
      });
    });

    const slotCards = Array.from(container.querySelectorAll(".slot-card"));
    expect(slotCards[1].textContent).toContain(secondSelection);
    expect(slotCards[6].textContent).toContain(firstSelection);

    await user.clear(within(firstCard).getByRole("spinbutton"));
    await user.type(within(firstCard).getByRole("spinbutton"), "4");
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/clipboard/clipboard-1", {
        content: firstSelection,
        is_pinned: true,
        slot: 4,
      });
    });

    expect(slotCards[3].textContent).toContain(firstSelection);
    expect(slotCards[6].textContent).toContain("leer");

    await user.click(screen.getByRole("button", { name: "Rich" }));
    expect(await screen.findByTestId("editor-pane")).toHaveTextContent("Rich Editor: Kapitel 1");

    await user.click(screen.getByRole("button", { name: "Markdown" }));
    const returnedTextarea = await screen.findByPlaceholderText(MARKDOWN_PLACEHOLDER);

    await waitFor(() => {
      expect(returnedTextarea).toHaveValue(extendedMarkdown);
    });

    expect(slotCards[1].textContent).toContain(secondSelection);
    expect(slotCards[3].textContent).toContain(firstSelection);
  });

  it("opens the floating clipboard list and assigns a fixed slot directly there", async () => {
    const user = userEvent.setup();
    const { container } = render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const selectedText = "Alter Text";
    selectMarkdownText(textarea, selectedText);

    const clipboardButton = screen.getByRole("button", { name: "In Clipboard uebernehmen" });
    await waitFor(() => {
      expect(clipboardButton).toBeEnabled();
    });
    await user.click(clipboardButton);

    const paletteToggle = await screen.findByRole("button", { name: "Clipboard-Floating-Liste" });
    await user.click(paletteToggle);

    const palette = await screen.findByRole("dialog", { name: "Clipboard-Liste" });
    expect(within(palette).getByText("Gesammelte Ausschnitte und feste Slots")).toBeInTheDocument();

    await user.click(within(palette).getByRole("button", { name: "Slot 6 zuweisen" }));

    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/clipboard/clipboard-1", {
        content: selectedText,
        is_pinned: true,
        slot: 6,
      });
    });

    const slotCards = Array.from(container.querySelectorAll(".slot-card"));
    expect(slotCards[5].textContent).toContain(selectedText);

    await user.click(within(palette).getByRole("button", { name: "Schliessen" }));
    await waitFor(() => {
      expect(screen.queryByRole("dialog", { name: "Clipboard-Liste" })).not.toBeInTheDocument();
    });
  });

  it("inserts pinned slot content directly from the slot card", async () => {
    const user = userEvent.setup();
    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const selectedText = "Alter Text";
    selectMarkdownText(textarea, selectedText);

    await user.click(screen.getByRole("button", { name: "In Clipboard uebernehmen" }));

    const pinCheckbox = await screen.findByRole("checkbox", { name: "anpinnen" });
    await user.click(pinCheckbox);

    const slotInput = screen.getByRole("spinbutton");
    await user.clear(slotInput);
    await user.type(slotInput, "2");

    await act(async () => {
      textarea.focus();
      textarea.setSelectionRange(0, 0);
      fireEvent.select(textarea);
    });

    await user.click(screen.getByRole("button", { name: "Slot 2 einfuegen" }));

    await waitFor(() => {
      expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(`Alter Text${chapter.markdown_content}`);
    });
  });

  it("anchors a selection directly from the workflow cockpit without prompting", async () => {
    const user = userEvent.setup();
    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    selectMarkdownText(textarea, "Alter Text");

    const workflowCockpit = screen.getByText("Zielbox aktiv").closest(".workflow-target-card");
    expect(workflowCockpit).toBeTruthy();

    await user.click(within(workflowCockpit).getByRole("button", { name: "Auswahl ankern" }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/api/chapters/${chapter.id}/anchors`, expect.objectContaining({
        selected_text: "Alter Text",
        workflow_box_id: workflowBox.id,
        note: "",
      }));
    });

    await waitFor(() => {
      expect(within(workflowCockpit).getAllByText("Alter Text").length).toBeGreaterThan(0);
    });
  });

  it("deletes pinned clipboard entries cleanly and completes markdown autosave", async () => {
    const user = userEvent.setup();
    const { container } = render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const firstSelection = "Alter Text";
    selectMarkdownText(textarea, firstSelection);

    const clipboardButton = screen.getByRole("button", { name: "In Clipboard uebernehmen" });
    await waitFor(() => {
      expect(clipboardButton).toBeEnabled();
    });
    await user.click(clipboardButton);

    const extendedMarkdown = `${chapter.markdown_content}\n\nZusatz Szene`;
    await act(async () => {
      fireEvent.change(textarea, {
        target: { value: extendedMarkdown },
      });
    });

    const secondSelection = "Zusatz Szene";
    selectMarkdownText(textarea, secondSelection);
    await waitFor(() => {
      expect(clipboardButton).toBeEnabled();
    });
    await user.click(clipboardButton);

    const clipboardCards = () => Array.from(container.querySelectorAll(".clipboard-list article"));
    const slotCards = () => Array.from(container.querySelectorAll(".slot-card"));

    const firstCard = clipboardCards().find((card) => card.textContent.includes(firstSelection));
    const secondCard = clipboardCards().find((card) => card.textContent.includes(secondSelection));

    expect(firstCard).toBeTruthy();
    expect(secondCard).toBeTruthy();

    await user.click(within(firstCard).getByRole("checkbox", { name: "anpinnen" }));
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/clipboard/clipboard-1", {
        content: firstSelection,
        is_pinned: true,
        slot: 0,
      });
    });
    fireEvent.change(within(firstCard).getByRole("spinbutton"), { target: { value: "2" } });
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/clipboard/clipboard-1", {
        content: firstSelection,
        is_pinned: true,
        slot: 2,
      });
    });

    await user.click(within(secondCard).getByRole("checkbox", { name: "anpinnen" }));
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/clipboard/clipboard-2", {
        content: secondSelection,
        is_pinned: true,
        slot: 0,
      });
    });
    fireEvent.change(within(secondCard).getByRole("spinbutton"), { target: { value: "5" } });
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/clipboard/clipboard-2", {
        content: secondSelection,
        is_pinned: true,
        slot: 5,
      });
    });

    expect(slotCards()[1].textContent).toContain(firstSelection);
    expect(slotCards()[4].textContent).toContain(secondSelection);

    await user.click(within(firstCard).getByRole("button", { name: "loeschen" }));
    await waitFor(() => {
      expect(api.delete).toHaveBeenCalledWith("/api/clipboard/clipboard-1");
    });

    expect(slotCards()[1].textContent).toContain("leer");
    expect(slotCards()[4].textContent).toContain(secondSelection);

    const remainingCard = clipboardCards().find((card) => card.textContent.includes(secondSelection));
    expect(remainingCard).toBeTruthy();

    await user.click(within(remainingCard).getByRole("button", { name: "loeschen" }));
    await waitFor(() => {
      expect(api.delete).toHaveBeenCalledWith("/api/clipboard/clipboard-2");
    });

    expect(await screen.findByText("Noch keine Clipboard-Eintraege vorhanden.")).toBeInTheDocument();
    expect(slotCards()[4].textContent).toContain("leer");

    const autosaveMarkdown = `${extendedMarkdown}\n\nAutosave Probe`;
    await act(async () => {
      fireEvent.change(textarea, {
        target: { value: autosaveMarkdown },
      });
    });

    expect(screen.getByText("Autosave ausstehend")).toBeInTheDocument();

    await waitFor(
      () => {
        expect(api.put).toHaveBeenCalledWith(`/api/chapters/${chapter.id}`, expect.objectContaining({
          title: chapter.title,
          markdown_content: autosaveMarkdown,
          editor_json: JSON.stringify(markdownToDoc(autosaveMarkdown)),
          save_mode: "autosave",
          create_revision: false,
          autosave_reason: "idle_autosave",
          session_id: expect.any(String),
        }));
      },
      { timeout: 4000 },
    );

    await act(async () => {
      await new Promise((resolve) => window.setTimeout(resolve, 1300));
    });

    await waitFor(() => {
      expect(screen.getByText("Synchron")).toBeInTheDocument();
    });
  }, 10000);

  it("cancels pending autosave on chapter switch and avoids duplicate save after manual save", async () => {
    const user = userEvent.setup();

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const chapterOneDraft = `${chapter.markdown_content}\n\nUngelesener Entwurf`;
    await act(async () => {
      fireEvent.change(textarea, {
        target: { value: chapterOneDraft },
      });
    });

    expect(screen.getByText("Autosave ausstehend")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Kapitel 2/ }));

    await waitFor(() => {
      expect(screen.getByDisplayValue("Kapitel 2")).toBeInTheDocument();
      expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(chapterTwo.markdown_content);
    });

    await act(async () => {
      await new Promise((resolve) => window.setTimeout(resolve, 1700));
    });

    expect(api.put).not.toHaveBeenCalledWith(`/api/chapters/${chapter.id}`, expect.objectContaining({ markdown_content: chapterOneDraft }));

    const chapterTwoTextarea = screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER);
    const chapterTwoDraft = `${chapterTwo.markdown_content}\n\nManuell vor Autosave`;
    await act(async () => {
      fireEvent.change(chapterTwoTextarea, {
        target: { value: chapterTwoDraft },
      });
    });

    expect(screen.getByText("Autosave ausstehend")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Kapitel speichern" }));

    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith(`/api/chapters/${chapterTwo.id}`, expect.objectContaining({
        title: chapterTwo.title,
        markdown_content: chapterTwoDraft,
        editor_json: JSON.stringify(markdownToDoc(chapterTwoDraft)),
        save_mode: "manual",
        create_revision: true,
        revision_type: "manual",
        created_by: "easy-author-editor",
        autosave_reason: "manual_save",
        session_id: expect.any(String),
      }));
    });

    expect(api.put).toHaveBeenCalledTimes(1);

    await act(async () => {
      await new Promise((resolve) => window.setTimeout(resolve, 1700));
    });

    expect(api.put).toHaveBeenCalledTimes(1);

    await act(async () => {
      await new Promise((resolve) => window.setTimeout(resolve, 1300));
    });

    await waitFor(() => {
      expect(screen.getByText("Synchron")).toBeInTheDocument();
      expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(chapterTwoDraft);
    });
  }, 12000);

  it("shows autosave failures, keeps the draft, and recovers through manual retry", async () => {
    const user = userEvent.setup();

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const failedAutosaveDraft = `${chapter.markdown_content}\n\nFehlerfall`;
    const saveError = new Error("Speicherziel nicht erreichbar");
    api.put.mockImplementationOnce(async () => {
      throw saveError;
    });

    await act(async () => {
      fireEvent.change(textarea, {
        target: { value: failedAutosaveDraft },
      });
    });

    expect(screen.getByText("Autosave ausstehend")).toBeInTheDocument();

    await waitFor(
      () => {
        expect(screen.getByText("Fehler beim Speichern")).toBeInTheDocument();
        expect(screen.getByText("Speicherziel nicht erreichbar")).toBeInTheDocument();
      },
      { timeout: 4000 },
    );

    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(failedAutosaveDraft);
    expect(api.put).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole("button", { name: "Kapitel speichern" }));

    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith(`/api/chapters/${chapter.id}`, expect.objectContaining({
        title: chapter.title,
        markdown_content: failedAutosaveDraft,
        editor_json: JSON.stringify(markdownToDoc(failedAutosaveDraft)),
        save_mode: "manual",
        create_revision: true,
        revision_type: "manual",
        created_by: "easy-author-editor",
        autosave_reason: "manual_save",
        session_id: expect.any(String),
      }));
    });

    expect(api.put).toHaveBeenCalledTimes(2);
    expect(screen.queryByText("Speicherziel nicht erreichbar")).not.toBeInTheDocument();

    await act(async () => {
      await new Promise((resolve) => window.setTimeout(resolve, 1300));
    });

    await waitFor(() => {
      expect(screen.getByText("Synchron")).toBeInTheDocument();
      expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(failedAutosaveDraft);
    });
  }, 9000);

  it("updates repeated save errors and clears stale failure state after switching chapters", async () => {
    const user = userEvent.setup();

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const unstableDraft = `${chapter.markdown_content}\n\nMehrfachfehler`;
    const autosaveError = new Error("Autosave Verbindung verloren");
    const manualError = new Error("Manuelles Speichern weiterhin blockiert");

    api.put.mockImplementationOnce(async () => {
      throw autosaveError;
    });
    api.put.mockImplementationOnce(async () => {
      throw manualError;
    });

    await act(async () => {
      fireEvent.change(textarea, {
        target: { value: unstableDraft },
      });
    });

    await waitFor(
      () => {
        expect(screen.getByText("Fehler beim Speichern")).toBeInTheDocument();
        expect(screen.getByText("Autosave Verbindung verloren")).toBeInTheDocument();
      },
      { timeout: 4000 },
    );

    await user.click(screen.getByRole("button", { name: "Kapitel speichern" }));

    await waitFor(() => {
      expect(screen.getByText("Manuelles Speichern weiterhin blockiert")).toBeInTheDocument();
    });

    expect(screen.queryByText("Autosave Verbindung verloren")).not.toBeInTheDocument();
    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(unstableDraft);
    expect(api.put).toHaveBeenCalledTimes(2);

    await user.click(screen.getByRole("button", { name: /Kapitel 2/ }));

    await waitFor(() => {
      expect(screen.getByDisplayValue("Kapitel 2")).toBeInTheDocument();
      expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(chapterTwo.markdown_content);
      expect(screen.getByText("Synchron")).toBeInTheDocument();
    });

    expect(screen.queryByText("Manuelles Speichern weiterhin blockiert")).not.toBeInTheDocument();

    const chapterTwoTextarea = screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER);
    const recoveredDraft = `${chapterTwo.markdown_content}\n\nRecovery nach Wechsel`;
    await act(async () => {
      fireEvent.change(chapterTwoTextarea, {
        target: { value: recoveredDraft },
      });
    });

    await user.click(screen.getByRole("button", { name: "Kapitel speichern" }));

    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith(`/api/chapters/${chapterTwo.id}`, expect.objectContaining({
        title: chapterTwo.title,
        markdown_content: recoveredDraft,
        editor_json: JSON.stringify(markdownToDoc(recoveredDraft)),
        save_mode: "manual",
        create_revision: true,
        revision_type: "manual",
        created_by: "easy-author-editor",
        autosave_reason: "manual_save",
        session_id: expect.any(String),
      }));
    });

    expect(api.put).toHaveBeenCalledTimes(3);
    expect(screen.queryByText("Manuelles Speichern weiterhin blockiert")).not.toBeInTheDocument();

    await act(async () => {
      await new Promise((resolve) => window.setTimeout(resolve, 1300));
    });

    await waitFor(() => {
      expect(screen.getByText("Synchron")).toBeInTheDocument();
      expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(recoveredDraft);
    });
  }, 10000);

  it("recovers from anchor and clipboard failures without losing the current markdown draft", async () => {
    const user = userEvent.setup();
    const promptSpy = vi.spyOn(window, "prompt").mockReturnValue("Fehlernotiz");

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const resilientDraft = `${chapter.markdown_content}\n\nStabiler Entwurf`;
    await act(async () => {
      fireEvent.change(textarea, {
        target: { value: resilientDraft },
      });
    });

    const selectedText = "Alter Text";
    selectMarkdownText(textarea, selectedText);

    const anchorError = new Error("Anker konnte nicht gespeichert werden");
    const clipboardError = new Error("Clipboard ist temporaer gesperrt");
    const defaultPost = api.post.getMockImplementation();

    api.post.mockImplementationOnce(async (path, payload) => {
      if (path === `/api/chapters/${chapter.id}/anchors`) {
        throw anchorError;
      }
      return defaultPost(path, payload);
    });
    api.post.mockImplementationOnce(async (path, payload) => {
      if (path === `/api/books/${book.id}/clipboard`) {
        throw clipboardError;
      }
      return defaultPost(path, payload);
    });

    const anchorButton = screen.getByRole("button", { name: "Anker setzen" });
    const clipboardButton = screen.getByRole("button", { name: "In Clipboard uebernehmen" });

    await waitFor(() => {
      expect(anchorButton).toBeEnabled();
      expect(clipboardButton).toBeEnabled();
    });

    await user.click(anchorButton);

    await waitFor(() => {
      expect(screen.getByText("Anker konnte nicht gespeichert werden")).toBeInTheDocument();
    });

    expect(screen.queryByText("passage | Fehlernotiz")).not.toBeInTheDocument();
    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);

    selectMarkdownText(textarea, selectedText);
    await user.click(clipboardButton);

    await waitFor(() => {
      expect(screen.getByText("Clipboard ist temporaer gesperrt")).toBeInTheDocument();
    });

    expect(screen.queryByText("Anker konnte nicht gespeichert werden")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "einfuegen" })).not.toBeInTheDocument();
    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);

    selectMarkdownText(textarea, selectedText);
    await user.click(anchorButton);

    await waitFor(() => {
      expect(screen.getByText("passage | Fehlernotiz")).toBeInTheDocument();
    });

    expect(screen.queryByText("Clipboard ist temporaer gesperrt")).not.toBeInTheDocument();
    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);

    selectMarkdownText(textarea, selectedText);
    await user.click(clipboardButton);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "einfuegen" })).toBeInTheDocument();
    });

    expect(screen.queryByText("Clipboard ist temporaer gesperrt")).not.toBeInTheDocument();
    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);

    promptSpy.mockRestore();
  }, 9000);

  it("recovers from clipboard pin and delete failures while keeping slot state consistent", async () => {
    const user = userEvent.setup();
    const { container } = render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const resilientDraft = `${chapter.markdown_content}\n\nSlot Recovery`;
    await act(async () => {
      fireEvent.change(textarea, {
        target: { value: resilientDraft },
      });
    });

    const selectedText = "Alter Text";
    selectMarkdownText(textarea, selectedText);

    const clipboardButton = screen.getByRole("button", { name: "In Clipboard uebernehmen" });
    await waitFor(() => {
      expect(clipboardButton).toBeEnabled();
    });
    await user.click(clipboardButton);

    const insertButton = await screen.findByRole("button", { name: "einfuegen" });
    const clipboardCard = insertButton.closest("article");
    expect(clipboardCard).toBeTruthy();

    const pinError = new Error("Clipboard-Slot konnte nicht aktualisiert werden");
    api.put.mockImplementationOnce(async () => {
      throw pinError;
    });

    await user.click(within(clipboardCard).getByRole("checkbox", { name: "anpinnen" }));

    await waitFor(() => {
      expect(screen.getByText("Clipboard-Slot konnte nicht aktualisiert werden")).toBeInTheDocument();
    });

    const slotCards = Array.from(container.querySelectorAll(".slot-card"));
    expect(slotCards.every((slot) => slot.textContent.includes("leer"))).toBe(true);
    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);

    await user.click(within(clipboardCard).getByRole("checkbox", { name: "anpinnen" }));
    await waitFor(() => {
      expect(screen.queryByText("Clipboard-Slot konnte nicht aktualisiert werden")).not.toBeInTheDocument();
      expect(within(clipboardCard).getByRole("checkbox", { name: "anpinnen" })).toBeChecked();
    });

    fireEvent.change(within(clipboardCard).getByRole("spinbutton"), { target: { value: "4" } });
    await waitFor(() => {
      expect(slotCards[3].textContent).toContain(selectedText);
    });

    const deleteError = new Error("Clipboard-Eintrag konnte nicht geloescht werden");
    api.delete.mockImplementationOnce(async (path) => {
      if (path === "/api/clipboard/clipboard-1") {
        throw deleteError;
      }
      return null;
    });

    await user.click(within(clipboardCard).getByRole("button", { name: "loeschen" }));

    await waitFor(() => {
      expect(screen.getByText("Clipboard-Eintrag konnte nicht geloescht werden")).toBeInTheDocument();
    });

    expect(within(clipboardCard).getByRole("button", { name: "einfuegen" })).toBeInTheDocument();
    expect(slotCards[3].textContent).toContain(selectedText);
    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);

    await user.click(within(clipboardCard).getByRole("button", { name: "loeschen" }));

    await waitFor(() => {
      expect(screen.queryByText("Clipboard-Eintrag konnte nicht geloescht werden")).not.toBeInTheDocument();
      expect(screen.getByText("Noch keine Clipboard-Eintraege vorhanden.")).toBeInTheDocument();
    });

    expect(slotCards[3].textContent).toContain("leer");
    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);
  }, 9000);

  it("recovers from workflow and knowledge update failures without losing the chapter draft", async () => {
    const user = userEvent.setup();
    const { container } = render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const resilientDraft = `${chapter.markdown_content}\n\nKontext bleibt stabil`;
    await act(async () => {
      fireEvent.change(textarea, {
        target: { value: resilientDraft },
      });
    });

    const workflowTitleInput = container.querySelector(".workflow-card input");
    expect(workflowTitleInput).toBeTruthy();

    const workflowError = new Error("Workflow-Box konnte nicht gespeichert werden");
    api.put.mockImplementationOnce(async () => {
      throw workflowError;
    });

    fireEvent.change(workflowTitleInput, { target: { value: "Notizen Plus" } });
    await act(async () => {
      fireEvent.blur(workflowTitleInput);
    });

    await waitFor(() => {
      expect(screen.getByText("Workflow-Box konnte nicht gespeichert werden")).toBeInTheDocument();
    });

    expect(workflowTitleInput).toHaveValue("Notizen Plus");
    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);

    await act(async () => {
      fireEvent.blur(workflowTitleInput);
    });

    await waitFor(() => {
      expect(screen.queryByText("Workflow-Box konnte nicht gespeichert werden")).not.toBeInTheDocument();
      expect(workflowTitleInput).toHaveValue("Notizen Plus");
    });

    const knowledgeSummary = screen.getByPlaceholderText("Kurze Zusammenfassung");
    const knowledgeError = new Error("Wissenseintrag konnte nicht gespeichert werden");
    api.put.mockImplementationOnce(async () => {
      throw knowledgeError;
    });

    fireEvent.change(knowledgeSummary, { target: { value: "Neue Wissensnotiz" } });
    await act(async () => {
      fireEvent.blur(knowledgeSummary);
    });

    await waitFor(() => {
      expect(screen.getByText("Wissenseintrag konnte nicht gespeichert werden")).toBeInTheDocument();
    });

    expect(knowledgeSummary).toHaveValue("Neue Wissensnotiz");
    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);

    await act(async () => {
      fireEvent.blur(knowledgeSummary);
    });

    await waitFor(() => {
      expect(screen.queryByText("Wissenseintrag konnte nicht gespeichert werden")).not.toBeInTheDocument();
      expect(knowledgeSummary).toHaveValue("Neue Wissensnotiz");
    });

    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);
  }, 9000);

  it("keeps the working draft stable across failed project and book switches, then clears errors on recovery", async () => {
    const user = userEvent.setup();

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const resilientDraft = `${chapter.markdown_content}\n\nWechsel bleibt stabil`;
    await act(async () => {
      fireEvent.change(textarea, {
        target: { value: resilientDraft },
      });
    });

    const bookSwitchError = new Error("Buchwechsel momentan nicht moeglich");
    const defaultGet = api.get.getMockImplementation();
    api.get.mockImplementationOnce(async (path) => {
      if (path === `/api/books/${bookTwo.id}`) {
        throw bookSwitchError;
      }
      return defaultGet(path);
    });

    await user.click(screen.getByRole("button", { name: /Buch wechseln/ }));
    await user.click(screen.getByRole("button", { name: /Buch Zwei/ }));

    await waitFor(() => {
      expect(screen.getByText("Buchwechsel momentan nicht moeglich")).toBeInTheDocument();
    });

    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);
    expect(screen.getByDisplayValue("Kapitel 1")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Buch wechseln/ }));
    await user.click(screen.getByRole("button", { name: /Buch Eins/ }));

    await waitFor(() => {
      expect(screen.queryByText("Buchwechsel momentan nicht moeglich")).not.toBeInTheDocument();
      expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);
    });

    const projectSwitchError = new Error("Projektwechsel momentan nicht moeglich");
    api.get.mockImplementationOnce(async (path) => {
      if (path === `/api/projects/${projectTwo.id}`) {
        throw projectSwitchError;
      }
      return defaultGet(path);
    });

    await user.click(screen.getByRole("button", { name: /Projekt wechseln/ }));
    await user.click(screen.getByRole("button", { name: /Sachbuchprojekt/ }));

    await waitFor(() => {
      expect(screen.getByText("Projektwechsel momentan nicht moeglich")).toBeInTheDocument();
    });

    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);
    expect(screen.getByDisplayValue("Kapitel 1")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Projekt wechseln/ }));
    await user.click(screen.getAllByRole("button", { name: /Romanprojekt/ }).at(-1));

    await waitFor(() => {
      expect(screen.queryByText("Projektwechsel momentan nicht moeglich")).not.toBeInTheDocument();
      expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);
      expect(screen.getByDisplayValue("Kapitel 1")).toBeInTheDocument();
    });
  }, 10000);

  it("keeps the current work stable across failed project, book, and chapter creation attempts", async () => {
    const user = userEvent.setup();
    const promptSpy = vi.spyOn(window, "prompt");

    render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const resilientDraft = `${chapter.markdown_content}\n\nCreate bleibt stabil`;
    await act(async () => {
      fireEvent.change(textarea, {
        target: { value: resilientDraft },
      });
    });

    const defaultPost = api.post.getMockImplementation();

    promptSpy.mockReturnValueOnce("Projekt Fehlerfall");
    api.post.mockImplementationOnce(async (path, payload) => {
      if (path === "/api/projects") {
        throw new Error("Projekt konnte nicht angelegt werden");
      }
      return defaultPost(path, payload);
    });

    await user.click(screen.getByRole("button", { name: "+ Projekt" }));

    await waitFor(() => {
      expect(screen.getByText("Projekt konnte nicht angelegt werden")).toBeInTheDocument();
    });

    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);
    expect(screen.getByDisplayValue("Kapitel 1")).toBeInTheDocument();

    promptSpy.mockReturnValueOnce("Buch Fehlerfall");
    api.post.mockImplementationOnce(async (path, payload) => {
      if (path === `/api/projects/${project.id}/books`) {
        throw new Error("Buch konnte nicht angelegt werden");
      }
      return defaultPost(path, payload);
    });

    await user.click(screen.getByRole("button", { name: "+ Buch" }));

    await waitFor(() => {
      expect(screen.getByText("Buch konnte nicht angelegt werden")).toBeInTheDocument();
    });

    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);
    expect(screen.getByDisplayValue("Kapitel 1")).toBeInTheDocument();

    promptSpy.mockReturnValueOnce("Kapitel Fehlerfall");
    api.post.mockImplementationOnce(async (path, payload) => {
      if (path === `/api/books/${book.id}/chapters`) {
        throw new Error("Kapitel konnte nicht angelegt werden");
      }
      return defaultPost(path, payload);
    });

    await user.click(screen.getByRole("button", { name: "+ Kapitel" }));

    await waitFor(() => {
      expect(screen.getByText("Kapitel konnte nicht angelegt werden")).toBeInTheDocument();
    });

    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(resilientDraft);
    expect(screen.getByDisplayValue("Kapitel 1")).toBeInTheDocument();

    promptSpy.mockReturnValueOnce("Kapitel Bonus");
    await user.click(screen.getByRole("button", { name: "+ Kapitel" }));

    await waitFor(() => {
      expect(screen.queryByText("Kapitel konnte nicht angelegt werden")).not.toBeInTheDocument();
      expect(screen.getByDisplayValue("Kapitel Bonus")).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /Kapitel 1/ }));
    await waitFor(() => {
      expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(chapter.markdown_content);
    });

    promptSpy.mockReturnValueOnce("Buch Neu");
    await user.click(screen.getByRole("button", { name: "+ Buch" }));

    await waitFor(() => {
      expect(screen.queryByText("Buch konnte nicht angelegt werden")).not.toBeInTheDocument();
      expect(screen.getByRole("button", { name: /Buch Neu/ })).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /Buch wechseln/ }));
    await user.click(screen.getByRole("button", { name: /Buch Eins/ }));
    await waitFor(() => {
      expect(screen.getByDisplayValue("Kapitel 1")).toBeInTheDocument();
    });

    promptSpy.mockReturnValueOnce("Projekt Neu");
    await user.click(screen.getByRole("button", { name: "+ Projekt" }));

    await waitFor(() => {
      expect(screen.queryByText("Projekt konnte nicht angelegt werden")).not.toBeInTheDocument();
      expect(screen.getByRole("button", { name: /Projekt Neu/ })).toBeInTheDocument();
    });

    promptSpy.mockRestore();
  }, 12000);

  it("keeps autosave stable while workflow sidebar actions run in parallel", async () => {
    const user = userEvent.setup();
    const promptSpy = vi.spyOn(window, "prompt");
    const { container } = render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const parallelDraft = `${chapter.markdown_content}\n\nAutosave und Workflow laufen parallel`;
    await act(async () => {
      fireEvent.change(textarea, {
        target: { value: parallelDraft },
      });
    });

    expect(screen.getByText("Autosave ausstehend")).toBeInTheDocument();

    promptSpy.mockReturnValueOnce("Sprint Box");
    await user.click(screen.getByRole("button", { name: "+ Box" }));

    await waitFor(() => {
      expect(screen.getByText(/Zielbox: Sprint Box/)).toBeInTheDocument();
    });

    await waitFor(() => {
      expect(screen.getByDisplayValue("Sprint Box")).toBeInTheDocument();
    });

    const workflowCards = () => Array.from(container.querySelectorAll(".workflow-card"));
    const sprintTitleInput = screen.getByDisplayValue("Sprint Box");
    const sprintCard = workflowCards().find((card) => card.contains(sprintTitleInput));
    expect(sprintCard).toBeTruthy();

    fireEvent.change(sprintTitleInput, { target: { value: "Sprint Box Plus" } });
    await act(async () => {
      fireEvent.blur(sprintTitleInput);
    });

    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/workflow-boxes/box-3", {
        title: "Sprint Box Plus",
        type: "custom",
        tags: [],
        is_collapsed: false,
      });
    });

    const sprintTypeSelect = within(sprintCard).getByDisplayValue("custom");
    fireEvent.change(sprintTypeSelect, { target: { value: "research" } });
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/workflow-boxes/box-3", {
        title: "Sprint Box Plus",
        type: "research",
        tags: [],
        is_collapsed: false,
      });
    });

    const collapsedCheckbox = within(sprintCard).getByRole("checkbox");
    fireEvent.click(collapsedCheckbox);
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/workflow-boxes/box-3", {
        title: "Sprint Box Plus",
        type: "research",
        tags: [],
        is_collapsed: true,
      });
    });

    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(parallelDraft);

    await waitFor(
      () => {
        expect(api.put).toHaveBeenCalledWith(`/api/chapters/${chapter.id}`, expect.objectContaining({
          title: chapter.title,
          markdown_content: parallelDraft,
          editor_json: JSON.stringify(markdownToDoc(parallelDraft)),
          save_mode: "autosave",
          create_revision: false,
          autosave_reason: "idle_autosave",
          session_id: expect.any(String),
        }));
      },
      { timeout: 4000 },
    );

    expect(
      api.put.mock.calls.filter(([path]) => path === `/api/chapters/${chapter.id}` && path.startsWith("/api/chapters/")),
    ).toHaveLength(1);

    await act(async () => {
      await new Promise((resolve) => window.setTimeout(resolve, 1300));
    });

    await waitFor(() => {
      expect(screen.getByText("Synchron")).toBeInTheDocument();
      expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(parallelDraft);
      expect(screen.getByText(/Zielbox: Sprint Box Plus/)).toBeInTheDocument();
    });

    promptSpy.mockRestore();
  }, 12000);

  it("keeps multiple workflow boxes stable under full editor activity", async () => {
    const user = userEvent.setup();
    const promptSpy = vi.spyOn(window, "prompt");
    const { container } = render(<App />);

    expect(await screen.findByDisplayValue("Kapitel 1")).toBeInTheDocument();
    const textarea = await openMarkdownEditor(user);

    const fullLoadDraft = `${chapter.markdown_content}\n\nWorkflow Vollgas fuer alle Boxen`;
    await act(async () => {
      fireEvent.change(textarea, {
        target: { value: fullLoadDraft },
      });
    });

    expect(screen.getByText("Autosave ausstehend")).toBeInTheDocument();

    const workflowCards = () => Array.from(container.querySelectorAll(".workflow-card"));
    const cardByValue = (value) => {
      const input = screen.getByDisplayValue(value);
      return workflowCards().find((card) => card.contains(input));
    };

    const rootTitleInput = screen.getByDisplayValue("Notizen");
    await act(async () => {
      fireEvent.change(rootTitleInput, { target: { value: "Notizen Master" } });
      fireEvent.blur(rootTitleInput);
    });

    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/workflow-boxes/box-1", {
        title: "Notizen Master",
        type: "notes",
        tags: ["Alter", "Text", "idee"],
        is_collapsed: false,
      });
    });

    promptSpy.mockReturnValueOnce("Figurenfokus");
    await user.click(screen.getByRole("button", { name: "+ Box" }));
    await waitFor(() => {
      expect(screen.getByDisplayValue("Figurenfokus")).toBeInTheDocument();
    });

    promptSpy.mockReturnValueOnce("Research Sprint");
    await user.click(screen.getByRole("button", { name: "+ Box" }));
    await waitFor(() => {
      expect(screen.getByDisplayValue("Research Sprint")).toBeInTheDocument();
      expect(screen.getByText(/Zielbox: Research Sprint/)).toBeInTheDocument();
    });

    const figuresCard = cardByValue("Figurenfokus");
    expect(figuresCard).toBeTruthy();
    await user.click(figuresCard);
    await waitFor(() => {
      expect(screen.getByText(/Zielbox: Figurenfokus/)).toBeInTheDocument();
    });

    const figuresTypeSelect = within(figuresCard).getByDisplayValue("custom");
    await act(async () => {
      fireEvent.change(figuresTypeSelect, { target: { value: "persons" } });
    });
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/workflow-boxes/box-3", {
        title: "Figurenfokus",
        type: "persons",
        tags: [],
        is_collapsed: false,
      });
    });

    selectMarkdownText(textarea, "Alter Text");
    promptSpy.mockReturnValueOnce("Anker zu Figuren");
    await user.click(screen.getByRole("button", { name: "Anker setzen" }));
    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/api/chapters/${chapter.id}/anchors`, expect.objectContaining({
        selected_text: "Alter Text",
        workflow_box_id: "box-3",
        note: "Anker zu Figuren",
      }));
    });

    const researchCard = cardByValue("Research Sprint");
    expect(researchCard).toBeTruthy();
    await user.click(researchCard);
    await waitFor(() => {
      expect(screen.getByText(/Zielbox: Research Sprint/)).toBeInTheDocument();
    });

    const researchTitleInput = screen.getByDisplayValue("Research Sprint");
    await act(async () => {
      fireEvent.change(researchTitleInput, { target: { value: "Research Sprint Plus" } });
      fireEvent.blur(researchTitleInput);
    });

    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/workflow-boxes/box-4", {
        title: "Research Sprint Plus",
        type: "custom",
        tags: [],
        is_collapsed: false,
      });
      expect(screen.getByText(/Zielbox: Research Sprint Plus/)).toBeInTheDocument();
    });

    const researchCardUpdated = cardByValue("Research Sprint Plus");
    expect(researchCardUpdated).toBeTruthy();
    const researchTypeSelect = within(researchCardUpdated).getByDisplayValue("custom");
    await act(async () => {
      fireEvent.change(researchTypeSelect, { target: { value: "research" } });
    });
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/workflow-boxes/box-4", {
        title: "Research Sprint Plus",
        type: "research",
        tags: [],
        is_collapsed: false,
      });
    });

    const researchCollapsedCheckbox = within(researchCardUpdated).getByRole("checkbox");
    await act(async () => {
      fireEvent.click(researchCollapsedCheckbox);
    });
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith("/api/workflow-boxes/box-4", {
        title: "Research Sprint Plus",
        type: "research",
        tags: [],
        is_collapsed: true,
      });
    });

    selectMarkdownText(textarea, "Workflow Vollgas");
    promptSpy.mockReturnValueOnce("Anker zu Research");
    await user.click(screen.getByRole("button", { name: "Anker setzen" }));
    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/api/chapters/${chapter.id}/anchors`, expect.objectContaining({
        selected_text: "Workflow Vollgas",
        workflow_box_id: "box-4",
        note: "Anker zu Research",
      }));
    });

    const masterCard = cardByValue("Notizen Master");
    expect(masterCard).toBeTruthy();
    await user.click(masterCard);
    await waitFor(() => {
      expect(screen.getByText(/Zielbox: Notizen Master/)).toBeInTheDocument();
    });

    selectMarkdownText(textarea, "Kapitel 1");
    promptSpy.mockReturnValueOnce("Anker zu Notizen");
    await user.click(screen.getByRole("button", { name: "Anker setzen" }));
    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/api/chapters/${chapter.id}/anchors`, expect.objectContaining({
        selected_text: "Kapitel 1",
        workflow_box_id: "box-1",
        note: "Anker zu Notizen",
      }));
    });

    await waitFor(() => {
      expect(screen.getByText("passage | Anker zu Figuren")).toBeInTheDocument();
      expect(screen.getByText("passage | Anker zu Research")).toBeInTheDocument();
      expect(screen.getByText("passage | Anker zu Notizen")).toBeInTheDocument();
      expect(screen.getAllByText("Alter Text").length).toBeGreaterThan(0);
      expect(screen.getAllByText("Workflow Vollgas").length).toBeGreaterThan(0);
      expect(screen.getAllByText("Kapitel 1").length).toBeGreaterThan(0);
    });

    expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(fullLoadDraft);

    await waitFor(
      () => {
        expect(api.put).toHaveBeenCalledWith(`/api/chapters/${chapter.id}`, expect.objectContaining({
          title: chapter.title,
          markdown_content: fullLoadDraft,
          editor_json: JSON.stringify(markdownToDoc(fullLoadDraft)),
          save_mode: "autosave",
          create_revision: false,
          autosave_reason: "idle_autosave",
          session_id: expect.any(String),
        }));
      },
      { timeout: 4000 },
    );

    expect(api.put.mock.calls.filter(([path]) => path === `/api/chapters/${chapter.id}`)).toHaveLength(1);

    await act(async () => {
      await new Promise((resolve) => window.setTimeout(resolve, 1300));
    });

    await waitFor(() => {
      expect(screen.getByText("Synchron")).toBeInTheDocument();
      expect(screen.getByPlaceholderText(MARKDOWN_PLACEHOLDER)).toHaveValue(fullLoadDraft);
      expect(screen.getByText(/Zielbox: Notizen Master/)).toBeInTheDocument();
    });

    promptSpy.mockRestore();
  }, 15000);
});
