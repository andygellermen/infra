export const GLOBAL_APPEARANCE_STORAGE_KEY = "easy-author.editor-appearance.v1";
export const SESSION_WORK_VIEW_STORAGE_KEY = "easy-author.work-mode.v1";

export const DEFAULT_GLOBAL_APPEARANCE = Object.freeze({
  themeMode: "system",
  fontFamily: "serif",
  googleFontName: "Cormorant Garamond",
  fontSize: 18,
  lineHeight: 1.8,
  contentWidth: 860,
  fullscreenContentWidth: 1040,
  fullscreenBackdrop: "linen",
  surfacePreset: "warm",
  caretColor: "#76c7ff",
});

const THEMES = new Set(["light", "dark", "system"]);
const FONT_FAMILIES = new Set(["serif", "sans", "mono", "google"]);
const SURFACES = new Set(["warm", "paper", "night"]);
const BACKDROPS = new Set(["linen", "paper", "dusk", "night"]);
const CONTENT_WIDTHS = new Set([640, 720, 860, 960, 1040, 1160]);
const FULLSCREEN_WIDTHS = new Set([860, 1040, 1200, 1360]);
const WORK_VIEWS = new Set(["clean", "intense", "review"]);
const LEGACY_WORK_VIEWS = new Map([
  ["write", "clean"],
  ["structure", "intense"],
  ["review", "review"],
]);

export function normalizeGlobalAppearance(value) {
  const candidate = value && typeof value === "object" ? value : {};
  const next = { ...DEFAULT_GLOBAL_APPEARANCE, ...candidate };
  next.themeMode = THEMES.has(next.themeMode) ? next.themeMode : DEFAULT_GLOBAL_APPEARANCE.themeMode;
  next.fontFamily = FONT_FAMILIES.has(next.fontFamily) ? next.fontFamily : DEFAULT_GLOBAL_APPEARANCE.fontFamily;
  next.googleFontName =
    String(next.googleFontName || DEFAULT_GLOBAL_APPEARANCE.googleFontName).trim().slice(0, 80) ||
    DEFAULT_GLOBAL_APPEARANCE.googleFontName;
  next.surfacePreset = SURFACES.has(next.surfacePreset) ? next.surfacePreset : DEFAULT_GLOBAL_APPEARANCE.surfacePreset;
  next.fullscreenBackdrop = BACKDROPS.has(next.fullscreenBackdrop)
    ? next.fullscreenBackdrop
    : DEFAULT_GLOBAL_APPEARANCE.fullscreenBackdrop;
  next.fontSize = Math.min(24, Math.max(16, Number(next.fontSize) || DEFAULT_GLOBAL_APPEARANCE.fontSize));
  next.lineHeight = Math.min(2.2, Math.max(1.5, Number(next.lineHeight) || DEFAULT_GLOBAL_APPEARANCE.lineHeight));
  next.contentWidth = CONTENT_WIDTHS.has(Number(next.contentWidth))
    ? Number(next.contentWidth)
    : DEFAULT_GLOBAL_APPEARANCE.contentWidth;
  next.fullscreenContentWidth = FULLSCREEN_WIDTHS.has(Number(next.fullscreenContentWidth))
    ? Number(next.fullscreenContentWidth)
    : DEFAULT_GLOBAL_APPEARANCE.fullscreenContentWidth;
  next.caretColor = /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.test(String(next.caretColor || "").trim())
    ? String(next.caretColor).trim()
    : DEFAULT_GLOBAL_APPEARANCE.caretColor;
  return next;
}

export function loadGlobalAppearance(storage) {
  if (!storage) {
    return { ...DEFAULT_GLOBAL_APPEARANCE };
  }
  try {
    const raw = storage.getItem(GLOBAL_APPEARANCE_STORAGE_KEY);
    return raw ? normalizeGlobalAppearance(JSON.parse(raw)) : { ...DEFAULT_GLOBAL_APPEARANCE };
  } catch {
    return { ...DEFAULT_GLOBAL_APPEARANCE };
  }
}

export function saveGlobalAppearance(storage, value) {
  try {
    storage?.setItem(GLOBAL_APPEARANCE_STORAGE_KEY, JSON.stringify(normalizeGlobalAppearance(value)));
  } catch {
    // Browsers may deny local persistence; the in-memory preference remains usable.
  }
}

export function loadSessionWorkView(storage) {
  try {
    const stored = storage?.getItem(SESSION_WORK_VIEW_STORAGE_KEY);
    if (!stored) {
      return null;
    }
    const normalized = LEGACY_WORK_VIEWS.get(stored) || stored;
    return WORK_VIEWS.has(normalized) ? normalized : null;
  } catch {
    return null;
  }
}

export function saveSessionWorkView(storage, value) {
  if (!WORK_VIEWS.has(value)) {
    return;
  }
  try {
    storage?.setItem(SESSION_WORK_VIEW_STORAGE_KEY, value);
  } catch {
    // Browsers may deny local persistence; the session state remains usable.
  }
}
