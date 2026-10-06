import { describe, expect, it, vi } from "vitest";
import { DEFAULT_TYPOGRAPHY, loadGlobalTypography, normalizeTypographyOverrides, resolveTypography, saveGlobalTypography } from "./typography";

describe("book typography", () => {
  it("inherits each unset field without mutating either input", () => {
    const globalDefaults = { ...DEFAULT_TYPOGRAPHY, bodyFont: "Literata", bodySize: 18 };
    const overrides = { h1Size: 42, firstLineIndent: 24 };
    const beforeDefaults = { ...globalDefaults };
    const beforeOverrides = { ...overrides };
    expect(resolveTypography(globalDefaults, overrides)).toEqual({
      ...globalDefaults,
      h1Size: 42,
      firstLineIndent: 24,
    });
    expect(globalDefaults).toEqual(beforeDefaults);
    expect(overrides).toEqual(beforeOverrides);
  });

  it("keeps independent sizes for headings, body, quotes, and tables", () => {
    const sizes = { h1Size: 44, h2Size: 38, h3Size: 32, h4Size: 27, h5Size: 23, h6Size: 20, bodySize: 18, quoteSize: 17, tableSize: 15 };
    expect(normalizeTypographyOverrides(sizes)).toEqual(sizes);
  });

  it("accepts layout values within safe ranges and rejects invalid values field by field", () => {
    expect(normalizeTypographyOverrides({
      bodyFont: "  Literata  ", headingFont: "Inter", textWidth: 980,
      firstLineIndent: 32, lineHeight: 1.75, paragraphSpacing: 20,
      bodySize: 200, h1Size: "large", unknown: 42,
    })).toEqual({
      bodyFont: "Literata", headingFont: "Inter", textWidth: 980,
      firstLineIndent: 32, lineHeight: 1.75, paragraphSpacing: 20,
    });
  });

  it("resets a book to inheritance with empty overrides", () => {
    expect(normalizeTypographyOverrides({})).toEqual({});
    expect(resolveTypography(DEFAULT_TYPOGRAPHY, {})).toEqual(DEFAULT_TYPOGRAPHY);
  });

  it("loads and saves the global standard defensively", () => {
    const storage = { getItem: () => JSON.stringify({ bodySize: 20, h1Size: 500 }), setItem: vi.fn() };
    expect(loadGlobalTypography(storage)).toEqual({ ...DEFAULT_TYPOGRAPHY, bodySize: 20 });
    saveGlobalTypography(storage, { ...DEFAULT_TYPOGRAPHY, bodySize: 21 });
    expect(storage.setItem).toHaveBeenCalledWith("easy-author.global-typography.v1", expect.stringContaining('"bodySize":21'));
  });
});
