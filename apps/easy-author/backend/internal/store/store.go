package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/model"
)

type Store struct {
	db         *sql.DB
	libraryDir string
}

type CreateProjectInput struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	LastOpenedAt string `json:"last_opened_at"`
}

type UpdateProjectInput struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	LastOpenedAt string `json:"last_opened_at"`
}

type CreateBookInput struct {
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Author       string `json:"author"`
	Visibility   string `json:"visibility"`
	WorkType     string `json:"work_type"`
	ThemeID      string `json:"theme_id"`
	CoverAssetID string `json:"cover_asset_id"`
}

type UpdateBookInput struct {
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Author       string `json:"author"`
	Visibility   string `json:"visibility"`
	WorkType     string `json:"work_type"`
	ThemeID      string `json:"theme_id"`
	CoverAssetID string `json:"cover_asset_id"`
}

type CreateChapterInput struct {
	Title           string `json:"title"`
	Summary         string `json:"summary"`
	MarkdownContent string `json:"markdown_content"`
	EditorJSON      string `json:"editor_json"`
	SectionType     string `json:"section_type"`
	Status          string `json:"status"`
}

type UpdateChapterInput struct {
	Title               string `json:"title"`
	Summary             string `json:"summary"`
	MarkdownContent     string `json:"markdown_content"`
	EditorJSON          string `json:"editor_json"`
	SectionType         string `json:"section_type"`
	Status              string `json:"status"`
	SaveMode            string `json:"save_mode"`
	AutosaveReason      string `json:"autosave_reason"`
	SessionID           string `json:"session_id"`
	CreateRevision      bool   `json:"create_revision"`
	RevisionType        string `json:"revision_type"`
	RevisionTitle       string `json:"revision_title"`
	RevisionDescription string `json:"revision_description"`
	CreatedBy           string `json:"created_by"`
}

type UpdateChapterOptionsInput struct {
	SectionType        string `json:"section_type"`
	Status             string `json:"status"`
	IsIncludedInExport bool   `json:"is_included_in_export"`
	IsVisibleInTOC     bool   `json:"is_visible_in_toc"`
	IsSampleContent    bool   `json:"is_sample_content"`
}

type CreateReusableBookPageInput struct {
	Title           string `json:"title"`
	PageType        string `json:"page_type"`
	ContentMarkdown string `json:"content_markdown"`
	ContentJSON     string `json:"content_json"`
	IsGlobal        bool   `json:"is_global"`
}

type UpdateReusableBookPageInput struct {
	Title           string `json:"title"`
	PageType        string `json:"page_type"`
	ContentMarkdown string `json:"content_markdown"`
	ContentJSON     string `json:"content_json"`
	IsGlobal        bool   `json:"is_global"`
}

type CreateThemeInput struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Target       string `json:"target"`
	SettingsJSON string `json:"settings_json"`
}

type UpdateThemeInput struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Target       string `json:"target"`
	SettingsJSON string `json:"settings_json"`
}

type CreateRevisionInput struct {
	RevisionType string `json:"revision_type"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	SessionID    string `json:"session_id"`
	CreatedBy    string `json:"created_by"`
}

type RestoreRevisionInput struct {
	CreatedBy string `json:"created_by"`
}

type CreateMilestoneInput struct {
	RevisionID    string `json:"revision_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	MilestoneType string `json:"milestone_type"`
	Locked        bool   `json:"locked"`
	CreatedBy     string `json:"created_by"`
}

type UpdateMilestoneInput struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	MilestoneType string `json:"milestone_type"`
	Locked        bool   `json:"locked"`
	CreatedBy     string `json:"created_by"`
}

type ReorderChaptersInput struct {
	ChapterIDs []string `json:"chapter_ids"`
}

type CreateWorkflowBoxInput struct {
	Title       string   `json:"title"`
	Type        string   `json:"type"`
	Tags        []string `json:"tags"`
	IsCollapsed bool     `json:"is_collapsed"`
}

type UpdateWorkflowBoxInput struct {
	Title       string   `json:"title"`
	Type        string   `json:"type"`
	Tags        []string `json:"tags"`
	IsCollapsed bool     `json:"is_collapsed"`
}

type CreateAnchorInput struct {
	WorkflowBoxID string `json:"workflow_box_id"`
	Title         string `json:"title"`
	AnchorType    string `json:"anchor_type"`
	SelectedText  string `json:"selected_text"`
	StartOffset   int    `json:"start_offset"`
	EndOffset     int    `json:"end_offset"`
	ContextBefore string `json:"context_before"`
	ContextAfter  string `json:"context_after"`
	Note          string `json:"note"`
}

type CreateClipboardItemInput struct {
	ChapterID      string `json:"chapter_id"`
	Content        string `json:"content"`
	ContentType    string `json:"content_type"`
	SourceAnchorID string `json:"source_anchor_id"`
	Slot           int    `json:"slot"`
	IsPinned       bool   `json:"is_pinned"`
}

type UpdateClipboardItemInput struct {
	Content  string `json:"content"`
	Slot     int    `json:"slot"`
	IsPinned bool   `json:"is_pinned"`
}

type CreateReviewCommentInput struct {
	RevisionID    string `json:"revision_id"`
	CommentType   string `json:"comment_type"`
	Author        string `json:"author"`
	Body          string `json:"body"`
	SuggestedText string `json:"suggested_text"`
	SelectedText  string `json:"selected_text"`
	StartOffset   int    `json:"start_offset"`
	EndOffset     int    `json:"end_offset"`
	ContextBefore string `json:"context_before"`
	ContextAfter  string `json:"context_after"`
	Status        string `json:"status"`
	IsTodoDone    bool   `json:"is_todo_done"`
}

type UpdateReviewCommentInput struct {
	RevisionID    string `json:"revision_id"`
	CommentType   string `json:"comment_type"`
	Author        string `json:"author"`
	Body          string `json:"body"`
	SuggestedText string `json:"suggested_text"`
	Status        string `json:"status"`
	IsTodoDone    bool   `json:"is_todo_done"`
}

type CreateKnowledgeItemInput struct {
	Type    string   `json:"type"`
	Name    string   `json:"name"`
	Summary string   `json:"summary"`
	Body    string   `json:"body"`
	Tags    []string `json:"tags"`
}

type UpdateKnowledgeItemInput struct {
	Type    string   `json:"type"`
	Name    string   `json:"name"`
	Summary string   `json:"summary"`
	Body    string   `json:"body"`
	Tags    []string `json:"tags"`
}

func New(db *sql.DB, libraryDir string) *Store {
	return &Store{db: db, libraryDir: libraryDir}
}

func (s *Store) Init(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'active',
			last_opened_at TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS books (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			title TEXT NOT NULL,
			subtitle TEXT NOT NULL DEFAULT '',
			author TEXT NOT NULL DEFAULT '',
			visibility TEXT NOT NULL DEFAULT 'private',
			work_type TEXT NOT NULL DEFAULT 'book',
			theme_id TEXT NOT NULL DEFAULT '',
			cover_asset_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS chapters (
			id TEXT PRIMARY KEY,
			book_id TEXT NOT NULL,
			title TEXT NOT NULL,
			summary TEXT NOT NULL DEFAULT '',
			position INTEGER NOT NULL,
			markdown_content TEXT NOT NULL DEFAULT '',
			editor_json TEXT NOT NULL DEFAULT '',
			section_type TEXT NOT NULL DEFAULT 'body',
			status TEXT NOT NULL DEFAULT 'draft',
			is_included_in_export INTEGER NOT NULL DEFAULT 1,
			is_visible_in_toc INTEGER NOT NULL DEFAULT 1,
			is_sample_content INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY(book_id) REFERENCES books(id) ON DELETE CASCADE
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS chapters_book_position_idx ON chapters(book_id, position);`,
		`CREATE TABLE IF NOT EXISTS workflow_boxes (
			id TEXT PRIMARY KEY,
			book_id TEXT NOT NULL,
			title TEXT NOT NULL,
			type TEXT NOT NULL,
			tags_json TEXT NOT NULL DEFAULT '[]',
			position INTEGER NOT NULL,
			is_collapsed INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY(book_id) REFERENCES books(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS anchors (
			id TEXT PRIMARY KEY,
			chapter_id TEXT NOT NULL,
			workflow_box_id TEXT NOT NULL,
			title TEXT NOT NULL DEFAULT '',
			anchor_type TEXT NOT NULL,
			selected_text TEXT NOT NULL DEFAULT '',
			start_offset INTEGER NOT NULL DEFAULT 0,
			end_offset INTEGER NOT NULL DEFAULT 0,
			context_before TEXT NOT NULL DEFAULT '',
			context_after TEXT NOT NULL DEFAULT '',
			note TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY(chapter_id) REFERENCES chapters(id) ON DELETE CASCADE,
			FOREIGN KEY(workflow_box_id) REFERENCES workflow_boxes(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS review_comments (
			id TEXT PRIMARY KEY,
			chapter_id TEXT NOT NULL,
			revision_id TEXT NOT NULL DEFAULT '',
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
			updated_at TEXT NOT NULL,
			FOREIGN KEY(chapter_id) REFERENCES chapters(id) ON DELETE CASCADE,
			FOREIGN KEY(revision_id) REFERENCES revisions(id) ON DELETE SET DEFAULT
		);`,
		`CREATE TABLE IF NOT EXISTS clipboard_items (
			id TEXT PRIMARY KEY,
			book_id TEXT NOT NULL,
			chapter_id TEXT NOT NULL DEFAULT '',
			content TEXT NOT NULL,
			content_type TEXT NOT NULL DEFAULT 'text/plain',
			source_anchor_id TEXT NOT NULL DEFAULT '',
			slot INTEGER NOT NULL DEFAULT 0,
			is_pinned INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY(book_id) REFERENCES books(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS knowledge_items (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			summary TEXT NOT NULL DEFAULT '',
			body TEXT NOT NULL DEFAULT '',
			tags_json TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS reusable_book_pages (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL DEFAULT '',
			title TEXT NOT NULL,
			page_type TEXT NOT NULL DEFAULT 'custom',
			content_markdown TEXT NOT NULL DEFAULT '',
			content_json TEXT NOT NULL DEFAULT '',
			is_global INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS reusable_book_pages_project_updated_idx ON reusable_book_pages(project_id, updated_at DESC);`,
		`CREATE TABLE IF NOT EXISTS themes (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			target TEXT NOT NULL DEFAULT 'general',
			settings_json TEXT NOT NULL DEFAULT '{}',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS themes_target_updated_idx ON themes(target, updated_at DESC);`,
	}
	statements = append(statements, versioningSchemaStatements()...)

	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("init schema: %w", err)
		}
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE workflow_boxes ADD COLUMN tags_json TEXT NOT NULL DEFAULT '[]'`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("migrate workflow tags: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE projects ADD COLUMN status TEXT NOT NULL DEFAULT 'active'`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("migrate project status: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE projects ADD COLUMN last_opened_at TEXT NOT NULL DEFAULT ''`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("migrate project last_opened_at: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE books ADD COLUMN work_type TEXT NOT NULL DEFAULT 'book'`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("migrate book work_type: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE books ADD COLUMN theme_id TEXT NOT NULL DEFAULT ''`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("migrate book theme_id: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE books ADD COLUMN cover_asset_id TEXT NOT NULL DEFAULT ''`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("migrate book cover_asset_id: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE chapters ADD COLUMN summary TEXT NOT NULL DEFAULT ''`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("migrate chapter summary: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE chapters ADD COLUMN section_type TEXT NOT NULL DEFAULT 'body'`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("migrate chapter section_type: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE chapters ADD COLUMN status TEXT NOT NULL DEFAULT 'draft'`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("migrate chapter status: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE chapters ADD COLUMN is_included_in_export INTEGER NOT NULL DEFAULT 1`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("migrate chapter is_included_in_export: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE chapters ADD COLUMN is_visible_in_toc INTEGER NOT NULL DEFAULT 1`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("migrate chapter is_visible_in_toc: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE chapters ADD COLUMN is_sample_content INTEGER NOT NULL DEFAULT 0`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("migrate chapter is_sample_content: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE review_comments ADD COLUMN revision_id TEXT NOT NULL DEFAULT ''`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("migrate review comment revision_id: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS review_comments_chapter_revision_created_idx ON review_comments(chapter_id, revision_id, created_at DESC)`); err != nil {
		return fmt.Errorf("migrate review comment revision index: %w", err)
	}

	return s.ensureDemoContent(ctx)
}

func (s *Store) ensureDemoContent(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM projects`).Scan(&count); err != nil {
		return fmt.Errorf("count projects: %w", err)
	}
	if count > 0 {
		return nil
	}

	project, err := s.CreateProject(ctx, CreateProjectInput{
		Title:       "easy-author Demo-Projekt",
		Description: "Erster Spike mit Schreibraum, Workflow-Boxen und Clipboard.",
	})
	if err != nil {
		return err
	}
	book, err := s.CreateBook(ctx, project.ID, CreateBookInput{
		Title:      "Die leise Karte",
		Subtitle:   "Ein Beispielmanuskript fuer den MVP-Spike",
		Author:     "easy-author",
		Visibility: "private",
	})
	if err != nil {
		return err
	}
	if _, err := s.CreateChapter(ctx, book.ID, CreateChapterInput{
		Title: "Kapitel 1",
		MarkdownContent: strings.TrimSpace(`# Kapitel 1

Der Garten war still, doch in Mara begann etwas zu kippen.

Sie blieb noch einen Moment am Tor stehen und notierte sich im Kopf, was spaeter wichtig werden koennte.`),
		EditorJSON: "",
	}); err != nil {
		return err
	}
	for index, boxType := range []string{"notes", "persons", "research", "clipboard"} {
		title := map[string]string{
			"notes":     "Notizen",
			"persons":   "Figuren",
			"research":  "Recherche",
			"clipboard": "Clipboard",
		}[boxType]
		tags := map[string][]string{
			"notes":     {"idee", "notiz", "frage", "offen"},
			"persons":   {"figur", "person", "charakter", "mara"},
			"research":  {"quelle", "fakt", "recherche", "datum", "jahr"},
			"clipboard": {"snippet", "textbaustein", "clipboard"},
		}[boxType]
		if _, err := s.createWorkflowBoxWithPosition(ctx, book.ID, CreateWorkflowBoxInput{
			Title:       title,
			Type:        boxType,
			Tags:        tags,
			IsCollapsed: index > 1,
		}, index+1); err != nil {
			return err
		}
	}
	for _, item := range []CreateKnowledgeItemInput{
		{Type: "person", Name: "Mara", Summary: "Hauptfigur im Demokapitel."},
		{Type: "location", Name: "Alter Garten", Summary: "Ort der stillen Eröffnungsszene."},
		{Type: "event", Name: "Erste Begegnung", Summary: "Tragendes Ereignis fuer den spaeteren Plot."},
	} {
		if _, err := s.CreateKnowledgeItem(ctx, project.ID, item); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ListProjects(ctx context.Context) ([]model.Project, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, description, status, last_opened_at, created_at, updated_at FROM projects ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var item model.Project
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Status, &item.LastOpenedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, item)
	}
	return projects, rows.Err()
}

func (s *Store) CreateProject(ctx context.Context, input CreateProjectInput) (model.Project, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "Unbenanntes Projekt"
	}
	now := nowUTC()
	item := model.Project{
		ID:           newID(),
		Title:        title,
		Description:  strings.TrimSpace(input.Description),
		Status:       normalizeProjectStatus(input.Status),
		LastOpenedAt: strings.TrimSpace(input.LastOpenedAt),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO projects (id, title, description, status, last_opened_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.Title, item.Description, item.Status, item.LastOpenedAt, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return model.Project{}, fmt.Errorf("create project: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateProject(ctx context.Context, id string, input UpdateProjectInput) (model.Project, error) {
	project, _, err := s.GetProject(ctx, id)
	if err != nil {
		return model.Project{}, err
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = project.Title
	}
	project.Title = title
	project.Description = strings.TrimSpace(input.Description)
	project.Status = normalizeProjectStatus(fallback(strings.TrimSpace(input.Status), project.Status))
	if strings.TrimSpace(input.LastOpenedAt) != "" {
		project.LastOpenedAt = strings.TrimSpace(input.LastOpenedAt)
	}
	project.UpdatedAt = nowUTC()
	_, err = s.db.ExecContext(
		ctx,
		`UPDATE projects SET title = ?, description = ?, status = ?, last_opened_at = ?, updated_at = ? WHERE id = ?`,
		project.Title,
		project.Description,
		project.Status,
		project.LastOpenedAt,
		project.UpdatedAt,
		project.ID,
	)
	if err != nil {
		return model.Project{}, fmt.Errorf("update project: %w", err)
	}
	return project, nil
}

func (s *Store) GetProject(ctx context.Context, id string) (model.Project, []model.Book, error) {
	var item model.Project
	err := s.db.QueryRowContext(ctx, `SELECT id, title, description, status, last_opened_at, created_at, updated_at FROM projects WHERE id = ?`, id).
		Scan(&item.ID, &item.Title, &item.Description, &item.Status, &item.LastOpenedAt, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Project{}, nil, ErrNotFound
		}
		return model.Project{}, nil, fmt.Errorf("get project: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `SELECT id, project_id, title, subtitle, author, visibility, work_type, theme_id, cover_asset_id, created_at, updated_at FROM books WHERE project_id = ? ORDER BY updated_at DESC`, id)
	if err != nil {
		return model.Project{}, nil, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()

	var books []model.Book
	for rows.Next() {
		var book model.Book
		if err := rows.Scan(&book.ID, &book.ProjectID, &book.Title, &book.Subtitle, &book.Author, &book.Visibility, &book.WorkType, &book.ThemeID, &book.CoverAssetID, &book.CreatedAt, &book.UpdatedAt); err != nil {
			return model.Project{}, nil, fmt.Errorf("scan book: %w", err)
		}
		books = append(books, book)
	}
	return item, books, rows.Err()
}

func (s *Store) CreateBook(ctx context.Context, projectID string, input CreateBookInput) (model.Book, error) {
	if _, _, err := s.GetProject(ctx, projectID); err != nil {
		return model.Book{}, err
	}
	visibility := normalizeVisibility(input.Visibility)
	now := nowUTC()
	item := model.Book{
		ID:           newID(),
		ProjectID:    projectID,
		Title:        fallback(strings.TrimSpace(input.Title), "Neues Buch"),
		Subtitle:     strings.TrimSpace(input.Subtitle),
		Author:       strings.TrimSpace(input.Author),
		Visibility:   visibility,
		WorkType:     normalizeWorkType(input.WorkType),
		ThemeID:      strings.TrimSpace(input.ThemeID),
		CoverAssetID: strings.TrimSpace(input.CoverAssetID),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO books (id, project_id, title, subtitle, author, visibility, work_type, theme_id, cover_asset_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.ProjectID, item.Title, item.Subtitle, item.Author, item.Visibility, item.WorkType, item.ThemeID, item.CoverAssetID, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return model.Book{}, fmt.Errorf("create book: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE projects SET updated_at = ? WHERE id = ?`, now, projectID); err != nil {
		return model.Book{}, fmt.Errorf("touch project: %w", err)
	}
	return item, nil
}

func (s *Store) GetBook(ctx context.Context, bookID string) (model.BookBundle, error) {
	var book model.Book
	err := s.db.QueryRowContext(ctx, `SELECT id, project_id, title, subtitle, author, visibility, work_type, theme_id, cover_asset_id, created_at, updated_at FROM books WHERE id = ?`, bookID).
		Scan(&book.ID, &book.ProjectID, &book.Title, &book.Subtitle, &book.Author, &book.Visibility, &book.WorkType, &book.ThemeID, &book.CoverAssetID, &book.CreatedAt, &book.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.BookBundle{}, ErrNotFound
		}
		return model.BookBundle{}, fmt.Errorf("get book: %w", err)
	}

	chapters, err := s.ListChapters(ctx, bookID)
	if err != nil {
		return model.BookBundle{}, err
	}
	boxes, err := s.ListWorkflowBoxes(ctx, bookID)
	if err != nil {
		return model.BookBundle{}, err
	}
	clipboard, err := s.ListClipboardItems(ctx, bookID)
	if err != nil {
		return model.BookBundle{}, err
	}
	reusablePages, err := s.ListReusableBookPages(ctx, bookID)
	if err != nil {
		return model.BookBundle{}, err
	}
	themes, err := s.ListThemes(ctx)
	if err != nil {
		return model.BookBundle{}, err
	}

	return model.BookBundle{
		Book:          book,
		Chapters:      chapters,
		WorkflowBoxes: boxes,
		Clipboard:     clipboard,
		ReusablePages: reusablePages,
		Themes:        themes,
	}, nil
}

func (s *Store) UpdateBook(ctx context.Context, id string, input UpdateBookInput) (model.Book, error) {
	var item model.Book
	err := s.db.QueryRowContext(ctx, `SELECT id, project_id, title, subtitle, author, visibility, work_type, theme_id, cover_asset_id, created_at, updated_at FROM books WHERE id = ?`, id).
		Scan(&item.ID, &item.ProjectID, &item.Title, &item.Subtitle, &item.Author, &item.Visibility, &item.WorkType, &item.ThemeID, &item.CoverAssetID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Book{}, ErrNotFound
		}
		return model.Book{}, fmt.Errorf("get book: %w", err)
	}

	item.Title = fallback(strings.TrimSpace(input.Title), item.Title)
	item.Subtitle = strings.TrimSpace(input.Subtitle)
	item.Author = strings.TrimSpace(input.Author)
	item.Visibility = normalizeVisibility(input.Visibility)
	item.WorkType = normalizeWorkType(fallback(strings.TrimSpace(input.WorkType), item.WorkType))
	item.ThemeID = strings.TrimSpace(input.ThemeID)
	item.CoverAssetID = strings.TrimSpace(input.CoverAssetID)
	item.UpdatedAt = nowUTC()

	_, err = s.db.ExecContext(ctx, `UPDATE books SET title = ?, subtitle = ?, author = ?, visibility = ?, work_type = ?, theme_id = ?, cover_asset_id = ?, updated_at = ? WHERE id = ?`,
		item.Title, item.Subtitle, item.Author, item.Visibility, item.WorkType, item.ThemeID, item.CoverAssetID, item.UpdatedAt, item.ID,
	)
	if err != nil {
		return model.Book{}, fmt.Errorf("update book: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE projects SET updated_at = ? WHERE id = ?`, item.UpdatedAt, item.ProjectID); err != nil {
		return model.Book{}, fmt.Errorf("touch project: %w", err)
	}
	return item, nil
}

func (s *Store) ListChapters(ctx context.Context, bookID string) ([]model.Chapter, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, book_id, title, summary, position, markdown_content, editor_json, section_type, status, is_included_in_export, is_visible_in_toc, is_sample_content, created_at, updated_at FROM chapters WHERE book_id = ? ORDER BY position ASC`, bookID)
	if err != nil {
		return nil, fmt.Errorf("list chapters: %w", err)
	}
	defer rows.Close()

	var chapters []model.Chapter
	for rows.Next() {
		var item model.Chapter
		var includedExport, visibleTOC, sampleContent int
		if err := rows.Scan(&item.ID, &item.BookID, &item.Title, &item.Summary, &item.Position, &item.MarkdownContent, &item.EditorJSON, &item.SectionType, &item.Status, &includedExport, &visibleTOC, &sampleContent, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan chapter: %w", err)
		}
		item.IsIncludedInExport = includedExport == 1
		item.IsVisibleInTOC = visibleTOC == 1
		item.IsSampleContent = sampleContent == 1
		chapters = append(chapters, item)
	}
	return chapters, rows.Err()
}

func (s *Store) CreateChapter(ctx context.Context, bookID string, input CreateChapterInput) (model.Chapter, error) {
	var nextPosition int
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(position), 0) + 1 FROM chapters WHERE book_id = ?`, bookID).Scan(&nextPosition); err != nil {
		return model.Chapter{}, fmt.Errorf("next chapter position: %w", err)
	}
	now := nowUTC()
	item := model.Chapter{
		ID:                 newID(),
		BookID:             bookID,
		Title:              fallback(strings.TrimSpace(input.Title), fmt.Sprintf("Kapitel %d", nextPosition)),
		Summary:            strings.TrimSpace(input.Summary),
		Position:           nextPosition,
		MarkdownContent:    strings.TrimSpace(input.MarkdownContent),
		EditorJSON:         strings.TrimSpace(input.EditorJSON),
		SectionType:        normalizeSectionType(input.SectionType),
		Status:             normalizeChapterStatus(input.Status),
		IsIncludedInExport: true,
		IsVisibleInTOC:     true,
		IsSampleContent:    false,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO chapters (id, book_id, title, summary, position, markdown_content, editor_json, section_type, status, is_included_in_export, is_visible_in_toc, is_sample_content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.BookID, item.Title, item.Summary, item.Position, item.MarkdownContent, item.EditorJSON, item.SectionType, item.Status, boolToInt(item.IsIncludedInExport), boolToInt(item.IsVisibleInTOC), boolToInt(item.IsSampleContent), item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return model.Chapter{}, fmt.Errorf("create chapter: %w", err)
	}
	if err := s.syncChapterSnapshot(ctx, item); err != nil {
		return model.Chapter{}, err
	}
	return item, nil
}

func (s *Store) GetChapter(ctx context.Context, id string) (model.Chapter, error) {
	var item model.Chapter
	var includedExport, visibleTOC, sampleContent int
	err := s.db.QueryRowContext(ctx, `SELECT id, book_id, title, summary, position, markdown_content, editor_json, section_type, status, is_included_in_export, is_visible_in_toc, is_sample_content, created_at, updated_at FROM chapters WHERE id = ?`, id).
		Scan(&item.ID, &item.BookID, &item.Title, &item.Summary, &item.Position, &item.MarkdownContent, &item.EditorJSON, &item.SectionType, &item.Status, &includedExport, &visibleTOC, &sampleContent, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Chapter{}, ErrNotFound
		}
		return model.Chapter{}, fmt.Errorf("get chapter: %w", err)
	}
	item.IsIncludedInExport = includedExport == 1
	item.IsVisibleInTOC = visibleTOC == 1
	item.IsSampleContent = sampleContent == 1
	return item, nil
}

func (s *Store) UpdateChapter(ctx context.Context, id string, input UpdateChapterInput) (model.Chapter, error) {
	item, err := s.GetChapter(ctx, id)
	if err != nil {
		return model.Chapter{}, err
	}
	item.Title = fallback(strings.TrimSpace(input.Title), item.Title)
	item.Summary = strings.TrimSpace(input.Summary)
	item.MarkdownContent = strings.TrimSpace(input.MarkdownContent)
	item.EditorJSON = strings.TrimSpace(input.EditorJSON)
	if strings.TrimSpace(input.SectionType) != "" {
		item.SectionType = normalizeSectionType(input.SectionType)
	}
	if strings.TrimSpace(input.Status) != "" {
		item.Status = normalizeChapterStatus(input.Status)
	}
	item.UpdatedAt = nowUTC()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Chapter{}, fmt.Errorf("begin update chapter tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `UPDATE chapters SET title = ?, summary = ?, markdown_content = ?, editor_json = ?, section_type = ?, status = ?, updated_at = ? WHERE id = ?`,
		item.Title, item.Summary, item.MarkdownContent, item.EditorJSON, item.SectionType, item.Status, item.UpdatedAt, item.ID,
	)
	if err != nil {
		return model.Chapter{}, fmt.Errorf("update chapter: %w", err)
	}

	saveMode := normalizeSaveMode(input.SaveMode)
	if shouldPersistAutosaveDraft(saveMode, input) {
		if _, err := s.createAutosaveDraftRecord(ctx, tx, item, createAutosaveDraftRecordInput{
			Reason:    defaultAutosaveReason(saveMode, input.AutosaveReason),
			SessionID: strings.TrimSpace(input.SessionID),
		}); err != nil {
			return model.Chapter{}, err
		}
	}

	if shouldCreateRevisionForSave(saveMode, input) {
		revisionType := revisionTypeForSave(saveMode, input.RevisionType)
		if _, err := s.createRevisionRecord(ctx, tx, item, createRevisionRecordInput{
			RevisionType: revisionType,
			Title:        strings.TrimSpace(input.RevisionTitle),
			Description:  strings.TrimSpace(input.RevisionDescription),
			SessionID:    strings.TrimSpace(input.SessionID),
			CreatedBy:    strings.TrimSpace(input.CreatedBy),
			EventType:    eventTypeForRevisionType(revisionType),
			EventTitle:   revisionEventTitleForSave(saveMode, item.Title, input.RevisionTitle),
		}); err != nil {
			return model.Chapter{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.Chapter{}, fmt.Errorf("commit update chapter tx: %w", err)
	}

	if err := s.syncChapterSnapshot(ctx, item); err != nil {
		return model.Chapter{}, err
	}
	return item, nil
}

func (s *Store) ReorderChapters(ctx context.Context, bookID string, input ReorderChaptersInput) ([]model.Chapter, error) {
	chapters, err := s.ListChapters(ctx, bookID)
	if err != nil {
		return nil, err
	}
	if len(chapters) == 0 {
		return nil, nil
	}
	if len(input.ChapterIDs) != len(chapters) {
		return nil, fmt.Errorf("chapter_ids must contain every chapter exactly once")
	}

	existingIDs := make(map[string]model.Chapter, len(chapters))
	for _, chapter := range chapters {
		existingIDs[chapter.ID] = chapter
	}

	seen := make(map[string]bool, len(input.ChapterIDs))
	for _, id := range input.ChapterIDs {
		if _, ok := existingIDs[id]; !ok {
			return nil, fmt.Errorf("chapter %s does not belong to book", id)
		}
		if seen[id] {
			return nil, fmt.Errorf("chapter_ids must not contain duplicates")
		}
		seen[id] = true
	}

	now := nowUTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin reorder chapters tx: %w", err)
	}
	defer tx.Rollback()

	for index, id := range input.ChapterIDs {
		if _, err := tx.ExecContext(ctx, `UPDATE chapters SET position = ?, updated_at = ? WHERE id = ? AND book_id = ?`,
			1000+index, now, id, bookID,
		); err != nil {
			return nil, fmt.Errorf("stage chapter reorder: %w", err)
		}
	}

	for index, id := range input.ChapterIDs {
		if _, err := tx.ExecContext(ctx, `UPDATE chapters SET position = ?, updated_at = ? WHERE id = ? AND book_id = ?`,
			index+1, now, id, bookID,
		); err != nil {
			return nil, fmt.Errorf("apply chapter reorder: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `UPDATE books SET updated_at = ? WHERE id = ?`, now, bookID); err != nil {
		return nil, fmt.Errorf("touch book: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit reorder chapters: %w", err)
	}

	if err := s.syncBookChapterSnapshots(ctx, bookID); err != nil {
		return nil, err
	}

	return s.ListChapters(ctx, bookID)
}

func (s *Store) ListWorkflowBoxes(ctx context.Context, bookID string) ([]model.WorkflowBox, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, book_id, title, type, tags_json, position, is_collapsed, created_at, updated_at FROM workflow_boxes WHERE book_id = ? ORDER BY position ASC`, bookID)
	if err != nil {
		return nil, fmt.Errorf("list workflow boxes: %w", err)
	}
	defer rows.Close()

	var items []model.WorkflowBox
	for rows.Next() {
		var item model.WorkflowBox
		var collapsed int
		var tagsJSON string
		if err := rows.Scan(&item.ID, &item.BookID, &item.Title, &item.Type, &tagsJSON, &item.Position, &collapsed, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan workflow box: %w", err)
		}
		item.Tags = decodeTags(tagsJSON)
		item.IsCollapsed = collapsed == 1
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CreateWorkflowBox(ctx context.Context, bookID string, input CreateWorkflowBoxInput) (model.WorkflowBox, error) {
	var nextPosition int
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(position), 0) + 1 FROM workflow_boxes WHERE book_id = ?`, bookID).Scan(&nextPosition); err != nil {
		return model.WorkflowBox{}, fmt.Errorf("next workflow position: %w", err)
	}
	return s.createWorkflowBoxWithPosition(ctx, bookID, input, nextPosition)
}

func (s *Store) createWorkflowBoxWithPosition(ctx context.Context, bookID string, input CreateWorkflowBoxInput, position int) (model.WorkflowBox, error) {
	now := nowUTC()
	item := model.WorkflowBox{
		ID:          newID(),
		BookID:      bookID,
		Title:       fallback(strings.TrimSpace(input.Title), "Neue Workflow-Box"),
		Type:        normalizeWorkflowType(input.Type),
		Tags:        normalizeTags(input.Tags),
		Position:    position,
		IsCollapsed: input.IsCollapsed,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO workflow_boxes (id, book_id, title, type, tags_json, position, is_collapsed, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.BookID, item.Title, item.Type, encodeTags(item.Tags), item.Position, boolToInt(item.IsCollapsed), item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return model.WorkflowBox{}, fmt.Errorf("create workflow box: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateWorkflowBox(ctx context.Context, id string, input UpdateWorkflowBoxInput) (model.WorkflowBox, error) {
	var item model.WorkflowBox
	var collapsed int
	var tagsJSON string
	err := s.db.QueryRowContext(ctx, `SELECT id, book_id, title, type, tags_json, position, is_collapsed, created_at, updated_at FROM workflow_boxes WHERE id = ?`, id).
		Scan(&item.ID, &item.BookID, &item.Title, &item.Type, &tagsJSON, &item.Position, &collapsed, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.WorkflowBox{}, ErrNotFound
		}
		return model.WorkflowBox{}, fmt.Errorf("get workflow box: %w", err)
	}
	item.Tags = decodeTags(tagsJSON)
	item.IsCollapsed = collapsed == 1
	item.Title = fallback(strings.TrimSpace(input.Title), item.Title)
	item.Type = normalizeWorkflowType(input.Type)
	item.Tags = normalizeTags(input.Tags)
	item.IsCollapsed = input.IsCollapsed
	item.UpdatedAt = nowUTC()
	_, err = s.db.ExecContext(ctx, `UPDATE workflow_boxes SET title = ?, type = ?, tags_json = ?, is_collapsed = ?, updated_at = ? WHERE id = ?`,
		item.Title, item.Type, encodeTags(item.Tags), boolToInt(item.IsCollapsed), item.UpdatedAt, item.ID,
	)
	if err != nil {
		return model.WorkflowBox{}, fmt.Errorf("update workflow box: %w", err)
	}
	return item, nil
}

func (s *Store) ListAnchors(ctx context.Context, chapterID string) ([]model.Anchor, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, chapter_id, workflow_box_id, title, anchor_type, selected_text, start_offset, end_offset, context_before, context_after, note, created_at, updated_at FROM anchors WHERE chapter_id = ? ORDER BY created_at DESC`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list anchors: %w", err)
	}
	defer rows.Close()

	var items []model.Anchor
	for rows.Next() {
		var item model.Anchor
		if err := rows.Scan(&item.ID, &item.ChapterID, &item.WorkflowBoxID, &item.Title, &item.AnchorType, &item.SelectedText, &item.StartOffset, &item.EndOffset, &item.ContextBefore, &item.ContextAfter, &item.Note, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan anchor: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CreateAnchor(ctx context.Context, chapterID string, input CreateAnchorInput) (model.Anchor, error) {
	now := nowUTC()
	item := model.Anchor{
		ID:            newID(),
		ChapterID:     chapterID,
		WorkflowBoxID: strings.TrimSpace(input.WorkflowBoxID),
		Title:         strings.TrimSpace(input.Title),
		AnchorType:    normalizeAnchorType(input.AnchorType),
		SelectedText:  strings.TrimSpace(input.SelectedText),
		StartOffset:   max(input.StartOffset, 0),
		EndOffset:     max(input.EndOffset, 0),
		ContextBefore: strings.TrimSpace(input.ContextBefore),
		ContextAfter:  strings.TrimSpace(input.ContextAfter),
		Note:          strings.TrimSpace(input.Note),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if item.WorkflowBoxID == "" {
		return model.Anchor{}, fmt.Errorf("workflow_box_id must not be empty")
	}
	if item.Title == "" {
		item.Title = truncateRunes(item.SelectedText, 48)
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO anchors (id, chapter_id, workflow_box_id, title, anchor_type, selected_text, start_offset, end_offset, context_before, context_after, note, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.ChapterID, item.WorkflowBoxID, item.Title, item.AnchorType, item.SelectedText, item.StartOffset, item.EndOffset, item.ContextBefore, item.ContextAfter, item.Note, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return model.Anchor{}, fmt.Errorf("create anchor: %w", err)
	}
	return item, nil
}

func (s *Store) DeleteAnchor(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM anchors WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete anchor: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete anchor rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListReviewComments(ctx context.Context, chapterID string) ([]model.ReviewComment, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, chapter_id, revision_id, comment_type, author, body, suggested_text, selected_text, start_offset, end_offset, context_before, context_after, status, is_todo_done, created_at, updated_at FROM review_comments WHERE chapter_id = ? ORDER BY created_at DESC`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list review comments: %w", err)
	}
	defer rows.Close()

	var items []model.ReviewComment
	for rows.Next() {
		var item model.ReviewComment
		var todoDone int
		if err := rows.Scan(
			&item.ID,
			&item.ChapterID,
			&item.RevisionID,
			&item.CommentType,
			&item.Author,
			&item.Body,
			&item.SuggestedText,
			&item.SelectedText,
			&item.StartOffset,
			&item.EndOffset,
			&item.ContextBefore,
			&item.ContextAfter,
			&item.Status,
			&todoDone,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan review comment: %w", err)
		}
		item.IsTodoDone = todoDone == 1
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CreateReviewComment(ctx context.Context, chapterID string, input CreateReviewCommentInput) (model.ReviewComment, error) {
	now := nowUTC()
	revisionID, err := s.normalizeReviewCommentRevisionID(ctx, chapterID, input.RevisionID)
	if err != nil {
		return model.ReviewComment{}, err
	}
	item := model.ReviewComment{
		ID:            newID(),
		ChapterID:     chapterID,
		RevisionID:    revisionID,
		CommentType:   normalizeReviewCommentType(input.CommentType),
		Author:        fallback(strings.TrimSpace(input.Author), "Review"),
		Body:          strings.TrimSpace(input.Body),
		SuggestedText: strings.TrimSpace(input.SuggestedText),
		SelectedText:  strings.TrimSpace(input.SelectedText),
		StartOffset:   max(input.StartOffset, 0),
		EndOffset:     max(input.EndOffset, 0),
		ContextBefore: strings.TrimSpace(input.ContextBefore),
		ContextAfter:  strings.TrimSpace(input.ContextAfter),
		Status:        normalizeReviewCommentStatus(input.Status),
		IsTodoDone:    input.IsTodoDone,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if item.Body == "" && item.SuggestedText == "" {
		return model.ReviewComment{}, fmt.Errorf("body or suggested_text must not be empty")
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO review_comments (id, chapter_id, revision_id, comment_type, author, body, suggested_text, selected_text, start_offset, end_offset, context_before, context_after, status, is_todo_done, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.ChapterID, item.RevisionID, item.CommentType, item.Author, item.Body, item.SuggestedText, item.SelectedText, item.StartOffset, item.EndOffset, item.ContextBefore, item.ContextAfter, item.Status, boolToInt(item.IsTodoDone), item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return model.ReviewComment{}, fmt.Errorf("create review comment: %w", err)
	}
	if item.RevisionID != "" {
		if _, err := s.createRevisionEventRecord(ctx, s.db, item.RevisionID, item.ChapterID, createRevisionEventInput{
			EventType:   "review_started",
			EntityType:  "review_comment",
			EntityID:    item.ID,
			Title:       reviewEventTitleForCommentType(item.CommentType),
			Description: reviewEventDescription(item),
			CreatedBy:   item.Author,
			Metadata: map[string]any{
				"comment_type": item.CommentType,
				"status":       item.Status,
			},
		}); err != nil {
			return model.ReviewComment{}, err
		}
	}
	return item, nil
}

func (s *Store) UpdateReviewComment(ctx context.Context, id string, input UpdateReviewCommentInput) (model.ReviewComment, error) {
	var item model.ReviewComment
	var todoDone int
	err := s.db.QueryRowContext(ctx, `SELECT id, chapter_id, revision_id, comment_type, author, body, suggested_text, selected_text, start_offset, end_offset, context_before, context_after, status, is_todo_done, created_at, updated_at FROM review_comments WHERE id = ?`, id).
		Scan(
			&item.ID,
			&item.ChapterID,
			&item.RevisionID,
			&item.CommentType,
			&item.Author,
			&item.Body,
			&item.SuggestedText,
			&item.SelectedText,
			&item.StartOffset,
			&item.EndOffset,
			&item.ContextBefore,
			&item.ContextAfter,
			&item.Status,
			&todoDone,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ReviewComment{}, ErrNotFound
		}
		return model.ReviewComment{}, fmt.Errorf("get review comment: %w", err)
	}

	item.IsTodoDone = todoDone == 1
	previousStatus := item.Status
	item.CommentType = normalizeReviewCommentType(input.CommentType)
	item.Author = fallback(strings.TrimSpace(input.Author), item.Author)
	if strings.TrimSpace(input.RevisionID) != "" || item.RevisionID != "" {
		nextRevisionID, err := s.normalizeReviewCommentRevisionID(ctx, item.ChapterID, input.RevisionID)
		if err != nil {
			return model.ReviewComment{}, err
		}
		if nextRevisionID != "" {
			item.RevisionID = nextRevisionID
		}
	}
	if strings.TrimSpace(input.Body) != "" || strings.TrimSpace(input.SuggestedText) == "" {
		item.Body = strings.TrimSpace(input.Body)
	}
	if strings.TrimSpace(input.SuggestedText) != "" || item.CommentType == "suggestion" || item.CommentType == "delete_request" {
		item.SuggestedText = strings.TrimSpace(input.SuggestedText)
	}
	item.Status = normalizeReviewCommentStatus(input.Status)
	item.IsTodoDone = input.IsTodoDone
	item.UpdatedAt = nowUTC()

	_, err = s.db.ExecContext(ctx, `UPDATE review_comments SET revision_id = ?, comment_type = ?, author = ?, body = ?, suggested_text = ?, status = ?, is_todo_done = ?, updated_at = ? WHERE id = ?`,
		item.RevisionID, item.CommentType, item.Author, item.Body, item.SuggestedText, item.Status, boolToInt(item.IsTodoDone), item.UpdatedAt, item.ID,
	)
	if err != nil {
		return model.ReviewComment{}, fmt.Errorf("update review comment: %w", err)
	}
	if item.RevisionID != "" && previousStatus == "open" && item.Status != "open" {
		if _, err := s.createRevisionEventRecord(ctx, s.db, item.RevisionID, item.ChapterID, createRevisionEventInput{
			EventType:   "review_completed",
			EntityType:  "review_comment",
			EntityID:    item.ID,
			Title:       "Review abgeschlossen",
			Description: reviewEventDescription(item),
			CreatedBy:   item.Author,
			Metadata: map[string]any{
				"comment_type": item.CommentType,
				"status":       item.Status,
			},
		}); err != nil {
			return model.ReviewComment{}, err
		}
	}
	return item, nil
}

func (s *Store) DeleteReviewComment(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM review_comments WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete review comment: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete review comment rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListClipboardItems(ctx context.Context, bookID string) ([]model.ClipboardItem, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, book_id, chapter_id, content, content_type, source_anchor_id, slot, is_pinned, created_at, updated_at FROM clipboard_items WHERE book_id = ? ORDER BY is_pinned DESC, slot ASC, updated_at DESC`, bookID)
	if err != nil {
		return nil, fmt.Errorf("list clipboard items: %w", err)
	}
	defer rows.Close()

	var items []model.ClipboardItem
	for rows.Next() {
		var item model.ClipboardItem
		var pinned int
		if err := rows.Scan(&item.ID, &item.BookID, &item.ChapterID, &item.Content, &item.ContentType, &item.SourceAnchorID, &item.Slot, &pinned, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan clipboard item: %w", err)
		}
		item.IsPinned = pinned == 1
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CreateClipboardItem(ctx context.Context, bookID string, input CreateClipboardItemInput) (model.ClipboardItem, error) {
	now := nowUTC()
	item := model.ClipboardItem{
		ID:             newID(),
		BookID:         bookID,
		ChapterID:      strings.TrimSpace(input.ChapterID),
		Content:        strings.TrimSpace(input.Content),
		ContentType:    fallback(strings.TrimSpace(input.ContentType), "text/markdown"),
		SourceAnchorID: strings.TrimSpace(input.SourceAnchorID),
		Slot:           normalizeSlot(input.Slot),
		IsPinned:       input.IsPinned,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if item.Content == "" {
		return model.ClipboardItem{}, fmt.Errorf("content must not be empty")
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO clipboard_items (id, book_id, chapter_id, content, content_type, source_anchor_id, slot, is_pinned, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.BookID, item.ChapterID, item.Content, item.ContentType, item.SourceAnchorID, item.Slot, boolToInt(item.IsPinned), item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return model.ClipboardItem{}, fmt.Errorf("create clipboard item: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateClipboardItem(ctx context.Context, id string, input UpdateClipboardItemInput) (model.ClipboardItem, error) {
	var item model.ClipboardItem
	var pinned int
	err := s.db.QueryRowContext(ctx, `SELECT id, book_id, chapter_id, content, content_type, source_anchor_id, slot, is_pinned, created_at, updated_at FROM clipboard_items WHERE id = ?`, id).
		Scan(&item.ID, &item.BookID, &item.ChapterID, &item.Content, &item.ContentType, &item.SourceAnchorID, &item.Slot, &pinned, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ClipboardItem{}, ErrNotFound
		}
		return model.ClipboardItem{}, fmt.Errorf("get clipboard item: %w", err)
	}
	item.IsPinned = pinned == 1
	if trimmed := strings.TrimSpace(input.Content); trimmed != "" {
		item.Content = trimmed
	}
	item.Slot = normalizeSlot(input.Slot)
	item.IsPinned = input.IsPinned
	item.UpdatedAt = nowUTC()
	_, err = s.db.ExecContext(ctx, `UPDATE clipboard_items SET content = ?, slot = ?, is_pinned = ?, updated_at = ? WHERE id = ?`,
		item.Content, item.Slot, boolToInt(item.IsPinned), item.UpdatedAt, item.ID,
	)
	if err != nil {
		return model.ClipboardItem{}, fmt.Errorf("update clipboard item: %w", err)
	}
	return item, nil
}

func (s *Store) DeleteClipboardItem(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM clipboard_items WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete clipboard item: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete clipboard item rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListKnowledgeItems(ctx context.Context, projectID string) ([]model.KnowledgeItem, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, project_id, type, name, summary, body, tags_json, created_at, updated_at FROM knowledge_items WHERE project_id = ? ORDER BY type ASC, name COLLATE NOCASE ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list knowledge items: %w", err)
	}
	defer rows.Close()

	var items []model.KnowledgeItem
	for rows.Next() {
		item, err := scanKnowledgeItem(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CreateKnowledgeItem(ctx context.Context, projectID string, input CreateKnowledgeItemInput) (model.KnowledgeItem, error) {
	if _, _, err := s.GetProject(ctx, projectID); err != nil {
		return model.KnowledgeItem{}, err
	}

	now := nowUTC()
	item := model.KnowledgeItem{
		ID:        newID(),
		ProjectID: projectID,
		Type:      normalizeKnowledgeType(input.Type),
		Name:      fallback(strings.TrimSpace(input.Name), "Neuer Wissenseintrag"),
		Summary:   strings.TrimSpace(input.Summary),
		Body:      strings.TrimSpace(input.Body),
		Tags:      normalizeTags(input.Tags),
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := s.db.ExecContext(ctx, `INSERT INTO knowledge_items (id, project_id, type, name, summary, body, tags_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.ProjectID, item.Type, item.Name, item.Summary, item.Body, encodeTags(item.Tags), item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return model.KnowledgeItem{}, fmt.Errorf("create knowledge item: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE projects SET updated_at = ? WHERE id = ?`, now, projectID); err != nil {
		return model.KnowledgeItem{}, fmt.Errorf("touch project: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateKnowledgeItem(ctx context.Context, id string, input UpdateKnowledgeItemInput) (model.KnowledgeItem, error) {
	item, err := s.getKnowledgeItem(ctx, id)
	if err != nil {
		return model.KnowledgeItem{}, err
	}

	item.Type = normalizeKnowledgeType(input.Type)
	item.Name = fallback(strings.TrimSpace(input.Name), item.Name)
	item.Summary = strings.TrimSpace(input.Summary)
	item.Body = strings.TrimSpace(input.Body)
	item.Tags = normalizeTags(input.Tags)
	item.UpdatedAt = nowUTC()

	_, err = s.db.ExecContext(ctx, `UPDATE knowledge_items SET type = ?, name = ?, summary = ?, body = ?, tags_json = ?, updated_at = ? WHERE id = ?`,
		item.Type, item.Name, item.Summary, item.Body, encodeTags(item.Tags), item.UpdatedAt, item.ID,
	)
	if err != nil {
		return model.KnowledgeItem{}, fmt.Errorf("update knowledge item: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE projects SET updated_at = ? WHERE id = ?`, item.UpdatedAt, item.ProjectID); err != nil {
		return model.KnowledgeItem{}, fmt.Errorf("touch project: %w", err)
	}
	return item, nil
}

func (s *Store) getKnowledgeItem(ctx context.Context, id string) (model.KnowledgeItem, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, project_id, type, name, summary, body, tags_json, created_at, updated_at FROM knowledge_items WHERE id = ?`, id)
	item, err := scanKnowledgeItem(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.KnowledgeItem{}, ErrNotFound
		}
		return model.KnowledgeItem{}, err
	}
	return item, nil
}

func scanKnowledgeItem(scan func(dest ...any) error) (model.KnowledgeItem, error) {
	var item model.KnowledgeItem
	var tagsJSON string
	if err := scan(&item.ID, &item.ProjectID, &item.Type, &item.Name, &item.Summary, &item.Body, &tagsJSON, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return model.KnowledgeItem{}, fmt.Errorf("scan knowledge item: %w", err)
	}
	item.Tags = decodeTags(tagsJSON)
	return item, nil
}

func (s *Store) syncChapterSnapshot(ctx context.Context, chapter model.Chapter) error {
	projectID, err := s.lookupBookProjectID(ctx, chapter.BookID)
	if err != nil {
		return err
	}
	targetDir := filepath.Join(s.libraryDir, projectID, chapter.BookID, "chapters")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("mkdir chapter snapshot dir: %w", err)
	}
	return s.writeChapterSnapshotFile(filepath.Join(targetDir, fmt.Sprintf("%03d-%s.md", chapter.Position, chapter.ID[:8])), chapter)
}

func (s *Store) syncBookChapterSnapshots(ctx context.Context, bookID string) error {
	projectID, err := s.lookupBookProjectID(ctx, bookID)
	if err != nil {
		return err
	}
	targetDir := filepath.Join(s.libraryDir, projectID, bookID, "chapters")
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("reset chapter snapshot dir: %w", err)
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("mkdir chapter snapshot dir: %w", err)
	}

	chapters, err := s.ListChapters(ctx, bookID)
	if err != nil {
		return err
	}
	for _, chapter := range chapters {
		target := filepath.Join(targetDir, fmt.Sprintf("%03d-%s.md", chapter.Position, chapter.ID[:8]))
		if err := s.writeChapterSnapshotFile(target, chapter); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) lookupBookProjectID(ctx context.Context, bookID string) (string, error) {
	var projectID string
	err := s.db.QueryRowContext(ctx, `SELECT books.project_id FROM books WHERE books.id = ?`, bookID).Scan(&projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("lookup chapter project: %w", err)
	}
	return projectID, nil
}

func (s *Store) writeChapterSnapshotFile(target string, chapter model.Chapter) error {
	content := strings.TrimSpace(chapter.MarkdownContent)
	if content == "" {
		content = "# " + chapter.Title + "\n"
	}
	if strings.TrimSpace(chapter.Summary) != "" {
		content = fmt.Sprintf("<!-- summary: %s -->\n\n%s", strings.TrimSpace(chapter.Summary), content)
	}
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write chapter snapshot: %w", err)
	}
	return nil
}

func fallback(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func normalizeProjectStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "paused", "archived":
		return strings.TrimSpace(value)
	default:
		return "active"
	}
}

func normalizeVisibility(value string) string {
	switch strings.TrimSpace(value) {
	case "shared", "registered":
		return "shared"
	case "public":
		return "public"
	default:
		return "private"
	}
}

func normalizeWorkType(value string) string {
	switch strings.TrimSpace(value) {
	case "series", "freebie", "course", "article-series", "ghost-series":
		return strings.TrimSpace(value)
	default:
		return "book"
	}
}

func normalizeSectionType(value string) string {
	switch strings.TrimSpace(value) {
	case "front_matter", "back_matter", "fragment":
		return strings.TrimSpace(value)
	default:
		return "body"
	}
}

func normalizeChapterStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "idea", "revision", "review", "final", "archived":
		return strings.TrimSpace(value)
	default:
		return "draft"
	}
}

func normalizeReusablePageType(value string) string {
	switch strings.TrimSpace(value) {
	case "author_bio", "copyright", "imprint", "dedication", "review_request", "newsletter", "also_by", "donation_note", "publisher_contact":
		return strings.TrimSpace(value)
	default:
		return "custom"
	}
}

func normalizeThemeTarget(value string) string {
	switch strings.TrimSpace(value) {
	case "pdf", "epub", "docx", "web":
		return strings.TrimSpace(value)
	default:
		return "general"
	}
}

func normalizeWorkflowType(value string) string {
	switch strings.TrimSpace(value) {
	case "notes", "persons", "events", "threads", "reminders", "research", "clipboard", "custom":
		return strings.TrimSpace(value)
	default:
		return "notes"
	}
}

func normalizeAnchorType(value string) string {
	switch strings.TrimSpace(value) {
	case "sentence", "passage", "invisible", "manual":
		return strings.TrimSpace(value)
	default:
		return "passage"
	}
}

func normalizeReviewCommentType(value string) string {
	switch strings.TrimSpace(value) {
	case "comment", "todo", "suggestion", "delete_request", "warning":
		return strings.TrimSpace(value)
	default:
		return "comment"
	}
}

func (s *Store) normalizeReviewCommentRevisionID(ctx context.Context, chapterID, revisionID string) (string, error) {
	revisionID = strings.TrimSpace(revisionID)
	if revisionID == "" {
		return "", nil
	}
	revision, err := s.GetRevision(ctx, revisionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", fmt.Errorf("revision_id is invalid")
		}
		return "", err
	}
	if revision.ChapterID != chapterID {
		return "", fmt.Errorf("revision_id does not belong to chapter")
	}
	return revisionID, nil
}

func normalizeReviewCommentStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "open", "resolved", "applied", "rejected":
		return strings.TrimSpace(value)
	default:
		return "open"
	}
}

func reviewEventTitleForCommentType(value string) string {
	switch normalizeReviewCommentType(value) {
	case "suggestion":
		return "Korrekturvorschlag erstellt"
	case "todo":
		return "Review-To-do erstellt"
	case "delete_request":
		return "Loeschbitte erstellt"
	case "warning":
		return "Review-Hinweis erstellt"
	default:
		return "Kommentar erstellt"
	}
}

func reviewEventDescription(comment model.ReviewComment) string {
	if strings.TrimSpace(comment.Body) != "" {
		return strings.TrimSpace(comment.Body)
	}
	if strings.TrimSpace(comment.SuggestedText) != "" {
		return strings.TrimSpace(comment.SuggestedText)
	}
	return strings.TrimSpace(comment.SelectedText)
}

func normalizeSlot(value int) int {
	if value < 1 || value > 9 {
		return 0
	}
	return value
}

func normalizeKnowledgeType(value string) string {
	switch strings.TrimSpace(value) {
	case "person", "location", "event", "thread", "motif", "term", "reminder", "research_note", "custom":
		return strings.TrimSpace(value)
	default:
		return "person"
	}
}

func normalizeTags(values []string) []string {
	var tags []string
	seen := make(map[string]struct{})
	for _, value := range values {
		tag := strings.TrimSpace(value)
		if tag == "" {
			continue
		}
		key := strings.ToLower(tag)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		tags = append(tags, tag)
	}
	return tags
}

func encodeTags(values []string) string {
	encoded, err := json.Marshal(normalizeTags(values))
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func decodeTags(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var tags []string
	if err := json.Unmarshal([]byte(value), &tags); err != nil {
		return nil
	}
	return normalizeTags(tags)
}

func truncateRunes(value string, limit int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit])
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func newID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buffer)
}

func max(left, right int) int {
	if left > right {
		return left
	}
	return right
}

var ErrNotFound = errors.New("not found")
