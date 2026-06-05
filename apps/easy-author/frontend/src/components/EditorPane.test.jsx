import { render } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import EditorPane from "./EditorPane";
import { markdownToDoc } from "../lib/markdown";

const setContent = vi.fn();
const focus = vi.fn();
let currentDocument = null;
let currentSelectionEmpty = true;
let currentTextBetween = vi.fn(() => "");
const domListeners = new Map();

vi.mock("@tiptap/react", () => ({
  EditorContent: ({ editor }) => <div data-testid="editor-content">{editor ? "editor" : "leer"}</div>,
  useEditor: vi.fn((config) => {
    if (currentDocument === null) {
      currentDocument = config.content;
    }
    return {
      commands: {
        setContent: (nextContent) => {
          currentDocument = nextContent;
          setContent(nextContent);
        },
        focus,
      },
      chain: () => ({
        focus: () => ({
          insertContent: () => ({
            run: () => {},
          }),
        }),
      }),
      getJSON: () => currentDocument,
      state: {
        selection: {
          empty: currentSelectionEmpty,
          from: 3,
          to: 12,
        },
        doc: {
          textBetween: (...args) => currentTextBetween(...args),
          content: {
            size: 42,
          },
        },
      },
      view: {
        dom: {
          addEventListener: (name, handler) => {
            domListeners.set(name, handler);
          },
          removeEventListener: (name) => {
            domListeners.delete(name);
          },
        },
      },
    };
  }),
}));

vi.mock("@tiptap/starter-kit", () => ({
  default: {},
}));

vi.mock("@tiptap/extension-placeholder", () => ({
  default: {
    configure: () => ({}),
  },
}));

vi.mock("@tiptap/extension-table", () => ({
  default: {
    configure: () => ({}),
  },
}));

vi.mock("@tiptap/extension-table-row", () => ({
  default: {},
}));

vi.mock("@tiptap/extension-table-header", () => ({
  default: {},
}));

vi.mock("@tiptap/extension-table-cell", () => ({
  default: {},
}));

describe("EditorPane synchronization", () => {
  beforeEach(() => {
    setContent.mockClear();
    focus.mockClear();
    currentDocument = null;
    currentSelectionEmpty = true;
    currentTextBetween = vi.fn(() => "");
    domListeners.clear();
  });

  it("does not push identical chapter content back into the editor on local rerenders", () => {
    const chapterDoc = markdownToDoc("# Kapitel 1\n\nAlter Text");
    const chapter = {
      id: "chapter-1",
      title: "Kapitel 1",
      markdown_content: "# Kapitel 1\n\nAlter Text",
      editor_json: JSON.stringify(chapterDoc),
    };

    const { rerender } = render(
      <EditorPane
        chapter={chapter}
        pinnedSlots={[]}
        onDocumentChange={() => {}}
        onSelectionChange={() => {}}
      />,
    );

    expect(setContent).toHaveBeenCalledTimes(1);

    rerender(
      <EditorPane
        chapter={{ ...chapter }}
        pinnedSlots={[]}
        onDocumentChange={() => {}}
        onSelectionChange={() => {}}
      />,
    );

    expect(setContent).toHaveBeenCalledTimes(1);

    const updatedMarkdown = "# Kapitel 1\n\nAlter Text\n\nNeuer Absatz";
    const updatedDoc = markdownToDoc(updatedMarkdown);

    rerender(
      <EditorPane
        chapter={{
          ...chapter,
          markdown_content: updatedMarkdown,
          editor_json: JSON.stringify(updatedDoc),
        }}
        pinnedSlots={[]}
        onDocumentChange={() => {}}
        onSelectionChange={() => {}}
      />,
    );

    expect(setContent).toHaveBeenCalledTimes(2);
    expect(setContent).toHaveBeenLastCalledWith(updatedDoc);
  });

  it("forwards native copy events from the rich editor selection", () => {
    currentSelectionEmpty = false;
    currentTextBetween = vi
      .fn()
      .mockReturnValueOnce("Rich Auswahl")
      .mockReturnValueOnce("vorher")
      .mockReturnValueOnce("nachher");

    const onClipboardCapture = vi.fn();

    render(
      <EditorPane
        chapter={{
          id: "chapter-1",
          title: "Kapitel 1",
          markdown_content: "# Kapitel 1\n\nAlter Text",
          editor_json: JSON.stringify(markdownToDoc("# Kapitel 1\n\nAlter Text")),
        }}
        pinnedSlots={[]}
        onDocumentChange={() => {}}
        onSelectionChange={() => {}}
        onClipboardCapture={onClipboardCapture}
      />,
    );

    domListeners.get("copy")?.();

    expect(onClipboardCapture).toHaveBeenCalledWith({
      selected_text: "Rich Auswahl",
      start_offset: 3,
      end_offset: 12,
      context_before: "vorher",
      context_after: "nachher",
    });
  });
});
