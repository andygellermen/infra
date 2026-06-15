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
import ReviewComment from "../extensions/ReviewComment";

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

function selectionAnchorFromEditor(editor) {
  if (typeof window === "undefined") {
    return null;
  }
  const selection = window.getSelection?.();
  if (selection?.rangeCount) {
    const rect = selection.getRangeAt(0).getBoundingClientRect();
    if (rect.width || rect.height) {
      return {
        x: rect.left + rect.width / 2,
        y: rect.top - 14,
      };
    }
  }
  const position = editor?.state?.selection?.from;
  if (!editor?.view?.coordsAtPos || typeof position !== "number") {
    return null;
  }
  const coords = editor.view.coordsAtPos(position);
  return {
    x: coords.left,
    y: coords.top - 14,
  };
}

function resolveChapterContent(chapter) {
  if (!chapter) {
    return markdownToDoc("");
  }
  const markdownContent = String(chapter.markdown_content || "");
  if (chapter.editor_json) {
    try {
      const parsed = JSON.parse(chapter.editor_json);
      if (markdownContent.trim()) {
        const serialized = docToMarkdown(parsed).replace(/\r\n/g, "\n").trim();
        const normalizedMarkdown = markdownContent.replace(/\r\n/g, "\n").trim();
        if (serialized !== normalizedMarkdown) {
          return markdownToDoc(markdownContent);
        }
      }
      return parsed;
    } catch {
      return markdownToDoc(markdownContent);
    }
  }
  return markdownToDoc(markdownContent);
}

function serializeDocument(document) {
  return JSON.stringify(document || null);
}

const EditorPane = forwardRef(function EditorPane(
  {
    chapter,
    pinnedSlots,
    activeReviewCommentId,
    reviewComments,
    onDocumentChange,
    onSelectionChange,
    onClipboardCapture,
    onSelectionContextChange,
    onReviewCommentActivate,
  },
  ref,
) {
  const pinnedSlotsRef = useRef(pinnedSlots);
  const clipboardCaptureRef = useRef(onClipboardCapture);
  const selectionContextRef = useRef(onSelectionContextChange);
  const reviewActivationRef = useRef(onReviewCommentActivate);
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

  useEffect(() => {
    selectionContextRef.current = onSelectionContextChange;
  }, [onSelectionContextChange]);

  useEffect(() => {
    reviewActivationRef.current = onReviewCommentActivate;
  }, [onReviewCommentActivate]);

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
      ReviewComment,
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
      const table = activeEditor.isActive("table");
      const payload = selectionPayloadFromState(activeEditor.state);
      const anchor = selectionAnchorFromEditor(activeEditor) || {};
      onSelectionChange?.(!activeEditor.state.selection.empty);
      setActiveStates({
        table,
        blockquote: activeEditor.isActive("blockquote"),
      });
      selectionContextRef.current?.(
        !payload?.selected_text && !table
          ? null
          : {
              kind: table && !payload?.selected_text ? "table" : "text",
              source: "rich",
              tableActive: table,
              payload,
              ...anchor,
            },
      );
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
    onSelectionContextChange?.(null);
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

  useEffect(() => {
    if (!editor?.view?.dom?.querySelectorAll) {
      return;
    }
    const commentsById = new Map((reviewComments || []).map((comment) => [comment.id, comment]));
    const nodes = editor.view.dom.querySelectorAll("[data-review-comment-id]");
    nodes.forEach((node) => {
      const commentId = node.getAttribute("data-review-comment-id");
      const isActive = commentId === activeReviewCommentId;
      const comment = commentsById.get(commentId || "");
      const phase = comment?.comment_phase || "unlinked";
      node.classList.toggle("is-active-review-comment", isActive);
      node.setAttribute("data-review-comment-phase", phase);
    });
  }, [activeReviewCommentId, editor, chapter?.id, chapter?.editor_json, reviewComments]);

  useEffect(() => {
    if (!editor?.view?.dom?.addEventListener) {
      return undefined;
    }
    const target = editor.view.dom;
    const handler = (event) => {
      const marker = event.target?.closest?.("[data-review-comment-id]");
      if (!marker) {
        return;
      }
      const rect = marker.getBoundingClientRect();
      reviewActivationRef.current?.(marker.getAttribute("data-review-comment-id") || "", {
        x: rect.left + rect.width / 2,
        y: rect.bottom + 16,
      });
    };
    target.addEventListener("click", handler);
    return () => target.removeEventListener("click", handler);
  }, [editor]);

  function rangesForReviewComment(commentId) {
    if (!editor || !commentId) {
      return [];
    }
    const ranges = [];
    editor.state.doc.descendants((node, position) => {
      if (!node.isText || !Array.isArray(node.marks)) {
        return;
      }
      const hasComment = node.marks.some(
        (mark) => mark.type.name === "reviewComment" && mark.attrs?.commentId === commentId,
      );
      if (hasComment) {
        ranges.push({
          from: position,
          to: position + node.nodeSize,
        });
      }
    });
    return ranges;
  }

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
      applyReviewCommentMark(comment) {
        if (!editor || !comment?.id) {
          return;
        }
        const from = Math.max(1, Number(comment.start_offset) || 0);
        const to = Math.max(from, Number(comment.end_offset) || 0);
        if (from === to) {
          return;
        }
        const markType = editor.state.schema.marks.reviewComment;
        if (!markType) {
          return;
        }
        let transaction = editor.state.tr.removeMark(from, to, markType);
        transaction = transaction.addMark(
          from,
          to,
              markType.create({
                commentId: comment.id,
                commentType: comment.comment_type || "comment",
                commentState: comment.status || "open",
                commentPhase: comment.comment_phase || "unlinked",
              }),
            );
        editor.view.dispatch(transaction);
      },
      removeReviewCommentMark(commentId) {
        if (!editor || !commentId) {
          return;
        }
        const markType = editor.state.schema.marks.reviewComment;
        if (!markType) {
          return;
        }
        const ranges = rangesForReviewComment(commentId);
        if (ranges.length === 0) {
          return;
        }
        let transaction = editor.state.tr;
        ranges.forEach((range) => {
          transaction = transaction.removeMark(range.from, range.to, markType);
        });
        editor.view.dispatch(transaction);
      },
      replaceReviewCommentText({ commentId, text, keepMark = false, commentType = "comment", commentState = "open", commentPhase = "unlinked" }) {
        if (!editor || !commentId) {
          return;
        }
        const ranges = rangesForReviewComment(commentId);
        if (ranges.length === 0) {
          return;
        }
        const from = Math.min(...ranges.map((range) => range.from));
        const to = Math.max(...ranges.map((range) => range.to));
        let transaction = editor.state.tr.insertText(text || "", from, to);
        if (keepMark && text) {
          const markType = editor.state.schema.marks.reviewComment;
          if (markType) {
            transaction = transaction.addMark(
              from,
              from + String(text).length,
              markType.create({
                commentId,
                commentType,
                commentState,
                commentPhase,
              }),
            );
          }
        }
        editor.view.dispatch(transaction);
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
