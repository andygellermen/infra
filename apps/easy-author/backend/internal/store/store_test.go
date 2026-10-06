package store

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/db"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/model"
)

func TestWorkItemsUseFivePhasesAndThreadCardsStayUnique(t *testing.T) {
	t.Parallel()
	appStore, ctx, chapter := newContextTestStore(t)

	standalone, err := appStore.CreateWorkItem(ctx, chapter.BookID, CreateWorkItemInput{Kind: "task", Title: "Kapitel prüfen"})
	if err != nil {
		t.Fatalf("create work item: %v", err)
	}
	if standalone.Phase != "backlog" {
		t.Fatalf("new work item should start in backlog, got %#v", standalone)
	}
	for _, phase := range []string{"backlog", "todo", "in_progress", "review", "done"} {
		moved, err := appStore.MoveWorkItem(ctx, standalone.ID, phase)
		if err != nil {
			t.Fatalf("move to %s: %v", phase, err)
		}
		if moved.Phase != phase {
			t.Fatalf("move returned phase %q, want %q", moved.Phase, phase)
		}
	}
	if _, err := appStore.MoveWorkItem(ctx, standalone.ID, "finished"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("invalid phase should be rejected, got %v", err)
	}

	comment, err := appStore.CreateContext(ctx, chapter.ID, CreateContextInput{
		ContextType: "comment",
		Anchor:      DocumentAnchorInput{SelectedText: "Passage", StartOffset: 5, EndOffset: 12},
		Message:     &CreateThreadMessageInput{Author: "Andy", Body: "Bitte prüfen"},
	})
	if err != nil {
		t.Fatalf("create comment context: %v", err)
	}
	board, err := appStore.ListKanban(ctx, KanbanQuery{BookIDs: []string{chapter.BookID}, LimitPerPhase: 12})
	if err != nil {
		t.Fatalf("list board: %v", err)
	}
	var threadCards int
	for _, item := range board.Items["backlog"] {
		if item.ThreadID == comment.Thread.ID {
			threadCards++
		}
	}
	if threadCards != 1 {
		t.Fatalf("one thread must produce exactly one card, got %#v", board.Items["backlog"])
	}
	if board.Totals["done"] != 1 || len(board.Items["done"]) != 0 {
		t.Fatalf("hidden done cards must still contribute to totals: %#v", board)
	}
	if _, err := appStore.CreateWorkItem(ctx, chapter.BookID, CreateWorkItemInput{Kind: "comment", Title: "Duplikat", ThreadID: comment.Thread.ID}); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate thread card should conflict, got %v", err)
	}
}

func TestKanbanFiltersSortsAndLimitsEachPhase(t *testing.T) {
	t.Parallel()
	appStore, ctx, chapter := newContextTestStore(t)
	project, err := appStore.CreateProject(ctx, CreateProjectInput{Title: "Zweites Projekt"})
	if err != nil {
		t.Fatalf("create second project: %v", err)
	}
	otherBook, err := appStore.CreateBook(ctx, project.ID, CreateBookInput{Title: "Anderes Buch"})
	if err != nil {
		t.Fatalf("create second book: %v", err)
	}

	overdue := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
	items := []CreateWorkItemInput{
		{Kind: "task", Title: "Normal neu", Priority: 0},
		{Kind: "reminder", Title: "Überfällig", Priority: 0, DueAt: overdue},
		{Kind: "task", Title: "Hohe Priorität", Priority: 3},
	}
	for _, input := range items {
		if _, err := appStore.CreateWorkItem(ctx, chapter.BookID, input); err != nil {
			t.Fatalf("create %s: %v", input.Title, err)
		}
	}
	if _, err := appStore.CreateWorkItem(ctx, otherBook.ID, CreateWorkItemInput{Kind: "task", Title: "Fremdes Buch", Priority: 3}); err != nil {
		t.Fatalf("create foreign item: %v", err)
	}

	board, err := appStore.ListKanban(ctx, KanbanQuery{BookIDs: []string{chapter.BookID}, LimitPerPhase: 2})
	if err != nil {
		t.Fatalf("list filtered board: %v", err)
	}
	if board.Totals["backlog"] != 3 || len(board.Items["backlog"]) != 2 {
		t.Fatalf("limit must not change total: %#v", board)
	}
	if board.Items["backlog"][0].Title != "Hohe Priorität" || board.Items["backlog"][1].Title != "Überfällig" {
		t.Fatalf("priority and overdue ordering is wrong: %#v", board.Items["backlog"])
	}
	for _, item := range board.Items["backlog"] {
		if item.BookID != chapter.BookID || item.Title == "Fremdes Buch" {
			t.Fatalf("book filter leaked another book: %#v", board.Items["backlog"])
		}
	}
}

func TestMoveWorkItemRollsBackWhenLinkedThreadUpdateFails(t *testing.T) {
	t.Parallel()
	appStore, ctx, chapter := newContextTestStore(t)
	comment, err := appStore.CreateContext(ctx, chapter.ID, CreateContextInput{
		ContextType: "comment",
		Anchor:      DocumentAnchorInput{SelectedText: "Passage", StartOffset: 5, EndOffset: 12},
		Message:     &CreateThreadMessageInput{Author: "Andy", Body: "Bitte prüfen"},
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	board, err := appStore.ListKanban(ctx, KanbanQuery{BookIDs: []string{chapter.BookID}, LimitPerPhase: 12})
	if err != nil || len(board.Items["backlog"]) != 1 {
		t.Fatalf("load linked card: %#v %v", board, err)
	}
	item := board.Items["backlog"][0]
	if _, err := appStore.db.ExecContext(ctx, `CREATE TRIGGER fail_thread_move BEFORE UPDATE ON comment_threads BEGIN SELECT RAISE(FAIL, 'thread update blocked'); END`); err != nil {
		t.Fatalf("create failure trigger: %v", err)
	}
	if _, err := appStore.MoveWorkItem(ctx, item.ID, "in_progress"); err == nil {
		t.Fatal("move should fail when linked thread update fails")
	}
	var phase, threadStatus, contextStatus string
	if err := appStore.db.QueryRowContext(ctx, `SELECT phase FROM work_items WHERE id = ?`, item.ID).Scan(&phase); err != nil {
		t.Fatalf("read item phase: %v", err)
	}
	if err := appStore.db.QueryRowContext(ctx, `SELECT status FROM comment_threads WHERE id = ?`, comment.Thread.ID).Scan(&threadStatus); err != nil {
		t.Fatalf("read thread status: %v", err)
	}
	if err := appStore.db.QueryRowContext(ctx, `SELECT status FROM anchor_contexts WHERE id = ?`, comment.ID).Scan(&contextStatus); err != nil {
		t.Fatalf("read context status: %v", err)
	}
	if phase != "backlog" || threadStatus != "open" || contextStatus != "open" {
		t.Fatalf("failed move was not rolled back: phase=%s thread=%s context=%s", phase, threadStatus, contextStatus)
	}
}

func newContextTestStore(t *testing.T) (*Store, context.Context, model.Chapter) {
	t.Helper()
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "easy-author.sqlite")
	database, err := db.OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	appStore := New(database, filepath.Dir(databasePath))
	if err := appStore.Init(ctx); err != nil {
		t.Fatalf("init store: %v", err)
	}
	project, err := appStore.CreateProject(ctx, CreateProjectInput{Title: "Kontextprojekt"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	book, err := appStore.CreateBook(ctx, project.ID, CreateBookInput{Title: "Kontextbuch"})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}
	chapter, err := appStore.CreateChapter(ctx, book.ID, CreateChapterInput{Title: "Kapitel", MarkdownContent: "Eine wichtige Passage."})
	if err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	return appStore, ctx, chapter
}

func TestContextsShareAnchorAndDeleteIndependently(t *testing.T) {
	t.Parallel()
	appStore, ctx, chapter := newContextTestStore(t)

	comment, err := appStore.CreateContext(ctx, chapter.ID, CreateContextInput{
		ContextType: "comment",
		Anchor:      DocumentAnchorInput{SelectedText: "wichtige Passage", StartOffset: 5, EndOffset: 22, ContextBefore: "Eine ", ContextAfter: "."},
		Message:     &CreateThreadMessageInput{Author: "Andy", Body: "Bitte genauer prüfen."},
	})
	if err != nil {
		t.Fatalf("create comment context: %v", err)
	}
	link, err := appStore.CreateContext(ctx, chapter.ID, CreateContextInput{
		ContextType: "link",
		AnchorID:    comment.AnchorID,
		TargetID:    "chapter-ziel",
	})
	if err != nil {
		t.Fatalf("create link context: %v", err)
	}
	if link.AnchorID != comment.AnchorID {
		t.Fatalf("contexts should share anchor: %#v %#v", comment, link)
	}

	items, err := appStore.ListChapterContexts(ctx, chapter.ID)
	if err != nil {
		t.Fatalf("list contexts: %v", err)
	}
	if len(items) != 2 || items[0].Anchor.ID != items[1].Anchor.ID {
		t.Fatalf("expected two contexts on one anchor, got %#v", items)
	}
	if err := appStore.DeleteContext(ctx, link.ID); err != nil {
		t.Fatalf("delete sibling context: %v", err)
	}
	remaining, err := appStore.ListChapterContexts(ctx, chapter.ID)
	if err != nil {
		t.Fatalf("list remaining contexts: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != comment.ID || remaining[0].Anchor.ID != comment.AnchorID {
		t.Fatalf("deleting one context removed its sibling or anchor: %#v", remaining)
	}
}

func TestContextThreadRepliesAndResolvedStatus(t *testing.T) {
	t.Parallel()
	appStore, ctx, chapter := newContextTestStore(t)
	created, err := appStore.CreateContext(ctx, chapter.ID, CreateContextInput{
		ContextType: "comment",
		Anchor:      DocumentAnchorInput{SelectedText: "Passage", StartOffset: 14, EndOffset: 21},
		Message:     &CreateThreadMessageInput{Author: "Andy", Body: "Erste Frage"},
	})
	if err != nil {
		t.Fatalf("create threaded context: %v", err)
	}
	if _, err := appStore.CreateThreadMessage(ctx, created.Thread.ID, CreateThreadMessageInput{Author: "Cody", Body: "Erste Antwort"}); err != nil {
		t.Fatalf("create reply: %v", err)
	}
	if _, err := appStore.CreateThreadMessage(ctx, created.Thread.ID, CreateThreadMessageInput{Author: "Andy", Body: "Zweite Antwort"}); err != nil {
		t.Fatalf("create second reply: %v", err)
	}
	thread, err := appStore.UpdateThreadStatus(ctx, created.Thread.ID, "resolved")
	if err != nil {
		t.Fatalf("resolve thread: %v", err)
	}
	if thread.Status != "resolved" || len(thread.Messages) != 3 {
		t.Fatalf("unexpected resolved thread: %#v", thread)
	}
	for index, body := range []string{"Erste Frage", "Erste Antwort", "Zweite Antwort"} {
		if thread.Messages[index].Body != body {
			t.Fatalf("messages not in stable order: %#v", thread.Messages)
		}
	}
	items, err := appStore.ListChapterContexts(ctx, chapter.ID)
	if err != nil || len(items) != 1 || items[0].Status != "resolved" || items[0].Thread.Status != "resolved" {
		t.Fatalf("resolved status not reflected by context: items=%#v err=%v", items, err)
	}
	board, err := appStore.ListKanban(ctx, KanbanQuery{BookIDs: []string{chapter.BookID}, LimitPerPhase: 12, IncludeDone: true})
	if err != nil || board.Totals["done"] != 1 || len(board.Items["done"]) != 1 || board.Items["done"][0].ThreadID != created.Thread.ID {
		t.Fatalf("resolved thread did not move its card atomically: board=%#v err=%v", board, err)
	}
}

func TestExistingDatabaseFixtureRemainsReadableAndGainsNormalizedViews(t *testing.T) {
	t.Parallel()
	appStore, ctx, chapter := newContextTestStore(t)
	box, err := appStore.CreateWorkflowBox(ctx, chapter.BookID, CreateWorkflowBoxInput{Title: "Legacy-Ziel", Type: "notes"})
	if err != nil {
		t.Fatalf("create workflow box: %v", err)
	}
	legacyAnchor, err := appStore.CreateAnchor(ctx, chapter.ID, CreateAnchorInput{
		WorkflowBoxID: box.ID, SelectedText: "wichtige Passage", StartOffset: 5, EndOffset: 22,
	})
	if err != nil {
		t.Fatalf("create legacy anchor: %v", err)
	}
	revision, err := appStore.CreateRevision(ctx, chapter.ID, CreateRevisionInput{RevisionType: "manual", CreatedBy: "test"})
	if err != nil {
		t.Fatalf("create legacy revision: %v", err)
	}
	legacyComment, err := appStore.CreateReviewComment(ctx, chapter.ID, CreateReviewCommentInput{
		RevisionID: revision.ID, Author: "Andy", Body: "Legacy-Kommentar", SelectedText: "Passage", StartOffset: 14, EndOffset: 21, Status: "resolved",
	})
	if err != nil {
		t.Fatalf("create legacy comment: %v", err)
	}
	clipboard, err := appStore.CreateClipboardItem(ctx, chapter.BookID, CreateClipboardItemInput{ChapterID: chapter.ID, Content: "Bewahrter Ausschnitt"})
	if err != nil {
		t.Fatalf("create clipboard fixture: %v", err)
	}
	if err := appStore.Init(ctx); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}
	items, err := appStore.ListChapterContexts(ctx, chapter.ID)
	if err != nil {
		t.Fatalf("list migrated contexts: %v", err)
	}
	var foundAnchor, foundComment bool
	for _, item := range items {
		if item.Anchor.ID == legacyAnchor.ID && item.LegacyEntityID == legacyAnchor.ID && item.ContextType == "link" {
			foundAnchor = true
		}
		if item.ID == legacyComment.ID && item.LegacyEntityID == legacyComment.ID && item.ContextType == "comment" && item.Status == "resolved" {
			foundComment = true
		}
	}
	if !foundAnchor || !foundComment {
		t.Fatalf("legacy records were not preserved in context view: %#v", items)
	}
	bundle, err := appStore.GetBook(ctx, chapter.BookID)
	if err != nil || bundle.Book.ID != chapter.BookID || len(bundle.Chapters) == 0 {
		t.Fatalf("book and chapter were not preserved: bundle=%#v err=%v", bundle, err)
	}
	clips, err := appStore.ListClipboardItems(ctx, chapter.BookID)
	if err != nil || len(clips) != 1 || clips[0].ID != clipboard.ID {
		t.Fatalf("clipboard fixture was not preserved: clips=%#v err=%v", clips, err)
	}
	revisions, err := appStore.ListRevisionsByChapter(ctx, chapter.ID)
	if err != nil || len(revisions) == 0 || revisions[0].ID != revision.ID {
		t.Fatalf("revision fixture was not preserved: revisions=%#v err=%v", revisions, err)
	}
	presentation, err := appStore.GetBookPresentation(ctx, chapter.BookID)
	if err != nil || presentation.DefaultWorkView != "clean" {
		t.Fatalf("legacy book did not gain presentation defaults: %#v err=%v", presentation, err)
	}
	board, err := appStore.ListKanban(ctx, KanbanQuery{BookIDs: []string{chapter.BookID}, LimitPerPhase: 12, IncludeDone: true})
	if err != nil || board.Totals["done"] != 1 || len(board.Items["done"]) != 1 {
		t.Fatalf("legacy resolved comment did not gain a work-item view: %#v err=%v", board, err)
	}
}

func TestBookPresentationDefaultsAndPersistsAcrossReopen(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "easy-author.sqlite")
	database, err := db.OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	appStore := New(database, filepath.Dir(databasePath))
	if err := appStore.Init(ctx); err != nil {
		t.Fatalf("init store: %v", err)
	}
	project, err := appStore.CreateProject(ctx, CreateProjectInput{Title: "Romanprojekt"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	book, err := appStore.CreateBook(ctx, project.ID, CreateBookInput{Title: "Testbuch"})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}

	presentation, err := appStore.GetBookPresentation(ctx, book.ID)
	if err != nil {
		t.Fatalf("get default presentation: %v", err)
	}
	if presentation.DefaultWorkView != "clean" || string(presentation.TypographyOverrides) != "{}" {
		t.Fatalf("unexpected legacy default: %#v", presentation)
	}

	wantTypography := json.RawMessage(`{"bodyFont":"Literata"}`)
	if _, err := appStore.UpdateBookPresentation(ctx, model.BookPresentation{
		BookID:              book.ID,
		DefaultWorkView:     "review",
		TypographyOverrides: wantTypography,
	}); err != nil {
		t.Fatalf("update presentation: %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("close sqlite: %v", err)
	}

	reopened, err := db.OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("reopen sqlite: %v", err)
	}
	t.Cleanup(func() { _ = reopened.Close() })
	reopenedStore := New(reopened, filepath.Dir(databasePath))
	if err := reopenedStore.Init(ctx); err != nil {
		t.Fatalf("re-init store: %v", err)
	}
	persisted, err := reopenedStore.GetBookPresentation(ctx, book.ID)
	if err != nil {
		t.Fatalf("get persisted presentation: %v", err)
	}
	if persisted.DefaultWorkView != "review" || string(persisted.TypographyOverrides) != string(wantTypography) {
		t.Fatalf("presentation did not persist: %#v", persisted)
	}
}

func TestInitMigratesLegacyReviewCommentsRevisionColumn(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	database, err := db.OpenSQLite(filepath.Join(tempDir, "easy-author.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	ctx := context.Background()
	legacyStatements := []string{
		`CREATE TABLE projects (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE books (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			title TEXT NOT NULL,
			subtitle TEXT NOT NULL DEFAULT '',
			author TEXT NOT NULL DEFAULT '',
			visibility TEXT NOT NULL DEFAULT 'private',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE chapters (
			id TEXT PRIMARY KEY,
			book_id TEXT NOT NULL,
			title TEXT NOT NULL,
			position INTEGER NOT NULL,
			markdown_content TEXT NOT NULL DEFAULT '',
			editor_json TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE review_comments (
			id TEXT PRIMARY KEY,
			chapter_id TEXT NOT NULL,
			comment_type TEXT NOT NULL DEFAULT 'comment',
			author TEXT NOT NULL DEFAULT '',
			body TEXT NOT NULL DEFAULT '',
			suggested_text TEXT NOT NULL DEFAULT '',
			selected_text TEXT NOT NULL DEFAULT '',
			start_offset INTEGER NOT NULL DEFAULT 0,
			end_offset INTEGER NOT NULL DEFAULT 0,
			context_before TEXT NOT NULL DEFAULT '',
			context_after TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'open',
			is_todo_done INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
	}
	for _, statement := range legacyStatements {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("seed legacy schema: %v", err)
		}
	}

	appStore := New(database, filepath.Join(tempDir, "library"))
	if err := appStore.Init(ctx); err != nil {
		t.Fatalf("init store: %v", err)
	}

	var revisionColumnCount int
	if err := database.QueryRowContext(ctx, `SELECT COUNT(*) FROM pragma_table_info('review_comments') WHERE name = 'revision_id'`).Scan(&revisionColumnCount); err != nil {
		t.Fatalf("query pragma_table_info: %v", err)
	}
	if revisionColumnCount != 1 {
		t.Fatalf("expected migrated revision_id column, got %d", revisionColumnCount)
	}

	var revisionIndexCount int
	if err := database.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = 'review_comments_chapter_revision_created_idx'`).Scan(&revisionIndexCount); err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	if revisionIndexCount != 1 {
		t.Fatalf("expected migrated review comment index, got %d", revisionIndexCount)
	}
}
