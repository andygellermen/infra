import { forwardRef, useEffect, useImperativeHandle, useMemo, useRef } from "react";
import { EditorContent, useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Placeholder from "@tiptap/extension-placeholder";
import { docToMarkdown, markdownToDoc } from "../lib/markdown";

const EditorPane = forwardRef(function EditorPane(
  { chapter, pinnedSlots, onDocumentChange, onSelectionChange },
  ref,
) {
  const pinnedSlotsRef = useRef(pinnedSlots);

  useEffect(() => {
    pinnedSlotsRef.current = pinnedSlots;
  }, [pinnedSlots]);

  const initialContent = useMemo(() => {
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
  }, [chapter]);

  const editor = useEditor({
    immediatelyRender: false,
    extensions: [
      StarterKit,
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
    },
    onUpdate: ({ editor: activeEditor }) => {
      const json = activeEditor.getJSON();
      onDocumentChange?.({
        editor_json: JSON.stringify(json),
        markdown_content: docToMarkdown(json),
      });
    },
  });

  useEffect(() => {
    if (!editor || !chapter) {
      return;
    }
    const nextContent = chapter.editor_json
      ? (() => {
          try {
            return JSON.parse(chapter.editor_json);
          } catch {
            return markdownToDoc(chapter.markdown_content || "");
          }
        })()
      : markdownToDoc(chapter.markdown_content || "");
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

  useImperativeHandle(
    ref,
    () => ({
      getSelectionPayload() {
        if (!editor || editor.state.selection.empty) {
          return null;
        }
        const { from, to } = editor.state.selection;
        const selectedText = editor.state.doc.textBetween(from, to, " ");
        const contextBefore = editor.state.doc.textBetween(Math.max(0, from - 60), from, " ");
        const contextAfter = editor.state.doc.textBetween(to, Math.min(editor.state.doc.content.size, to + 60), " ");
        return {
          selected_text: selectedText,
          start_offset: from,
          end_offset: to,
          context_before: contextBefore,
          context_after: contextAfter,
        };
      },
      insertClipboardContent(content) {
        if (!editor) {
          return;
        }
        editor.chain().focus().insertContent(content).run();
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

  return <EditorContent editor={editor} />;
});

export default EditorPane;
