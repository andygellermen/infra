function textWithMarks(node) {
  const content = node.text || "";
  const marks = node.marks || [];
  return marks.reduce((value, mark) => {
    switch (mark.type) {
      case "bold":
        return `**${value}**`;
      case "italic":
        return `*${value}*`;
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

function paragraphNode(text) {
  return {
    type: "paragraph",
    content: text
      ? [
          {
            type: "text",
            text,
          },
        ]
      : [],
  };
}

function parseInlineText(text) {
  return text ? [{ type: "text", text }] : [];
}

export function markdownToDoc(markdown) {
  if (!markdown || !markdown.trim()) {
    return {
      type: "doc",
      content: [paragraphNode("")],
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
        content: parseInlineText(headingMatch[2]),
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
        content: parseInlineText(codeLines.join("\n")),
      });
      continue;
    }

    if (trimmed.startsWith("> ")) {
      content.push({
        type: "blockquote",
        content: [paragraphNode(trimmed.slice(2))],
      });
      continue;
    }

    if (/^[-*]\s+/.test(trimmed)) {
      const items = [{ type: "listItem", content: [paragraphNode(trimmed.replace(/^[-*]\s+/, ""))] }];
      while (index + 1 < lines.length && /^[-*]\s+/.test(lines[index + 1].trim())) {
        index += 1;
        items.push({
          type: "listItem",
          content: [paragraphNode(lines[index].trim().replace(/^[-*]\s+/, ""))],
        });
      }
      content.push({ type: "bulletList", content: items });
      continue;
    }

    if (/^\d+\.\s+/.test(trimmed)) {
      const items = [{ type: "listItem", content: [paragraphNode(trimmed.replace(/^\d+\.\s+/, ""))] }];
      while (index + 1 < lines.length && /^\d+\.\s+/.test(lines[index + 1].trim())) {
        index += 1;
        items.push({
          type: "listItem",
          content: [paragraphNode(lines[index].trim().replace(/^\d+\.\s+/, ""))],
        });
      }
      content.push({ type: "orderedList", content: items });
      continue;
    }

    const paragraphLines = [trimmed];
    while (index + 1 < lines.length && lines[index + 1].trim()) {
      const next = lines[index + 1].trim();
      if (/^(#{1,6})\s+/.test(next) || /^[-*]\s+/.test(next) || /^\d+\.\s+/.test(next) || next.startsWith("> ") || next.startsWith("```") || next === "---") {
        break;
      }
      index += 1;
      paragraphLines.push(next);
    }
    content.push(paragraphNode(paragraphLines.join(" ")));
  }

  return {
    type: "doc",
    content: content.length > 0 ? content : [paragraphNode("")],
  };
}

export function previewText(value, limit = 96) {
  if (!value) {
    return "";
  }
  return value.length > limit ? `${value.slice(0, limit - 3)}...` : value;
}
