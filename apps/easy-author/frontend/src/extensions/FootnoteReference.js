import { Node, mergeAttributes } from "@tiptap/core";

const FootnoteReference = Node.create({
  name: "footnoteReference",
  group: "inline",
  inline: true,
  atom: true,
  selectable: true,

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
        tag: "sup[data-footnote-reference]",
      },
    ];
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "sup",
      mergeAttributes(HTMLAttributes, {
        "data-footnote-reference": HTMLAttributes.noteId || "",
        class: "footnote-reference",
      }),
      ["button", { type: "button" }, `[^${HTMLAttributes.noteId || ""}]`],
    ];
  },

  renderText({ node }) {
    return `[^${node.attrs.noteId || ""}]`;
  },
});

export default FootnoteReference;
