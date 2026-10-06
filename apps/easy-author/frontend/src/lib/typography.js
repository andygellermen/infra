export const TYPOGRAPHY_RANGES = Object.freeze({
  h1Size: [24, 72], h2Size: [22, 64], h3Size: [20, 56], h4Size: [18, 48], h5Size: [16, 40], h6Size: [14, 36],
  bodySize: [12, 32], quoteSize: [12, 32], tableSize: [10, 28], textWidth: [480, 1400],
  firstLineIndent: [0, 80], lineHeight: [1, 2.5], paragraphSpacing: [0, 64],
});

export const DEFAULT_TYPOGRAPHY = Object.freeze({
  bodyFont: "Source Serif 4", headingFont: "Source Serif 4",
  h1Size: 40, h2Size: 34, h3Size: 29, h4Size: 25, h5Size: 21, h6Size: 18,
  bodySize: 18, quoteSize: 18, tableSize: 15, textWidth: 760,
  firstLineIndent: 0, lineHeight: 1.7, paragraphSpacing: 14,
});

const FONT_FIELDS = new Set(["bodyFont", "headingFont"]);
const GLOBAL_TYPOGRAPHY_KEY = "easy-author.global-typography.v1";

export function normalizeTypographyOverrides(value) {
  if (!value || typeof value !== "object" || Array.isArray(value)) return {};
  const normalized = {};
  for (const [key, raw] of Object.entries(value)) {
    if (FONT_FIELDS.has(key)) {
      if (typeof raw === "string" && raw.trim() && raw.trim().length <= 100) normalized[key] = raw.trim();
      continue;
    }
    const range = TYPOGRAPHY_RANGES[key];
    if (range && typeof raw === "number" && Number.isFinite(raw) && raw >= range[0] && raw <= range[1]) normalized[key] = raw;
  }
  return normalized;
}

export function resolveTypography(globalDefaults, bookOverrides) {
  return {
    ...DEFAULT_TYPOGRAPHY,
    ...normalizeTypographyOverrides(globalDefaults),
    ...normalizeTypographyOverrides(bookOverrides),
  };
}

export function loadGlobalTypography(storage) {
  try {
    const stored = JSON.parse(storage?.getItem?.(GLOBAL_TYPOGRAPHY_KEY) || "{}");
    return resolveTypography(DEFAULT_TYPOGRAPHY, stored);
  } catch {
    return { ...DEFAULT_TYPOGRAPHY };
  }
}

export function saveGlobalTypography(storage, value) {
  try {
    storage?.setItem?.(GLOBAL_TYPOGRAPHY_KEY, JSON.stringify(resolveTypography(DEFAULT_TYPOGRAPHY, value)));
  } catch {
    // Storage may be unavailable; the in-memory preference remains usable.
  }
}
