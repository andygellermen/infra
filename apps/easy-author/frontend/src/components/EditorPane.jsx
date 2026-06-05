import { forwardRef, useEffect, useImperativeHandle, useMemo, useRef, useState } from "react";
import { EditorContent, useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Placeholder from "@tiptap/extension-placeholder";
import Table from "@tiptap/extension-table";
import TableCell from "@tiptap/extension-table-cell";
import TableHeader from "@tiptap/extension-table-header";
import TableRow from "@tiptap/extension-table-row";
import { TextSelection } from "@tiptap/pm/state";
import { docToMarkdown, markdownToDoc, normalizeRichTableMarkdown } from "../lib/markdown";
import FootnoteReference from "../extensions/FootnoteReference";
import FootnoteDefinition from "../extensions/FootnoteDefinition";

function selectionPayloadFromState(state) {
  if (!state || state.selection.empty) {
    return null;
  }
  const { from, to } = state.selection;
  const selectedText = state.doc.textBetween(from, to, " ");
  const contextBefore = state.doc.textBetween(Math.max(0, from - 60), from, " ");
  const contextAfter = state.doc.textBetween(to, Math.min(state.doc.content.size, to + 60), " ");
  return {
    selected_text: selectedText,
    start_offset: from,
    end_offset: to,
    context_before: contextBefore,
    context_after: contextAfter,
  };
}

function resolveChapterContent(chapter) {
  if (!chapter) {
    return markdownToDoc("");
  }
  if (chapter.editor_json) {
    try {
      return JSON.parse(chapter.editor_json);
    } catch {
      return markdownToDoc(chapter.markdown_content || "");
    }
  }
  return markdownToDoc(chapter.markdown_content || "");
}

function serializeDocument(document) {
  return JSON.stringify(document || null);
}

const EditorPane = forwardRef(function EditorPane(
  { chapter, pinnedSlots, onDocumentChange, onSelectionChange, onClipboardCapture },
  ref,
) {
  const pinnedSlotsRef = useRef(pinnedSlots);
  const clipboardCaptureRef = useRef(onClipboardCapture);
  const lastChapterIdRef = useRef(null);
  const [activeStates, setActiveStates] = useState({
    table: false,
    blockquote: false,
  });

  useEffect(() => {
    pinnedSlotsRef.current = pinnedSlots;
  }, [pinnedSlots]);

  useEffect(() => {
    clipboardCaptureRef.current = onClipboardCapture;
  }, [onClipboardCapture]);

  const initialContent = useMemo(() => resolveChapterContent(chapter), [chapter]);

  const editor = useEditor({
    immediatelyRender: false,
    extensions: [
      StarterKit,
      Table.configure({
        resizable: true,
      }),
      TableRow,
      TableHeader,
      TableCell,
      FootnoteReference,
      FootnoteDefinition,
      Placeholder.configure({
        placeholder: "Schreibe hier an deinem Kapitel weiter ...",
      }),
    ],
    content: initialContent,
    editorProps: {
      attributes: {
        class: "editor-surface",
      },
    },
    onSelectionUpdate: ({ editor: activeEditor }) => {
      onSelectionChange?.(!activeEditor.state.selection.empty);
      setActiveStates({
        table: activeEditor.isActive("table"),
        blockquote: activeEditor.isActive("blockquote"),
      });
    },
    onUpdate: ({ editor: activeEditor }) => {
      const json = activeEditor.getJSON();
      const normalized = normalizeRichTableMarkdown(json);
      const nextJson = normalized.changed ? normalized.doc : json;
      if (normalized.changed) {
        activeEditor.commands.setContent(nextJson, false);
      }
      onDocumentChange?.({
        editor_json: JSON.stringify(nextJson),
        markdown_content: docToMarkdown(nextJson),
      });
      setActiveStates({
        table: activeEditor.isActive("table"),
        blockquote: activeEditor.isActive("blockquote"),
      });
    },
  });

  useEffect(() => {
    if (!editor || !chapter) {
      lastChapterIdRef.current = chapter?.id || null;
      return;
    }
    const nextContent = resolveChapterContent(chapter);
    const chapterChanged = lastChapterIdRef.current !== chapter.id;
    const contentChanged = serializeDocument(editor.getJSON()) !== serializeDocument(nextContent);
    lastChapterIdRef.current = chapter.id;

    if (!chapterChanged && !contentChanged) {
      return;
    }

    editor.commands.setContent(nextContent, false);
    onSelectionChange?.(false);
  }, [chapter?.id, chapter?.editor_json, chapter?.markdown_content, editor, onSelectionChange]);

  useEffect(() => {
    if (!editor) {
      return undefined;
    }
    const target = editor.view.dom;
    const handler = (event) => {
      if (!(event.metaKey || event.ctrlKey) || !event.shiftKey) {
        return;
      }
      const slot = Number(event.key);
      if (Number.isNaN(slot) || slot < 1 || slot > 9) {
        return;
      }
      const item = pinnedSlotsRef.current.find((entry) => entry.slot === slot && entry.is_pinned);
      if (!item) {
        return;
      }
      event.preventDefault();
      editor.chain().focus().insertContent(item.content).run();
    };
    target.addEventListener("keydown", handler);
    return () => target.removeEventListener("keydown", handler);
  }, [editor]);

  useEffect(() => {
    if (!editor) {
      return undefined;
    }
    const target = editor.view.dom;
    const handler = () => {
      const payload = selectionPayloadFromState(editor.state);
      if (!payload?.selected_text) {
        return;
      }
      clipboardCaptureRef.current?.(payload);
    };
    target.addEventListener("copy", handler);
    target.addEventListener("cut", handler);
    return () => {
      target.removeEventListener("copy", handler);
      target.removeEventListener("cut", handler);
    };
  }, [editor]);

  useImperativeHandle(
    ref,
    () => ({
      getSelectionPayload() {
        if (!editor) {
          return null;
        }
        return selectionPayloadFromState(editor.state);
      },
      insertClipboardContent(content) {
        if (!editor) {
          return;
        }
        editor.chain().focus().insertContent(content).run();
      },
      insertText(content) {
        if (!editor) {
          return;
        }
        editor.chain().focus().insertContent(content).run();
      },
      getDocumentSnapshot() {
        if (!editor) {
          return null;
        }
        const json = editor.getJSON();
        return {
          editor_json: JSON.stringify(json),
          markdown_content: docToMarkdown(json),
        };
      },
      insertTable() {
        if (!editor) {
          return;
        }
        editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run();
      },
      toggleBlockquote() {
        editor?.chain().focus().toggleBlockquote().run();
      },
      insertFootnote() {
        if (!editor) {
          return;
        }

        const noteIds = [];
        editor.state.doc.descendants((node) => {
          if (node.type.name === "footnoteReference" || node.type.name === "footnoteDefinition") {
            const numericId = Number(node.attrs?.noteId);
            if (!Number.isNaN(numericId)) {
              noteIds.push(numericId);
            }
          }
        });

        const nextId = String((noteIds.length ? Math.max(...noteIds) : 0) + 1);
        const referenceNode = editor.state.schema.nodes.footnoteReference.create({ noteId: nextId });
        const definitionNode = editor.state.schema.nodes.footnoteDefinition.create(
          { noteId: nextId },
          [editor.state.schema.nodes.paragraph.create()],
        );

        let transaction = editor.state.tr.replaceSelectionWith(referenceNode);
        const selectionAnchor = transaction.selection.from;
        transaction = transaction.insert(transaction.doc.content.size, definitionNode);
        transaction = transaction.setSelection(TextSelection.create(transaction.doc, selectionAnchor));
        editor.view.dispatch(transaction);
        editor.commands.focus();
      },
      addColumnAfter() {
        editor?.chain().focus().addColumnAfter().run();
      },
      addRowAfter() {
        editor?.chain().focus().addRowAfter().run();
      },
      deleteColumn() {
        editor?.chain().focus().deleteColumn().run();
      },
      deleteRow() {
        editor?.chain().focus().deleteRow().run();
      },
      toggleHeaderRow() {
        editor?.chain().focus().toggleHeaderRow().run();
      },
      deleteTable() {
        editor?.chain().focus().deleteTable().run();
      },
      isTableActive() {
        return editor?.isActive("table") || false;
      },
      isBlockquoteActive() {
        return editor?.isActive("blockquote") || false;
      },
      focus() {
        editor?.commands.focus();
      },
    }),
    [editor],
  );

  if (!chapter) {
    return (
      <div className="editor-empty">
        <p>Waehle links ein Kapitel aus oder lege ein neues Kapitel an.</p>
      </div>
    );
  }

  return (
    <div className="editor-content-shell">
      {activeStates.table ? (
        <div className="context-table-toolbar" aria-label="Tabellen-Kontextmenue">
          <button type="button" className="ghost-button" onClick={() => editor?.chain().focus().addColumnAfter().run()}>
            + Spalte
          </button>
          <button type="button" className="ghost-button" onClick={() => editor?.chain().focus().addRowAfter().run()}>
            + Zeile
          </button>
          <button type="button" className="ghost-button" onClick={() => editor?.chain().focus().toggleHeaderRow().run()}>
            Kopfzeile
          </button>
          <button type="button" className="ghost-button" onClick={() => editor?.chain().focus().deleteColumn().run()}>
            Spalte -
          </button>
          <button type="button" className="ghost-button" onClick={() => editor?.chain().focus().deleteRow().run()}>
            Zeile -
          </button>
          <button type="button" className="ghost-button" onClick={() => editor?.chain().focus().deleteTable().run()}>
            Tabelle loeschen
          </button>
        </div>
      ) : null}
      <EditorContent editor={editor} />
    </div>
  );
});

export default EditorPane;
