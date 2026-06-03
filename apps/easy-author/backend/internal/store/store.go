package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
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
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CreateBookInput struct {
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	Author     string `json:"author"`
	Visibility string `json:"visibility"`
}

type CreateChapterInput struct {
	Title           string `json:"title"`
	MarkdownContent string `json:"markdown_content"`
	EditorJSON      string `json:"editor_json"`
}

type UpdateChapterInput struct {
	Title           string `json:"title"`
	MarkdownContent string `json:"markdown_content"`
	EditorJSON      string `json:"editor_json"`
}

type CreateWorkflowBoxInput struct {
	Title       string `json:"title"`
	Type        string `json:"type"`
	IsCollapsed bool   `json:"is_collapsed"`
}

type UpdateWorkflowBoxInput struct {
	Title       string `json:"title"`
	Type        string `json:"type"`
	IsCollapsed bool   `json:"is_collapsed"`
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

func New(db *sql.DB, libraryDir string) *Store {
	return &Store{db: db, libraryDir: libraryDir}
}

func (s *Store) Init(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
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
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS chapters (
			id TEXT PRIMARY KEY,
			book_id TEXT NOT NULL,
			title TEXT NOT NULL,
			position INTEGER NOT NULL,
			markdown_content TEXT NOT NULL DEFAULT '',
			editor_json TEXT NOT NULL DEFAULT '',
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
	}

	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("init schema: %w", err)
		}
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
		if _, err := s.createWorkflowBoxWithPosition(ctx, book.ID, CreateWorkflowBoxInput{
			Title:       title,
			Type:        boxType,
			IsCollapsed: index > 1,
		}, index+1); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ListProjects(ctx context.Context) ([]model.Project, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, description, created_at, updated_at FROM projects ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var item model.Project
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
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
		ID:          newID(),
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO projects (id, title, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		item.ID, item.Title, item.Description, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return model.Project{}, fmt.Errorf("create project: %w", err)
	}
	return item, nil
}

func (s *Store) GetProject(ctx context.Context, id string) (model.Project, []model.Book, error) {
	var item model.Project
	err := s.db.QueryRowContext(ctx, `SELECT id, title, description, created_at, updated_at FROM projects WHERE id = ?`, id).
		Scan(&item.ID, &item.Title, &item.Description, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Project{}, nil, ErrNotFound
		}
		return model.Project{}, nil, fmt.Errorf("get project: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `SELECT id, project_id, title, subtitle, author, visibility, created_at, updated_at FROM books WHERE project_id = ? ORDER BY updated_at DESC`, id)
	if err != nil {
		return model.Project{}, nil, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()

	var books []model.Book
	for rows.Next() {
		var book model.Book
		if err := rows.Scan(&book.ID, &book.ProjectID, &book.Title, &book.Subtitle, &book.Author, &book.Visibility, &book.CreatedAt, &book.UpdatedAt); err != nil {
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
		ID:         newID(),
		ProjectID:  projectID,
		Title:      fallback(strings.TrimSpace(input.Title), "Neues Buch"),
		Subtitle:   strings.TrimSpace(input.Subtitle),
		Author:     strings.TrimSpace(input.Author),
		Visibility: visibility,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO books (id, project_id, title, subtitle, author, visibility, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.ProjectID, item.Title, item.Subtitle, item.Author, item.Visibility, item.CreatedAt, item.UpdatedAt,
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
	err := s.db.QueryRowContext(ctx, `SELECT id, project_id, title, subtitle, author, visibility, created_at, updated_at FROM books WHERE id = ?`, bookID).
		Scan(&book.ID, &book.ProjectID, &book.Title, &book.Subtitle, &book.Author, &book.Visibility, &book.CreatedAt, &book.UpdatedAt)
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

	return model.BookBundle{
		Book:          book,
		Chapters:      chapters,
		WorkflowBoxes: boxes,
		Clipboard:     clipboard,
	}, nil
}

func (s *Store) ListChapters(ctx context.Context, bookID string) ([]model.Chapter, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, book_id, title, position, markdown_content, editor_json, created_at, updated_at FROM chapters WHERE book_id = ? ORDER BY position ASC`, bookID)
	if err != nil {
		return nil, fmt.Errorf("list chapters: %w", err)
	}
	defer rows.Close()

	var chapters []model.Chapter
	for rows.Next() {
		var item model.Chapter
		if err := rows.Scan(&item.ID, &item.BookID, &item.Title, &item.Position, &item.MarkdownContent, &item.EditorJSON, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan chapter: %w", err)
		}
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
		ID:              newID(),
		BookID:          bookID,
		Title:           fallback(strings.TrimSpace(input.Title), fmt.Sprintf("Kapitel %d", nextPosition)),
		Position:        nextPosition,
		MarkdownContent: strings.TrimSpace(input.MarkdownContent),
		EditorJSON:      strings.TrimSpace(input.EditorJSON),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO chapters (id, book_id, title, position, markdown_content, editor_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.BookID, item.Title, item.Position, item.MarkdownContent, item.EditorJSON, item.CreatedAt, item.UpdatedAt,
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
	err := s.db.QueryRowContext(ctx, `SELECT id, book_id, title, position, markdown_content, editor_json, created_at, updated_at FROM chapters WHERE id = ?`, id).
		Scan(&item.ID, &item.BookID, &item.Title, &item.Position, &item.MarkdownContent, &item.EditorJSON, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Chapter{}, ErrNotFound
		}
		return model.Chapter{}, fmt.Errorf("get chapter: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateChapter(ctx context.Context, id string, input UpdateChapterInput) (model.Chapter, error) {
	item, err := s.GetChapter(ctx, id)
	if err != nil {
		return model.Chapter{}, err
	}
	item.Title = fallback(strings.TrimSpace(input.Title), item.Title)
	item.MarkdownContent = strings.TrimSpace(input.MarkdownContent)
	item.EditorJSON = strings.TrimSpace(input.EditorJSON)
	item.UpdatedAt = nowUTC()
	_, err = s.db.ExecContext(ctx, `UPDATE chapters SET title = ?, markdown_content = ?, editor_json = ?, updated_at = ? WHERE id = ?`,
		item.Title, item.MarkdownContent, item.EditorJSON, item.UpdatedAt, item.ID,
	)
	if err != nil {
		return model.Chapter{}, fmt.Errorf("update chapter: %w", err)
	}
	if err := s.syncChapterSnapshot(ctx, item); err != nil {
		return model.Chapter{}, err
	}
	return item, nil
}

func (s *Store) ListWorkflowBoxes(ctx context.Context, bookID string) ([]model.WorkflowBox, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, book_id, title, type, position, is_collapsed, created_at, updated_at FROM workflow_boxes WHERE book_id = ? ORDER BY position ASC`, bookID)
	if err != nil {
		return nil, fmt.Errorf("list workflow boxes: %w", err)
	}
	defer rows.Close()

	var items []model.WorkflowBox
	for rows.Next() {
		var item model.WorkflowBox
		var collapsed int
		if err := rows.Scan(&item.ID, &item.BookID, &item.Title, &item.Type, &item.Position, &collapsed, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan workflow box: %w", err)
		}
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
		Position:    position,
		IsCollapsed: input.IsCollapsed,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO workflow_boxes (id, book_id, title, type, position, is_collapsed, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.BookID, item.Title, item.Type, item.Position, boolToInt(item.IsCollapsed), item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return model.WorkflowBox{}, fmt.Errorf("create workflow box: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateWorkflowBox(ctx context.Context, id string, input UpdateWorkflowBoxInput) (model.WorkflowBox, error) {
	var item model.WorkflowBox
	var collapsed int
	err := s.db.QueryRowContext(ctx, `SELECT id, book_id, title, type, position, is_collapsed, created_at, updated_at FROM workflow_boxes WHERE id = ?`, id).
		Scan(&item.ID, &item.BookID, &item.Title, &item.Type, &item.Position, &collapsed, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.WorkflowBox{}, ErrNotFound
		}
		return model.WorkflowBox{}, fmt.Errorf("get workflow box: %w", err)
	}
	item.IsCollapsed = collapsed == 1
	item.Title = fallback(strings.TrimSpace(input.Title), item.Title)
	item.Type = normalizeWorkflowType(input.Type)
	item.IsCollapsed = input.IsCollapsed
	item.UpdatedAt = nowUTC()
	_, err = s.db.ExecContext(ctx, `UPDATE workflow_boxes SET title = ?, type = ?, is_collapsed = ?, updated_at = ? WHERE id = ?`,
		item.Title, item.Type, boolToInt(item.IsCollapsed), item.UpdatedAt, item.ID,
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

func (s *Store) syncChapterSnapshot(ctx context.Context, chapter model.Chapter) error {
	var projectID string
	err := s.db.QueryRowContext(ctx, `SELECT books.project_id FROM books WHERE books.id = ?`, chapter.BookID).Scan(&projectID)
	if err != nil {
		return fmt.Errorf("lookup chapter project: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(s.libraryDir, projectID, chapter.BookID, "chapters"), 0o755); err != nil {
		return fmt.Errorf("mkdir chapter snapshot dir: %w", err)
	}

	filename := fmt.Sprintf("%03d-%s.md", chapter.Position, chapter.ID[:8])
	target := filepath.Join(s.libraryDir, projectID, chapter.BookID, "chapters", filename)

	content := strings.TrimSpace(chapter.MarkdownContent)
	if content == "" {
		content = "# " + chapter.Title + "\n"
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

func normalizeVisibility(value string) string {
	switch strings.TrimSpace(value) {
	case "registered", "public":
		return strings.TrimSpace(value)
	default:
		return "private"
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

func normalizeSlot(value int) int {
	if value < 1 || value > 9 {
		return 0
	}
	return value
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
