import test from "node:test";
import assert from "node:assert/strict";
import { docToMarkdown, markdownToDoc } from "./markdown.js";

test("roundtrips nested bullet and ordered lists", () => {
  const input = [
    "- Alpha",
    "  - Beta",
    "    1. Eins",
    "    2. Zwei",
    "  - Gamma",
    "- Delta",
  ].join("\n");

  const output = docToMarkdown(markdownToDoc(input));
  assert.equal(output, input);
});

test("preserves escaped markdown control characters as plain text", () => {
  const input = [
    "\\# Keine Ueberschrift",
    "\\- Kein Listeneintrag",
    "\\1. Keine nummerierte Liste",
    "\\> Kein Zitat",
    "\\*Sterne\\* und \\~Wellen\\~ und \\`Code\\`",
  ].join("\n");

  const output = docToMarkdown(markdownToDoc(input));
  assert.equal(output, input);
});

test("keeps wiki links and inline marks intact", () => {
  const input = "Eintrag [[Person:Mara]] mit **Fokus**, *Nuance*, ~~Alt~~ und `Code`.";
  const output = docToMarkdown(markdownToDoc(input));
  assert.equal(output, input);
});

test("roundtrips multi paragraph list items", () => {
  const input = [
    "- Alpha",
    "",
    "  Beta weiter",
    "",
    "  Noch ein Absatz",
    "- Gamma",
  ].join("\n");

  const output = docToMarkdown(markdownToDoc(input));
  assert.equal(output, input);
});

test("roundtrips blockquotes with multiple paragraphs", () => {
  const input = [
    "> Alpha",
    ">",
    "> Beta",
  ].join("\n");

  const output = docToMarkdown(markdownToDoc(input));
  assert.equal(output, input);
});

test("roundtrips blockquotes with nested lists", () => {
  const input = [
    "> Alpha",
    "> - Beta",
    "> - Gamma",
  ].join("\n");

  const output = docToMarkdown(markdownToDoc(input));
  assert.equal(output, input);
});

test("preserves ordered list start numbers", () => {
  const input = [
    "9. Neun",
    "10. Zehn",
  ].join("\n");

  const output = docToMarkdown(markdownToDoc(input));
  assert.equal(output, input);
});
