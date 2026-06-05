import test from "node:test";
import assert from "node:assert/strict";
import { docToMarkdown, markdownToDoc, normalizeRichTableMarkdown } from "./markdown.js";

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

test("roundtrips simple pipe tables", () => {
  const input = [
    "| Kopf-Spalte 1 | Kopf-Spalte 2 | Kopf-Spalte 3 |",
    "| --- | --- | --- |",
    "| Textblock 1 | Textblock 2 | Textblock 3 |",
    "| Mehr Text 1 | Mehr Text 2 | Mehr Text 3 |",
  ].join("\n");

  const output = docToMarkdown(markdownToDoc(input));
  assert.equal(output, input);
});

test("normalizes typed pipe table paragraphs into a table node", () => {
  const input = {
    type: "doc",
    content: [
      { type: "paragraph", content: [{ type: "text", text: "| Kopf 1 | Kopf 2 |" }] },
      { type: "paragraph", content: [{ type: "text", text: "| --- | --- |" }] },
      { type: "paragraph", content: [{ type: "text", text: "| A | B |" }] },
    ],
  };

  const output = normalizeRichTableMarkdown(input);

  assert.equal(output.changed, true);
  assert.equal(output.doc.content[0].type, "table");
  assert.equal(docToMarkdown(output.doc), ["| Kopf 1 | Kopf 2 |", "| --- | --- |", "| A | B |"].join("\n"));
});

test("roundtrips markdown footnotes", () => {
  const input = [
    "Ein Gedanke mit Fussnote[^1].",
    "",
    "[^1]: Das ist die erste Fussnote.",
  ].join("\n");

  const output = docToMarkdown(markdownToDoc(input));
  assert.equal(output, input);
});
