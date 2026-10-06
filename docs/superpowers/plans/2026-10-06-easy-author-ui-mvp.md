# EasyAuthor UI MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the quiet EasyAuthor writing shell, per-book work views and typography, durable contextual annotations, and book/global Kanban through package 5 of the approved design.

**Architecture:** Keep React/Tiptap as the editor surface and Go/SQLite as the canonical persistence layer. Extract focused UI components and preference helpers from the current monolithic `App.jsx`, then introduce generic anchors, threaded comments, and work items behind explicit REST interfaces so manuscript, margin contexts, and Kanban render the same state. Preserve existing chapter, workflow, clipboard, revision, and Markdown behavior while migrating legacy review comments into the shared model.

**Tech Stack:** Go 1.22+, `net/http`, modernc SQLite, React 18, Tiptap 2, Vite 5, Vitest 4, Testing Library, CSS.

**Spec:** `docs/superpowers/specs/2026-10-06-easy-author-ui-reader-design.md`

## Global Constraints

- Implement packages 1 through 5 only; packages 6 and 7 remain prepared interfaces.
- Global appearance supports `light`, `dark`, and `system`; book changes never change it.
- Work views are exactly `clean`, `intense`, and `review`; a session override never rewrites the book default.
- The transient control bar hides after 60 seconds of inactivity and never hides while focused, hovered, or owning an open menu/dialog.
- Book typography inherits unset values from global defaults and reset removes overrides instead of copying defaults.
- Durable text contexts include comment, task/reminder, link, clipboard insertion, and clipboard source.
- Kanban phases are exactly `backlog`, `todo`, `in_progress`, `review`, and `done`.
- Book comparison colors are transient selection state and are never persisted on a book.
- No EasyReader fetching, source revision storage, image conversion, or browser extension code is part of this plan.
- All schema changes must migrate an existing SQLite database without deleting user data.

## Review Focus

- A user typing continuously or using a menu at the 60-second boundary must not lose controls or focus; Task 2 pins timer pause and reset behavior.
- Invalid or missing book typography overrides must fall back field-by-field without poisoning global preferences; Task 4 pins normalization and reset.
- Multiple contexts sharing one passage must remain independently deletable and keyboard reachable; Task 6 pins grouping and partial deletion.
- Moving a card and updating its linked thread must be one transaction so reload cannot show divergent phases; Task 8 pins rollback and synchronization.
- Selecting, deselecting, and reselecting several books must preserve existing color assignments during the selection and avoid persistence; Task 9 pins allocation stability.

---

## Planned File Structure

- `frontend/src/lib/uiPreferences.js`: validated global appearance and session work-view persistence.
- `frontend/src/hooks/useTransientControls.js`: 60-second visibility state machine.
- `frontend/src/components/TransientControlBar.jsx`: top-edge trigger, control bar, menus, and status.
- `frontend/src/components/WorkViewPicker.jsx`: reusable three-card first-run and switch dialog.
- `frontend/src/components/TypographySettings.jsx`: inherited global/book typography editor.
- `frontend/src/components/ContextRail.jsx`: grouped, keyboard-accessible margin contexts.
- `frontend/src/components/CommentThread.jsx`: messenger-style thread surface.
- `frontend/src/components/KanbanBoard.jsx`: five-column board and keyboard/drag movement.
- `frontend/src/components/BookKanbanFilter.jsx`: multi-book selection, counters, and transient colors.
- `frontend/src/lib/bookColorAllocation.js`: deterministic maximum-distance palette allocation for the active selection.
- `backend/internal/model/model.go`: book presentation, document anchor, comment message, and work-item contracts.
- `backend/internal/store/store.go`: additive schema migration and transactional persistence.
- `backend/internal/httpapi/preferences.go`: book presentation endpoints.
- `backend/internal/httpapi/contexts.go`: anchors, threads, messages, and context endpoints.
- `backend/internal/httpapi/workitems.go`: book/global Kanban query and phase mutation endpoints.
- Existing `frontend/src/App.jsx`, `frontend/src/components/EditorPane.jsx`, `frontend/src/styles.css`, `frontend/src/lib/api.js`, `backend/internal/httpapi/app.go`, and their tests integrate these units.

### Task 1: Lock Current Behavior and Extract UI Preferences

**Files:**
- Create: `apps/easy-author/frontend/src/lib/uiPreferences.js`
- Create: `apps/easy-author/frontend/src/lib/uiPreferences.test.js`
- Modify: `apps/easy-author/frontend/src/App.jsx`
- Test: `apps/easy-author/frontend/src/App.test.jsx`

**Interfaces:**
- Produces: `loadGlobalAppearance(storage) -> GlobalAppearance`, `saveGlobalAppearance(storage, value)`, `loadSessionWorkView(storage) -> "clean"|"intense"|"review"|null`, `saveSessionWorkView(storage, value)`.
- Consumes: existing `DEFAULT_EDITOR_APPEARANCE` values and local-storage behavior.

- [ ] **Step 1: Add failing preference tests** for absent storage, malformed JSON, unsupported theme/view values, and storage exceptions; assert safe defaults and no thrown error.
- [ ] **Step 2: Run `npm run check:node && npx vitest run src/lib/uiPreferences.test.js` from `apps/easy-author/frontend`** and verify failure because the module does not exist.
- [ ] **Step 3: Implement the four preference functions** and move the current appearance/work-mode parsing constants from `App.jsx` into the focused module; map legacy `write`, `structure`, `review` values to `clean`, `intense`, `review` once on read.
- [ ] **Step 4: Update `App.jsx` and its fixtures** to consume the new functions without changing rendered behavior.
- [ ] **Step 5: Run `npm test` and `npm run build`**; expect all existing and new tests to pass and Vite to exit 0.
- [ ] **Step 6: Commit** with `git commit -m "refactor: isolate EasyAuthor UI preferences"`.

### Task 2: Transient Quiet Control Shell

**Files:**
- Create: `apps/easy-author/frontend/src/hooks/useTransientControls.js`
- Create: `apps/easy-author/frontend/src/hooks/useTransientControls.test.jsx`
- Create: `apps/easy-author/frontend/src/components/TransientControlBar.jsx`
- Create: `apps/easy-author/frontend/src/components/TransientControlBar.test.jsx`
- Modify: `apps/easy-author/frontend/src/App.jsx`
- Modify: `apps/easy-author/frontend/src/styles.css`

**Interfaces:**
- Produces: `useTransientControls({ timeoutMs: 60000, blocked }) -> { visible, reveal, hide, onPointerEnter, onPointerLeave, onFocusCapture, onBlurCapture, resetTimer }` and `TransientControlBar` props `{ book, chapter, workView, saveState, appearance, blocked, onWorkView, onAppearance, onSettings, onCommand }`.
- Consumes: Task 1 appearance and work-view types.

- [ ] **Step 1: Write failing fake-timer hook tests** asserting top-edge reveal, explicit ellipsis reveal, hide at exactly 60 seconds, timer reset on interaction, and no hide while hovered, focused, or `blocked`.
- [ ] **Step 2: Run `npm run check:node && npx vitest run src/hooks/useTransientControls.test.jsx src/components/TransientControlBar.test.jsx`** and verify failure from missing hook/component.
- [ ] **Step 3: Implement the visibility hook** with one owned timeout, cleanup on unmount, and no timer-driven focus changes.
- [ ] **Step 4: Implement the ellipsis trigger and control bar** with left book/chapter, center view, and right command/appearance/settings controls; use native buttons, labeled menus, Escape close, and keyboard focus entry.
- [ ] **Step 5: Replace permanent top/editor chrome in `App.jsx`** while retaining save errors and making any open menu/dialog set `blocked=true`.
- [ ] **Step 6: Add responsive and reduced-motion CSS** for hidden, revealing, visible, and narrow layouts.
- [ ] **Step 7: Run `npm test` and `npm run build`**; expect 0 failures and exit 0.
- [ ] **Step 8: Commit** with `git commit -m "feat: add quiet transient writing controls"`.

### Task 3: Persist Per-Book Work View

**Files:**
- Modify: `apps/easy-author/backend/internal/model/model.go`
- Modify: `apps/easy-author/backend/internal/store/store.go`
- Create: `apps/easy-author/backend/internal/httpapi/preferences.go`
- Modify: `apps/easy-author/backend/internal/httpapi/app.go`
- Modify: `apps/easy-author/backend/internal/store/store_test.go`
- Modify: `apps/easy-author/backend/internal/httpapi/app_test.go`
- Create: `apps/easy-author/frontend/src/components/WorkViewPicker.jsx`
- Create: `apps/easy-author/frontend/src/components/WorkViewPicker.test.jsx`
- Modify: `apps/easy-author/frontend/src/App.jsx`

**Interfaces:**
- Produces: Go `BookPresentation{BookID string, DefaultWorkView string, TypographyOverrides json.RawMessage}`; `Store.GetBookPresentation(ctx, bookID)`; `Store.UpdateBookPresentation(ctx, value)`; `GET/PUT /api/books/{id}/presentation`.
- Consumes: Task 1 work-view values and Task 2 control bar callback.

- [ ] **Step 1: Add failing store/API tests** asserting existing books migrate with `clean`, only the three allowed values persist, invalid values return 400, and an update survives store reopen.
- [ ] **Step 2: Run `go test ./internal/store ./internal/httpapi` from `apps/easy-author/backend`** and verify the new tests fail.
- [ ] **Step 3: Add the additive `book_presentations` schema and store methods** without modifying existing book rows or content.
- [ ] **Step 4: Add typed GET/PUT handlers and routes**; return a normalized default document for a legacy book without a row.
- [ ] **Step 5: Add failing picker tests** asserting the three exact cards, initial creation prompt, saved book default, session-only switch, and no accidental default write.
- [ ] **Step 6: Implement `WorkViewPicker` and integrate it** into book creation and the transient control bar; translate the labels to Clean & Free, Intense, and Review.
- [ ] **Step 7: Run backend tests, `npm test`, and `npm run build`**; expect all pass.
- [ ] **Step 8: Commit** with `git commit -m "feat: add per-book EasyAuthor work views"`.

### Task 4: Inherited Global and Book Typography

**Files:**
- Create: `apps/easy-author/frontend/src/lib/typography.js`
- Create: `apps/easy-author/frontend/src/lib/typography.test.js`
- Create: `apps/easy-author/frontend/src/components/TypographySettings.jsx`
- Create: `apps/easy-author/frontend/src/components/TypographySettings.test.jsx`
- Modify: `apps/easy-author/frontend/src/App.jsx`
- Modify: `apps/easy-author/frontend/src/styles.css`
- Modify: `apps/easy-author/backend/internal/httpapi/app_test.go`

**Interfaces:**
- Produces: `resolveTypography(globalDefaults, bookOverrides) -> ResolvedTypography`, `normalizeTypographyOverrides(value) -> TypographyOverrides`, `TypographySettings` callbacks `{ onGlobalChange, onBookChange, onResetBook }`.
- Consumes: Task 3 `TypographyOverrides` JSON and PUT endpoint.

- [ ] **Step 1: Add failing resolver tests** for per-field inheritance, distinct H1–H6/body/quote/table sizes, width, first-line indent, line height, paragraph spacing, invalid numeric/string values, and reset to empty overrides.
- [ ] **Step 2: Run `npm run check:node && npx vitest run src/lib/typography.test.js`** and verify failure because the module is absent.
- [ ] **Step 3: Implement normalization and resolution** with exact safe ranges documented in exported constants and no mutation of defaults or overrides.
- [ ] **Step 4: Add failing component tests** for global editing, book override badges, inherited-value display, and `Auf Gesamtstandard zurücksetzen` sending `{}`.
- [ ] **Step 5: Implement settings and CSS custom properties** for editor headings, body, blockquote, tables, width, indentation, line height, and paragraph spacing.
- [ ] **Step 6: Extend API tests** to reject malformed typography JSON and round-trip valid partial overrides.
- [ ] **Step 7: Run backend tests, `npm test`, and `npm run build`**; expect all pass.
- [ ] **Step 8: Commit** with `git commit -m "feat: add inherited book typography"`.

### Task 5: Shared Durable Context Persistence

**Files:**
- Modify: `apps/easy-author/backend/internal/model/model.go`
- Modify: `apps/easy-author/backend/internal/store/store.go`
- Create: `apps/easy-author/backend/internal/httpapi/contexts.go`
- Modify: `apps/easy-author/backend/internal/httpapi/app.go`
- Modify: `apps/easy-author/backend/internal/store/store_test.go`
- Modify: `apps/easy-author/backend/internal/httpapi/app_test.go`

**Interfaces:**
- Produces: `DocumentAnchor`, `AnchorContext`, `CommentThread`, and `CommentMessage`; store methods `ListChapterContexts`, `CreateContext`, `DeleteContext`, `CreateThreadMessage`, `UpdateThreadStatus`; REST under `/api/chapters/{chapterId}/contexts`, `/api/contexts/{id}`, `/api/threads/{id}` and `/api/threads/{id}/messages`.
- Consumes: existing anchor and review-comment records for additive migration.

- [ ] **Step 1: Write failing migration/store tests** for one anchor with several independent contexts, legacy anchor/comment preservation, thread replies, gray resolved status, and deleting one context without deleting siblings.
- [ ] **Step 2: Run backend store tests** and verify the new cases fail.
- [ ] **Step 3: Add normalized tables** for document anchors, anchor contexts, comment threads, and messages; migration must be idempotent and preserve legacy IDs where referenced.
- [ ] **Step 4: Implement store transactions and validation** for context types `comment`, `work_item`, `link`, `clipboard_insert`, and `clipboard_source`.
- [ ] **Step 5: Add failing HTTP tests** for list/create/delete, thread reply ordering, invalid chapter/context IDs, and keyboard-client-compatible stable ordering.
- [ ] **Step 6: Implement handlers and routes** with 400 for invalid types/statuses, 404 for missing parents, and 409 when an anchor cannot safely be changed.
- [ ] **Step 7: Run `go test ./...`**; expect 0 failures.
- [ ] **Step 8: Commit** with `git commit -m "feat: add durable EasyAuthor text contexts"`.

### Task 6: Editor Context Marks, Rail, and Threads

**Files:**
- Create: `apps/easy-author/frontend/src/extensions/DocumentContext.js`
- Create: `apps/easy-author/frontend/src/components/ContextRail.jsx`
- Create: `apps/easy-author/frontend/src/components/ContextRail.test.jsx`
- Create: `apps/easy-author/frontend/src/components/CommentThread.jsx`
- Create: `apps/easy-author/frontend/src/components/CommentThread.test.jsx`
- Modify: `apps/easy-author/frontend/src/components/EditorPane.jsx`
- Modify: `apps/easy-author/frontend/src/components/EditorPane.test.jsx`
- Modify: `apps/easy-author/frontend/src/App.jsx`
- Modify: `apps/easy-author/frontend/src/styles.css`

**Interfaces:**
- Produces: Tiptap mark attributes `{ anchorId, contextTypes }`; editor commands `applyDocumentContext`, `removeDocumentContext`, `focusDocumentAnchor`; grouped rail model `groupContextsByAnchor(contexts)`.
- Consumes: Task 5 context REST objects.

- [ ] **Step 1: Write failing editor tests** for persistent marks after reload, yellow Markdown highlights, multiple context types on one mark, resolved gray rendering, focus by anchor, and partial context deletion leaving the mark until its last context is removed.
- [ ] **Step 2: Run targeted editor tests** and verify failure.
- [ ] **Step 3: Implement the generic Tiptap context mark and the independent Markdown highlight mark/commands** while reading legacy review-comment marks during migration; expose both actions through the existing selection popup.
- [ ] **Step 4: Write failing rail/thread tests** for grouped icons, counts, tooltip/accessible names, keyboard activation, reply ordering, and context routing.
- [ ] **Step 5: Implement the right-margin rail and messenger thread**; collision grouping must not change document width or mutate editor content.
- [ ] **Step 6: Replace direct review-mark removal in `App.jsx`** with status-based gray rendering and explicit delete behavior.
- [ ] **Step 7: Run `npm test` and `npm run build`**; expect all pass.
- [ ] **Step 8: Commit** with `git commit -m "feat: add persistent editor context rail"`.

### Task 7: Work Item Persistence and Atomic Thread Synchronization

**Files:**
- Modify: `apps/easy-author/backend/internal/model/model.go`
- Modify: `apps/easy-author/backend/internal/store/store.go`
- Create: `apps/easy-author/backend/internal/httpapi/workitems.go`
- Modify: `apps/easy-author/backend/internal/httpapi/app.go`
- Modify: `apps/easy-author/backend/internal/store/store_test.go`
- Modify: `apps/easy-author/backend/internal/httpapi/app_test.go`

**Interfaces:**
- Produces: `WorkItem{ID, ProjectID, BookID, ChapterID, AnchorID, ThreadID, Kind, Title, Phase, Priority, DueAt, UpdatedAt}`; `KanbanQuery{BookIDs, LimitPerPhase, IncludeDone}`; atomic `Store.MoveWorkItem(ctx, id, phase)`.
- Consumes: Task 5 anchor/thread IDs.

- [ ] **Step 1: Write failing store tests** for the five exact phases, default backlog creation, per-book/global filtering, priority/overdue/update ordering, per-phase limit, and transaction rollback when linked thread update fails.
- [ ] **Step 2: Run store tests** and verify failure.
- [ ] **Step 3: Add work-item schema and transactional store methods**; a thread-backed card is unique by `thread_id`, while standalone task/reminder cards have no thread.
- [ ] **Step 4: Write failing HTTP tests** for `GET /api/kanban`, `POST /api/books/{bookId}/work-items`, and `PUT /api/work-items/{id}/phase`, including malformed book lists and limits above 20.
- [ ] **Step 5: Implement routes and handlers** with a default limit of 12, maximum 20, and explicit totals per phase independent of the returned slice.
- [ ] **Step 6: Run `go test ./...`**; expect all pass.
- [ ] **Step 7: Commit** with `git commit -m "feat: add synchronized EasyAuthor work items"`.

### Task 8: Book and Global Kanban UI

**Files:**
- Create: `apps/easy-author/frontend/src/components/KanbanBoard.jsx`
- Create: `apps/easy-author/frontend/src/components/KanbanBoard.test.jsx`
- Create: `apps/easy-author/frontend/src/components/BookKanbanFilter.jsx`
- Create: `apps/easy-author/frontend/src/components/BookKanbanFilter.test.jsx`
- Create: `apps/easy-author/frontend/src/lib/bookColorAllocation.js`
- Create: `apps/easy-author/frontend/src/lib/bookColorAllocation.test.js`
- Modify: `apps/easy-author/frontend/src/App.jsx`
- Modify: `apps/easy-author/frontend/src/styles.css`

**Interfaces:**
- Produces: `KanbanBoard({ phases, items, totals, limit, onMove, onLoadMore, onOpenSource })`; `BookKanbanFilter({ books, selectedIds, counts, colors, onToggle })`; `allocateBookColor(existingAssignments, bookId, palette) -> Map`.
- Consumes: Task 7 Kanban endpoints and Task 6 anchor navigation.

- [ ] **Step 1: Write failing color-allocation tests** for maximum perceptual separation, stable existing assignments when adding a book, released color on removal, no persistence calls, and no ribbon/color for one selected book.
- [ ] **Step 2: Implement the allocator** using a fixed accessible palette and perceptual-distance comparison; assignment state lives only in the filter component.
- [ ] **Step 3: Write failing board/filter tests** for five columns, three aggregate counters, 12-card default, `Weitere anzeigen`, ribbons at multi-select, source navigation, pointer drag, and keyboard move actions.
- [ ] **Step 4: Implement book-local and overview-global boards** using the same components and API query; preserve the active multi-book selection while the overview remains mounted.
- [ ] **Step 5: Add quiet neutral surfaces and status accents** for red backlog, amber todo, orange in-progress, green review, and gray done; ensure book colors affect only selection circles and ribbon borders.
- [ ] **Step 6: Run `npm test` and `npm run build`**; expect all pass.
- [ ] **Step 7: Commit** with `git commit -m "feat: add EasyAuthor Kanban views"`.

### Task 9: Migration, Accessibility, and Load Stabilization

**Files:**
- Modify: `apps/easy-author/backend/internal/store/store_test.go`
- Modify: `apps/easy-author/backend/internal/httpapi/app_test.go`
- Modify: `apps/easy-author/frontend/src/App.test.jsx`
- Modify: `apps/easy-author/frontend/src/components/EditorPane.test.jsx`
- Modify: `apps/easy-author/frontend/src/styles.css`
- Modify: `apps/easy-author/README.md`

**Interfaces:**
- Consumes: all prior tasks.
- Produces: a package-5 acceptance baseline and documented migration/operation behavior.

- [ ] **Step 1: Add an existing-database migration fixture test** containing current books, chapters, anchors, comments, clipboard items, and revisions; assert all remain readable after schema initialization and gain normalized presentation/context/work-item views.
- [ ] **Step 2: Add frontend acceptance tests** for full keyboard traversal, Escape behavior, system theme changes, narrow layout, hidden resolved contexts, and 20 cards in each of five phases.
- [ ] **Step 3: Add integration tests** covering selection → comment/thread → card → phase move → gray resolved mark → reopen after reload.
- [ ] **Step 4: Fix only failures revealed by these tests** and document user-facing work views, controls, typography inheritance, annotations, and Kanban in the README.
- [ ] **Step 5: Run `gofmt` on changed Go files, `go test ./...`, `npm test`, `npm run build`, and `docker compose config --quiet`**; all commands must exit 0.
- [ ] **Step 6: Run the existing EasyAuthor smoke procedure from `apps/easy-author/README.md`** against a local stack; record any environment-bound limit if the stack cannot be started.
- [ ] **Step 7: Commit** with `git commit -m "test: stabilize EasyAuthor UI MVP"`.

### Task 10: Package-5 Acceptance and Package-6/7 Boundary Check

**Files:**
- Modify: `apps/easy-author/docs/business-case/easy-author-author-strategy-audit.md`
- Modify: `apps/easy-author/docs/business-case/easy-read-import/easy-read-import-addendum.md`
- Create: `apps/easy-author/docs/ui-mvp-validation.md`

**Interfaces:**
- Consumes: package-5 verified behavior and the approved design.
- Produces: acceptance record and explicit future contracts for `DocumentAnchor`, `WorkItem`, `KnowledgeClipboardItem`, `Source`, and `SourceRevision` without implementing source ingestion.

- [ ] **Step 1: Walk every package-1-through-5 requirement in the spec** and record its automated test or manual observation in `ui-mvp-validation.md`.
- [ ] **Step 2: Document the package-6/7 interface boundary** in the Easy Read addendum, referencing the implemented anchor/work-item contracts and listing source-specific fields that remain future work.
- [ ] **Step 3: Re-run the full backend/frontend/build/config verification suite** and paste concise command results into the validation document.
- [ ] **Step 4: Run `git diff --check` and inspect `git status --short`**; expect no whitespace errors and only intended documentation changes.
- [ ] **Step 5: Commit** with `git commit -m "docs: validate EasyAuthor UI MVP"`.
