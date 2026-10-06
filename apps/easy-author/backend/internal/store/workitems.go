package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/model"
)

var kanbanPhases = []string{"backlog", "todo", "in_progress", "review", "done"}

func validKanbanPhase(value string) bool {
	for _, phase := range kanbanPhases {
		if value == phase {
			return true
		}
	}
	return false
}

func normalizeWorkItemKind(value string) (string, error) {
	value = strings.TrimSpace(value)
	switch value {
	case "task", "comment", "reminder":
		return value, nil
	default:
		return "", fmt.Errorf("%w: invalid work item kind", ErrInvalid)
	}
}

func (s *Store) CreateWorkItem(ctx context.Context, bookID string, input CreateWorkItemInput) (model.WorkItem, error) {
	kind, err := normalizeWorkItemKind(input.Kind)
	if err != nil {
		return model.WorkItem{}, err
	}
	title := truncateRunes(input.Title, 160)
	if title == "" {
		return model.WorkItem{}, fmt.Errorf("%w: title must not be empty", ErrInvalid)
	}
	phase := strings.TrimSpace(input.Phase)
	if phase == "" {
		phase = "backlog"
	}
	if !validKanbanPhase(phase) || input.Priority < 0 || input.Priority > 3 {
		return model.WorkItem{}, fmt.Errorf("%w: invalid phase or priority", ErrInvalid)
	}
	if input.DueAt != "" {
		if _, err := time.Parse(time.RFC3339, input.DueAt); err != nil {
			return model.WorkItem{}, fmt.Errorf("%w: due_at must be RFC3339", ErrInvalid)
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.WorkItem{}, err
	}
	defer tx.Rollback()
	var projectID string
	if err := tx.QueryRowContext(ctx, `SELECT project_id FROM books WHERE id = ?`, bookID).Scan(&projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.WorkItem{}, ErrNotFound
		}
		return model.WorkItem{}, err
	}
	chapterID := strings.TrimSpace(input.ChapterID)
	anchorID := strings.TrimSpace(input.AnchorID)
	threadID := strings.TrimSpace(input.ThreadID)
	if threadID != "" {
		var linkedBookID, linkedChapterID, linkedAnchorID string
		err := tx.QueryRowContext(ctx, `SELECT ch.book_id, ch.id, a.id FROM comment_threads t
			JOIN anchor_contexts c ON c.id = t.context_id JOIN document_anchors a ON a.id = c.anchor_id
			JOIN chapters ch ON ch.id = a.chapter_id WHERE t.id = ?`, threadID).Scan(&linkedBookID, &linkedChapterID, &linkedAnchorID)
		if errors.Is(err, sql.ErrNoRows) {
			return model.WorkItem{}, ErrNotFound
		}
		if err != nil {
			return model.WorkItem{}, err
		}
		if linkedBookID != bookID {
			return model.WorkItem{}, fmt.Errorf("%w: thread belongs to another book", ErrConflict)
		}
		chapterID, anchorID = linkedChapterID, linkedAnchorID
	}
	if chapterID != "" {
		var chapterBookID string
		if err := tx.QueryRowContext(ctx, `SELECT book_id FROM chapters WHERE id = ?`, chapterID).Scan(&chapterBookID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return model.WorkItem{}, ErrNotFound
			}
			return model.WorkItem{}, err
		}
		if chapterBookID != bookID {
			return model.WorkItem{}, fmt.Errorf("%w: chapter belongs to another book", ErrConflict)
		}
	}
	now := nowUTC()
	item := model.WorkItem{ID: newID(), ProjectID: projectID, BookID: bookID, ChapterID: chapterID, AnchorID: anchorID, ThreadID: threadID, Kind: kind, Title: title, Phase: phase, Priority: input.Priority, DueAt: input.DueAt, CreatedAt: now, UpdatedAt: now}
	_, err = tx.ExecContext(ctx, `INSERT INTO work_items (id, project_id, book_id, chapter_id, anchor_id, thread_id, kind, title, phase, priority, due_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.ProjectID, item.BookID, item.ChapterID, item.AnchorID, item.ThreadID, item.Kind, item.Title, item.Phase, item.Priority, item.DueAt, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return model.WorkItem{}, fmt.Errorf("%w: thread already has a work item", ErrConflict)
		}
		return model.WorkItem{}, fmt.Errorf("create work item: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return model.WorkItem{}, err
	}
	return item, nil
}

func (s *Store) ListKanban(ctx context.Context, query KanbanQuery) (model.KanbanResult, error) {
	limit := query.LimitPerPhase
	if limit == 0 {
		limit = 12
	}
	if limit < 1 || limit > 20 {
		return model.KanbanResult{}, fmt.Errorf("%w: limit_per_phase must be between 1 and 20", ErrInvalid)
	}
	items := make(map[string][]model.WorkItem, len(kanbanPhases))
	totals := make(map[string]int, len(kanbanPhases))
	for _, phase := range kanbanPhases {
		items[phase] = []model.WorkItem{}
		totals[phase] = 0
	}
	clauses := []string{"1 = 1"}
	args := []any{}
	if len(query.BookIDs) > 0 {
		placeholders := make([]string, len(query.BookIDs))
		for index, id := range query.BookIDs {
			if strings.TrimSpace(id) == "" {
				return model.KanbanResult{}, fmt.Errorf("%w: empty book id", ErrInvalid)
			}
			placeholders[index] = "?"
			args = append(args, id)
		}
		clauses = append(clauses, "book_id IN ("+strings.Join(placeholders, ",")+")")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, project_id, book_id, chapter_id, anchor_id, thread_id, kind, title, phase, priority, due_at, created_at, updated_at
		FROM work_items WHERE `+strings.Join(clauses, " AND ")+`
		ORDER BY phase, priority DESC, CASE WHEN due_at <> '' AND due_at < ? THEN 0 ELSE 1 END, updated_at DESC, created_at, id`, append(args, nowUTC())...)
	if err != nil {
		return model.KanbanResult{}, fmt.Errorf("list kanban: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item model.WorkItem
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.BookID, &item.ChapterID, &item.AnchorID, &item.ThreadID, &item.Kind, &item.Title, &item.Phase, &item.Priority, &item.DueAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return model.KanbanResult{}, err
		}
		totals[item.Phase]++
		if item.Phase == "done" && !query.IncludeDone {
			continue
		}
		if len(items[item.Phase]) < limit {
			items[item.Phase] = append(items[item.Phase], item)
		}
	}
	if err := rows.Err(); err != nil {
		return model.KanbanResult{}, err
	}
	return model.KanbanResult{Items: items, Totals: totals}, nil
}

func (s *Store) MoveWorkItem(ctx context.Context, id, phase string) (model.WorkItem, error) {
	phase = strings.TrimSpace(phase)
	if !validKanbanPhase(phase) {
		return model.WorkItem{}, fmt.Errorf("%w: invalid kanban phase", ErrInvalid)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.WorkItem{}, err
	}
	defer tx.Rollback()
	item, err := scanWorkItem(tx.QueryRowContext(ctx, `SELECT id, project_id, book_id, chapter_id, anchor_id, thread_id, kind, title, phase, priority, due_at, created_at, updated_at FROM work_items WHERE id = ?`, id))
	if err != nil {
		return model.WorkItem{}, err
	}
	now := nowUTC()
	if _, err := tx.ExecContext(ctx, `UPDATE work_items SET phase = ?, updated_at = ? WHERE id = ?`, phase, now, id); err != nil {
		return model.WorkItem{}, err
	}
	if item.ThreadID != "" {
		threadStatus := map[string]string{"backlog": "open", "todo": "planned", "in_progress": "in_progress", "review": "review", "done": "resolved"}[phase]
		if _, err := tx.ExecContext(ctx, `UPDATE comment_threads SET status = ?, updated_at = ? WHERE id = ?`, threadStatus, now, item.ThreadID); err != nil {
			return model.WorkItem{}, fmt.Errorf("update linked thread: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE anchor_contexts SET status = ?, updated_at = ? WHERE id = (SELECT context_id FROM comment_threads WHERE id = ?)`, threadStatus, now, item.ThreadID); err != nil {
			return model.WorkItem{}, fmt.Errorf("update linked context: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return model.WorkItem{}, err
	}
	item.Phase, item.UpdatedAt = phase, now
	return item, nil
}

func scanWorkItem(row *sql.Row) (model.WorkItem, error) {
	var item model.WorkItem
	if err := row.Scan(&item.ID, &item.ProjectID, &item.BookID, &item.ChapterID, &item.AnchorID, &item.ThreadID, &item.Kind, &item.Title, &item.Phase, &item.Priority, &item.DueAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.WorkItem{}, ErrNotFound
		}
		return model.WorkItem{}, err
	}
	return item, nil
}
