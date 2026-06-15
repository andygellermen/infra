package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/db"
)

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
