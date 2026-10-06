import { describe, expect, it } from "vitest";
import {
  DEFAULT_GLOBAL_APPEARANCE,
  loadGlobalAppearance,
  loadSessionWorkView,
  saveGlobalAppearance,
  saveSessionWorkView,
} from "./uiPreferences";

function memoryStorage(initial = {}) {
  const values = new Map(Object.entries(initial));
  return {
    getItem: (key) => values.get(key) ?? null,
    setItem: (key, value) => values.set(key, String(value)),
    value: (key) => values.get(key),
  };
}

describe("global appearance preferences", () => {
  it("uses safe defaults when storage is absent or malformed", () => {
    expect(loadGlobalAppearance(null)).toEqual(DEFAULT_GLOBAL_APPEARANCE);
    expect(loadGlobalAppearance(memoryStorage({ "easy-author.editor-appearance.v1": "{" }))).toEqual(
      DEFAULT_GLOBAL_APPEARANCE,
    );
  });

  it("normalizes unsupported theme and appearance values", () => {
    const storage = memoryStorage({
      "easy-author.editor-appearance.v1": JSON.stringify({
        themeMode: "neon",
        fontFamily: "comic",
        surfacePreset: "laser",
        fontSize: 99,
      }),
    });

    expect(loadGlobalAppearance(storage)).toMatchObject({
      themeMode: "system",
      fontFamily: "serif",
      surfacePreset: "warm",
      fontSize: 24,
    });
  });

  it("does not throw when storage access fails", () => {
    const storage = {
      getItem: () => {
        throw new Error("blocked");
      },
      setItem: () => {
        throw new Error("blocked");
      },
    };

    expect(() => loadGlobalAppearance(storage)).not.toThrow();
    expect(() => saveGlobalAppearance(storage, DEFAULT_GLOBAL_APPEARANCE)).not.toThrow();
  });
});

describe("session work view preferences", () => {
  it("returns null for absent and unsupported values", () => {
    expect(loadSessionWorkView(memoryStorage())).toBeNull();
    expect(loadSessionWorkView(memoryStorage({ "easy-author.work-mode.v1": "unknown" }))).toBeNull();
  });

  it.each([
    ["write", "clean"],
    ["structure", "intense"],
    ["review", "review"],
    ["clean", "clean"],
    ["intense", "intense"],
  ])("maps stored %s to %s", (stored, expected) => {
    expect(loadSessionWorkView(memoryStorage({ "easy-author.work-mode.v1": stored }))).toBe(expected);
  });

  it("saves only supported work views and ignores storage failures", () => {
    const storage = memoryStorage();
    saveSessionWorkView(storage, "intense");
    expect(storage.value("easy-author.work-mode.v1")).toBe("intense");
    saveSessionWorkView(storage, "unknown");
    expect(storage.value("easy-author.work-mode.v1")).toBe("intense");
    expect(() => saveSessionWorkView({ setItem: () => { throw new Error("blocked"); } }, "clean")).not.toThrow();
  });
});
