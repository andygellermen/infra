import { useEffect, useMemo, useRef, useState } from "react";
import EditorPane from "./components/EditorPane";
import SidebarSection from "./components/SidebarSection";
import { api } from "./lib/api";
import { markdownToDoc, previewText } from "./lib/markdown";
import {
  extractWikiLinks,
  formatTagInput,
  knowledgeReference,
  knowledgeTypeLabel,
  normalizeKnowledgeKey,
  splitTagInput,
} from "./lib/knowledge";

const EMPTY_DRAFT = {
  title: "",
  summary: "",
  markdown_content: "",
  editor_json: "",
};

const DEFAULT_EDITOR_APPEARANCE = {
  fontFamily: "serif",
  googleFontName: "Cormorant Garamond",
  fontSize: 18,
  lineHeight: 1.8,
  contentWidth: 860,
  fullscreenContentWidth: 1040,
  fullscreenBackdrop: "linen",
  surfacePreset: "warm",
  caretColor: "#76c7ff",
};

const EDITOR_APPEARANCE_STORAGE_KEY = "easy-author.editor-appearance.v1";
const POPUP_HOLD_DELAY_MS = 2000;
const POPUP_FADE_IN_DELAY_MS = 24;
const ALLOWED_FULLSCREEN_BACKDROPS = new Set(["linen", "paper", "dusk", "night"]);
const ALLOWED_SURFACE_PRESETS = new Set(["warm", "paper", "night"]);
const ALLOWED_FONT_FAMILIES = new Set(["serif", "sans", "mono", "google"]);
const GOOGLE_FONT_PRESETS = [
  "Cormorant Garamond",
  "Crimson Pro",
  "EB Garamond",
  "Libre Baskerville",
  "Lora",
  "Merriweather",
  "Newsreader",
  "Playfair Display",
  "Source Serif 4",
  "Spectral",
  "Inter",
  "Work Sans",
];
const FULLSCREEN_BACKDROP_THEME = {
  linen: {
    glow: "rgba(228, 183, 137, 0.12)",
    start: "#f7f0e5",
    end: "#efe4d4",
    cardStart: "rgba(255, 252, 247, 0.78)",
    cardEnd: "rgba(250, 243, 231, 0.72)",
    cardBorder: "rgba(110, 82, 54, 0.12)",
    cardShadow: "0 26px 60px rgba(72, 42, 18, 0.12)",
  },
  paper: {
    glow: "rgba(255, 255, 255, 0.36)",
    start: "#fbfaf6",
    end: "#f0ebe1",
    cardStart: "rgba(255, 255, 255, 0.82)",
    cardEnd: "rgba(251, 249, 244, 0.72)",
    cardBorder: "rgba(129, 118, 100, 0.12)",
    cardShadow: "0 26px 60px rgba(70, 60, 44, 0.1)",
  },
  dusk: {
    glow: "rgba(164, 118, 88, 0.18)",
    start: "#ede1d9",
    end: "#d9c9c1",
    cardStart: "rgba(250, 243, 239, 0.72)",
    cardEnd: "rgba(239, 229, 223, 0.62)",
    cardBorder: "rgba(122, 84, 60, 0.12)",
    cardShadow: "0 28px 64px rgba(78, 48, 32, 0.14)",
  },
  night: {
    glow: "rgba(110, 90, 82, 0.18)",
    start: "#201b18",
    end: "#141110",
    cardStart: "rgba(33, 28, 25, 0.74)",
    cardEnd: "rgba(24, 20, 18, 0.66)",
    cardBorder: "rgba(243, 237, 229, 0.12)",
    cardShadow: "0 28px 68px rgba(0, 0, 0, 0.34)",
  },
};

const WORKFLOW_TYPE_META = {
  notes: {
    label: "Notizen",
    hint: "Sammelt lose Gedanken, Formulierungen und Zwischenideen direkt am Kapitel.",
  },
  persons: {
    label: "Figuren",
    hint: "Bindet Passagen an Figurenentwicklung, Eigenschaften und offene Charakterfragen.",
  },
  events: {
    label: "Ereignisse",
    hint: "Markiert Schluesselmomente, Wendepunkte und Folgen im Kapitelverlauf.",
  },
  threads: {
    label: "Handlungsfaeden",
    hint: "Verknuepft Passagen mit roten Faeden, Konflikten und offenen Spannungen.",
  },
  reminders: {
    label: "Erinnerungen",
    hint: "Haelt spaetere To-dos, Rueckfragen und Ueberarbeitungsmarken fest.",
  },
  research: {
    label: "Recherche",
    hint: "Sammelt Stellen, die Faktencheck, Quellen oder Vertiefung benoetigen.",
  },
  clipboard: {
    label: "Clipboard",
    hint: "Fokussiert wiederverwendbare Textbausteine und schnelle Rueckgriffe im Schreibfluss.",
  },
  custom: {
    label: "Eigene Box",
    hint: "Freier Arbeitsraum fuer deinen individuellen Schreib- oder Review-Prozess.",
  },
};

const REVIEW_COMMENT_TYPE_META = {
  comment: {
    label: "Kommentar",
    hint: "Freier Hinweis oder Lektoratskommentar.",
    actionLabel: "Kommentar",
  },
  todo: {
    label: "To-do",
    hint: "Offener Punkt mit spaeterer Erledigung.",
    actionLabel: "To-do",
  },
  suggestion: {
    label: "Korrekturvorschlag",
    hint: "Konkreter Ersatztext zum Uebernehmen oder Anpassen.",
    actionLabel: "Vorschlag",
  },
  delete_request: {
    label: "Loeschbitte",
    hint: "Diese Passage sollte entfernt werden.",
    actionLabel: "Loeschung",
  },
  warning: {
    label: "Hinweis",
    hint: "Auffaelligkeit oder Warnung im Text.",
    actionLabel: "Hinweis",
  },
};

function workflowTypeMeta(type) {
  return WORKFLOW_TYPE_META[type] || WORKFLOW_TYPE_META.custom;
}

function reviewCommentTypeMeta(type) {
  return REVIEW_COMMENT_TYPE_META[type] || REVIEW_COMMENT_TYPE_META.comment;
}

const STORY_TIME_KEYWORDS = [
  "zeit",
  "zeitpunkt",
  "datum",
  "jahr",
  "jahre",
  "uhr",
  "morgen",
  "abend",
  "nacht",
  "fruehling",
  "sommer",
  "herbst",
  "winter",
  "spaeter",
  "damals",
  "heute",
  "gestern",
  "morgen",
];

const STORY_DATE_REGEX = /\b(?:\d{1,2}[./-]\d{1,2}[./-]\d{2,4}|\d{4})\b/g;
const STORY_TIME_RANGE_REGEX =
  /\b(?:\d{1,2}[./-]\d{1,2}[./-]\d{2,4}\s*(?:bis|-)\s*\d{1,2}[./-]\d{1,2}[./-]\d{2,4}|\d{4}\s*(?:bis|-)\s*\d{4})\b/g;
const WORKFLOW_TYPE_HINTS = {
  notes: ["idee", "gedanke", "notiz", "motiv", "bild", "frage", "ton", "szene"],
  persons: ["figur", "person", "charakter", "protagonist", "antagonist", "beziehung", "stimme"],
  events: ["ereignis", "wendepunkt", "szene", "ankunft", "abschied", "begegnung", "unfall", "entscheidung"],
  threads: ["konflikt", "spur", "ziel", "motiv", "geheimnis", "handlung", "subplot", "folge"],
  reminders: ["todo", "offen", "spaeter", "später", "pruefen", "prüfen", "merken", "ueberarbeiten", "überarbeiten"],
  research: ["quelle", "fakt", "datum", "jahr", "histor", "ort", "beleg", "referenz", "wiki"],
  clipboard: ["snippet", "zitat", "dialog", "formulierung", "textbaustein", "wortlaut"],
};
const REMINDER_REGEX = /\b(?:todo|fixme|offen|pruefen|prüfen|nachtragen|spaeter|später|ueberarbeiten|überarbeiten|merken)\b/gi;
const QUOTE_REGEX = /["“”„‚'`].+?["“”„‚'`]/g;
const HASHTAG_REGEX = /#([\p{L}\p{N}_-]+)/gu;
const TOKEN_REGEX = /[\p{L}\p{N}_-]+/gu;
const PROPER_NAME_REGEX = /\b[A-ZÄÖÜ][\p{L}-]{2,}\b/gu;
const EVENT_REGEX = /\b(?:begann|beginnt|begonnen|traf|treffen|ankam|ankommt|verlor|fand|fanden|entdeckte|entschied|passierte|stirbt|starb|kündigte|kuendigte|explodierte|verschwand)\b/gi;
const THREAD_REGEX = /\b(?:konflikt|ziel|spur|folge|geheimnis|hindernis|motiv|frage|spannung|subplot|offenbarung)\b/gi;
const WORKFLOW_MIN_SCORES = {
  notes: 2,
  persons: 4,
  events: 4,
  threads: 4,
  reminders: 3,
  research: 4,
  clipboard: 3,
  custom: 4,
};
const WORKFLOW_COMBINATION_RULES = [
  {
    key: "scene-triad",
    label: "Zeit · Figur · Ereignis",
    description: "Szenische Kombination",
    types: ["research", "persons", "events"],
    when: (signals) => signals.hasTimeCue && signals.properNames.length > 0 && signals.eventHits.length > 0,
  },
  {
    key: "story-thread",
    label: "Figur · Konflikt · Handlung",
    description: "Handlungsfaden aktiv",
    types: ["persons", "threads", "events"],
    when: (signals) => signals.properNames.length > 0 && signals.threadHits.length > 0,
  },
  {
    key: "research-reminder",
    label: "Zeit · Frage · Recherche",
    description: "Pruefbedarf aktiv",
    types: ["research", "reminders"],
    when: (signals) => signals.hasTimeCue && signals.hasQuestion,
  },
  {
    key: "reference-pack",
    label: "Zitat · Wissen · Notiz",
    description: "Referenzpaket aktiv",
    types: ["clipboard", "research", "notes"],
    when: (signals) => signals.quoteHits.length > 0 && (signals.wikiLinks.length > 0 || /\d/.test(signals.source)),
  },
];

function normalizeWorkflowTag(value) {
  return String(value || "")
    .trim()
    .toLowerCase();
}

function normalizeTagsLocal(values) {
  const seen = new Set();
  return (values || [])
    .map((value) => String(value || "").trim())
    .filter(Boolean)
    .filter((value) => {
      const key = value.toLowerCase();
      if (seen.has(key)) {
        return false;
      }
      seen.add(key);
      return true;
    });
}

function detectStoryTimeCues(text) {
  const source = String(text || "");
  const lowered = source.toLowerCase();
  const keywordHits = STORY_TIME_KEYWORDS.filter((keyword) => lowered.includes(keyword));
  const dateHits = source.match(STORY_DATE_REGEX) || [];
  const rangeHits = source.match(STORY_TIME_RANGE_REGEX) || [];
  return {
    keywordHits,
    dateHits,
    rangeHits,
    hasTimeCue: keywordHits.length > 0 || dateHits.length > 0 || rangeHits.length > 0,
  };
}

function sanitizeEditorAppearance(value) {
  const next = {
    ...DEFAULT_EDITOR_APPEARANCE,
    ...(value && typeof value === "object" ? value : {}),
  };
  next.fontFamily = ALLOWED_FONT_FAMILIES.has(next.fontFamily) ? next.fontFamily : DEFAULT_EDITOR_APPEARANCE.fontFamily;
  next.googleFontName = String(next.googleFontName || DEFAULT_EDITOR_APPEARANCE.googleFontName).trim().slice(0, 80) || DEFAULT_EDITOR_APPEARANCE.googleFontName;
  next.surfacePreset = ALLOWED_SURFACE_PRESETS.has(next.surfacePreset)
    ? next.surfacePreset
    : DEFAULT_EDITOR_APPEARANCE.surfacePreset;
  next.fullscreenBackdrop = ALLOWED_FULLSCREEN_BACKDROPS.has(next.fullscreenBackdrop)
    ? next.fullscreenBackdrop
    : DEFAULT_EDITOR_APPEARANCE.fullscreenBackdrop;
  next.fontSize = Math.min(24, Math.max(16, Number(next.fontSize) || DEFAULT_EDITOR_APPEARANCE.fontSize));
  next.lineHeight = Math.min(2.2, Math.max(1.5, Number(next.lineHeight) || DEFAULT_EDITOR_APPEARANCE.lineHeight));
  next.contentWidth = [640, 720, 860, 960, 1040, 1160].includes(Number(next.contentWidth))
    ? Number(next.contentWidth)
    : DEFAULT_EDITOR_APPEARANCE.contentWidth;
  next.fullscreenContentWidth = [860, 1040, 1200, 1360].includes(Number(next.fullscreenContentWidth))
    ? Number(next.fullscreenContentWidth)
    : DEFAULT_EDITOR_APPEARANCE.fullscreenContentWidth;
  next.caretColor =
    /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.test(String(next.caretColor || "").trim())
      ? String(next.caretColor).trim()
      : DEFAULT_EDITOR_APPEARANCE.caretColor;
  return next;
}

function loadStoredEditorAppearance() {
  if (typeof window === "undefined") {
    return DEFAULT_EDITOR_APPEARANCE;
  }
  try {
    const raw = window.localStorage.getItem(EDITOR_APPEARANCE_STORAGE_KEY);
    if (!raw) {
      return DEFAULT_EDITOR_APPEARANCE;
    }
    return sanitizeEditorAppearance(JSON.parse(raw));
  } catch {
    return DEFAULT_EDITOR_APPEARANCE;
  }
}

function emptyReviewCommentDraft() {
  return {
    comment_type: "comment",
    author: "Review",
    body: "",
    suggested_text: "",
    selected_text: "",
    start_offset: 0,
    end_offset: 0,
    context_before: "",
    context_after: "",
    status: "open",
    is_todo_done: false,
  };
}

function googleFontHref(fontName) {
  return `https://fonts.googleapis.com/css2?family=${String(fontName || "").trim().split(/\s+/).join("+")}&display=swap`;
}

function formatReviewTimestamp(value) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString("de-DE", {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

function tokenizeSelectionText(text) {
  return Array.from(String(text || "").matchAll(TOKEN_REGEX), (match) => normalizeWorkflowTag(match[0])).filter(Boolean);
}

function detectWorkflowSignals(text) {
  const source = String(text || "");
  const lowered = source.toLowerCase();
  const tokens = tokenizeSelectionText(source);
  const hashtags = Array.from(source.matchAll(HASHTAG_REGEX), (match) => normalizeWorkflowTag(match[1])).filter(Boolean);
  const wikiLinks = extractWikiLinks(source);
  const reminderHits = Array.from(source.matchAll(REMINDER_REGEX), (match) => normalizeWorkflowTag(match[0]));
  const quoteHits = source.match(QUOTE_REGEX) || [];
  const properNames = Array.from(source.matchAll(PROPER_NAME_REGEX), (match) => match[0]).filter(Boolean);
  const eventHits = Array.from(source.matchAll(EVENT_REGEX), (match) => normalizeWorkflowTag(match[0]));
  const threadHits = Array.from(source.matchAll(THREAD_REGEX), (match) => normalizeWorkflowTag(match[0]));
  const time = detectStoryTimeCues(source);
  return {
    source,
    lowered,
    tokens,
    hashtags,
    wikiLinks,
    reminderHits,
    quoteHits,
    properNames,
    eventHits,
    threadHits,
    lineCount: source.split("\n").filter((line) => line.trim()).length,
    hasQuestion: source.includes("?"),
    hasTimeCue: time.hasTimeCue,
    timeKeywordHits: time.keywordHits,
    timeDateHits: time.dateHits,
    timeRangeHits: time.rangeHits,
    hasPipeTable: source.includes("|"),
  };
}

function detectWorkflowCombinations(signals) {
  return WORKFLOW_COMBINATION_RULES.filter((rule) => rule.when(signals));
}

function combinationReasonForType(combinations, type) {
  const match = (combinations || []).find((entry) => entry.types.includes(type));
  return match ? `${match.description} · ${match.label}` : "";
}

function scoreWorkflowSuggestion(box, chapterText, selectionText) {
  const tags = (box.tags || []).map(normalizeWorkflowTag).filter(Boolean);
  const title = normalizeWorkflowTag(box.title);
  if (tags.length === 0 && !title && !box.type) {
    return 0;
  }
  const chapterSource = String(chapterText || "").toLowerCase();
  const selectionSignals = detectWorkflowSignals(selectionText);
  const chapterSignals = detectWorkflowSignals(chapterText);
  const selectionCombinations = detectWorkflowCombinations(selectionSignals);
  const chapterCombinations = detectWorkflowCombinations(chapterSignals);
  const reasons = new Set();
  const hasSelectionFocus = Boolean(selectionSignals.source.trim());

  let score = 0;
  const uniqueSelectionWords = new Set(selectionSignals.tokens);
  const uniqueSelectionTags = new Set(selectionSignals.hashtags);

  tags.forEach((tag) => {
    if (uniqueSelectionTags.has(tag)) {
      score += 5;
      reasons.add(`#${tag}`);
      return;
    }
    if (uniqueSelectionWords.has(tag)) {
      score += 4;
      reasons.add(`Tag ${tag}`);
      return;
    }
    if (selectionSignals.lowered.includes(tag)) {
      score += 2;
      reasons.add(`Kontext ${tag}`);
      return;
    }
    if (!hasSelectionFocus && chapterSource.includes(tag)) {
      score += 1;
    }
  });

  if (title && selectionSignals.lowered.includes(title)) {
    score += 2;
    reasons.add(`Titel ${box.title}`);
  }

  const typeHints = WORKFLOW_TYPE_HINTS[box.type] || [];
  const matchedTypeHints = typeHints.filter((hint) => selectionSignals.lowered.includes(hint));
  if (matchedTypeHints.length > 0) {
    score += Math.min(6, matchedTypeHints.length * 2);
    reasons.add(matchedTypeHints[0]);
  }

  const timeTagged = tags.some((tag) =>
    ["zeit", "datum", "jahr", "jahreszahl", "kalender", "uhr", "timeline", "timeline-story"].includes(tag),
  );
  if (timeTagged) {
    if (selectionSignals.hasTimeCue) {
      score += 4;
      reasons.add("Zeitbezug");
    } else if (chapterSignals.hasTimeCue) {
      score += 2;
    }
  }

  if ((box.type === "research" || box.type === "events" || box.type === "threads") && selectionSignals.timeRangeHits.length > 0) {
    score += 3;
    reasons.add("Zeitraum");
  }

  if (box.type === "persons") {
    if (selectionSignals.properNames.length > 0 || selectionSignals.wikiLinks.some((item) => item.type === "person")) {
      score += 3;
      reasons.add("Figurenhinweis");
    }
  }

  if (box.type === "events" && (selectionSignals.eventHits.length > 0 || selectionSignals.hasTimeCue)) {
    score += selectionSignals.eventHits.length > 0 ? 3 : 2;
    reasons.add(selectionSignals.eventHits.length > 0 ? "Ereignis" : "Zeitfenster");
  }

  if (box.type === "threads" && (selectionSignals.threadHits.length > 0 || selectionSignals.hasQuestion)) {
    score += selectionSignals.threadHits.length > 0 ? 3 : 2;
    reasons.add(selectionSignals.threadHits.length > 0 ? "Handlungsfaden" : "offene Frage");
  }

  if (box.type === "reminders" && (selectionSignals.hasQuestion || selectionSignals.reminderHits.length > 0)) {
    score += selectionSignals.reminderHits.length > 0 ? 4 : 2;
    reasons.add(selectionSignals.reminderHits.length > 0 ? "To-do" : "Frage");
  }

  if (box.type === "research" && (selectionSignals.wikiLinks.length > 0 || /\d/.test(selectionText || ""))) {
    score += 2;
    reasons.add(selectionSignals.wikiLinks.length > 0 ? "Wiki-Bezug" : "Faktenbezug");
  }

  if (box.type === "clipboard" && (selectionSignals.quoteHits.length > 0 || selectionSignals.lineCount <= 3)) {
    score += 2;
    reasons.add(selectionSignals.quoteHits.length > 0 ? "Zitat" : "Snippet");
  }

  const selectionComboReason = combinationReasonForType(selectionCombinations, box.type);
  if (selectionComboReason) {
    score += 3;
    reasons.add(selectionComboReason);
  }

  if (!hasSelectionFocus) {
    const chapterComboReason = combinationReasonForType(chapterCombinations, box.type);
    if (chapterComboReason) {
      score += 1;
      reasons.add(chapterComboReason);
    }
  }

  if (box.type === "notes" && score === 0 && hasSelectionFocus && selectionSignals.lineCount <= 5) {
    score += 1;
    reasons.add("Allgemein");
  }

  if (!hasSelectionFocus && score > 0) {
    score = Math.max(0, score - 1);
  }

  return {
    score,
    reasons: Array.from(reasons).slice(0, 3),
  };
}

function minimumWorkflowScore(type, hasSelectionFocus) {
  const base = WORKFLOW_MIN_SCORES[type] || WORKFLOW_MIN_SCORES.custom;
  return hasSelectionFocus ? base : base + 2;
}

function activationReasonText(reasons) {
  return Array.isArray(reasons) && reasons.length > 0 ? reasons.join(" · ") : "Noch keine aktiven Signale";
}

function App() {
  const editorRef = useRef(null);
  const markdownTextareaRef = useRef(null);
  const autosaveRef = useRef(null);
  const skipAutosaveRef = useRef(true);
  const selectionPopupDelayRef = useRef(null);
  const selectionPopupFadeRef = useRef(null);
  const editorHeaderHideRef = useRef(null);
  const projectSectionRef = useRef(null);
  const bookSectionRef = useRef(null);
  const chapterSectionRef = useRef(null);
  const workflowSectionRef = useRef(null);
  const knowledgeSectionRef = useRef(null);
  const contextSectionRef = useRef(null);
  const anchorSectionRef = useRef(null);
  const clipboardSectionRef = useRef(null);

  const [projects, setProjects] = useState([]);
  const [projectDetail, setProjectDetail] = useState(null);
  const [selectedProjectId, setSelectedProjectId] = useState("");
  const [selectedBookId, setSelectedBookId] = useState("");
  const [bookBundle, setBookBundle] = useState(null);
  const [selectedChapterId, setSelectedChapterId] = useState("");
  const [chapterDraft, setChapterDraft] = useState(EMPTY_DRAFT);
  const [anchors, setAnchors] = useState([]);
  const [reviewComments, setReviewComments] = useState([]);
  const [clipboardItems, setClipboardItems] = useState([]);
  const [knowledgeItems, setKnowledgeItems] = useState([]);
  const [knowledgeQuery, setKnowledgeQuery] = useState("");
  const [selectedKnowledgeItemId, setSelectedKnowledgeItemId] = useState("");
  const [selectedKnowledgeRefKey, setSelectedKnowledgeRefKey] = useState("");
  const [selectedWorkflowBoxId, setSelectedWorkflowBoxId] = useState("");
  const [hasSelection, setHasSelection] = useState(false);
  const [editorMode, setEditorMode] = useState("rich");
  const [saveState, setSaveState] = useState("Synchron");
  const [errorMessage, setErrorMessage] = useState("");
  const [showEditorHelp, setShowEditorHelp] = useState(false);
  const [showEditorSettings, setShowEditorSettings] = useState(false);
  const [showWritingTools, setShowWritingTools] = useState(false);
  const [showClipboardPalette, setShowClipboardPalette] = useState(false);
  const [selectionContext, setSelectionContext] = useState(null);
  const [selectionPopupVisible, setSelectionPopupVisible] = useState(false);
  const [showReviewComposer, setShowReviewComposer] = useState(false);
  const [reviewCommentDraft, setReviewCommentDraft] = useState(emptyReviewCommentDraft);
  const [activeReviewCommentId, setActiveReviewCommentId] = useState("");
  const [reviewBubblePosition, setReviewBubblePosition] = useState(null);
  const [isEditingReviewSuggestion, setIsEditingReviewSuggestion] = useState(false);
  const [reviewSuggestionDraft, setReviewSuggestionDraft] = useState("");
  const [isEditorFullscreen, setIsEditorFullscreen] = useState(false);
  const [draggedChapterId, setDraggedChapterId] = useState("");
  const [chapterDropTargetId, setChapterDropTargetId] = useState("");
  const [showProjectPicker, setShowProjectPicker] = useState(false);
  const [showBookPicker, setShowBookPicker] = useState(false);
  const [showProjectDetails, setShowProjectDetails] = useState(false);
  const [showBookDetails, setShowBookDetails] = useState(false);
  const [showProjectEdit, setShowProjectEdit] = useState(false);
  const [showBookEdit, setShowBookEdit] = useState(false);
  const [showLeftOverlay, setShowLeftOverlay] = useState(false);
  const [showRightOverlay, setShowRightOverlay] = useState(false);
  const [activeLeftSection, setActiveLeftSection] = useState("chapter");
  const [activeRightSection, setActiveRightSection] = useState("clipboard");
  const [editorAppearance, setEditorAppearance] = useState(loadStoredEditorAppearance);
  const [showEditorHeader, setShowEditorHeader] = useState(false);
  const slotNumbers = Array.from({ length: 9 }, (_, index) => index + 1);

  const currentChapter = useMemo(
    () => bookBundle?.chapters?.find((chapter) => chapter.id === selectedChapterId) || null,
    [bookBundle, selectedChapterId],
  );
  const currentChapterId = currentChapter?.id || "";
  const currentBook = useMemo(
    () => projectDetail?.books?.find((book) => book.id === selectedBookId) || bookBundle?.book || null,
    [projectDetail, selectedBookId, bookBundle],
  );
  const currentProject = useMemo(
    () => projectDetail?.project || projects.find((project) => project.id === selectedProjectId) || null,
    [projectDetail, projects, selectedProjectId],
  );
  const chaptersById = useMemo(
    () => new Map((bookBundle?.chapters || []).map((chapter) => [chapter.id, chapter])),
    [bookBundle?.chapters],
  );

  const pinnedSlots = useMemo(
    () =>
      clipboardItems
        .filter((item) => item.is_pinned && item.slot >= 1 && item.slot <= 9)
        .sort((left, right) => left.slot - right.slot),
    [clipboardItems],
  );
  const latestClipboardItems = useMemo(() => clipboardItems.slice(0, 3), [clipboardItems]);

  const chapterKnowledgeRefs = useMemo(
    () => extractWikiLinks(chapterDraft.markdown_content || currentChapter?.markdown_content || ""),
    [chapterDraft.markdown_content, currentChapter?.markdown_content],
  );

  const chapterKnowledgeMatches = useMemo(() => {
    const lookup = new Map(
      knowledgeItems.map((item) => [normalizeKnowledgeKey(item.type, item.name), item]),
    );
    return chapterKnowledgeRefs
      .map((reference) => ({
        reference,
        item: lookup.get(reference.key) || null,
      }))
      .filter(
        (entry, index, items) =>
          items.findIndex((candidate) => candidate.reference.key === entry.reference.key) === index,
      );
  }, [chapterKnowledgeRefs, knowledgeItems]);

  const unresolvedKnowledgeRefs = useMemo(
    () => chapterKnowledgeMatches.filter((entry) => !entry.item),
    [chapterKnowledgeMatches],
  );

  const filteredKnowledgeItems = useMemo(() => {
    const query = knowledgeQuery.trim().toLowerCase();
    if (!query) {
      return knowledgeItems;
    }
    return knowledgeItems.filter((item) =>
      [item.name, item.summary, item.type, ...(item.tags || [])].join(" ").toLowerCase().includes(query),
    );
  }, [knowledgeItems, knowledgeQuery]);

  const selectedKnowledgeItem = useMemo(
    () => filteredKnowledgeItems.find((item) => item.id === selectedKnowledgeItemId) || filteredKnowledgeItems[0] || null,
    [filteredKnowledgeItems, selectedKnowledgeItemId],
  );

  const selectedKnowledgeRef = useMemo(
    () =>
      chapterKnowledgeMatches.find((entry) => entry.reference.key === selectedKnowledgeRefKey) ||
      chapterKnowledgeMatches[0] ||
      null,
    [chapterKnowledgeMatches, selectedKnowledgeRefKey],
  );

  const manualWorkflowBox = useMemo(
    () => bookBundle?.workflow_boxes?.find((item) => item.id === selectedWorkflowBoxId) || null,
    [bookBundle, selectedWorkflowBoxId],
  );

  const anchorCountByWorkflowBox = useMemo(() => {
    const counts = new Map();
    anchors.forEach((anchor) => {
      counts.set(anchor.workflow_box_id, (counts.get(anchor.workflow_box_id) || 0) + 1);
    });
    return counts;
  }, [anchors]);
  const anchorsByWorkflowBox = useMemo(() => {
    const grouped = new Map();
    anchors.forEach((anchor) => {
      grouped.set(anchor.workflow_box_id, [...(grouped.get(anchor.workflow_box_id) || []), anchor]);
    });
    return grouped;
  }, [anchors]);
  const latestAnchorByWorkflowBox = useMemo(() => {
    const latest = new Map();
    anchors.forEach((anchor) => {
      latest.set(anchor.workflow_box_id, anchor);
    });
    return latest;
  }, [anchors]);
  const activeSelectionPayload = selectionContext?.payload || null;
  const activeReviewComment = useMemo(
    () => reviewComments.find((comment) => comment.id === activeReviewCommentId) || null,
    [activeReviewCommentId, reviewComments],
  );
  const hasSelectionFocus = Boolean(activeSelectionPayload?.selected_text?.trim());
  const chapterTextForWorkflow = chapterDraft.markdown_content || currentChapter?.markdown_content || "";
  const workflowSuggestions = useMemo(() => {
    const selectionText = activeSelectionPayload?.selected_text || "";
    return (bookBundle?.workflow_boxes || [])
      .map((box) => ({
        box,
        ...scoreWorkflowSuggestion(box, chapterTextForWorkflow, selectionText),
      }))
      .filter((entry) => entry.score >= minimumWorkflowScore(entry.box.type, Boolean(selectionText.trim())))
      .sort((left, right) => right.score - left.score)
      .slice(0, hasSelectionFocus ? 3 : 2);
  }, [bookBundle?.workflow_boxes, chapterTextForWorkflow, activeSelectionPayload?.selected_text, hasSelectionFocus]);
  const showWorkflowSuggestionCloud =
    workflowSuggestions.length > 0 && (selectionPopupVisible || hasSelectionFocus || workflowSuggestions[0]?.score >= 7);
  const primaryWorkflowSuggestion = workflowSuggestions[0] || null;
  const activeSelectionCombinations = useMemo(
    () => (hasSelectionFocus ? detectWorkflowCombinations(detectWorkflowSignals(activeSelectionPayload?.selected_text || "")) : []),
    [activeSelectionPayload?.selected_text, hasSelectionFocus],
  );
  const autoWorkflowBoxId = useMemo(() => {
    if (!hasSelectionFocus || !primaryWorkflowSuggestion) {
      return "";
    }
    const threshold = minimumWorkflowScore(primaryWorkflowSuggestion.box.type, true) + 1;
    return primaryWorkflowSuggestion.score >= threshold ? primaryWorkflowSuggestion.box.id : "";
  }, [hasSelectionFocus, primaryWorkflowSuggestion]);
  const hasTemporaryAutoTarget = Boolean(autoWorkflowBoxId && autoWorkflowBoxId !== selectedWorkflowBoxId);
  const effectiveWorkflowBoxId = autoWorkflowBoxId || selectedWorkflowBoxId;
  const activeWorkflowBox = useMemo(
    () => bookBundle?.workflow_boxes?.find((item) => item.id === effectiveWorkflowBoxId) || null,
    [bookBundle, effectiveWorkflowBoxId],
  );
  const activeWorkflowAnchors = useMemo(
    () => anchorsByWorkflowBox.get(effectiveWorkflowBoxId) || [],
    [anchorsByWorkflowBox, effectiveWorkflowBoxId],
  );
  const workflowActivationById = useMemo(() => {
    const selectionText = activeSelectionPayload?.selected_text || "";
    const suggestionLookup = new Map(workflowSuggestions.map((entry, index) => [entry.box.id, { ...entry, rank: index }]));
    const selectionSignals = detectWorkflowSignals(selectionText);
    const chapterSignals = detectWorkflowSignals(chapterTextForWorkflow);
    const selectionCombinations = hasSelectionFocus ? detectWorkflowCombinations(selectionSignals) : [];
    const chapterCombinations = detectWorkflowCombinations(chapterSignals);
    return new Map(
      (bookBundle?.workflow_boxes || []).map((box) => {
        const anchorCount = anchorCountByWorkflowBox.get(box.id) || 0;
        const chapterActivation = scoreWorkflowSuggestion(box, chapterTextForWorkflow, "");
        const selectionActivation = hasSelectionFocus
          ? scoreWorkflowSuggestion(box, chapterTextForWorkflow, selectionText)
          : { score: 0, reasons: [] };
        const suggestion = suggestionLookup.get(box.id) || null;
        const selectionCombinationReason = combinationReasonForType(selectionCombinations, box.type);
        const chapterCombinationReason = combinationReasonForType(chapterCombinations, box.type);
        let tone = "idle";
        let label = "Ruhend";
        let reason = activationReasonText(chapterActivation.reasons);

        if (hasTemporaryAutoTarget && autoWorkflowBoxId === box.id) {
          tone = "focus";
          label = "Auto-Ziel";
          reason = activationReasonText(selectionActivation.reasons);
        } else if (hasTemporaryAutoTarget && selectedWorkflowBoxId === box.id) {
          tone = "selected";
          label = "Basis";
          reason = `Manuelles Ziel · kehrt nach der Auswahl zu ${box.title} zurueck`;
        } else if (selectedWorkflowBoxId === box.id && hasSelectionFocus && selectionActivation.score > 0) {
          tone = "focus";
          label = "Im Fokus";
          reason = activationReasonText(selectionActivation.reasons);
        } else if (selectedWorkflowBoxId === box.id) {
          tone = "selected";
          label = "Ziel";
          reason = anchorCount > 0 ? `${anchorCount} Anker im Kapitel` : "Manuell als Ziel gesetzt";
        } else if (selectionCombinationReason) {
          tone = "combo";
          label = "Kombi";
          reason = selectionCombinationReason;
        } else if (suggestion && suggestion.rank === 0) {
          tone = "focus";
          label = "Im Fokus";
          reason = activationReasonText(suggestion.reasons);
        } else if (suggestion) {
          tone = "ready";
          label = "Bereit";
          reason = activationReasonText(suggestion.reasons);
        } else if (chapterCombinationReason) {
          tone = "combo";
          label = "Kombi";
          reason = chapterCombinationReason;
        } else if (chapterActivation.score >= minimumWorkflowScore(box.type, false)) {
          tone = "context";
          label = "Im Kontext";
          reason = activationReasonText(chapterActivation.reasons);
        } else if (anchorCount > 0) {
          tone = "linked";
          label = "Verbunden";
          reason = `${anchorCount} Anker im Kapitel`;
        }

        return [
          box.id,
          {
            tone,
            label,
            reason,
            anchorCount,
            selectionScore: selectionActivation.score,
            chapterScore: chapterActivation.score,
            isSuggested: Boolean(suggestion),
            isPrimary: suggestion?.rank === 0,
            comboReason: selectionCombinationReason || chapterCombinationReason || "",
          },
        ];
      }),
    );
  }, [
    activeSelectionPayload?.selected_text,
    anchorCountByWorkflowBox,
    autoWorkflowBoxId,
    bookBundle?.workflow_boxes,
    chapterTextForWorkflow,
    hasSelectionFocus,
    hasTemporaryAutoTarget,
    selectedWorkflowBoxId,
    workflowSuggestions,
  ]);
  const fullscreenBackdropTheme = FULLSCREEN_BACKDROP_THEME[editorAppearance.fullscreenBackdrop] || FULLSCREEN_BACKDROP_THEME.linen;
  const activeWorkflowState = activeWorkflowBox ? workflowActivationById.get(activeWorkflowBox.id) || null : null;
  const showWorkflowTargetDetails = Boolean(
    activeWorkflowBox &&
      (hasSelection ||
        hasSelectionFocus ||
        hasTemporaryAutoTarget ||
        activeWorkflowAnchors.length > 0 ||
        activeWorkflowState?.comboReason ||
        activeWorkflowState?.tone === "focus" ||
        activeWorkflowState?.tone === "combo"),
  );
  const showFloatingStatus = !["Synchron", "Gespeichert", "Autosave gespeichert"].includes(saveState);
  const editorMetaItems = [
    effectiveWorkflowBoxId
      ? hasTemporaryAutoTarget
        ? `Auto-Ziel: ${activeWorkflowBox?.title || ""} · Basis: ${manualWorkflowBox?.title || ""}`
        : `Zielbox: ${activeWorkflowBox?.title || ""}`
      : "",
    editorMode === "markdown" ? "Modus · Markdown" : "",
  ].filter(Boolean);

  const editorSurfaceStyle = useMemo(
    () => ({
      "--editor-font-family":
        editorAppearance.fontFamily === "google"
          ? `"${editorAppearance.googleFontName}", "Iowan Old Style", "Palatino Linotype", "Book Antiqua", Palatino, serif`
          : editorAppearance.fontFamily === "sans"
          ? '"Avenir Next", "Segoe UI", sans-serif'
          : editorAppearance.fontFamily === "mono"
            ? '"SFMono-Regular", "Menlo", "Monaco", monospace'
            : '"Iowan Old Style", "Palatino Linotype", "Book Antiqua", Palatino, serif',
      "--editor-font-size": `${editorAppearance.fontSize}px`,
      "--editor-line-height": String(editorAppearance.lineHeight),
      "--editor-table-cell-padding-y": `${Math.max(10, Math.round(editorAppearance.fontSize * editorAppearance.lineHeight * 0.34))}px`,
      "--editor-table-cell-padding-x": `${Math.max(12, Math.round(editorAppearance.fontSize * editorAppearance.lineHeight * 0.4))}px`,
      "--editor-max-width": `${editorAppearance.contentWidth}px`,
      "--editor-fullscreen-max-width": `${editorAppearance.fullscreenContentWidth}px`,
      "--editor-caret-color": editorAppearance.caretColor,
      "--editor-selection-bg":
        editorAppearance.surfacePreset === "night" ? "rgba(243, 237, 229, 0.16)" : "rgba(158, 91, 33, 0.14)",
      "--editor-selection-text":
        editorAppearance.surfacePreset === "night" ? "#f7f1ea" : "var(--editor-ink)",
      "--fullscreen-backdrop-glow": fullscreenBackdropTheme.glow,
      "--fullscreen-backdrop-start": fullscreenBackdropTheme.start,
      "--fullscreen-backdrop-end": fullscreenBackdropTheme.end,
      "--fullscreen-card-start": fullscreenBackdropTheme.cardStart,
      "--fullscreen-card-end": fullscreenBackdropTheme.cardEnd,
      "--fullscreen-card-border": fullscreenBackdropTheme.cardBorder,
      "--fullscreen-card-shadow": fullscreenBackdropTheme.cardShadow,
    }),
    [editorAppearance, fullscreenBackdropTheme],
  );

  function clearSelectionPopup() {
    window.clearTimeout(selectionPopupDelayRef.current);
    window.clearTimeout(selectionPopupFadeRef.current);
    setSelectionPopupVisible(false);
    setSelectionContext(null);
  }

  function closeTransientPanels() {
    clearSelectionPopup();
    setShowLeftOverlay(false);
    setShowRightOverlay(false);
    setShowEditorHelp(false);
    setShowEditorSettings(false);
    setShowClipboardPalette(false);
    setShowReviewComposer(false);
    setShowEditorHeader(false);
    setActiveReviewCommentId("");
    setReviewBubblePosition(null);
    setIsEditingReviewSuggestion(false);
    setReviewSuggestionDraft("");
  }

  function showSelectionPopup(nextContext, delay = 0) {
    window.clearTimeout(selectionPopupDelayRef.current);
    window.clearTimeout(selectionPopupFadeRef.current);
    setSelectionPopupVisible(false);

    const reveal = () => {
      setSelectionContext(nextContext);
      selectionPopupFadeRef.current = window.setTimeout(() => {
        setSelectionPopupVisible(true);
      }, POPUP_FADE_IN_DELAY_MS);
    };

    if (delay > 0) {
      setSelectionContext(null);
      selectionPopupDelayRef.current = window.setTimeout(reveal, delay);
      return;
    }

    reveal();
  }

  useEffect(() => {
    loadProjects();
  }, []);

  useEffect(() => {
    try {
      window.localStorage.setItem(EDITOR_APPEARANCE_STORAGE_KEY, JSON.stringify(editorAppearance));
    } catch {
      // ignore local persistence failures
    }
  }, [editorAppearance]);

  useEffect(
    () => () => {
      window.clearTimeout(selectionPopupDelayRef.current);
      window.clearTimeout(selectionPopupFadeRef.current);
    },
    [],
  );

  useEffect(() => {
    if (!selectedProjectId) {
      return;
    }
    setShowProjectPicker(false);
    setShowProjectDetails(false);
    setShowProjectEdit(false);
    loadProject(selectedProjectId);
    loadKnowledgeItems(selectedProjectId);
  }, [selectedProjectId]);

  useEffect(() => {
    if (!filteredKnowledgeItems.some((item) => item.id === selectedKnowledgeItemId)) {
      setSelectedKnowledgeItemId(filteredKnowledgeItems[0]?.id || "");
    }
  }, [filteredKnowledgeItems, selectedKnowledgeItemId]);

  useEffect(() => {
    if (!chapterKnowledgeMatches.some((entry) => entry.reference.key === selectedKnowledgeRefKey)) {
      setSelectedKnowledgeRefKey(chapterKnowledgeMatches[0]?.reference.key || "");
    }
  }, [chapterKnowledgeMatches, selectedKnowledgeRefKey]);

  useEffect(() => {
    if (!selectedBookId) {
      setBookBundle(null);
      setSelectedChapterId("");
      setClipboardItems([]);
      setShowClipboardPalette(false);
      setShowBookDetails(false);
      setShowBookEdit(false);
      setShowBookPicker(false);
      clearSelectionPopup();
      return;
    }
    setShowBookPicker(false);
    setShowBookDetails(false);
    setShowBookEdit(false);
    loadBook(selectedBookId);
  }, [selectedBookId]);

  useEffect(() => {
    const lookup = {
      project: projectSectionRef,
      book: bookSectionRef,
      chapter: chapterSectionRef,
      workflow: workflowSectionRef,
      knowledge: knowledgeSectionRef,
    };
    if (showLeftOverlay && lookup[activeLeftSection]?.current) {
      lookup[activeLeftSection].current.scrollIntoView({ block: "start", behavior: "smooth" });
    }
  }, [activeLeftSection, showLeftOverlay]);

  useEffect(() => {
    const lookup = {
      context: contextSectionRef,
      anchor: anchorSectionRef,
      clipboard: clipboardSectionRef,
    };
    if (showRightOverlay && lookup[activeRightSection]?.current) {
      lookup[activeRightSection].current.scrollIntoView({ block: "start", behavior: "smooth" });
    }
  }, [activeRightSection, showRightOverlay]);

  useEffect(() => {
    if (clipboardItems.length === 0) {
      setShowClipboardPalette(false);
    }
  }, [clipboardItems.length]);

  useEffect(() => {
    if (activeReviewCommentId && !reviewComments.some((comment) => comment.id === activeReviewCommentId)) {
      setActiveReviewCommentId("");
      setReviewBubblePosition(null);
    }
  }, [activeReviewCommentId, reviewComments]);

  useEffect(() => {
    if (!activeReviewComment || activeReviewComment.comment_type !== "suggestion") {
      setIsEditingReviewSuggestion(false);
      setReviewSuggestionDraft("");
      return;
    }
    setReviewSuggestionDraft(activeReviewComment.suggested_text || activeReviewComment.selected_text || "");
  }, [activeReviewComment]);

  useEffect(() => {
    if (!selectedChapterId || !currentChapter) {
      setAnchors([]);
      setReviewComments([]);
      setChapterDraft(EMPTY_DRAFT);
      setErrorMessage("");
      setSaveState("Synchron");
      clearSelectionPopup();
      setActiveReviewCommentId("");
      setReviewBubblePosition(null);
      setShowReviewComposer(false);
      return;
    }
    skipAutosaveRef.current = true;
    setErrorMessage("");
    setSaveState("Synchron");
    clearSelectionPopup();
    setChapterDraft({
      title: currentChapter.title,
      summary: currentChapter.summary || "",
      markdown_content: currentChapter.markdown_content || "",
      editor_json: currentChapter.editor_json || "",
    });
    loadAnchors(currentChapter.id);
    loadReviewComments(currentChapter.id);
  }, [selectedChapterId, currentChapterId]);

  useEffect(() => {
    if (!currentChapter) {
      return undefined;
    }
    if (skipAutosaveRef.current) {
      skipAutosaveRef.current = false;
      return undefined;
    }

    const changed =
      chapterDraft.title !== currentChapter.title ||
      chapterDraft.summary !== (currentChapter.summary || "") ||
      chapterDraft.markdown_content !== (currentChapter.markdown_content || "") ||
      chapterDraft.editor_json !== (currentChapter.editor_json || "");

    if (!changed) {
      setSaveState("Synchron");
      return undefined;
    }

    setSaveState("Autosave ausstehend");
    clearTimeout(autosaveRef.current);
    autosaveRef.current = window.setTimeout(() => {
      saveChapter(false);
    }, 1500);

    return () => clearTimeout(autosaveRef.current);
  }, [chapterDraft, currentChapter]);

  useEffect(() => {
    function handleGlobalShortcuts(event) {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "s") {
        event.preventDefault();
        void saveChapter(true);
        return;
      }

      if (event.key === "Escape" && (showLeftOverlay || showRightOverlay || showEditorHelp || showEditorSettings || showClipboardPalette || showReviewComposer || activeReviewCommentId || selectionContext)) {
        event.preventDefault();
        closeTransientPanels();
        return;
      }

      if (event.key === "Escape" && isEditorFullscreen) {
        event.preventDefault();
        setIsEditorFullscreen(false);
      }
    }

    window.addEventListener("keydown", handleGlobalShortcuts);
    return () => window.removeEventListener("keydown", handleGlobalShortcuts);
  }, [
    showLeftOverlay,
    showRightOverlay,
    showEditorHelp,
    showEditorSettings,
    showClipboardPalette,
    showReviewComposer,
    activeReviewCommentId,
    selectionContext,
    isEditorFullscreen,
    currentChapter,
    chapterDraft,
    editorMode,
  ]);

  useEffect(() => () => clearTimeout(editorHeaderHideRef.current), []);

  useEffect(() => {
    if (!isEditorFullscreen) {
      return;
    }
    closeTransientPanels();
    setShowEditorHeader(false);
  }, [isEditorFullscreen]);

  useEffect(() => {
    if (typeof document === "undefined") {
      return;
    }
    const linkId = "easy-author-google-font";
    const existing = document.getElementById(linkId);
    if (editorAppearance.fontFamily !== "google" || !editorAppearance.googleFontName) {
      existing?.remove();
      return;
    }
    const href = googleFontHref(editorAppearance.googleFontName);
    if (existing) {
      existing.setAttribute("href", href);
      return;
    }
    const link = document.createElement("link");
    link.id = linkId;
    link.rel = "stylesheet";
    link.href = href;
    document.head.appendChild(link);
  }, [editorAppearance.fontFamily, editorAppearance.googleFontName]);

  async function loadProjects() {
    try {
      setErrorMessage("");
      const response = await api.get("/api/projects");
      const items = response.projects || [];
      setProjects(items);
      if (!selectedProjectId && items.length > 0) {
        setSelectedProjectId(items[0].id);
      }
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function loadProject(projectId) {
    try {
      const response = await api.get(`/api/projects/${projectId}`);
      setProjectDetail(response);
      setErrorMessage("");
      const firstBook = response.books?.[0];
      if (!selectedBookId || !response.books?.some((book) => book.id === selectedBookId)) {
        setSelectedBookId(firstBook?.id || "");
      }
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function loadKnowledgeItems(projectId) {
    try {
      const response = await api.get(`/api/projects/${projectId}/knowledge-items`);
      setKnowledgeItems(response.knowledge_items || []);
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function loadBook(bookId) {
    try {
      const response = await api.get(`/api/books/${bookId}`);
      setBookBundle(response);
      setClipboardItems(response.clipboard || []);
      setErrorMessage("");
      if (!selectedChapterId || !response.chapters?.some((chapter) => chapter.id === selectedChapterId)) {
        setSelectedChapterId(response.chapters?.[0]?.id || "");
      }
      if (!selectedWorkflowBoxId || !response.workflow_boxes?.some((item) => item.id === selectedWorkflowBoxId)) {
        setSelectedWorkflowBoxId(response.workflow_boxes?.[0]?.id || "");
      }
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function loadAnchors(chapterId) {
    try {
      const response = await api.get(`/api/chapters/${chapterId}/anchors`);
      setAnchors(response.anchors || []);
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function loadReviewComments(chapterId) {
    try {
      const response = await api.get(`/api/chapters/${chapterId}/comments`);
      setReviewComments(response.comments || []);
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function saveChapter(manual) {
    if (!currentChapter) {
      return;
    }
    try {
      const richSnapshot =
        editorMode === "rich"
          ? editorRef.current?.getDocumentSnapshot?.() || null
          : null;
      const liveDraft =
        richSnapshot && (richSnapshot.markdown_content || richSnapshot.editor_json)
          ? {
              ...chapterDraft,
              ...richSnapshot,
            }
          : chapterDraft;
      const payload =
        editorMode === "markdown"
          ? {
              ...liveDraft,
              editor_json: JSON.stringify(markdownToDoc(liveDraft.markdown_content || "")),
            }
          : liveDraft;
      if (payload !== chapterDraft) {
        setChapterDraft(payload);
      }
      const requestPayload = { ...payload };
      if (!requestPayload.summary && !(currentChapter?.summary || "")) {
        delete requestPayload.summary;
      }
      setErrorMessage("");
      setSaveState(manual ? "Speichert ..." : "Autosave laeuft ...");
      const updated = await api.put(`/api/chapters/${currentChapter.id}`, requestPayload);
      setBookBundle((previous) => ({
        ...previous,
        chapters: previous.chapters.map((chapter) => (chapter.id === updated.id ? updated : chapter)),
      }));
      const nextDraft = {
        title: updated.title,
        summary: updated.summary || "",
        markdown_content: updated.markdown_content || "",
        editor_json: updated.editor_json || "",
      };
      setChapterDraft((previous) =>
        previous.title === nextDraft.title &&
        previous.markdown_content === nextDraft.markdown_content &&
        previous.editor_json === nextDraft.editor_json
          ? previous
          : nextDraft,
      );
      setErrorMessage("");
      setSaveState(manual ? "Gespeichert" : "Autosave gespeichert");
      window.setTimeout(() => setSaveState("Synchron"), 1200);
    } catch (error) {
      setSaveState("Fehler beim Speichern");
      setErrorMessage(error.message);
    }
  }

  async function createProject() {
    const title = window.prompt("Titel des neuen Projekts", "Neues Projekt");
    if (!title) {
      return;
    }
    try {
      const project = await api.post("/api/projects", { title, description: "" });
      await loadProjects();
      setSelectedProjectId(project.id);
      setShowProjectPicker(false);
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function createBook() {
    if (!selectedProjectId) {
      return;
    }
    const title = window.prompt("Titel des neuen Buchs", "Neues Buch");
    if (!title) {
      return;
    }
    try {
      const book = await api.post(`/api/projects/${selectedProjectId}/books`, {
        title,
        subtitle: "",
        author: "",
        visibility: "private",
      });
      await loadProject(selectedProjectId);
      setSelectedBookId(book.id);
      setShowBookPicker(false);
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function createChapter() {
    if (!selectedBookId) {
      return;
    }
    const title = window.prompt("Titel des neuen Kapitels", `Kapitel ${(bookBundle?.chapters?.length || 0) + 1}`);
    if (!title) {
      return;
    }
    try {
      const chapter = await api.post(`/api/books/${selectedBookId}/chapters`, {
        title,
        summary: "",
        markdown_content: `# ${title}\n`,
        editor_json: "",
      });
      await loadBook(selectedBookId);
      setSelectedChapterId(chapter.id);
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function reorderChapters(sourceChapterId, targetChapterId) {
    if (!selectedBookId || !bookBundle?.chapters?.length || sourceChapterId === targetChapterId) {
      return;
    }

    const previousChapters = bookBundle.chapters;
    const sourceIndex = previousChapters.findIndex((chapter) => chapter.id === sourceChapterId);
    const targetIndex = previousChapters.findIndex((chapter) => chapter.id === targetChapterId);
    if (sourceIndex === -1 || targetIndex === -1) {
      return;
    }

    const nextChapters = [...previousChapters];
    const [movedChapter] = nextChapters.splice(sourceIndex, 1);
    nextChapters.splice(targetIndex, 0, movedChapter);
    const positionedChapters = nextChapters.map((chapter, index) => ({
      ...chapter,
      position: index + 1,
    }));

    setBookBundle((previous) => ({
      ...previous,
      chapters: positionedChapters,
    }));

    try {
      const response = await api.put(`/api/books/${selectedBookId}/chapters/reorder`, {
        chapter_ids: positionedChapters.map((chapter) => chapter.id),
      });
      setBookBundle((previous) => ({
        ...previous,
        chapters: response.chapters || positionedChapters,
      }));
      setErrorMessage("");
    } catch (error) {
      setBookBundle((previous) => ({
        ...previous,
        chapters: previousChapters,
      }));
      if (error.status === 404) {
        setErrorMessage("Kapitel konnten noch nicht neu sortiert werden. Bitte den easy-author-Backend-Prozess einmal neu starten.");
        return;
      }
      setErrorMessage(error.message);
    }
  }

  async function createWorkflowBox(defaults = {}) {
    if (!selectedBookId) {
      return;
    }
    const title = window.prompt("Name der Workflow-Box", defaults.title || "Neue Box");
    if (!title) {
      return;
    }
    try {
      const box = await api.post(`/api/books/${selectedBookId}/workflow-boxes`, {
        title,
        type: defaults.type || "custom",
        tags: defaults.tags || [],
        is_collapsed: false,
      });
      await loadBook(selectedBookId);
      setSelectedWorkflowBoxId(box.id);
      return box;
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function createKnowledgeItem() {
    if (!selectedProjectId) {
      return;
    }
    const name = window.prompt("Name des Wissenseintrags", "Neuer Eintrag");
    if (!name) {
      return;
    }
    try {
      const created = await api.post(`/api/projects/${selectedProjectId}/knowledge-items`, {
        type: "person",
        name,
        summary: "",
        body: "",
        tags: [],
      });
      setKnowledgeItems((previous) =>
        [...previous, created].sort((left, right) => left.name.localeCompare(right.name, "de")),
      );
      setSelectedKnowledgeItemId(created.id);
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function updateKnowledgeItem(item, nextFields) {
    try {
      const updated = await api.put(`/api/knowledge-items/${item.id}`, {
        type: nextFields.type ?? item.type,
        name: nextFields.name ?? item.name,
        summary: nextFields.summary ?? item.summary,
        body: nextFields.body ?? item.body,
        tags: nextFields.tags ?? item.tags,
      });
      setKnowledgeItems((previous) =>
        previous
          .map((entry) => (entry.id === updated.id ? updated : entry))
          .sort((left, right) => left.name.localeCompare(right.name, "de")),
      );
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  function patchLocalBook(bookId, nextFields) {
    setProjectDetail((previous) =>
      previous
        ? {
            ...previous,
            books: previous.books.map((entry) => (entry.id === bookId ? { ...entry, ...nextFields } : entry)),
          }
        : previous,
    );
    setBookBundle((previous) =>
      previous?.book?.id === bookId
        ? {
            ...previous,
            book: {
              ...previous.book,
              ...nextFields,
            },
          }
        : previous,
    );
  }

  function patchLocalProject(projectId, nextFields) {
    setProjects((previous) =>
      previous.map((entry) => (entry.id === projectId ? { ...entry, ...nextFields } : entry)),
    );
    setProjectDetail((previous) =>
      previous?.project?.id === projectId
        ? {
            ...previous,
            project: {
              ...previous.project,
              ...nextFields,
            },
          }
        : previous,
    );
  }

  async function updateProject(projectId, nextFields) {
    const project = currentProject?.id === projectId ? currentProject : projects.find((entry) => entry.id === projectId);
    if (!project) {
      return;
    }
    try {
      const updated = await api.put(`/api/projects/${projectId}`, {
        title: nextFields.title ?? project.title,
        description: nextFields.description ?? project.description ?? "",
      });
      patchLocalProject(projectId, updated);
      setErrorMessage("");
    } catch (error) {
      if (error.status === 405) {
        setErrorMessage("Projekt-Details konnten noch nicht gespeichert werden. Bitte den easy-author-Backend-Prozess einmal neu starten.");
        return;
      }
      setErrorMessage(error.message);
    }
  }

  async function updateBook(bookId, nextFields) {
    const book = currentBook?.id === bookId ? currentBook : projectDetail?.books?.find((entry) => entry.id === bookId);
    if (!book) {
      return;
    }
    try {
      const updated = await api.put(`/api/books/${bookId}`, {
        title: nextFields.title ?? book.title,
        subtitle: nextFields.subtitle ?? book.subtitle ?? "",
        author: nextFields.author ?? book.author ?? "",
        visibility: nextFields.visibility ?? book.visibility ?? "private",
      });
      patchLocalBook(bookId, updated);
      setErrorMessage("");
    } catch (error) {
      if (error.status === 405) {
        setErrorMessage("Buch-Details konnten noch nicht gespeichert werden. Bitte den easy-author-Backend-Prozess einmal neu starten.");
        return;
      }
      setErrorMessage(error.message);
    }
  }

  async function updateWorkflowBox(id, nextFields) {
    const box = bookBundle?.workflow_boxes?.find((entry) => entry.id === id);
    if (!box) {
      return;
    }
    try {
      const updated = await api.put(`/api/workflow-boxes/${id}`, {
        title: nextFields.title ?? box.title,
        type: nextFields.type ?? box.type,
        tags: nextFields.tags ?? box.tags ?? [],
        is_collapsed: nextFields.is_collapsed ?? box.is_collapsed,
      });
      setBookBundle((previous) => ({
        ...previous,
        workflow_boxes: previous.workflow_boxes.map((entry) => (entry.id === updated.id ? updated : entry)),
      }));
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function createAnchor(workflowBoxId = selectedWorkflowBoxId, options = {}) {
    const { promptForNote = true } = options;
    if (!selectedChapterId || !workflowBoxId) {
      setErrorMessage("Bitte zuerst ein Kapitel und eine Workflow-Box waehlen.");
      return;
    }
    const payload =
      editorMode === "markdown" ? getMarkdownSelectionPayload() : editorRef.current?.getSelectionPayload();
    if (!payload?.selected_text) {
      setErrorMessage("Bitte zuerst eine Textpassage im Editor markieren.");
      return;
    }
    const note = promptForNote ? window.prompt("Optionale Notiz fuer diesen Anker", "") || "" : "";
    try {
      await api.post(`/api/chapters/${selectedChapterId}/anchors`, {
        ...payload,
        workflow_box_id: workflowBoxId,
        anchor_type: "passage",
        title: previewText(payload.selected_text, 40),
        note,
      });
      await loadAnchors(selectedChapterId);
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function attachSelectionToNewWorkflowBox() {
    const selectedText = activeSelectionPayload?.selected_text || "";
    const suggestedTags = normalizeTagsLocal(
      detectStoryTimeCues(selectedText).hasTimeCue ? ["zeit", "datum", ...selectedText.split(/\s+/).slice(0, 2)] : selectedText.split(/\s+/).slice(0, 3),
    );
    const created = await createWorkflowBox({
      title: previewText(selectedText || "Neue Box", 32),
      tags: suggestedTags,
    });
    if (created?.id) {
      await createAnchor(created.id, { promptForNote: false });
    }
  }

  async function deleteAnchor(anchorId) {
    try {
      await api.delete(`/api/anchors/${anchorId}`);
      await loadAnchors(selectedChapterId);
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  function openReviewComposerFromSelection(defaultType = "comment") {
    if (editorMode !== "rich") {
      setErrorMessage("Kommentare und Review-Markierungen stehen aktuell im Rich-Editor zur Verfuegung.");
      return;
    }
    const payload = activeSelectionPayload;
    if (!payload?.selected_text) {
      setErrorMessage("Bitte zuerst eine Textpassage im Editor markieren.");
      return;
    }
    setReviewCommentDraft({
      ...emptyReviewCommentDraft(),
      comment_type: defaultType,
      author: "Review",
      selected_text: payload.selected_text,
      start_offset: payload.start_offset,
      end_offset: payload.end_offset,
      context_before: payload.context_before,
      context_after: payload.context_after,
    });
    setActiveReviewCommentId("");
    setReviewBubblePosition(null);
    setShowReviewComposer(true);
  }

  function closeReviewComposer() {
    setShowReviewComposer(false);
    setReviewCommentDraft(emptyReviewCommentDraft());
  }

  function startReviewSuggestionEdit(comment) {
    if (!comment) {
      return;
    }
    setReviewSuggestionDraft(comment.suggested_text || comment.selected_text || "");
    setIsEditingReviewSuggestion(true);
  }

  function cancelReviewSuggestionEdit() {
    setIsEditingReviewSuggestion(false);
    setReviewSuggestionDraft(activeReviewComment?.suggested_text || activeReviewComment?.selected_text || "");
  }

  async function createReviewComment() {
    if (!selectedChapterId) {
      return;
    }
    if (!reviewCommentDraft.selected_text?.trim()) {
      setErrorMessage("Bitte zuerst eine Textstelle fuer den Kommentar markieren.");
      return;
    }
    if (!reviewCommentDraft.body.trim() && !reviewCommentDraft.suggested_text.trim()) {
      setErrorMessage("Bitte einen Kommentartext oder einen Vorschlag eintragen.");
      return;
    }

    try {
      const created = await api.post(`/api/chapters/${selectedChapterId}/comments`, reviewCommentDraft);
      setReviewComments((previous) => [created, ...previous]);
      editorRef.current?.applyReviewCommentMark?.(created);
      setActiveReviewCommentId(created.id);
      setReviewBubblePosition(
        selectionContext
          ? {
              x: selectionContext.x,
              y: selectionContext.y + 22,
            }
          : null,
      );
      closeReviewComposer();
      clearSelectionPopup();
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function updateReviewComment(commentId, nextFields) {
    const comment = reviewComments.find((entry) => entry.id === commentId);
    if (!comment) {
      return null;
    }
    try {
      const updated = await api.put(`/api/comments/${commentId}`, {
        comment_type: nextFields.comment_type ?? comment.comment_type,
        author: nextFields.author ?? comment.author,
        body: nextFields.body ?? comment.body,
        suggested_text: nextFields.suggested_text ?? comment.suggested_text,
        status: nextFields.status ?? comment.status,
        is_todo_done: nextFields.is_todo_done ?? comment.is_todo_done,
      });
      setReviewComments((previous) => previous.map((entry) => (entry.id === updated.id ? updated : entry)));
      setErrorMessage("");
      return updated;
    } catch (error) {
      setErrorMessage(error.message);
      return null;
    }
  }

  async function removeReviewComment(commentId) {
    try {
      await api.delete(`/api/comments/${commentId}`);
      editorRef.current?.removeReviewCommentMark?.(commentId);
      setReviewComments((previous) => previous.filter((entry) => entry.id !== commentId));
      if (activeReviewCommentId === commentId) {
        setActiveReviewCommentId("");
        setReviewBubblePosition(null);
      }
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  function activateReviewComment(commentId, coords = null) {
    if (!commentId) {
      setActiveReviewCommentId("");
      setReviewBubblePosition(null);
      return;
    }
    clearSelectionPopup();
    setActiveReviewCommentId(commentId);
    setReviewBubblePosition(coords);
  }

  async function applyReviewSuggestion(comment, replacementText = comment?.suggested_text || "") {
    if (!comment || !replacementText.trim()) {
      return;
    }
    editorRef.current?.replaceReviewCommentText?.({
      commentId: comment.id,
      text: replacementText.trim(),
      keepMark: false,
    });
    await updateReviewComment(comment.id, {
      status: "applied",
      body: comment.body,
      suggested_text: replacementText.trim(),
      is_todo_done: true,
    });
    editorRef.current?.removeReviewCommentMark?.(comment.id);
    setActiveReviewCommentId("");
    setReviewBubblePosition(null);
    setIsEditingReviewSuggestion(false);
    setReviewSuggestionDraft("");
  }

  async function applyDeleteRequest(comment) {
    if (!comment) {
      return;
    }
    editorRef.current?.replaceReviewCommentText?.({
      commentId: comment.id,
      text: "",
      keepMark: false,
    });
    await updateReviewComment(comment.id, {
      status: "applied",
      body: comment.body,
      suggested_text: comment.suggested_text,
      is_todo_done: true,
    });
    editorRef.current?.removeReviewCommentMark?.(comment.id);
    setActiveReviewCommentId("");
    setReviewBubblePosition(null);
    setIsEditingReviewSuggestion(false);
    setReviewSuggestionDraft("");
  }

  async function resolveReviewComment(comment) {
    if (!comment) {
      return;
    }
    await updateReviewComment(comment.id, {
      status: "resolved",
      body: comment.body,
      suggested_text: comment.suggested_text,
      is_todo_done: true,
    });
    editorRef.current?.removeReviewCommentMark?.(comment.id);
    setActiveReviewCommentId("");
    setReviewBubblePosition(null);
    setIsEditingReviewSuggestion(false);
    setReviewSuggestionDraft("");
  }

  async function rejectReviewComment(comment) {
    if (!comment) {
      return;
    }
    const updated = await updateReviewComment(comment.id, {
      status: "rejected",
      body: comment.body,
      suggested_text: comment.suggested_text,
      is_todo_done: false,
    });
    if (!updated) {
      return;
    }
    editorRef.current?.removeReviewCommentMark?.(comment.id);
    setActiveReviewCommentId("");
    setReviewBubblePosition(null);
    setIsEditingReviewSuggestion(false);
    setReviewSuggestionDraft("");
  }

  async function createClipboardItem() {
    const payload =
      editorMode === "markdown" ? getMarkdownSelectionPayload() : editorRef.current?.getSelectionPayload();
    if (!payload?.selected_text) {
      setErrorMessage("Bitte zuerst eine Textpassage im Editor markieren.");
      return;
    }
    await captureClipboardPayload(payload);
  }

  async function captureClipboardPayload(payload) {
    if (!selectedBookId || !payload?.selected_text) {
      return;
    }
    try {
      const created = await api.post(`/api/books/${selectedBookId}/clipboard`, {
        chapter_id: selectedChapterId,
        content: payload.selected_text,
        content_type: "text/markdown",
        source_anchor_id: "",
        is_pinned: false,
        slot: 0,
      });
      setClipboardItems((previous) => [created, ...previous]);
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  function handleEditorCopy(payload) {
    void captureClipboardPayload(payload);
  }

  async function updateClipboard(item, nextFields) {
    try {
      const updated = await api.put(`/api/clipboard/${item.id}`, {
        content: nextFields.content ?? item.content,
        is_pinned: nextFields.is_pinned ?? item.is_pinned,
        slot: nextFields.slot ?? item.slot,
      });
      setClipboardItems((previous) => previous.map((entry) => (entry.id === updated.id ? updated : entry)));
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  async function deleteClipboard(itemId) {
    try {
      await api.delete(`/api/clipboard/${itemId}`);
      setClipboardItems((previous) => previous.filter((entry) => entry.id !== itemId));
      setErrorMessage("");
    } catch (error) {
      setErrorMessage(error.message);
    }
  }

  function nextFreeClipboardSlot(excludedItemId = "") {
    return (
      slotNumbers.find(
        (slot) => !clipboardItems.some((item) => item.id !== excludedItemId && item.is_pinned && item.slot === slot),
      ) || 1
    );
  }

  async function assignClipboardSlot(item, slot) {
    if (slot <= 0) {
      await updateClipboard(item, {
        is_pinned: false,
        slot: 0,
      });
      return;
    }

    await updateClipboard(item, {
      is_pinned: true,
      slot,
    });
  }

  async function toggleClipboardPinFromPalette(item) {
    if (item.is_pinned && item.slot >= 1) {
      await assignClipboardSlot(item, 0);
      return;
    }

    await assignClipboardSlot(item, item.slot >= 1 ? item.slot : nextFreeClipboardSlot(item.id));
  }

  function getMarkdownSelectionPayload() {
    const textarea = markdownTextareaRef.current;
    if (!textarea) {
      return null;
    }
    return getMarkdownSelectionPayloadFromTarget(textarea);
  }

  function getMarkdownSelectionPayloadFromTarget(textarea) {
    if (!textarea) {
      return null;
    }
    const start = textarea.selectionStart ?? 0;
    const end = textarea.selectionEnd ?? 0;
    if (start === end) {
      return null;
    }
    const content = chapterDraft.markdown_content || "";
    return {
      selected_text: content.slice(start, end),
      start_offset: start,
      end_offset: end,
      context_before: content.slice(Math.max(0, start - 60), start),
      context_after: content.slice(end, Math.min(content.length, end + 60)),
    };
  }

  function updateMarkdownSelectionState(target) {
    const payload = getMarkdownSelectionPayloadFromTarget(target);
    const hasActiveSelection = Boolean(payload?.selected_text);
    setHasSelection(hasActiveSelection);
    if (!hasActiveSelection) {
      clearSelectionPopup();
      return;
    }
    const rect = target.getBoundingClientRect();
    showSelectionPopup({
      kind: "text",
      source: "markdown",
      payload,
      tableActive: false,
      x: rect.left + Math.min(rect.width / 2, 260),
      y: rect.top + 18,
    }, POPUP_HOLD_DELAY_MS);
  }

  function insertIntoMarkdown(content) {
    const textarea = markdownTextareaRef.current;
    if (!textarea) {
      clearSelectionPopup();
      setChapterDraft((previous) => ({
        ...previous,
        markdown_content: `${previous.markdown_content || ""}${content}`,
        editor_json: "",
      }));
      return;
    }
    const start = textarea.selectionStart ?? chapterDraft.markdown_content.length;
    const end = textarea.selectionEnd ?? chapterDraft.markdown_content.length;
    const current = chapterDraft.markdown_content || "";
    const nextValue = `${current.slice(0, start)}${content}${current.slice(end)}`;
    clearSelectionPopup();
    setChapterDraft((previous) => ({
      ...previous,
      markdown_content: nextValue,
      editor_json: "",
    }));
    window.requestAnimationFrame(() => {
      const cursor = start + content.length;
      textarea.focus();
      textarea.setSelectionRange(cursor, cursor);
      setHasSelection(false);
    });
  }

  function insertIntoActiveEditor(content) {
    if (editorMode === "markdown") {
      insertIntoMarkdown(content);
      return;
    }
    editorRef.current?.insertText(content);
  }

  function insertWikiLinkFromSelection() {
    const selectedText = activeSelectionPayload?.selected_text;
    if (!selectedText) {
      return;
    }
    insertIntoActiveEditor(`[[${selectedText.trim()}]]`);
    clearSelectionPopup();
    setHasSelection(false);
  }

  function applyEditorSelectionContext(nextContext) {
    if (!nextContext) {
      clearSelectionPopup();
      return;
    }
    const hasTextSelection = Boolean(nextContext.payload?.selected_text);
    setHasSelection(hasTextSelection);
    const normalizedContext = {
      x: nextContext.x ?? 360,
      y: nextContext.y ?? 160,
      ...nextContext,
    };
    showSelectionPopup(normalizedContext, hasTextSelection ? POPUP_HOLD_DELAY_MS : 0);
  }

  function clipboardSourceLabel(item) {
    if (!item.chapter_id) {
      return "Allgemein";
    }
    return chaptersById.get(item.chapter_id)?.title || "Kapitel";
  }

  function nextFootnoteId(content) {
    const matches = Array.from(String(content || "").matchAll(/\[\^(\d+)\]/g)).map((match) => Number(match[1]));
    return String((matches.length ? Math.max(...matches) : 0) + 1);
  }

  function insertMarkdownFootnote() {
    const noteId = nextFootnoteId(chapterDraft.markdown_content);
    const reference = `[^${noteId}]`;
    const definition = `\n\n[^${noteId}]: `;
    const existing = String(chapterDraft.markdown_content || "");
    const hasDefinition = existing.includes(`[^${noteId}]:`);
    const textarea = markdownTextareaRef.current;
    const start = textarea?.selectionStart ?? existing.length;
    const end = textarea?.selectionEnd ?? existing.length;
    const nextValue = `${existing.slice(0, start)}${reference}${existing.slice(end)}${hasDefinition ? "" : definition}`;
    setChapterDraft((previous) => ({
      ...previous,
      markdown_content: nextValue,
      editor_json: "",
    }));
    window.requestAnimationFrame(() => {
      if (!textarea) {
        return;
      }
      const cursor = start + reference.length;
      textarea.focus();
      textarea.setSelectionRange(cursor, cursor);
    });
  }

  function toggleMarkdownBlockquote() {
    const textarea = markdownTextareaRef.current;
    if (!textarea) {
      insertIntoMarkdown("> ");
      return;
    }
    const content = chapterDraft.markdown_content || "";
    const start = textarea.selectionStart ?? 0;
    const end = textarea.selectionEnd ?? 0;
    const lineStart = content.lastIndexOf("\n", Math.max(0, start - 1)) + 1;
    const lineEndIndex = content.indexOf("\n", end);
    const lineEnd = lineEndIndex === -1 ? content.length : lineEndIndex;
    const segment = content.slice(lineStart, lineEnd);
    const lines = segment.split("\n");
    const alreadyQuoted = lines.every((line) => line.startsWith("> "));
    const nextSegment = lines.map((line) => (alreadyQuoted ? line.replace(/^> /, "") : `> ${line}`)).join("\n");
    const nextValue = `${content.slice(0, lineStart)}${nextSegment}${content.slice(lineEnd)}`;
    setChapterDraft((previous) => ({
      ...previous,
      markdown_content: nextValue,
      editor_json: "",
    }));
  }

  function updateEditorAppearance(field, value) {
    setEditorAppearance((previous) =>
      sanitizeEditorAppearance({
        ...previous,
        [field]: value,
      }),
    );
  }

  function insertTable() {
    if (editorMode === "markdown") {
      insertIntoMarkdown(["| Kopf-Spalte 1 | Kopf-Spalte 2 |", "| --- | --- |", "|  |  |"].join("\n"));
      return;
    }
    editorRef.current?.insertTable?.();
  }

  function toggleQuote() {
    if (editorMode === "markdown") {
      toggleMarkdownBlockquote();
      return;
    }
    editorRef.current?.toggleBlockquote?.();
  }

  function insertFootnote() {
    if (editorMode === "markdown") {
      insertMarkdownFootnote();
      return;
    }
    editorRef.current?.insertFootnote?.();
  }

  function switchEditorMode(nextMode) {
    if (nextMode === editorMode) {
      return;
    }
    if (nextMode === "markdown") {
      const snapshot = editorRef.current?.getDocumentSnapshot?.();
      if (snapshot) {
        setChapterDraft((previous) => ({
          ...previous,
          ...snapshot,
        }));
      }
    }
    setHasSelection(false);
    clearSelectionPopup();
    setActiveReviewCommentId("");
    setReviewBubblePosition(null);
    setEditorMode(nextMode);
  }

  function revealEditorHeader() {
    clearTimeout(editorHeaderHideRef.current);
    setShowEditorHeader(true);
  }

  function hideEditorHeaderSoon() {
    clearTimeout(editorHeaderHideRef.current);
    editorHeaderHideRef.current = window.setTimeout(() => {
      if (!showEditorHelp && !showEditorSettings) {
        setShowEditorHeader(false);
      }
    }, 220);
  }

  function toggleLeftOverlay(section) {
    if (showLeftOverlay && activeLeftSection === section) {
      setShowLeftOverlay(false);
      return;
    }
    setActiveLeftSection(section);
    setShowLeftOverlay(true);
  }

  function toggleRightOverlay(section) {
    if (showRightOverlay && activeRightSection === section) {
      setShowRightOverlay(false);
      return;
    }
    setActiveRightSection(section);
    setShowRightOverlay(true);
  }

  const bookTitle = bookBundle?.book?.title || "Kein Buch geladen";
  const showFocusScrim =
    showLeftOverlay ||
    showRightOverlay ||
    showEditorHelp ||
    showEditorSettings ||
    showReviewComposer ||
    showClipboardPalette ||
    Boolean(activeReviewCommentId);
  const isWidgetFocusActive = showFocusScrim || showEditorHelp || showEditorSettings;
  const showTopRail = !isEditorFullscreen && (showEditorHeader || showEditorHelp || showEditorSettings || showFloatingStatus);
  const leftSectionClass = (section) =>
    `widget-panel-section ${activeLeftSection === section ? "is-active-widget is-selected" : "is-inactive-widget"}`;
  const rightSectionClass = (section) =>
    `widget-panel-section ${activeRightSection === section ? "is-active-widget is-selected" : "is-inactive-widget"}`;

  return (
    <div
      className={`app-shell ${isEditorFullscreen ? "editor-fullscreen-shell" : ""} ${showLeftOverlay ? "has-left-overlay" : ""} ${showRightOverlay ? "has-right-overlay" : ""} ${isWidgetFocusActive ? "has-widget-focus" : ""}`}
    >
      {errorMessage ? <div className="error-banner">{errorMessage}</div> : null}
      {showFocusScrim ? <div className="focus-scrim" aria-hidden="true" onClick={closeTransientPanels} /> : null}

      {!isEditorFullscreen ? (
        <div className="floating-rail floating-rail--left" aria-label="Navigation">
          <button type="button" className={`icon-button ${showLeftOverlay && activeLeftSection === "project" ? "active" : ""}`} aria-label="Projekt" onClick={() => toggleLeftOverlay("project")}>⌂</button>
          <button type="button" className={`icon-button ${showLeftOverlay && activeLeftSection === "book" ? "active" : ""}`} aria-label="Buch" onClick={() => toggleLeftOverlay("book")}>📘</button>
          <button type="button" className={`icon-button ${showLeftOverlay && activeLeftSection === "chapter" ? "active" : ""}`} aria-label="Kapitel" onClick={() => toggleLeftOverlay("chapter")}>☰</button>
          <button type="button" className={`icon-button ${showLeftOverlay && activeLeftSection === "workflow" ? "active" : ""}`} aria-label="Workflow" onClick={() => toggleLeftOverlay("workflow")}>◎</button>
          <button type="button" className={`icon-button ${showLeftOverlay && activeLeftSection === "knowledge" ? "active" : ""}`} aria-label="Wissen" onClick={() => toggleLeftOverlay("knowledge")}>✦</button>
        </div>
      ) : null}

      {!isEditorFullscreen ? (
        <div className={`floating-rail floating-rail--top ${showTopRail ? "is-revealed" : "is-dormant"}`} aria-label="Editor-Steuerung">
          <span className={`floating-status-pill ${showFloatingStatus ? "is-visible" : "is-idle"}`} title={saveState}>
            {saveState}
          </span>
          <button className="icon-button top-icon top-icon--save" type="button" aria-label="Kapitel speichern" onClick={() => saveChapter(true)} disabled={!currentChapter}>💾</button>
          <button type="button" className={`icon-button top-icon top-icon--mode ${editorMode === "rich" ? "active" : ""}`} aria-label="Rich" onClick={() => switchEditorMode("rich")}>✍</button>
          <button type="button" className={`icon-button top-icon top-icon--mode ${editorMode === "markdown" ? "active" : ""}`} aria-label="Markdown" onClick={() => switchEditorMode("markdown")}>#</button>
          <button type="button" className={`icon-button top-icon top-icon--utility ${showWritingTools ? "active" : ""}`} aria-label={showWritingTools ? "Werkzeuge ausblenden" : "Werkzeuge"} onClick={() => setShowWritingTools((previous) => !previous)}>✚</button>
          <button type="button" className={`icon-button top-icon top-icon--utility ${showEditorHelp ? "active" : ""}`} aria-label={showEditorHelp ? "Hilfe ausblenden" : "Hilfe"} onClick={() => setShowEditorHelp((previous) => !previous)}>?</button>
          <button type="button" className={`icon-button top-icon top-icon--utility ${showEditorSettings ? "active" : ""}`} aria-label={showEditorSettings ? "⚙ Einstellungen ausblenden" : "⚙ Einstellungen"} onClick={() => setShowEditorSettings((previous) => !previous)}>⚙</button>
          <button type="button" className={`icon-button top-icon top-icon--focus ${isEditorFullscreen ? "active" : ""}`} aria-label={isEditorFullscreen ? "Vollbild aus" : "Vollbild"} aria-pressed={isEditorFullscreen} onClick={() => setIsEditorFullscreen((previous) => !previous)}>⛶</button>
        </div>
      ) : null}

      {!isEditorFullscreen ? (
        <div className="floating-rail floating-rail--right" aria-label="Kontext">
          <button type="button" className={`icon-button ${showRightOverlay && activeRightSection === "context" ? "active" : ""}`} aria-label="Wiki-Links" onClick={() => toggleRightOverlay("context")}>🔗</button>
          <button type="button" className={`icon-button ${showRightOverlay && activeRightSection === "anchor" ? "active" : ""}`} aria-label="Anker" onClick={() => toggleRightOverlay("anchor")}>⚓</button>
          <button type="button" className={`icon-button ${showRightOverlay && activeRightSection === "clipboard" ? "active" : ""}`} aria-label="Clipboard" onClick={() => toggleRightOverlay("clipboard")}>📋</button>
        </div>
      ) : null}

      <main
        className={`workspace-grid ${isEditorFullscreen ? "editor-fullscreen" : ""}`}
        data-fullscreen-backdrop={editorAppearance.fullscreenBackdrop}
      >
        <aside className={`workspace-panel left-panel ${showLeftOverlay ? "is-open" : ""}`}>
          <div className="overlay-panel-header">
            <strong>{activeLeftSection === "project" ? "Projekt" : activeLeftSection === "book" ? "Buch" : activeLeftSection === "chapter" ? "Kapitel" : activeLeftSection === "workflow" ? "Workflow" : "Wissen"}</strong>
            <button type="button" className="icon-button" aria-label="Navigation schließen" onClick={() => setShowLeftOverlay(false)}>×</button>
          </div>
          <SidebarSection sectionRef={projectSectionRef} className={leftSectionClass("project")} eyebrow="Projekt" title="Arbeitsraum" actionLabel="+ Projekt" onAction={createProject}>
            {currentProject ? (
              <>
                <button
                  type="button"
                  className={`pill-button active picker-toggle-button ${showProjectPicker ? "is-open" : ""}`}
                  onClick={() => setShowProjectPicker((previous) => !previous)}
                >
                  <div>
                    <strong>{currentProject.title}</strong>
                    <span>{currentProject.description || "Ohne Beschreibung"}</span>
                  </div>
                  <small>{showProjectPicker ? "Auswahl ausblenden" : "Projekt wechseln"}</small>
                </button>
                {showProjectPicker ? (
                  <div className="pill-list compact-picker-list">
                    {projects
                      .filter((project) => project.id !== selectedProjectId)
                      .map((project) => (
                        <button
                          key={project.id}
                          type="button"
                          className="pill-button"
                          onClick={() => {
                            setSelectedProjectId(project.id);
                            setShowProjectPicker(false);
                          }}
                        >
                          <strong>{project.title}</strong>
                          <span>{project.description || "Ohne Beschreibung"}</span>
                        </button>
                      ))}
                  </div>
                ) : null}
                <div className="detail-toggle-row">
                  <button
                    type="button"
                    className="ghost-button"
                    aria-label="Projektdetails"
                    onClick={() => setShowProjectDetails((previous) => !previous)}
                  >
                    {showProjectDetails ? "Details schließen" : "Details"}
                  </button>
                </div>
                {showProjectDetails ? (
                  <div className="book-meta-card">
                    <div className="context-card-header">
                      <strong>Projektdetails</strong>
                      <span className="knowledge-chip">aktiv</span>
                    </div>
                    {showProjectEdit ? (
                      <>
                        <div className="detail-card-toolbar">
                          <button type="button" className="ghost-button" onClick={() => setShowProjectEdit(false)}>
                            Fertig
                          </button>
                        </div>
                        <label className="editor-setting">
                          <span>Titel</span>
                          <input
                            value={currentProject.title || ""}
                            onChange={(event) => patchLocalProject(currentProject.id, { title: event.target.value })}
                            onBlur={(event) => updateProject(currentProject.id, { title: event.target.value })}
                          />
                        </label>
                        <label className="editor-setting">
                          <span>Projektbeschreibung</span>
                          <textarea
                            rows="3"
                            value={currentProject.description || ""}
                            placeholder="Worum geht es in diesem Projekt?"
                            onChange={(event) => patchLocalProject(currentProject.id, { description: event.target.value })}
                            onBlur={(event) => updateProject(currentProject.id, { description: event.target.value })}
                          />
                        </label>
                      </>
                    ) : (
                      <div className="detail-card-summary">
                        <p>{currentProject.description || "Ohne Beschreibung"}</p>
                        <button type="button" className="ghost-button" onClick={() => setShowProjectEdit(true)}>
                          Projekt ändern
                        </button>
                      </div>
                    )}
                  </div>
                ) : null}
              </>
            ) : (
              <p className="empty-note">Noch kein Projekt ausgewaehlt.</p>
            )}
          </SidebarSection>

          <SidebarSection sectionRef={bookSectionRef} className={leftSectionClass("book")} eyebrow="Buch" title={projectDetail?.project?.title || "Noch kein Projekt"} actionLabel="+ Buch" onAction={createBook}>
            {currentBook ? (
              <>
                <button
                  type="button"
                  className={`book-card active picker-toggle-button ${showBookPicker ? "is-open" : ""}`}
                  onClick={() => setShowBookPicker((previous) => !previous)}
                >
                  <div>
                    <strong>{currentBook.title}</strong>
                    <span>{currentBook.subtitle || "Ohne Beschreibung"}</span>
                  </div>
                  <small>{showBookPicker ? "Auswahl ausblenden" : "Buch wechseln"}</small>
                </button>
                {showBookPicker ? (
                  <div className="book-stack compact-picker-list">
                    {(projectDetail?.books || [])
                      .filter((book) => book.id !== selectedBookId)
                      .map((book) => (
                        <button
                          key={book.id}
                          type="button"
                          className="book-card"
                          onClick={() => {
                            setSelectedBookId(book.id);
                            setShowBookPicker(false);
                          }}
                        >
                          <strong>{book.title}</strong>
                          <span>{book.subtitle || "Ohne Beschreibung"}</span>
                          <small>{book.visibility}</small>
                        </button>
                      ))}
                  </div>
                ) : null}
                <div className="detail-toggle-row">
                  <button
                    type="button"
                    className="ghost-button"
                    aria-label="Buchdetails"
                    onClick={() => setShowBookDetails((previous) => !previous)}
                  >
                    {showBookDetails ? "Details schließen" : "Details"}
                  </button>
                </div>
                {showBookDetails ? (
                  <div className="book-meta-card">
                    <div className="context-card-header">
                      <strong>Buchdetails</strong>
                      <span className="knowledge-chip">{currentBook.visibility}</span>
                    </div>
                    {showBookEdit ? (
                      <>
                        <div className="detail-card-toolbar">
                          <button type="button" className="ghost-button" onClick={() => setShowBookEdit(false)}>
                            Fertig
                          </button>
                        </div>
                        <label className="editor-setting">
                          <span>Titel</span>
                          <input
                            value={currentBook.title || ""}
                            onChange={(event) => patchLocalBook(currentBook.id, { title: event.target.value })}
                            onBlur={(event) => updateBook(currentBook.id, { title: event.target.value })}
                          />
                        </label>
                        <label className="editor-setting">
                          <span>Beschreibung</span>
                          <textarea
                            rows="3"
                            value={currentBook.subtitle || ""}
                            placeholder="Kurzbeschreibung oder Positionierung des Buchs"
                            onChange={(event) => patchLocalBook(currentBook.id, { subtitle: event.target.value })}
                            onBlur={(event) => updateBook(currentBook.id, { subtitle: event.target.value })}
                          />
                        </label>
                        <div className="book-meta-grid">
                          <label className="editor-setting">
                            <span>Autor</span>
                            <input
                              value={currentBook.author || ""}
                              placeholder="Autor oder Arbeitstitel"
                              onChange={(event) => patchLocalBook(currentBook.id, { author: event.target.value })}
                              onBlur={(event) => updateBook(currentBook.id, { author: event.target.value })}
                            />
                          </label>
                          <label className="editor-setting">
                            <span>Sichtbarkeit</span>
                            <select
                              value={currentBook.visibility || "private"}
                              onChange={(event) => {
                                patchLocalBook(currentBook.id, { visibility: event.target.value });
                                void updateBook(currentBook.id, { visibility: event.target.value });
                              }}
                            >
                              <option value="private">private</option>
                              <option value="shared">shared</option>
                              <option value="public">public</option>
                            </select>
                          </label>
                        </div>
                      </>
                    ) : (
                      <div className="detail-card-summary">
                        <p>{currentBook.subtitle || "Ohne Beschreibung"}</p>
                        <div className="detail-card-meta">
                          <span className="knowledge-chip">{currentBook.author || "ohne Autor"}</span>
                          <span className="knowledge-chip">{currentBook.visibility || "private"}</span>
                        </div>
                        <button type="button" className="ghost-button" onClick={() => setShowBookEdit(true)}>
                          Buch ändern
                        </button>
                      </div>
                    )}
                  </div>
                ) : null}
              </>
            ) : (
              <p className="empty-note">Noch kein Buch ausgewaehlt.</p>
            )}
          </SidebarSection>

          <SidebarSection sectionRef={chapterSectionRef} className={leftSectionClass("chapter")} eyebrow="Kapitel" title={bookTitle} actionLabel="+ Kapitel" onAction={createChapter}>
            <div className="chapter-list">
              {(bookBundle?.chapters || []).map((chapter) => (
                <button
                  key={chapter.id}
                  type="button"
                  draggable
                  className={`chapter-row ${chapter.id === selectedChapterId ? "active" : ""} ${chapter.id === draggedChapterId ? "dragging" : ""} ${chapter.id === chapterDropTargetId ? "drop-target" : ""}`}
                  onClick={() => setSelectedChapterId(chapter.id)}
                  onDragStart={() => {
                    setDraggedChapterId(chapter.id);
                    setChapterDropTargetId(chapter.id);
                  }}
                  onDragOver={(event) => {
                    event.preventDefault();
                    if (chapterDropTargetId !== chapter.id) {
                      setChapterDropTargetId(chapter.id);
                    }
                  }}
                  onDrop={async (event) => {
                    event.preventDefault();
                    const sourceChapterId = draggedChapterId;
                    setDraggedChapterId("");
                    setChapterDropTargetId("");
                    await reorderChapters(sourceChapterId, chapter.id);
                  }}
                  onDragEnd={() => {
                    setDraggedChapterId("");
                    setChapterDropTargetId("");
                  }}
                >
                  <span className="chapter-drag-handle" aria-hidden="true">
                    ⋮⋮
                  </span>
                  <span className="chapter-index">{String(chapter.position).padStart(2, "0")}</span>
                  <span>{chapter.title}</span>
                </button>
              ))}
            </div>
          </SidebarSection>

          <SidebarSection sectionRef={workflowSectionRef} className={leftSectionClass("workflow")} eyebrow="Workflow" title="Workflow-Boxen" actionLabel="+ Box" onAction={() => createWorkflowBox()}>
            {showWorkflowSuggestionCloud ? (
              <div className="workflow-suggestion-cloud">
                {workflowSuggestions.map(({ box, score, reasons }) => (
                  <button
                    key={box.id}
                    type="button"
                    className={`cloud-chip ${effectiveWorkflowBoxId === box.id ? "active" : ""}`}
                    onClick={() => setSelectedWorkflowBoxId(box.id)}
                    title={reasons?.length ? `Ausgeloest durch: ${reasons.join(", ")}` : "Vorgeschlagene Workflow-Box"}
                  >
                    <span>{box.title}</span>
                    <small>{reasons?.[0] || (score >= 4 ? "stark passend" : "vorgeschlagen")}</small>
                  </button>
                ))}
              </div>
            ) : null}
            {activeWorkflowBox ? (
              <div className={`workflow-target-card ${showWorkflowTargetDetails ? "is-expanded" : "is-compact"} tone-${workflowActivationById.get(activeWorkflowBox.id)?.tone || "selected"}`}>
                <div className="context-card-header">
                  <strong>Zielbox aktiv</strong>
                  <span className="knowledge-chip">{activeWorkflowBox.type}</span>
                </div>
                <p>{activeWorkflowBox.title}</p>
                <div className="workflow-target-meta">
                  <span className={`workflow-status-chip tone-${workflowActivationById.get(activeWorkflowBox.id)?.tone || "selected"}`}>
                    {workflowActivationById.get(activeWorkflowBox.id)?.label || "Ziel"}
                  </span>
                  {workflowActivationById.get(activeWorkflowBox.id)?.comboReason ? (
                    <span className="knowledge-chip">Kombi aktiv</span>
                  ) : null}
                  <span className="knowledge-chip">
                    {anchorCountByWorkflowBox.get(activeWorkflowBox.id) || 0} Anker im aktuellen Kapitel
                  </span>
                  <span className="knowledge-chip">
                    {activeWorkflowAnchors.length > 0 ? "bereits verbunden" : "bereit fuer erste Passage"}
                  </span>
                </div>
                <p className="workflow-activation-note">
                  {workflowActivationById.get(activeWorkflowBox.id)?.reason || "Diese Box ist als Ziel gesetzt."}
                </p>
                {showWorkflowTargetDetails ? (
                  <>
                    <small className="workflow-target-hint">{workflowTypeMeta(activeWorkflowBox.type).hint}</small>
                    <div className="workflow-target-actions">
                      <button
                        type="button"
                        className="secondary-button"
                        onClick={() => createAnchor(activeWorkflowBox.id, { promptForNote: false })}
                        disabled={!hasSelection || !selectedChapterId}
                      >
                        Auswahl ankern
                      </button>
                      <button type="button" className="ghost-button" onClick={() => setShowEditorHelp(true)}>
                        Workflow-Hilfe
                      </button>
                    </div>
                    <div className="workflow-anchor-list">
                      {activeWorkflowAnchors.length === 0 ? (
                        <p className="empty-note">Noch keine Passagen in dieser Box. Markiere Text und verankere ihn direkt hier.</p>
                      ) : (
                        activeWorkflowAnchors.slice(-3).reverse().map((anchor) => (
                          <article key={anchor.id} className="workflow-anchor-card">
                            <strong>{anchor.title || "Passage"}</strong>
                            <p>{previewText(anchor.selected_text, 120)}</p>
                            <small>{anchor.note || "ohne Notiz"}</small>
                          </article>
                        ))
                      )}
                    </div>
                  </>
                ) : null}
              </div>
            ) : null}
            <div className="workflow-list">
              {(bookBundle?.workflow_boxes || []).map((box) => {
                const activation = workflowActivationById.get(box.id) || {
                  tone: "idle",
                  label: "Ruhend",
                  reason: "Noch keine aktiven Signale",
                };
                return (
                  <div
                    key={box.id}
                    className={`workflow-card ${box.id === selectedWorkflowBoxId ? "active" : ""} tone-${activation.tone}`}
                    onClick={() => setSelectedWorkflowBoxId(box.id)}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(event) => {
                      if (event.key === "Enter" || event.key === " ") {
                        setSelectedWorkflowBoxId(box.id);
                      }
                    }}
                  >
                  <div className="workflow-card-header">
                    <div>
                      <strong>{workflowTypeMeta(box.type).label}</strong>
                      <small>{anchorCountByWorkflowBox.get(box.id) || 0} Anker im Kapitel</small>
                    </div>
                    <div className="workflow-card-actions">
                      <span className={`workflow-status-chip tone-${activation.tone}`}>{activation.label}</span>
                      {selectedWorkflowBoxId === box.id ? <span className="knowledge-chip">Ziel</span> : null}
                      <button
                        type="button"
                        className="ghost-button"
                        onClick={(event) => {
                          event.stopPropagation();
                          setSelectedWorkflowBoxId(box.id);
                        }}
                      >
                        {selectedWorkflowBoxId === box.id ? "aktiv" : "als Ziel"}
                      </button>
                    </div>
                  </div>
                  <input
                    className="workflow-title-input"
                    value={box.title}
                    onChange={(event) =>
                      setBookBundle((previous) => ({
                        ...previous,
                        workflow_boxes: previous.workflow_boxes.map((entry) =>
                          entry.id === box.id ? { ...entry, title: event.target.value } : entry,
                        ),
                      }))
                    }
                    onBlur={(event) => updateWorkflowBox(box.id, { title: event.target.value })}
                  />
                  <input
                    className="workflow-tag-input"
                    value={formatTagInput(box.tags)}
                    placeholder="Trigger-Tags, komma-getrennt"
                    onChange={(event) =>
                      setBookBundle((previous) => ({
                        ...previous,
                        workflow_boxes: previous.workflow_boxes.map((entry) =>
                          entry.id === box.id ? { ...entry, tags: splitTagInput(event.target.value) } : entry,
                        ),
                      }))
                    }
                    onBlur={(event) => updateWorkflowBox(box.id, { tags: splitTagInput(event.target.value) })}
                  />
                  <p className="workflow-activation-note">{activation.reason}</p>
                  {!box.is_collapsed ? (
                    <>
                      <div className="workflow-row">
                        <select value={box.type} onChange={(event) => updateWorkflowBox(box.id, { type: event.target.value })}>
                          <option value="notes">notes</option>
                          <option value="persons">persons</option>
                          <option value="events">events</option>
                          <option value="threads">threads</option>
                          <option value="reminders">reminders</option>
                          <option value="research">research</option>
                          <option value="clipboard">clipboard</option>
                          <option value="custom">custom</option>
                        </select>
                        <label className="checkbox-row">
                          <input
                            type="checkbox"
                            checked={box.is_collapsed}
                            onChange={(event) => updateWorkflowBox(box.id, { is_collapsed: event.target.checked })}
                          />
                          collapsed
                        </label>
                        <span className="knowledge-chip">{anchorCountByWorkflowBox.get(box.id) || 0} Anker</span>
                      </div>
                      <p className="workflow-card-hint">{workflowTypeMeta(box.type).hint}</p>
                      <div className="workflow-row">
                        <button
                          type="button"
                          className="secondary-button"
                          onClick={(event) => {
                            event.stopPropagation();
                            setSelectedWorkflowBoxId(box.id);
                            void createAnchor(box.id, { promptForNote: false });
                          }}
                          disabled={!hasSelection || !selectedChapterId}
                        >
                          Auswahl ankern
                        </button>
                        {latestAnchorByWorkflowBox.get(box.id) ? (
                          <span className="knowledge-chip">
                            zuletzt: {previewText(latestAnchorByWorkflowBox.get(box.id).selected_text, 28)}
                          </span>
                        ) : (
                          <span className="knowledge-chip">noch leer</span>
                        )}
                      </div>
                    </>
                  ) : (
                    <div className="workflow-row">
                      <label className="checkbox-row">
                        <input
                          type="checkbox"
                          checked={box.is_collapsed}
                          onChange={(event) => updateWorkflowBox(box.id, { is_collapsed: event.target.checked })}
                        />
                        collapsed
                      </label>
                      <span className="knowledge-chip">{workflowTypeMeta(box.type).label}</span>
                      {latestAnchorByWorkflowBox.get(box.id) ? (
                        <span className="knowledge-chip">
                          {previewText(latestAnchorByWorkflowBox.get(box.id).selected_text, 24)}
                        </span>
                      ) : null}
                    </div>
                  )}
                  </div>
                );
              })}
            </div>
          </SidebarSection>

          <SidebarSection sectionRef={knowledgeSectionRef} className={leftSectionClass("knowledge")} eyebrow="Wissen" title="Wissensbank" actionLabel="+ Eintrag" onAction={createKnowledgeItem}>
            <div className="knowledge-list compact">
              <input
                value={knowledgeQuery}
                placeholder="Begriffe filtern"
                onChange={(event) => setKnowledgeQuery(event.target.value)}
              />
              {knowledgeItems.length === 0 ? <p className="empty-note">Noch keine Wissenseintraege vorhanden.</p> : null}
              <div className="chip-cloud">
                {filteredKnowledgeItems.map((item) => {
                  const linkedInChapter = chapterKnowledgeMatches.some(
                    (entry) => entry.reference.key === normalizeKnowledgeKey(item.type, item.name),
                  );
                  return (
                    <button
                      key={item.id}
                      type="button"
                      className={`cloud-chip ${selectedKnowledgeItem?.id === item.id ? "active" : ""} ${linkedInChapter ? "linked" : ""}`}
                      onClick={() => setSelectedKnowledgeItemId(item.id)}
                    >
                      <span>{item.name}</span>
                      {linkedInChapter ? <small>im Kapitel</small> : null}
                    </button>
                  );
                })}
              </div>
              {selectedKnowledgeItem ? (
                <article className={`knowledge-card ${chapterKnowledgeMatches.some((entry) => entry.reference.key === normalizeKnowledgeKey(selectedKnowledgeItem.type, selectedKnowledgeItem.name)) ? "linked" : ""}`}>
                  <div className="context-card-header">
                    <strong>{knowledgeTypeLabel(selectedKnowledgeItem.type)}</strong>
                    <button type="button" className="ghost-button" onClick={() => insertIntoActiveEditor(knowledgeReference(selectedKnowledgeItem))}>
                      Link
                    </button>
                  </div>
                  <input
                    value={selectedKnowledgeItem.name}
                    onChange={(event) =>
                      setKnowledgeItems((previous) =>
                        previous.map((entry) =>
                          entry.id === selectedKnowledgeItem.id ? { ...entry, name: event.target.value } : entry,
                        ),
                      )
                    }
                    onBlur={(event) => updateKnowledgeItem(selectedKnowledgeItem, { name: event.target.value })}
                  />
                  <div className="workflow-row">
                    <select
                      value={selectedKnowledgeItem.type}
                      onChange={(event) => updateKnowledgeItem(selectedKnowledgeItem, { type: event.target.value })}
                    >
                      <option value="person">person</option>
                      <option value="location">location</option>
                      <option value="event">event</option>
                      <option value="thread">thread</option>
                      <option value="motif">motif</option>
                      <option value="term">term</option>
                      <option value="reminder">reminder</option>
                      <option value="research_note">research_note</option>
                      <option value="custom">custom</option>
                    </select>
                    <span className="knowledge-chip">{knowledgeReference(selectedKnowledgeItem)}</span>
                  </div>
                  <textarea
                    rows="2"
                    value={selectedKnowledgeItem.summary || ""}
                    placeholder="Kurze Zusammenfassung"
                    onChange={(event) =>
                      setKnowledgeItems((previous) =>
                        previous.map((entry) =>
                          entry.id === selectedKnowledgeItem.id ? { ...entry, summary: event.target.value } : entry,
                        ),
                      )
                    }
                    onBlur={(event) => updateKnowledgeItem(selectedKnowledgeItem, { summary: event.target.value })}
                  />
                  <input
                    value={formatTagInput(selectedKnowledgeItem.tags)}
                    placeholder="Tags, komma-getrennt"
                    onChange={(event) =>
                      setKnowledgeItems((previous) =>
                        previous.map((entry) =>
                          entry.id === selectedKnowledgeItem.id
                            ? { ...entry, tags: splitTagInput(event.target.value) }
                            : entry,
                        ),
                      )
                    }
                    onBlur={(event) => updateKnowledgeItem(selectedKnowledgeItem, { tags: splitTagInput(event.target.value) })}
                  />
                </article>
              ) : null}
            </div>
          </SidebarSection>
        </aside>

        <section className="editor-panel">
          {!isEditorFullscreen ? (
            <div
              className="editor-top-hover-zone"
              aria-hidden="true"
              onMouseEnter={revealEditorHeader}
              onMouseLeave={hideEditorHeaderSoon}
            />
          ) : null}
          <div
            className={`editor-header-shell ${!isEditorFullscreen && (showEditorHeader || showEditorHelp || showEditorSettings) ? "is-visible" : ""}`}
            onMouseEnter={revealEditorHeader}
            onMouseLeave={hideEditorHeaderSoon}
            onFocusCapture={revealEditorHeader}
            onBlurCapture={(event) => {
              if (!event.currentTarget.contains(event.relatedTarget)) {
                hideEditorHeaderSoon();
              }
            }}
          >
            <div className="editor-header">
              <div className="chapter-header-stack">
                <input
                  className="chapter-title-input"
                  value={chapterDraft.title}
                  onChange={(event) => setChapterDraft((previous) => ({ ...previous, title: event.target.value }))}
                  placeholder="Kapitelueberschrift"
                  disabled={!currentChapter}
                />
                <input
                  className="chapter-summary-input"
                  value={chapterDraft.summary}
                  onChange={(event) => setChapterDraft((previous) => ({ ...previous, summary: event.target.value }))}
                  placeholder="Kapitel-Merker oder Kurzbeschreibung fuer deinen Schreibfluss"
                  disabled={!currentChapter}
                />
              </div>
              <div className="editor-actions">
                <div className={`editor-action-group editor-action-group--context ${hasSelection ? "is-awake" : "is-muted"}`}>
                  <button type="button" className="secondary-button icon-button" aria-label="Anker setzen" onClick={() => createAnchor()} disabled={!hasSelection}>
                    ⚓
                  </button>
                  <button type="button" className="secondary-button icon-button" aria-label="In Clipboard uebernehmen" onClick={createClipboardItem} disabled={!hasSelection}>
                    📋
                  </button>
                </div>
                {showWritingTools ? (
                  <div className="editor-action-group editor-action-group--tools is-awake">
                    <button type="button" className="secondary-button subtle-button icon-button" aria-label="Tabelle" onClick={insertTable} disabled={!currentChapter}>
                      ▦
                    </button>
                    <button type="button" className="secondary-button subtle-button icon-button" aria-label="Zitat" onClick={toggleQuote} disabled={!currentChapter}>
                      ❝
                    </button>
                    <button type="button" className="secondary-button subtle-button icon-button" aria-label="Fussnote" onClick={insertFootnote} disabled={!currentChapter}>
                      †
                    </button>
                  </div>
                ) : null}
              </div>
            </div>
            {editorMetaItems.length > 0 ? (
              <div className="editor-meta">
                {editorMetaItems.map((item) => (
                  <span key={item} className="editor-meta-chip is-subtle">{item}</span>
                ))}
              </div>
            ) : null}
          </div>

          {selectionContext ? (
            <div
              className={`selection-context-popup ${selectionPopupVisible ? "is-visible" : ""}`}
              role="dialog"
              aria-label="Auswahl-Aktionen"
              aria-hidden={!selectionPopupVisible}
              style={{
                left: `${selectionContext.x}px`,
                top: `${selectionContext.y}px`,
              }}
            >
              <div className="selection-context-header">
                <strong>{selectionContext.tableActive && !selectionContext.payload?.selected_text ? "Tabellen-Kontext" : "Auswahl"}</strong>
                <small>
                  {selectionContext.payload?.selected_text
                    ? previewText(selectionContext.payload.selected_text, 42)
                    : "Werkzeuge fuer die aktuelle Tabelle"}
                </small>
                {primaryWorkflowSuggestion?.reasons?.length && selectionContext.payload?.selected_text ? (
                  <small className="selection-context-trigger">
                    Vorschlag: {primaryWorkflowSuggestion.box.title} · {primaryWorkflowSuggestion.reasons.join(" · ")}
                  </small>
                ) : null}
                {activeSelectionCombinations.length > 0 ? (
                  <small className="selection-context-trigger">
                    Kombi: {activeSelectionCombinations.map((entry) => entry.label).join(" + ")}
                  </small>
                ) : null}
              </div>
              <div className="selection-context-actions">
                {selectionContext.payload?.selected_text ? (
                  <>
                    <button type="button" className="secondary-button" onClick={() => createAnchor(effectiveWorkflowBoxId || selectedWorkflowBoxId, { promptForNote: true })}>
                      {hasTemporaryAutoTarget && activeWorkflowBox ? `Anker · ${activeWorkflowBox.title}` : "Anker"}
                    </button>
                    {editorMode === "rich" ? (
                      <button type="button" className="secondary-button" onClick={() => openReviewComposerFromSelection("comment")}>
                        Kommentar
                      </button>
                    ) : null}
                    <button type="button" className="secondary-button" onClick={createClipboardItem}>
                      Clipboard
                    </button>
                    <button type="button" className="ghost-button" onClick={insertWikiLinkFromSelection}>
                      Wiki-Link
                    </button>
                  </>
                ) : null}
                {workflowSuggestions.slice(0, 2).map(({ box, reasons }, index) => (
                  <button
                    key={box.id}
                    type="button"
                    className={index === 0 ? "secondary-button" : "ghost-button"}
                    title={reasons?.length ? `Ausgeloest durch: ${reasons.join(", ")}` : undefined}
                    onClick={() => {
                      setSelectedWorkflowBoxId(box.id);
                      void createAnchor(box.id, { promptForNote: false });
                    }}
                    disabled={!selectionContext.payload?.selected_text}
                  >
                    Zu {box.title}
                  </button>
                ))}
                <button
                  type="button"
                  className="ghost-button"
                  onClick={() => void attachSelectionToNewWorkflowBox()}
                  disabled={!selectionContext.payload?.selected_text}
                >
                  Neue Box
                </button>
                {selectionContext.tableActive ? (
                  <>
                    <button type="button" className="ghost-button" onClick={() => editorRef.current?.addColumnAfter()}>
                      + Spalte
                    </button>
                    <button type="button" className="ghost-button" onClick={() => editorRef.current?.addRowAfter()}>
                      + Zeile
                    </button>
                  </>
                ) : null}
              </div>
            </div>
          ) : null}

          <div
            className={`editor-frame surface-${editorAppearance.surfacePreset}`}
            style={editorSurfaceStyle}
            data-surface={editorAppearance.surfacePreset}
            data-fullscreen-backdrop={editorAppearance.fullscreenBackdrop}
          >
            <div className="editor-frame-inner">
              {editorMode === "markdown" ? (
                <textarea
                  ref={markdownTextareaRef}
                  className="markdown-textarea"
                  value={chapterDraft.markdown_content}
                  placeholder="Schreibe hier direkt in Markdown. Wiki-Links wie [[Mara]] oder [[Ort:Alter Garten]] bleiben erhalten."
                  onFocus={() => {
                    setHasSelection(false);
                    clearSelectionPopup();
                  }}
                  onSelect={(event) => {
                    updateMarkdownSelectionState(event.target);
                  }}
                  onKeyUp={(event) => {
                    updateMarkdownSelectionState(event.target);
                  }}
                  onClick={(event) => {
                    updateMarkdownSelectionState(event.target);
                  }}
                  onCopy={() => handleEditorCopy(getMarkdownSelectionPayload())}
                  onCut={() => handleEditorCopy(getMarkdownSelectionPayload())}
                  onChange={(event) =>
                    setChapterDraft((previous) => ({
                      ...previous,
                      markdown_content: event.target.value,
                      editor_json: "",
                    }))
                  }
                />
              ) : (
                <EditorPane
                  ref={editorRef}
                  chapter={currentChapter ? { ...currentChapter, ...chapterDraft } : null}
                  pinnedSlots={pinnedSlots}
                  activeReviewCommentId={activeReviewCommentId}
                  onSelectionChange={setHasSelection}
                  onSelectionContextChange={applyEditorSelectionContext}
                  onReviewCommentActivate={activateReviewComment}
                  onClipboardCapture={handleEditorCopy}
                  onDocumentChange={(nextDocument) =>
                    setChapterDraft((previous) => ({
                      ...previous,
                      ...nextDocument,
                    }))
                  }
                />
              )}
            </div>
          </div>
        </section>

        <aside className={`workspace-panel right-panel ${showRightOverlay ? "is-open" : ""}`}>
          <div className="overlay-panel-header">
            <strong>{activeRightSection === "context" ? "Wiki-Links" : activeRightSection === "anchor" ? "Anker" : "Clipboard"}</strong>
            <button type="button" className="icon-button" aria-label="Kontext schließen" onClick={() => setShowRightOverlay(false)}>×</button>
          </div>
          <SidebarSection sectionRef={contextSectionRef} className={rightSectionClass("context")} eyebrow="Kontext" title="Wiki-Links im Kapitel">
            <div className="knowledge-context-list compact">
              {chapterKnowledgeMatches.length === 0 ? (
                <p className="empty-note">Noch keine `[[...]]`-Referenzen im aktuellen Kapitel.</p>
              ) : null}
              <div className="chip-cloud">
                {chapterKnowledgeMatches.map((entry) => (
                  <button
                    key={entry.reference.key}
                    type="button"
                    className={`cloud-chip ${selectedKnowledgeRef?.reference.key === entry.reference.key ? "active" : ""} ${entry.item ? "linked" : "unresolved"}`}
                    onClick={() => setSelectedKnowledgeRefKey(entry.reference.key)}
                  >
                    <span>{entry.item ? entry.item.name : entry.reference.raw}</span>
                  </button>
                ))}
              </div>
              {selectedKnowledgeRef ? (
                <article className={`context-card ${selectedKnowledgeRef.item ? "" : "unresolved"}`}>
                  <div className="context-card-header">
                    <strong>{selectedKnowledgeRef.item ? selectedKnowledgeRef.item.name : selectedKnowledgeRef.reference.raw}</strong>
                    <span className="knowledge-chip">
                      {knowledgeTypeLabel(selectedKnowledgeRef.item?.type || selectedKnowledgeRef.reference.type)}
                    </span>
                  </div>
                  <p>{selectedKnowledgeRef.item?.summary || "Noch kein passender Wissenseintrag vorhanden."}</p>
                  <small>{selectedKnowledgeRef.item ? knowledgeReference(selectedKnowledgeRef.item) : `[[${selectedKnowledgeRef.reference.raw}]]`}</small>
                </article>
              ) : null}
              {unresolvedKnowledgeRefs.length > 0 ? (
                <p className="empty-note">Offene Referenzen koennen links als Wissenseintrag angelegt oder umbenannt werden.</p>
              ) : null}
            </div>
          </SidebarSection>

          <SidebarSection sectionRef={anchorSectionRef} className={rightSectionClass("anchor")} eyebrow="Anker" title="Aktuelle Textstelle">
            <div className="anchor-list">
              {anchors.length === 0 ? <p className="empty-note">Noch keine Anker fuer dieses Kapitel.</p> : null}
              {anchors.map((anchor) => (
                <article key={anchor.id} className="context-card">
                  <div className="context-card-header">
                    <strong>{anchor.title || "Anker"}</strong>
                    <button type="button" className="ghost-button" onClick={() => deleteAnchor(anchor.id)}>
                      loeschen
                    </button>
                  </div>
                  <p>{anchor.selected_text}</p>
                  <small>
                    {anchor.anchor_type} | {anchor.note || "ohne Notiz"}
                  </small>
                </article>
              ))}
            </div>
          </SidebarSection>

          <SidebarSection sectionRef={clipboardSectionRef} className={rightSectionClass("clipboard")} eyebrow="Workflow" title="Clipboard & Slots">
            <div className="clipboard-summary">
              <span>
                {clipboardItems.length} Eintraege · {pinnedSlots.length} Slots fixiert
              </span>
              <button
                type="button"
                className="ghost-button"
                onClick={() => setShowClipboardPalette(true)}
                disabled={clipboardItems.length === 0}
              >
                Floating-Liste
              </button>
            </div>

            <div className="slot-grid">
              {slotNumbers.map((slot) => {
                const item = pinnedSlots.find((entry) => entry.slot === slot);
                return (
                  <div key={slot} className={`slot-card ${item ? "filled" : ""}`}>
                    <strong>{slot}</strong>
                    <span>{item ? previewText(item.content, 36) : "leer"}</span>
                    {item ? (
                      <>
                        <small>{clipboardSourceLabel(item)}</small>
                        <button
                          type="button"
                          className="ghost-button"
                          aria-label={`Slot ${slot} einfuegen`}
                          onClick={() => insertIntoActiveEditor(item.content)}
                        >
                          einfuegen
                        </button>
                      </>
                    ) : (
                      <small>per Klick oder Shortcut belegbar</small>
                    )}
                  </div>
                );
              })}
            </div>

            {latestClipboardItems.length > 0 ? (
              <div className="clipboard-recent-row">
                {latestClipboardItems.map((item) => (
                  <button
                    key={item.id}
                    type="button"
                    className="cloud-chip"
                    onClick={() => insertIntoActiveEditor(item.content)}
                  >
                    <span>{previewText(item.content, 22)}</span>
                    <small>{clipboardSourceLabel(item)}</small>
                  </button>
                ))}
              </div>
            ) : null}

            <div className="clipboard-list">
              {clipboardItems.length === 0 ? <p className="empty-note">Noch keine Clipboard-Eintraege vorhanden.</p> : null}
              {clipboardItems.map((item) => (
                <article key={item.id} className="context-card">
                  <div className="context-card-header">
                    <strong>{previewText(item.content, 32)}</strong>
                    <div className="workflow-row">
                      <span className="knowledge-chip">{clipboardSourceLabel(item)}</span>
                      <button type="button" className="ghost-button" onClick={() => deleteClipboard(item.id)}>
                        loeschen
                      </button>
                    </div>
                  </div>
                  <p>{previewText(item.content, 120)}</p>
                  <div className="clipboard-controls">
                    <label className="checkbox-row">
                      <input
                        type="checkbox"
                        checked={item.is_pinned}
                        onChange={(event) => updateClipboard(item, { is_pinned: event.target.checked })}
                      />
                      anpinnen
                    </label>
                    <label className="slot-picker">
                      Slot
                      <input
                        type="number"
                        min="0"
                        max="9"
                        value={item.slot}
                        onChange={(event) => updateClipboard(item, { slot: Number(event.target.value) })}
                      />
                    </label>
                    <button type="button" className="ghost-button" onClick={() => insertIntoActiveEditor(item.content)}>
                      einfuegen
                    </button>
                    <button
                      type="button"
                      className="ghost-button"
                      onClick={() => assignClipboardSlot(item, nextFreeClipboardSlot(item.id))}
                    >
                      naechster Slot
                    </button>
                  </div>
                </article>
              ))}
            </div>
          </SidebarSection>
        </aside>
      </main>

      {showEditorHelp ? (
        <section className="editor-help" role="dialog" aria-label="Editor-Hilfe">
          <div className="editor-help-header">
            <div>
              <div className="panel-eyebrow">Editor-Hilfe</div>
              <strong>MVP-Referenz fuer Schreiben, Workflow und Einfuegen</strong>
            </div>
            <button type="button" className="ghost-button" onClick={() => setShowEditorHelp(false)}>
              Schliessen
            </button>
          </div>
          <div className="editor-help-grid">
            <article className="editor-help-card">
              <strong>Markdown-Modus</strong>
              <p>
                Aktuell voll unterstuetzt sind Ueberschriften, Listen, verschachtelte Listen, Zitate, Code-Fences,
                Trennlinien, harte Umbrueche, Escaping, Inline-Code, Fett, Kursiv, Durchgestrichen,
                `[[...]]`-Wiki-Links und einfache Pipe-Tabellen.
              </p>
            </article>
            <article className="editor-help-card">
              <strong>Rich-Editor</strong>
              <p>
                Der Tiptap-Teil verarbeitet die gespeicherten Strukturen und erkennt typische Schreibmuster wie
                Ueberschriften, Listen, Zitate, Code-Bloecke, Trennlinien und Tabellen. Abgetippte Pipe-Tabellen
                wie `| Kopf | Kopf |` werden beim Weiterschreiben automatisch in eine Rich-Tabelle uebernommen.
              </p>
            </article>
            <article className="editor-help-card">
              <strong>Clipboard & Slots</strong>
              <p>
                Markiere Text und uebernimm ihn ueber `In Clipboard uebernehmen`. Rechts kannst du Eintraege
                anpinnen, Slots `1-9` zuweisen und ueber `einfuegen` an der Cursorposition einsetzen. Im Rich-Editor
                gehen gepinnte Slots auch per `Cmd/Ctrl + Shift + 1-9`. Die Floating-Liste sammelt alle Snippets
                an einer Stelle und erlaubt schnelle Slot-Zuordnung ohne den Schreibfluss zu verlassen. Gefuellte
                Slot-Karten lassen sich auch direkt per Klick wieder einfuegen.
              </p>
            </article>
            <article className="editor-help-card">
              <strong>Workflow-Anker</strong>
              <p>
                Waehle links zuerst eine Workflow-Box. Die aktive Box erscheint als `Zielbox`. Wenn du danach Text
                markierst und `Anker setzen` klickst, wird die Passage genau dieser Workflow-Box zugeordnet. Im
                Workflow-Cockpit kannst du markierte Passagen auch direkt ohne Zusatzdialog verankern.
              </p>
            </article>
            <article className="editor-help-card">
              <strong>Tabellen-Werkzeuge</strong>
              <p>
                `Tabelle` legt schnell eine Grundstruktur an. Sobald der Cursor in einer Tabelle steht, erscheint
                direkt am Editor eine kontextuelle Table-Bar fuer Spalten, Zeilen, Kopfzeile und Loeschen.
              </p>
            </article>
            <article className="editor-help-card">
              <strong>Zitate & Fussnoten</strong>
              <p>
                `Zitat` schaltet Blockquotes um. `Fussnote` erzeugt im Rich-Editor eine Referenz samt Notizblock und
                im Markdown-Modus eine `[^1]`-Referenz mit passender Definition am Kapitelende.
              </p>
            </article>
          </div>
        </section>
      ) : null}

      {showEditorSettings ? (
        <section className="editor-settings" role="dialog" aria-label="Editor-Einstellungen">
          <div className="editor-help-header">
            <div>
              <div className="panel-eyebrow">Einstellungen</div>
              <strong>Look and Feel fuer konzentriertes Schreiben</strong>
            </div>
            <button type="button" className="ghost-button" onClick={() => setShowEditorSettings(false)}>
              Schliessen
            </button>
          </div>
          <div className="editor-settings-grid">
            <label className="editor-setting">
              <span>Schriftfamilie</span>
              <select value={editorAppearance.fontFamily} onChange={(event) => updateEditorAppearance("fontFamily", event.target.value)}>
                <option value="serif">Serif</option>
                <option value="sans">Sans</option>
                <option value="mono">Mono</option>
                <option value="google">Google Font</option>
              </select>
            </label>
            {editorAppearance.fontFamily === "google" ? (
              <>
                <label className="editor-setting">
                  <span>Google Font Preset</span>
                  <select
                    value={GOOGLE_FONT_PRESETS.includes(editorAppearance.googleFontName) ? editorAppearance.googleFontName : "__custom__"}
                    onChange={(event) => {
                      if (event.target.value !== "__custom__") {
                        updateEditorAppearance("googleFontName", event.target.value);
                      }
                    }}
                  >
                    {GOOGLE_FONT_PRESETS.map((fontName) => (
                      <option key={fontName} value={fontName}>
                        {fontName}
                      </option>
                    ))}
                    <option value="__custom__">Eigener Fontname</option>
                  </select>
                </label>
                <label className="editor-setting">
                  <span>Google Font Name</span>
                  <input
                    value={editorAppearance.googleFontName}
                    placeholder="z. B. EB Garamond"
                    onChange={(event) => updateEditorAppearance("googleFontName", event.target.value)}
                  />
                </label>
              </>
            ) : null}
            <label className="editor-setting">
              <span>Schriftgroesse</span>
              <input
                type="range"
                min="16"
                max="24"
                step="1"
                value={editorAppearance.fontSize}
                onChange={(event) => updateEditorAppearance("fontSize", Number(event.target.value))}
              />
            </label>
            <label className="editor-setting">
              <span>Zeilenhoehe</span>
              <input
                type="range"
                min="1.5"
                max="2.2"
                step="0.05"
                value={editorAppearance.lineHeight}
                onChange={(event) => updateEditorAppearance("lineHeight", Number(event.target.value))}
              />
            </label>
            <label className="editor-setting">
              <span>Textbreite</span>
              <select
                value={editorAppearance.contentWidth}
                onChange={(event) => updateEditorAppearance("contentWidth", Number(event.target.value))}
              >
                <option value="640">Sehr schmal</option>
                <option value="720">Schmal</option>
                <option value="860">Standard</option>
                <option value="960">Komfort</option>
                <option value="1040">Breit</option>
                <option value="1160">Sehr breit</option>
              </select>
            </label>
            <label className="editor-setting">
              <span>Vollbild-Breite</span>
              <select
                value={editorAppearance.fullscreenContentWidth}
                onChange={(event) => updateEditorAppearance("fullscreenContentWidth", Number(event.target.value))}
              >
                <option value="860">Kompakt</option>
                <option value="1040">Standard</option>
                <option value="1200">Breit</option>
                <option value="1360">Sehr breit</option>
              </select>
            </label>
            <label className="editor-setting">
              <span>Vollbild-Hintergrund</span>
              <select
                value={editorAppearance.fullscreenBackdrop}
                onChange={(event) => updateEditorAppearance("fullscreenBackdrop", event.target.value)}
              >
                <option value="linen">Leinen</option>
                <option value="paper">Papier</option>
                <option value="dusk">Dusk</option>
                <option value="night">Nacht</option>
              </select>
            </label>
            <label className="editor-setting">
              <span>Farbprofil</span>
              <select
                value={editorAppearance.surfacePreset}
                onChange={(event) => updateEditorAppearance("surfacePreset", event.target.value)}
              >
                <option value="warm">Warm</option>
                <option value="paper">Papier</option>
                <option value="night">Nacht</option>
              </select>
            </label>
            <label className="editor-setting">
              <span>Caret-Farbe</span>
              <input
                type="color"
                value={editorAppearance.caretColor}
                onChange={(event) => updateEditorAppearance("caretColor", event.target.value)}
              />
            </label>
          </div>
        </section>
      ) : null}

      {showReviewComposer ? (
        <section className="review-comment-composer" role="dialog" aria-label="Kommentar verfassen">
          <div className="editor-help-header">
            <div>
              <div className="panel-eyebrow">Review</div>
              <strong>Kommentar zur markierten Passage</strong>
            </div>
            <button type="button" className="ghost-button" onClick={closeReviewComposer}>
              Schliessen
            </button>
          </div>
          <div className="review-comment-composer-grid">
            <label className="editor-setting">
              <span>Typ</span>
              <select
                value={reviewCommentDraft.comment_type}
                onChange={(event) =>
                  setReviewCommentDraft((previous) => ({
                    ...previous,
                    comment_type: event.target.value,
                  }))
                }
              >
                {Object.entries(REVIEW_COMMENT_TYPE_META).map(([type, meta]) => (
                  <option key={type} value={type}>
                    {meta.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="editor-setting">
              <span>Autor</span>
              <input
                value={reviewCommentDraft.author}
                onChange={(event) =>
                  setReviewCommentDraft((previous) => ({
                    ...previous,
                    author: event.target.value,
                  }))
                }
              />
            </label>
            <label className="editor-setting review-comment-composer-span">
              <span>Markierte Passage</span>
              <textarea rows="3" value={reviewCommentDraft.selected_text} readOnly />
            </label>
            <label className="editor-setting review-comment-composer-span">
              <span>{reviewCommentTypeMeta(reviewCommentDraft.comment_type).label}</span>
              <textarea
                rows="4"
                value={reviewCommentDraft.body}
                placeholder={reviewCommentTypeMeta(reviewCommentDraft.comment_type).hint}
                onChange={(event) =>
                  setReviewCommentDraft((previous) => ({
                    ...previous,
                    body: event.target.value,
                  }))
                }
              />
            </label>
            {reviewCommentDraft.comment_type === "suggestion" ? (
              <label className="editor-setting review-comment-composer-span">
                <span>Ersatztext</span>
                <textarea
                  rows="4"
                  value={reviewCommentDraft.suggested_text}
                  placeholder='z. B. "Der neue Satzbau ist nun viel einfacher strukturiert und leichter lesbar."'
                  onChange={(event) =>
                    setReviewCommentDraft((previous) => ({
                      ...previous,
                      suggested_text: event.target.value,
                    }))
                  }
                />
              </label>
            ) : null}
            <div className="review-comment-composer-actions">
              <button type="button" className="secondary-button" onClick={createReviewComment}>
                Kommentar anlegen
              </button>
              <button type="button" className="ghost-button" onClick={closeReviewComposer}>
                Abbrechen
              </button>
            </div>
          </div>
        </section>
      ) : null}

      {activeReviewComment && reviewBubblePosition ? (
        <section
          className={`review-comment-bubble review-comment-bubble--${activeReviewComment.comment_type || "comment"}`}
          role="dialog"
          aria-label="Kommentar"
          style={{
            left: `${reviewBubblePosition.x}px`,
            top: `${reviewBubblePosition.y}px`,
          }}
        >
          <div className="review-comment-bubble-header">
            <div>
              <div className={`panel-eyebrow review-comment-type-pill review-comment-type-pill--${activeReviewComment.comment_type || "comment"}`}>
                {reviewCommentTypeMeta(activeReviewComment.comment_type).label}
              </div>
              <strong>{activeReviewComment.author || "Review"}</strong>
            </div>
            <button type="button" className="ghost-button" onClick={() => activateReviewComment("")}>
              ×
            </button>
          </div>
          <small className="review-comment-bubble-meta">
            {formatReviewTimestamp(activeReviewComment.created_at)}
            {activeReviewComment.status !== "open" ? ` · ${activeReviewComment.status}` : ""}
          </small>
          <div className="review-comment-quote">{activeReviewComment.selected_text}</div>
          {activeReviewComment.body ? <p>{activeReviewComment.body}</p> : null}
          {activeReviewComment.suggested_text ? (
            <div className="review-comment-suggestion">
              <strong>Vorschlag</strong>
              {isEditingReviewSuggestion ? (
                <div className="review-comment-suggestion-editor">
                  <textarea
                    rows="5"
                    value={reviewSuggestionDraft}
                    onChange={(event) => setReviewSuggestionDraft(event.target.value)}
                  />
                  <div className="review-comment-suggestion-editor-actions">
                    <button
                      type="button"
                      className="secondary-button"
                      onClick={() => applyReviewSuggestion(activeReviewComment, reviewSuggestionDraft)}
                    >
                      Uebernehmen
                    </button>
                    <button type="button" className="ghost-button" onClick={cancelReviewSuggestionEdit}>
                      Abbrechen
                    </button>
                  </div>
                </div>
              ) : (
                <p>{activeReviewComment.suggested_text}</p>
              )}
            </div>
          ) : null}
          <div className="review-comment-bubble-actions">
            {activeReviewComment.comment_type === "suggestion" ? (
              <>
                <button
                  type="button"
                  className="secondary-button"
                  onClick={() => applyReviewSuggestion(activeReviewComment, activeReviewComment.suggested_text)}
                >
                  Uebernehmen
                </button>
                <button
                  type="button"
                  className="ghost-button"
                  onClick={() => startReviewSuggestionEdit(activeReviewComment)}
                >
                  Anpassen
                </button>
                <button
                  type="button"
                  className="ghost-button"
                  onClick={() => rejectReviewComment(activeReviewComment)}
                >
                  Ablehnen
                </button>
              </>
            ) : null}
            {activeReviewComment.comment_type === "delete_request" ? (
              <>
                <button type="button" className="secondary-button" onClick={() => applyDeleteRequest(activeReviewComment)}>
                  Loeschen
                </button>
                <button
                  type="button"
                  className="ghost-button"
                  onClick={() => rejectReviewComment(activeReviewComment)}
                >
                  Ablehnen
                </button>
              </>
            ) : null}
            {["comment", "todo", "warning"].includes(activeReviewComment.comment_type) ? (
              <button type="button" className="secondary-button" onClick={() => resolveReviewComment(activeReviewComment)}>
                {activeReviewComment.comment_type === "todo" ? "Erledigt" : "Loesen"}
              </button>
            ) : null}
            <button type="button" className="ghost-button" onClick={() => removeReviewComment(activeReviewComment.id)}>
              Kommentar loeschen
            </button>
          </div>
        </section>
      ) : null}

      {clipboardItems.length > 0 && !isEditorFullscreen ? (
        <button
          type="button"
          className={`clipboard-fab ${showClipboardPalette ? "active" : ""}`}
          aria-expanded={showClipboardPalette}
          aria-controls="clipboard-palette"
          aria-label="Clipboard-Floating-Liste"
          onClick={() => setShowClipboardPalette((previous) => !previous)}
        >
          Clipboard · {clipboardItems.length}
        </button>
      ) : null}

      {showClipboardPalette && !isEditorFullscreen ? (
        <section id="clipboard-palette" className="clipboard-palette" role="dialog" aria-label="Clipboard-Liste">
          <div className="clipboard-palette-header">
            <div>
              <div className="panel-eyebrow">Floating Clipboard</div>
              <strong>Gesammelte Ausschnitte und feste Slots</strong>
            </div>
            <button type="button" className="ghost-button" onClick={() => setShowClipboardPalette(false)}>
              Schliessen
            </button>
          </div>

          <div className="clipboard-palette-list">
            {clipboardItems.map((item) => (
              <article
                key={item.id}
                className={`context-card clipboard-palette-card ${item.is_pinned && item.slot >= 1 ? "pinned" : ""}`}
              >
                <div className="context-card-header">
                  <strong>{previewText(item.content, 36)}</strong>
                  <span className="knowledge-chip">{item.is_pinned && item.slot >= 1 ? `Slot ${item.slot}` : "frei"}</span>
                </div>
                <p>{previewText(item.content, 180)}</p>
                <div className="clipboard-controls">
                  <button type="button" className="secondary-button" onClick={() => insertIntoActiveEditor(item.content)}>
                    einfuegen
                  </button>
                  <button type="button" className="ghost-button" onClick={() => toggleClipboardPinFromPalette(item)}>
                    {item.is_pinned && item.slot >= 1 ? "Slot loesen" : "fixieren"}
                  </button>
                  <button type="button" className="ghost-button" onClick={() => deleteClipboard(item.id)}>
                    loeschen
                  </button>
                </div>
                <small>{clipboardSourceLabel(item)}</small>
                <div className="clipboard-slot-pills" aria-label={`Slot-Zuordnung fuer ${previewText(item.content, 24)}`}>
                  {slotNumbers.map((slot) => (
                    <button
                      key={slot}
                      type="button"
                      className={`slot-pill ${item.slot === slot ? "active" : ""}`}
                      aria-label={`Slot ${slot} zuweisen`}
                      onClick={() => assignClipboardSlot(item, slot)}
                    >
                      {slot}
                    </button>
                  ))}
                  <button
                    type="button"
                    className={`slot-pill clear ${item.slot === 0 ? "active" : ""}`}
                    aria-label="Slot freigeben"
                    onClick={() => assignClipboardSlot(item, 0)}
                  >
                    frei
                  </button>
                </div>
              </article>
            ))}
          </div>
        </section>
      ) : null}
    </div>
  );
}

export default App;
