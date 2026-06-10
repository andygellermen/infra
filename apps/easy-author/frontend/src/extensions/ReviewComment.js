import { Mark, mergeAttributes } from "@tiptap/core";

const REVIEW_COMMENT_CLASS = "review-comment-mark";

const ReviewComment = Mark.create({
  name: "reviewComment",

  inclusive: false,

  addAttributes() {
    return {
      commentId: {
        default: "",
        parseHTML: (element) => element.getAttribute("data-review-comment-id") || "",
        renderHTML: (attributes) =>
          attributes.commentId ? { "data-review-comment-id": attributes.commentId } : {},
      },
      commentType: {
        default: "comment",
        parseHTML: (element) => element.getAttribute("data-review-comment-type") || "comment",
        renderHTML: (attributes) =>
          attributes.commentType ? { "data-review-comment-type": attributes.commentType } : {},
      },
      commentState: {
        default: "open",
        parseHTML: (element) => element.getAttribute("data-review-comment-state") || "open",
        renderHTML: (attributes) =>
          attributes.commentState ? { "data-review-comment-state": attributes.commentState } : {},
      },
    };
  },

  parseHTML() {
    return [{ tag: "span[data-review-comment-id]" }];
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "span",
      mergeAttributes(HTMLAttributes, {
        class: REVIEW_COMMENT_CLASS,
      }),
      0,
    ];
  },
});

export default ReviewComment;
