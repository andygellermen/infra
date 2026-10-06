import { describe, expect, it } from "vitest";
import { BOOK_COLOR_PALETTE, allocateBookColor } from "./bookColorAllocation";

describe("allocateBookColor", () => {
  it("keeps existing assignments stable and chooses the most distinct remaining color", () => {
    const existing = new Map([["book-a", BOOK_COLOR_PALETTE[0]]]);
    const allocated = allocateBookColor(existing, "book-b", BOOK_COLOR_PALETTE);
    expect(allocated.get("book-a")).toBe(BOOK_COLOR_PALETTE[0]);
    expect(allocated.get("book-b")).toBe(BOOK_COLOR_PALETTE[1]);
    expect(existing.has("book-b")).toBe(false);
  });

  it("reuses a released color without persisting assignments", () => {
    const reduced = new Map([["book-b", BOOK_COLOR_PALETTE[1]]]);
    const allocated = allocateBookColor(reduced, "book-c", BOOK_COLOR_PALETTE);
    expect(allocated.get("book-c")).toBe(BOOK_COLOR_PALETTE[0]);
    expect(() => JSON.stringify(allocated)).not.toThrow();
    expect(Object.keys(window.localStorage)).toHaveLength(0);
  });
});
