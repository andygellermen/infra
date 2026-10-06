import { Mark, mergeAttributes } from "@tiptap/core";

function splitTypes(value) {
  const values = Array.isArray(value) ? value : String(value || "").split(",");
  return values.map((item) => String(item).trim()).filter(Boolean);
}

export function mergeContextTypes(current, incoming, { remove = false } = {}) {
  const types = splitTypes(current);
  const changes = new Set(splitTypes(incoming));
  const next = remove ? types.filter((type) => !changes.has(type)) : [...types, ...[...changes].filter((type) => !types.includes(type))];
  return next.join(",");
}

const DocumentContext = Mark.create({
  name: "documentContext",
  inclusive: false,
  addAttributes() {
    return {
      anchorId: {
        default: "",
        parseHTML: (element) => element.getAttribute("data-document-anchor-id") || element.getAttribute("data-review-comment-id") || "",
        renderHTML: ({ anchorId }) => anchorId ? { "data-document-anchor-id": anchorId } : {},
      },
      contextTypes: {
        default: "",
        parseHTML: (element) => element.getAttribute("data-context-types") || (element.hasAttribute("data-review-comment-id") ? "comment" : ""),
        renderHTML: ({ contextTypes }) => contextTypes ? { "data-context-types": contextTypes } : {},
      },
      contextState: {
        default: "open",
        parseHTML: (element) => element.getAttribute("data-context-state") || element.getAttribute("data-review-comment-state") || "open",
        renderHTML: ({ contextState }) => contextState ? { "data-context-state": contextState } : {},
      },
    };
  },
  parseHTML() {
    return [{ tag: "span[data-document-anchor-id]" }, { tag: "span[data-review-comment-id]" }];
  },
  renderHTML({ HTMLAttributes }) {
    return ["span", mergeAttributes(HTMLAttributes, { class: "document-context-mark" }), 0];
  },
  addCommands() {
    return {
      applyDocumentContext: (attributes) => ({ commands }) => commands.setMark(this.name, attributes),
      removeDocumentContext: () => ({ commands }) => commands.unsetMark(this.name),
      focusDocumentAnchor: (anchorId) => ({ state, commands }) => {
        let range = null;
        state.doc.descendants((node, position) => {
          if (range || !node.isText) return;
          const mark = node.marks?.find((candidate) => candidate.type.name === this.name && candidate.attrs?.anchorId === anchorId);
          if (mark) range = { from: position, to: position + node.nodeSize };
        });
        return range ? commands.setTextSelection(range) : false;
      },
    };
  },
});

export const MarkdownHighlight = Mark.create({
  name: "markdownHighlight",
  inclusive: false,
  parseHTML() {
    return [{ tag: "mark[data-markdown-highlight]" }, { tag: "mark" }];
  },
  renderHTML({ HTMLAttributes }) {
    return ["mark", mergeAttributes(HTMLAttributes, { "data-markdown-highlight": "true" }), 0];
  },
  addCommands() {
    return {
      toggleMarkdownHighlight: () => ({ commands }) => commands.toggleMark(this.name),
    };
  },
});

export default DocumentContext;
