export const BOOK_COLOR_PALETTE = [
  "#006D77", "#C43D2F", "#6A4C93", "#8A5A00", "#007A3D", "#A23E8C", "#2F5DA8", "#7A4E2D",
];

function channels(hex) {
  const value = hex.replace("#", "");
  return [0, 2, 4].map((offset) => Number.parseInt(value.slice(offset, offset + 2), 16));
}

function distance(left, right) {
  const a = channels(left);
  const b = channels(right);
  const redMean = (a[0] + b[0]) / 2;
  const red = a[0] - b[0];
  const green = a[1] - b[1];
  const blue = a[2] - b[2];
  return Math.sqrt((2 + redMean / 256) * red ** 2 + 4 * green ** 2 + (2 + (255 - redMean) / 256) * blue ** 2);
}

export function allocateBookColor(existingAssignments, bookId, palette = BOOK_COLOR_PALETTE) {
  const result = new Map(existingAssignments || []);
  if (!bookId || result.has(bookId) || palette.length === 0) return result;
  const assigned = [...result.values()];
  const available = palette.filter((color) => !assigned.includes(color));
  const candidates = available.length ? available : palette;
  const chosen = assigned.length === 0
    ? candidates[0]
    : candidates.reduce((best, candidate) => {
      const score = Math.min(...assigned.map((color) => distance(candidate, color)));
      const bestScore = Math.min(...assigned.map((color) => distance(best, color)));
      return score > bestScore ? candidate : best;
    }, candidates[0]);
  result.set(bookId, chosen);
  return result;
}
