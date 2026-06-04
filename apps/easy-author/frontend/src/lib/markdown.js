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

function textWithMarks(node) {
  const content = node.text || "";
  const marks = node.marks || [];
  return marks.reduce((value, mark) => {
    switch (mark.type) {
      case "bold":
        return `**${value}**`;
      case "italic":
        return `*${value}*`;
      case "strike":
        return `~~${value}~~`;
      case "code":
        return `\`${value}\``;
      default:
        return value;
    }
  }, content);
}

function inlineContent(content = []) {
  return content
    .map((node) => {
      if (node.type === "text") {
        return textWithMarks(node);
      }
      if (node.type === "hardBreak") {
        return "  \n";
      }
      return "";
    })
    .join("");
}

function blockToMarkdown(node, depth = 0) {
  switch (node.type) {
    case "heading":
      return `${"#".repeat(node.attrs?.level || 1)} ${inlineContent(node.content)}`;
    case "paragraph":
      return inlineContent(node.content);
    case "bulletList":
      return (node.content || [])
        .map((item) => `${"  ".repeat(depth)}- ${blockToMarkdown(item, depth + 1).trimStart()}`)
        .join("\n");
    case "orderedList":
      return (node.content || [])
        .map((item, index) => `${"  ".repeat(depth)}${index + 1}. ${blockToMarkdown(item, depth + 1).trimStart()}`)
        .join("\n");
    case "listItem":
      return (node.content || []).map((child) => blockToMarkdown(child, depth)).join("\n");
    case "blockquote":
      return (node.content || [])
        .map((child) =>
          blockToMarkdown(child, depth)
            .split("\n")
            .map((line) => `> ${line}`)
            .join("\n"),
        )
        .join("\n");
    case "codeBlock":
      return `\`\`\`\n${inlineContent(node.content)}\n\`\`\``;
    case "horizontalRule":
      return "---";
    default:
      return "";
  }
}

export function docToMarkdown(doc) {
  if (!doc || !Array.isArray(doc.content)) {
    return "";
  }
  return doc.content
    .map((node) => blockToMarkdown(node))
    .filter(Boolean)
    .join("\n\n")
    .trim();
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

function findNextToken(text, fromIndex) {
  const patterns = [
    /\[\[[^[\]]+\]\]/g,
    /\*\*[^*]+\*\*/g,
    /\*[^*\n]+\*/g,
    /~~[^~]+~~/g,
    /`[^`\n]+`/g,
  ];

  let nextMatch = null;
  for (const pattern of patterns) {
    pattern.lastIndex = fromIndex;
    const match = pattern.exec(text);
    if (!match) {
      continue;
    }
    if (!nextMatch || match.index < nextMatch.index) {
      nextMatch = { index: match.index, value: match[0] };
    }
  }
  return nextMatch;
}

function parseInlineMarkdown(text) {
  if (!text) {
    return [];
  }

  const nodes = [];
  let cursor = 0;

  while (cursor < text.length) {
    const token = findNextToken(text, cursor);
    if (!token) {
      nodes.push(...wrapText(text.slice(cursor)));
      break;
    }

    if (token.index > cursor) {
      nodes.push(...wrapText(text.slice(cursor, token.index)));
    }

    const value = token.value;
    if (value.startsWith("[[")) {
      nodes.push(...wrapText(value));
    } else if (value.startsWith("**")) {
      nodes.push(...wrapText(value.slice(2, -2), [{ type: "bold" }]));
    } else if (value.startsWith("~~")) {
      nodes.push(...wrapText(value.slice(2, -2), [{ type: "strike" }]));
    } else if (value.startsWith("`")) {
      nodes.push(...wrapText(value.slice(1, -1), [{ type: "code" }]));
    } else if (value.startsWith("*")) {
      nodes.push(...wrapText(value.slice(1, -1), [{ type: "italic" }]));
    } else {
      nodes.push(...wrapText(value));
    }

    cursor = token.index + value.length;
  }

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

function isBlockBoundary(value) {
  return (
    /^(#{1,6})\s+/.test(value) ||
    /^[-*]\s+/.test(value) ||
    /^\d+\.\s+/.test(value) ||
    value.startsWith("> ") ||
    value.startsWith("```") ||
    value === "---"
  );
}

export function markdownToDoc(markdown) {
  if (!markdown || !markdown.trim()) {
    return {
      type: "doc",
      content: [paragraphNode([])],
    };
  }

  const lines = markdown.replace(/\r\n/g, "\n").split("\n");
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

    if (trimmed.startsWith("```")) {
      const codeLines = [];
      index += 1;
      while (index < lines.length && !lines[index].trim().startsWith("```")) {
        codeLines.push(lines[index]);
        index += 1;
      }
      content.push({
        type: "codeBlock",
        content: wrapText(codeLines.join("\n")),
      });
      continue;
    }

    if (trimmed.startsWith("> ")) {
      const quoteLines = [trimmed.slice(2)];
      while (index + 1 < lines.length && lines[index + 1].trim().startsWith("> ")) {
        index += 1;
        quoteLines.push(lines[index].trim().slice(2));
      }
      content.push({
        type: "blockquote",
        content: [parseParagraphLines(quoteLines)],
      });
      continue;
    }

    if (/^[-*]\s+/.test(trimmed)) {
      const items = [
        { type: "listItem", content: [parseParagraphLines([trimmed.replace(/^[-*]\s+/, "")])] },
      ];
      while (index + 1 < lines.length && /^[-*]\s+/.test(lines[index + 1].trim())) {
        index += 1;
        items.push({
          type: "listItem",
          content: [parseParagraphLines([lines[index].trim().replace(/^[-*]\s+/, "")])],
        });
      }
      content.push({ type: "bulletList", content: items });
      continue;
    }

    if (/^\d+\.\s+/.test(trimmed)) {
      const items = [
        { type: "listItem", content: [parseParagraphLines([trimmed.replace(/^\d+\.\s+/, "")])] },
      ];
      while (index + 1 < lines.length && /^\d+\.\s+/.test(lines[index + 1].trim())) {
        index += 1;
        items.push({
          type: "listItem",
          content: [parseParagraphLines([lines[index].trim().replace(/^\d+\.\s+/, "")])],
        });
      }
      content.push({ type: "orderedList", content: items });
      continue;
    }

    const paragraphLines = [line];
    while (index + 1 < lines.length) {
      const nextRaw = lines[index + 1];
      const nextTrimmed = nextRaw.trim();
      if (!nextTrimmed || isBlockBoundary(nextTrimmed)) {
        break;
      }
      index += 1;
      paragraphLines.push(nextRaw);
    }
    content.push(parseParagraphLines(paragraphLines));
  }

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
