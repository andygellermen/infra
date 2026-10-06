package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/model"
)

var contextTypes = map[string]bool{
	"comment": true, "work_item": true, "link": true,
	"clipboard_insert": true, "clipboard_source": true,
}

var threadStatuses = map[string]bool{
	"open": true, "planned": true, "in_progress": true, "review": true, "resolved": true,
}

func (s *Store) CreateContext(ctx context.Context, chapterID string, input CreateContextInput) (model.AnchorContext, error) {
	chapterID = strings.TrimSpace(chapterID)
	contextType := strings.TrimSpace(input.ContextType)
	if chapterID == "" || !contextTypes[contextType] {
		return model.AnchorContext{}, fmt.Errorf("%w: invalid chapter or context type", ErrInvalid)
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "open"
	}
	if !threadStatuses[status] {
		return model.AnchorContext{}, fmt.Errorf("%w: invalid context status", ErrInvalid)
	}
	if contextType == "comment" && (input.Message == nil || strings.TrimSpace(input.Message.Body) == "") {
		return model.AnchorContext{}, fmt.Errorf("%w: comments require an initial message", ErrInvalid)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.AnchorContext{}, err
	}
	defer tx.Rollback()

	anchorID := strings.TrimSpace(input.AnchorID)
	if anchorID != "" {
		var storedChapterID string
		if err := tx.QueryRowContext(ctx, `SELECT chapter_id FROM document_anchors WHERE id = ?`, anchorID).Scan(&storedChapterID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return model.AnchorContext{}, ErrNotFound
			}
			return model.AnchorContext{}, err
		}
		if storedChapterID != chapterID {
			return model.AnchorContext{}, fmt.Errorf("%w: anchor belongs to another chapter", ErrConflict)
		}
	} else {
		anchorID = newID()
		if input.Anchor.StartOffset < 0 || input.Anchor.EndOffset < input.Anchor.StartOffset || strings.TrimSpace(input.Anchor.SelectedText) == "" {
			return model.AnchorContext{}, fmt.Errorf("%w: invalid anchor range", ErrInvalid)
		}
		checksum := strings.TrimSpace(input.Anchor.Checksum)
		if checksum == "" {
			digest := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%d\x00%d\x00%s\x00%s", input.Anchor.SelectedText, input.Anchor.StartOffset, input.Anchor.EndOffset, input.Anchor.ContextBefore, input.Anchor.ContextAfter)))
			checksum = fmt.Sprintf("%x", digest)
		}
		now := nowUTC()
		result, err := tx.ExecContext(ctx, `INSERT INTO document_anchors (id, chapter_id, block_id, selected_text, start_offset, end_offset, context_before, context_after, checksum, created_at, updated_at)
			SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? WHERE EXISTS (SELECT 1 FROM chapters WHERE id = ?)`,
			anchorID, chapterID, strings.TrimSpace(input.Anchor.BlockID), strings.TrimSpace(input.Anchor.SelectedText), input.Anchor.StartOffset, input.Anchor.EndOffset,
			input.Anchor.ContextBefore, input.Anchor.ContextAfter, checksum, now, now, chapterID)
		if err != nil {
			return model.AnchorContext{}, fmt.Errorf("create document anchor: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return model.AnchorContext{}, ErrNotFound
		}
	}

	contextID := newID()
	now := nowUTC()
	if _, err := tx.ExecContext(ctx, `INSERT INTO anchor_contexts (id, anchor_id, context_type, target_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		contextID, anchorID, contextType, strings.TrimSpace(input.TargetID), status, now, now); err != nil {
		return model.AnchorContext{}, fmt.Errorf("create anchor context: %w", err)
	}
	if contextType == "comment" {
		threadID := newID()
		if _, err := tx.ExecContext(ctx, `INSERT INTO comment_threads (id, context_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, threadID, contextID, status, now, now); err != nil {
			return model.AnchorContext{}, fmt.Errorf("create comment thread: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO comment_messages (id, thread_id, author, body, position, created_at) VALUES (?, ?, ?, ?, 0, ?)`,
			newID(), threadID, strings.TrimSpace(input.Message.Author), strings.TrimSpace(input.Message.Body), now); err != nil {
			return model.AnchorContext{}, fmt.Errorf("create comment message: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return model.AnchorContext{}, err
	}
	return s.getContext(ctx, contextID)
}

func (s *Store) ListChapterContexts(ctx context.Context, chapterID string) ([]model.AnchorContext, error) {
	var chapterExists int
	if err := s.db.QueryRowContext(ctx, `SELECT 1 FROM chapters WHERE id = ?`, chapterID).Scan(&chapterExists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT c.id FROM anchor_contexts c JOIN document_anchors a ON a.id = c.anchor_id WHERE a.chapter_id = ? AND c.status <> 'deleted' ORDER BY a.start_offset, a.end_offset, c.created_at, c.id`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list chapter contexts: %w", err)
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	items := make([]model.AnchorContext, 0, len(ids))
	for _, id := range ids {
		item, err := s.getContext(ctx, id)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) getContext(ctx context.Context, id string) (model.AnchorContext, error) {
	var item model.AnchorContext
	err := s.db.QueryRowContext(ctx, `SELECT c.id, c.anchor_id, c.context_type, c.target_id, c.status, c.legacy_entity_id, c.created_at, c.updated_at,
		a.id, a.chapter_id, a.block_id, a.selected_text, a.start_offset, a.end_offset, a.context_before, a.context_after, a.checksum, a.created_at, a.updated_at
		FROM anchor_contexts c JOIN document_anchors a ON a.id = c.anchor_id WHERE c.id = ?`, id).Scan(
		&item.ID, &item.AnchorID, &item.ContextType, &item.TargetID, &item.Status, &item.LegacyEntityID, &item.CreatedAt, &item.UpdatedAt,
		&item.Anchor.ID, &item.Anchor.ChapterID, &item.Anchor.BlockID, &item.Anchor.SelectedText, &item.Anchor.StartOffset, &item.Anchor.EndOffset,
		&item.Anchor.ContextBefore, &item.Anchor.ContextAfter, &item.Anchor.Checksum, &item.Anchor.CreatedAt, &item.Anchor.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.AnchorContext{}, ErrNotFound
	}
	if err != nil {
		return model.AnchorContext{}, fmt.Errorf("get context: %w", err)
	}
	if item.ContextType == "comment" {
		thread, err := s.getThreadByContext(ctx, item.ID)
		if err != nil {
			return model.AnchorContext{}, err
		}
		item.Thread = thread
		item.Status = thread.Status
	}
	return item, nil
}

func (s *Store) getThreadByContext(ctx context.Context, contextID string) (model.CommentThread, error) {
	var thread model.CommentThread
	err := s.db.QueryRowContext(ctx, `SELECT id, context_id, status, created_at, updated_at FROM comment_threads WHERE context_id = ?`, contextID).Scan(
		&thread.ID, &thread.ContextID, &thread.Status, &thread.CreatedAt, &thread.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.CommentThread{}, ErrNotFound
	}
	if err != nil {
		return model.CommentThread{}, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, thread_id, author, body, created_at FROM comment_messages WHERE thread_id = ? ORDER BY position, id`, thread.ID)
	if err != nil {
		return model.CommentThread{}, err
	}
	defer rows.Close()
	thread.Messages = []model.CommentMessage{}
	for rows.Next() {
		var message model.CommentMessage
		if err := rows.Scan(&message.ID, &message.ThreadID, &message.Author, &message.Body, &message.CreatedAt); err != nil {
			return model.CommentThread{}, err
		}
		thread.Messages = append(thread.Messages, message)
	}
	return thread, rows.Err()
}

func (s *Store) DeleteContext(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var anchorID string
	if err := tx.QueryRowContext(ctx, `SELECT anchor_id FROM anchor_contexts WHERE id = ?`, id).Scan(&anchorID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM anchor_contexts WHERE id = ?`, id); err != nil {
		return err
	}
	var siblings int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM anchor_contexts WHERE anchor_id = ?`, anchorID).Scan(&siblings); err != nil {
		return err
	}
	if siblings == 0 {
		if _, err := tx.ExecContext(ctx, `DELETE FROM document_anchors WHERE id = ?`, anchorID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) CreateThreadMessage(ctx context.Context, threadID string, input CreateThreadMessageInput) (model.CommentMessage, error) {
	body := strings.TrimSpace(input.Body)
	if body == "" {
		return model.CommentMessage{}, fmt.Errorf("%w: message body must not be empty", ErrInvalid)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.CommentMessage{}, err
	}
	defer tx.Rollback()
	var nextPosition int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(position), -1) + 1 FROM comment_messages WHERE thread_id = ?`, threadID).Scan(&nextPosition); err != nil {
		return model.CommentMessage{}, err
	}
	message := model.CommentMessage{ID: newID(), ThreadID: threadID, Author: strings.TrimSpace(input.Author), Body: body, CreatedAt: nowUTC()}
	result, err := tx.ExecContext(ctx, `INSERT INTO comment_messages (id, thread_id, author, body, position, created_at)
		SELECT ?, ?, ?, ?, ?, ? WHERE EXISTS (SELECT 1 FROM comment_threads WHERE id = ?)`, message.ID, threadID, message.Author, message.Body, nextPosition, message.CreatedAt, threadID)
	if err != nil {
		return model.CommentMessage{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return model.CommentMessage{}, ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return model.CommentMessage{}, err
	}
	return message, nil
}

func (s *Store) UpdateThreadStatus(ctx context.Context, threadID, status string) (model.CommentThread, error) {
	status = strings.TrimSpace(status)
	if !threadStatuses[status] {
		return model.CommentThread{}, fmt.Errorf("%w: invalid thread status", ErrInvalid)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.CommentThread{}, err
	}
	defer tx.Rollback()
	now := nowUTC()
	result, err := tx.ExecContext(ctx, `UPDATE comment_threads SET status = ?, updated_at = ? WHERE id = ?`, status, now, threadID)
	if err != nil {
		return model.CommentThread{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return model.CommentThread{}, ErrNotFound
	}
	if _, err := tx.ExecContext(ctx, `UPDATE anchor_contexts SET status = ?, updated_at = ? WHERE id = (SELECT context_id FROM comment_threads WHERE id = ?)`, status, now, threadID); err != nil {
		return model.CommentThread{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.CommentThread{}, err
	}
	var contextID string
	if err := s.db.QueryRowContext(ctx, `SELECT context_id FROM comment_threads WHERE id = ?`, threadID).Scan(&contextID); err != nil {
		return model.CommentThread{}, err
	}
	return s.getThreadByContext(ctx, contextID)
}
