const prefixMap = {
  ort: "location",
  location: "location",
  figur: "person",
  person: "person",
  personae: "person",
  ereignis: "event",
  event: "event",
  handlungsstrang: "thread",
  thread: "thread",
  motiv: "motif",
  begriff: "term",
  term: "term",
  erinnerung: "reminder",
  reminder: "reminder",
  recherche: "research_note",
  research: "research_note",
};

const referencePrefixes = {
  location: "Ort",
  event: "Ereignis",
  thread: "Handlungsstrang",
  motif: "Motiv",
  term: "Begriff",
  reminder: "Erinnerung",
  research_note: "Recherche",
};

export function extractWikiLinks(markdown) {
  if (!markdown) {
    return [];
  }

  const matches = markdown.matchAll(/\[\[([^[\]]+)\]\]/g);
  return Array.from(matches, (match) => parseWikiReference(match[1])).filter(Boolean);
}

export function parseWikiReference(rawValue) {
  const value = rawValue?.trim();
  if (!value) {
    return null;
  }

  const separatorIndex = value.indexOf(":");
  if (separatorIndex >= 0) {
    const prefix = value.slice(0, separatorIndex).trim().toLowerCase();
    const name = value.slice(separatorIndex + 1).trim();
    return {
      raw: value,
      type: prefixMap[prefix] || "custom",
      name,
      key: normalizeKnowledgeKey(prefixMap[prefix] || "custom", name),
    };
  }

  return {
    raw: value,
    type: "person",
    name: value,
    key: normalizeKnowledgeKey("person", value),
  };
}

export function normalizeKnowledgeKey(type, name) {
  return `${type || "custom"}::${String(name || "")
    .trim()
    .toLowerCase()}`;
}

export function knowledgeReference(item) {
  const prefix = referencePrefixes[item.type];
  if (!prefix || item.type === "person") {
    return `[[${item.name}]]`;
  }
  return `[[${prefix}:${item.name}]]`;
}

export function splitTagInput(value) {
  return String(value || "")
    .split(",")
    .map((entry) => entry.trim())
    .filter(Boolean);
}

export function formatTagInput(tags) {
  return Array.isArray(tags) ? tags.join(", ") : "";
}

export function knowledgeTypeLabel(type) {
  return (
    {
      person: "Person",
      location: "Ort",
      event: "Ereignis",
      thread: "Handlungsstrang",
      motif: "Motiv",
      term: "Begriff",
      reminder: "Erinnerung",
      research_note: "Recherche",
      custom: "Custom",
    }[type] || type
  );
}
