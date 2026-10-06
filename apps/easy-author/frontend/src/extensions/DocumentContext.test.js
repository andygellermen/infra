import { describe, expect, it } from "vitest";
import DocumentContext, { MarkdownHighlight, mergeContextTypes } from "./DocumentContext";

describe("DocumentContext", () => {
  it("keeps several independent context types on one durable mark", () => {
    expect(mergeContextTypes("comment,link", ["work_item", "comment"])).toBe("comment,link,work_item");
    expect(mergeContextTypes("comment,link", ["comment"], { remove: true })).toBe("link");
  });

  it("declares persistent anchor attributes and separate markdown highlighting", () => {
    expect(DocumentContext.name).toBe("documentContext");
    expect(MarkdownHighlight.name).toBe("markdownHighlight");
    const attributes = DocumentContext.config.addAttributes();
    expect(attributes).toHaveProperty("anchorId");
    expect(attributes).toHaveProperty("contextTypes");
    expect(attributes).toHaveProperty("contextState");
  });
});
