package store

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/db"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/model"
)

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
}

func TestInitMigratesLegacyAnchorsAndCommentsIntoContexts(t *testing.T) {
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
