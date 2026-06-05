import { Node, mergeAttributes } from "@tiptap/core";

const FootnoteDefinition = Node.create({
  name: "footnoteDefinition",
  group: "block",
  content: "block+",
  defining: true,
  isolating: true,

  addAttributes() {
    return {
      noteId: {
        default: "",
      },
    };
  },

  parseHTML() {
    return [
      {
        tag: "div[data-footnote-definition]",
      },
    ];
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(HTMLAttributes, {
        "data-footnote-definition": HTMLAttributes.noteId || "",
        class: "footnote-definition",
      }),
      ["div", { class: "footnote-definition-label" }, `[^${HTMLAttributes.noteId || ""}]`],
      ["div", { class: "footnote-definition-body" }, 0],
    ];
  },
});

export default FootnoteDefinition;
