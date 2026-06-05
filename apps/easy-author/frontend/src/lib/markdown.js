function wrapText(text, marks = []) {
  if (!text) {
    return [];
  }
  return [
    {
      type: "text",
      text,
      ...(marks.length > 0 ? { marks } : {}),
    },
  ];
}

function paragraphNode(content) {
  return {
    type: "paragraph",
    content: Array.isArray(content) ? content : [],
  };
}

function hardBreakNode() {
  return {
    type: "hardBreak",
  };
}

function footnoteReferenceNode(noteId) {
  return {
    type: "footnoteReference",
    attrs: {
      noteId: String(noteId || ""),
    },
  };
}

function footnoteDefinitionNode(noteId, content) {
  return {
    type: "footnoteDefinition",
    attrs: {
      noteId: String(noteId || ""),
    },
    content: Array.isArray(content) && content.length > 0 ? content : [paragraphNode([])],
  };
}

function codeBlockNode(value) {
  return {
    type: "codeBlock",
    content: wrapText(value),
  };
}

function tableCellNode(content, isHeader = false) {
  return {
    type: isHeader ? "tableHeader" : "tableCell",
    content: [paragraphNode(Array.isArray(content) ? content : wrapText(String(content || "")))],
  };
}

function tableRowNode(cells, header = false) {
  return {
    type: "tableRow",
    content: cells.map((cell) => tableCellNode(parseInlineMarkdown(String(cell || "")), header)),
  };
}

function tableNode(headerCells, bodyRows) {
  return {
    type: "table",
    content: [
      tableRowNode(headerCells, true),
      ...bodyRows.map((cells) => tableRowNode(cells, false)),
    ],
  };
}

function escapeInlineText(text) {
  return String(text || "").replace(/([\\`*~])/g, "\\$1");
}

function escapeBlockStart(line) {
  if (!line) {
    return "";
  }
  if (
    /^#{1,6}\s/.test(line) ||
    /^>\s?/.test(line) ||
    /^[-*+]\s/.test(line) ||
    /^\d+\.\s/.test(line) ||
    /^```/.test(line) ||
    /^---$/.test(line)
  ) {
    return `\\${line}`;
  }
  return line;
}

function textWithMarks(node) {
  const content = node.text || "";
  const marks = node.marks || [];
  return marks.reduce((value, mark) => {
    switch (mark.type) {
      case "bold":
        return `**${escapeInlineText(value)}**`;
      case "italic":
        return `*${escapeInlineText(value)}*`;
      case "strike":
        return `~~${escapeInlineText(value)}~~`;
      case "code":
        return `\`${String(value).replace(/[`\\]/g, "\\$&")}\``;
      default:
        return value;
    }
  }, escapeInlineText(content));
}

function inlineLines(content = []) {
  const lines = [""];
  content.forEach((node) => {
    if (node.type === "text") {
      lines[lines.length - 1] += textWithMarks(node);
      return;
    }
    if (node.type === "footnoteReference") {
      lines[lines.length - 1] += `[^${node.attrs?.noteId || ""}]`;
      return;
    }
    if (node.type === "hardBreak") {
      lines.push("");
    }
  });
  return lines;
}

function paragraphToLines(node) {
  return inlineLines(node.content).map((line) => escapeBlockStart(line));
}

function indentLines(lines, indent) {
  return lines.map((line) => `${indent}${line}`);
}

function serializeListItem(item, depth, marker) {
  const indent = "  ".repeat(depth);
  const continuationIndent = `${indent}  `;
  const lines = [];
  const children = item.content || [];

  if (children.length === 0) {
    return [`${indent}${marker}`.trimEnd()];
  }

  children.forEach((child, index) => {
    const isFirst = index === 0;

    if (child.type === "paragraph") {
      const paragraphLines = paragraphToLines(child);
      if (isFirst) {
        const [firstLine = "", ...rest] = paragraphLines;
        lines.push(`${indent}${marker}${firstLine}`);
        rest.forEach((line) => lines.push(`${continuationIndent}${line}`));
        return;
      }
      lines.push("");
      paragraphLines.forEach((line) => lines.push(`${continuationIndent}${line}`));
      return;
    }

    if (isFirst) {
      lines.push(`${indent}${marker}`.trimEnd());
    }

    if (child.type === "bulletList" || child.type === "orderedList") {
      lines.push(...blockToLines(child, depth + 1));
      return;
    }

    if (!isFirst) {
      lines.push("");
    }
    lines.push(...indentLines(blockToLines(child, depth), continuationIndent));
  });

  return lines;
}

function blockquoteToLines(node, depth) {
  const children = node.content || [];
  if (children.length === 0) {
    return [">"];
  }

  const lines = [];
  children.forEach((child, index) => {
    const previous = children[index - 1];
    const needsSeparator =
      index > 0 &&
      !(
        previous?.type === "paragraph" &&
        (child.type === "bulletList" || child.type === "orderedList")
      );

    if (needsSeparator) {
      lines.push(">");
    }
    const childLines = blockToLines(child, depth);
    if (childLines.length === 0) {
      lines.push(">");
      return;
    }
    childLines.forEach((line) => {
      lines.push(line ? `> ${line}` : ">");
    });
  });
  return lines;
}

function cellContentToMarkdown(node) {
  const blocks = node.content || [];
  if (blocks.length === 0) {
    return "";
  }
  return blocks
    .map((child) => {
      if (child.type === "paragraph") {
        return inlineLines(child.content).join("<br>");
      }
      return blockToLines(child).join(" ");
    })
    .join("<br>");
}

function tableToLines(node) {
  const rows = node.content || [];
  if (rows.length === 0) {
    return [];
  }
  const headerRow = rows[0];
  const columnCount = headerRow.content?.length || 0;
  if (columnCount === 0) {
    return [];
  }
  const headerCells = headerRow.content.map((cell) => cellContentToMarkdown(cell));
  const separator = Array.from({ length: columnCount }, () => "---");
  const bodyLines = rows.slice(1).map((row) => row.content.map((cell) => cellContentToMarkdown(cell)));
  return [
    `| ${headerCells.join(" | ")} |`,
    `| ${separator.join(" | ")} |`,
    ...bodyLines.map((cells) => `| ${cells.join(" | ")} |`),
  ];
}

function footnoteDefinitionToLines(node) {
  const noteId = node.attrs?.noteId || "";
  const children = node.content || [];
  if (children.length === 0) {
    return [`[^${noteId}]: `];
  }

  const blockLines = children.flatMap((child, index) => {
    const lines = blockToLines(child);
    if (index === 0) {
      return lines;
    }
    return ["", ...lines];
  });

  const [firstLine = "", ...rest] = blockLines;
  return [`[^${noteId}]: ${firstLine}`, ...rest.map((line) => (line ? `  ${line}` : ""))];
}

function blockToLines(node, depth = 0) {
  switch (node.type) {
    case "heading":
      return [`${"#".repeat(node.attrs?.level || 1)} ${inlineLines(node.content).join("  \n")}`];
    case "paragraph":
      return paragraphToLines(node);
    case "bulletList":
      return (node.content || []).flatMap((item) => serializeListItem(item, depth, "- "));
    case "orderedList": {
      const start = Number(node.attrs?.start || 1);
      return (node.content || []).flatMap((item, index) => serializeListItem(item, depth, `${start + index}. `));
    }
    case "blockquote":
      return blockquoteToLines(node, depth);
    case "codeBlock": {
      const content = node.content?.map((child) => child.text || "").join("") || "";
      return ["```", ...content.split("\n"), "```"];
    }
    case "horizontalRule":
      return ["---"];
    case "table":
      return tableToLines(node);
    case "footnoteDefinition":
      return footnoteDefinitionToLines(node);
    default:
      return [];
  }
}

export function docToMarkdown(doc) {
  if (!doc || !Array.isArray(doc.content)) {
    return "";
  }

  const blocks = doc.content
    .map((node) => blockToLines(node))
    .filter((lines) => lines.length > 0)
    .map((lines) => lines.join("\n"));

  return blocks.join("\n\n").trim();
}

function isEscaped(text, index) {
  let slashCount = 0;
  for (let cursor = index - 1; cursor >= 0 && text[cursor] === "\\"; cursor -= 1) {
    slashCount += 1;
  }
  return slashCount % 2 === 1;
}

function findClosingToken(text, token, fromIndex) {
  let cursor = fromIndex;
  while (cursor <= text.length - token.length) {
    if (text.startsWith(token, cursor) && !isEscaped(text, cursor)) {
      return cursor;
    }
    cursor += 1;
  }
  return -1;
}

function flushPlainText(buffer, nodes) {
  if (!buffer.value) {
    return;
  }
  nodes.push(...wrapText(buffer.value));
  buffer.value = "";
}

function parseInlineMarkdown(text) {
  if (!text) {
    return [];
  }

  const nodes = [];
  const buffer = { value: "" };
  let cursor = 0;

  while (cursor < text.length) {
    const char = text[cursor];

    if (char === "\\" && cursor + 1 < text.length) {
      buffer.value += text[cursor + 1];
      cursor += 2;
      continue;
    }

    if (text.startsWith("[[", cursor)) {
      const closing = findClosingToken(text, "]]", cursor + 2);
      if (closing !== -1) {
        flushPlainText(buffer, nodes);
        nodes.push(...wrapText(text.slice(cursor, closing + 2)));
        cursor = closing + 2;
        continue;
      }
    }

    if (text.startsWith("[^", cursor)) {
      const closing = findClosingToken(text, "]", cursor + 2);
      if (closing !== -1) {
        const noteId = text.slice(cursor + 2, closing).trim();
        if (noteId) {
          flushPlainText(buffer, nodes);
          nodes.push(footnoteReferenceNode(noteId));
          cursor = closing + 1;
          continue;
        }
      }
    }

    if (text.startsWith("**", cursor)) {
      const closing = findClosingToken(text, "**", cursor + 2);
      if (closing !== -1) {
        flushPlainText(buffer, nodes);
        nodes.push(...wrapText(text.slice(cursor + 2, closing), [{ type: "bold" }]));
        cursor = closing + 2;
        continue;
      }
    }

    if (text.startsWith("~~", cursor)) {
      const closing = findClosingToken(text, "~~", cursor + 2);
      if (closing !== -1) {
        flushPlainText(buffer, nodes);
        nodes.push(...wrapText(text.slice(cursor + 2, closing), [{ type: "strike" }]));
        cursor = closing + 2;
        continue;
      }
    }

    if (text[cursor] === "`") {
      const closing = findClosingToken(text, "`", cursor + 1);
      if (closing !== -1) {
        flushPlainText(buffer, nodes);
        nodes.push(...wrapText(text.slice(cursor + 1, closing), [{ type: "code" }]));
        cursor = closing + 1;
        continue;
      }
    }

    if (text[cursor] === "*") {
      const closing = findClosingToken(text, "*", cursor + 1);
      if (closing !== -1) {
        flushPlainText(buffer, nodes);
        nodes.push(...wrapText(text.slice(cursor + 1, closing), [{ type: "italic" }]));
        cursor = closing + 1;
        continue;
      }
    }

    buffer.value += char;
    cursor += 1;
  }

  flushPlainText(buffer, nodes);
  return nodes;
}

function parseParagraphLines(lines) {
  const content = [];
  lines.forEach((line, index) => {
    if (index > 0) {
      content.push(hardBreakNode());
    }
    content.push(...parseInlineMarkdown(line));
  });
  return paragraphNode(content);
}

function countIndent(line) {
  const match = line.match(/^ */);
  return match ? match[0].length : 0;
}

function isBlank(line) {
  return !line || !line.trim();
}

function parseBulletMarker(line) {
  return line.match(/^(\s*)([-*])\s+(.*)$/);
}

function parseOrderedMarker(line) {
  return line.match(/^(\s*)(\d+)\.\s+(.*)$/);
}

function isBulletListLine(line) {
  return Boolean(parseBulletMarker(line));
}

function isOrderedListLine(line) {
  return Boolean(parseOrderedMarker(line));
}

function isListLine(line) {
  return isBulletListLine(line) || isOrderedListLine(line);
}

function isQuoteLine(line) {
  const trimmedStart = line.trimStart();
  return trimmedStart === ">" || trimmedStart.startsWith("> ");
}

function isBlockBoundary(line) {
  const trimmed = line.trim();
  if (!trimmed) {
    return false;
  }
  return (
    /^(#{1,6})\s+/.test(trimmed) ||
    /^\[\^[^\]]+\]:/.test(trimmed) ||
    isListLine(line) ||
    isQuoteLine(line) ||
    trimmed.startsWith("```") ||
    trimmed === "---"
  );
}

function stripQuoteMarker(line) {
  const trimmedStart = line.trimStart();
  if (trimmedStart === ">") {
    return "";
  }
  if (trimmedStart.startsWith("> ")) {
    return trimmedStart.slice(2);
  }
  return trimmedStart;
}

function splitPipeTableRow(line) {
  const trimmed = line.trim();
  if (!trimmed.includes("|")) {
    return null;
  }
  const normalized = trimmed.replace(/^\|/, "").replace(/\|$/, "");
  const cells = normalized.split("|").map((cell) => cell.trim());
  return cells.length >= 2 ? cells : null;
}

function isTableSeparatorLine(line, expectedColumns) {
  const cells = splitPipeTableRow(line);
  if (!cells || cells.length !== expectedColumns) {
    return false;
  }
  return cells.every((cell) => /^:?-{3,}:?$/.test(cell));
}

function isPipeTableRow(line, expectedColumns) {
  const cells = splitPipeTableRow(line);
  if (!cells || cells.length < 2) {
    return false;
  }
  return expectedColumns ? cells.length === expectedColumns : true;
}

function parsePipeTable(lines, startIndex) {
  const headerCells = splitPipeTableRow(lines[startIndex]);
  const separatorLine = lines[startIndex + 1];
  if (!headerCells || !separatorLine || !isTableSeparatorLine(separatorLine, headerCells.length)) {
    return null;
  }

  const bodyRows = [];
  let index = startIndex + 2;
  while (index < lines.length && isPipeTableRow(lines[index], headerCells.length)) {
    const rowCells = splitPipeTableRow(lines[index]);
    if (!rowCells) {
      break;
    }
    bodyRows.push(rowCells);
    index += 1;
  }

  if (bodyRows.length === 0) {
    bodyRows.push(Array.from({ length: headerCells.length }, () => ""));
  }

  return {
    node: tableNode(headerCells, bodyRows),
    nextIndex: index,
  };
}

function parseFootnoteDefinition(lines, startIndex) {
  const match = lines[startIndex].trim().match(/^\[\^([^\]]+)\]:\s*(.*)$/);
  if (!match) {
    return null;
  }

  const noteId = match[1];
  const bodyLines = [match[2] || ""];
  let index = startIndex + 1;

  while (index < lines.length) {
    const nextLine = lines[index];
    if (!nextLine.trim()) {
      bodyLines.push("");
      index += 1;
      continue;
    }
    if (/^\s{2,}/.test(nextLine)) {
      bodyLines.push(nextLine.replace(/^\s{2}/, ""));
      index += 1;
      continue;
    }
    break;
  }

  while (bodyLines.length > 0 && !bodyLines[bodyLines.length - 1].trim()) {
    bodyLines.pop();
  }

  return {
    node: footnoteDefinitionNode(noteId, parseBlocks(bodyLines)),
    nextIndex: index,
  };
}

function extractPlainParagraphLine(node) {
  if (!node || node.type !== "paragraph") {
    return null;
  }
  let value = "";
  for (const child of node.content || []) {
    if (child.type === "text") {
      value += child.text || "";
      continue;
    }
    if (child.type === "hardBreak") {
      return null;
    }
    return null;
  }
  return value;
}

export function normalizeRichTableMarkdown(doc) {
  if (!doc || doc.type !== "doc" || !Array.isArray(doc.content)) {
    return { doc, changed: false };
  }

  const nextContent = [];
  let changed = false;

  for (let index = 0; index < doc.content.length; index += 1) {
    const current = doc.content[index];
    const headerLine = extractPlainParagraphLine(current);
    const separatorLine = extractPlainParagraphLine(doc.content[index + 1]);

    if (
      headerLine &&
      separatorLine &&
      isPipeTableRow(headerLine) &&
      isTableSeparatorLine(separatorLine, splitPipeTableRow(headerLine)?.length || 0)
    ) {
      const bodyLines = [];
      let cursor = index + 2;
      const expectedColumns = splitPipeTableRow(headerLine)?.length || 0;
      while (cursor < doc.content.length) {
        const candidate = extractPlainParagraphLine(doc.content[cursor]);
        if (!candidate || !isPipeTableRow(candidate, expectedColumns)) {
          break;
        }
        bodyLines.push(candidate);
        cursor += 1;
      }

      if (bodyLines.length > 0) {
        const normalized = parsePipeTable([headerLine, separatorLine, ...bodyLines], 0);
        if (normalized) {
          nextContent.push(normalized.node);
          index = cursor - 1;
          changed = true;
          continue;
        }
      }
    }

    nextContent.push(current);
  }

  return changed
    ? {
        doc: {
          ...doc,
          content: nextContent,
        },
        changed: true,
      }
    : { doc, changed: false };
}

function parseBlocks(lines) {
  const content = [];

  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index];
    const trimmed = line.trim();

    if (!trimmed) {
      continue;
    }

    const headingMatch = trimmed.match(/^(#{1,6})\s+(.*)$/);
    if (headingMatch) {
      content.push({
        type: "heading",
        attrs: { level: headingMatch[1].length },
        content: parseInlineMarkdown(headingMatch[2]),
      });
      continue;
    }

    if (trimmed === "---") {
      content.push({ type: "horizontalRule" });
      continue;
    }

    const footnote = parseFootnoteDefinition(lines, index);
    if (footnote) {
      content.push(footnote.node);
      index = footnote.nextIndex - 1;
      continue;
    }

    const table = parsePipeTable(lines, index);
    if (table) {
      content.push(table.node);
      index = table.nextIndex - 1;
      continue;
    }

    if (trimmed.startsWith("```")) {
      const codeLines = [];
      index += 1;
      while (index < lines.length && !lines[index].trim().startsWith("```")) {
        codeLines.push(lines[index]);
        index += 1;
      }
      content.push(codeBlockNode(codeLines.join("\n")));
      continue;
    }

    if (isQuoteLine(line)) {
      const quoteLines = [stripQuoteMarker(line)];
      while (index + 1 < lines.length && isQuoteLine(lines[index + 1])) {
        index += 1;
        quoteLines.push(stripQuoteMarker(lines[index]));
      }
      const quoteContent = parseBlocks(quoteLines);
      content.push({
        type: "blockquote",
        content: quoteContent.length > 0 ? quoteContent : [paragraphNode([])],
      });
      continue;
    }

    if (isListLine(line)) {
      const list = parseList(lines, index);
      content.push(list.node);
      index = list.nextIndex - 1;
      continue;
    }

    const paragraphLines = [line];
    while (index + 1 < lines.length) {
      const nextRaw = lines[index + 1];
      if (isBlank(nextRaw) || isBlockBoundary(nextRaw)) {
        break;
      }
      index += 1;
      paragraphLines.push(nextRaw);
    }
    content.push(parseParagraphLines(paragraphLines));
  }

  return content;
}

function parseList(lines, startIndex) {
  const firstLine = lines[startIndex];
  const bulletMarker = parseBulletMarker(firstLine);
  const orderedMarker = parseOrderedMarker(firstLine);
  const ordered = Boolean(orderedMarker);
  const baseIndent = countIndent(firstLine);
  const marker = ordered ? orderedMarker : bulletMarker;
  const start = ordered ? Number(marker?.[2] || 1) : 1;
  const items = [];
  let index = startIndex;

  while (index < lines.length) {
    const currentLine = lines[index];
    const currentMarker = ordered ? parseOrderedMarker(currentLine) : parseBulletMarker(currentLine);

    if (!currentMarker || countIndent(currentLine) !== baseIndent) {
      break;
    }

    const itemLines = [];
    itemLines.push(currentMarker[3] || "");
    index += 1;

    while (index < lines.length) {
      const nextLine = lines[index];

      if (isBlank(nextLine)) {
        const lookahead = lines[index + 1];
        if (!lookahead) {
          index += 1;
          break;
        }
        if (
          (ordered ? parseOrderedMarker(lookahead) : parseBulletMarker(lookahead)) &&
          countIndent(lookahead) === baseIndent
        ) {
          break;
        }
        if (countIndent(lookahead) < baseIndent + 2 && !isBlank(lookahead)) {
          break;
        }
        itemLines.push("");
        index += 1;
        continue;
      }

      if (
        (ordered ? parseOrderedMarker(nextLine) : parseBulletMarker(nextLine)) &&
        countIndent(nextLine) === baseIndent
      ) {
        break;
      }

      if (countIndent(nextLine) < baseIndent + 2) {
        break;
      }

      itemLines.push(nextLine.slice(baseIndent + 2));
      index += 1;
    }

    while (itemLines.length > 0 && itemLines[0] === "") {
      itemLines.shift();
    }
    while (itemLines.length > 0 && itemLines[itemLines.length - 1] === "") {
      itemLines.pop();
    }

    const itemContent = parseBlocks(itemLines);
    items.push({
      type: "listItem",
      content: itemContent.length > 0 ? itemContent : [paragraphNode([])],
    });
  }

  return {
    node: {
      type: ordered ? "orderedList" : "bulletList",
      ...(ordered ? { attrs: { start } } : {}),
      content: items,
    },
    nextIndex: index,
  };
}

export function markdownToDoc(markdown) {
  if (!markdown || !markdown.trim()) {
    return {
      type: "doc",
      content: [paragraphNode([])],
    };
  }

  const lines = markdown.replace(/\r\n/g, "\n").split("\n");
  const content = parseBlocks(lines);

  return {
    type: "doc",
    content: content.length > 0 ? content : [paragraphNode([])],
  };
}

export function previewText(value, limit = 96) {
  if (!value) {
    return "";
  }
  return value.length > limit ? `${value.slice(0, limit - 3)}...` : value;
}
