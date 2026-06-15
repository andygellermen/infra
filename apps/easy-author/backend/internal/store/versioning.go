package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/model"
)

type RestoreRevisionResult struct {
	Chapter            model.Chapter  `json:"chapter"`
	ProtectionRevision model.Revision `json:"protection_revision"`
	RestoredRevision   model.Revision `json:"restored_revision"`
}

type createRevisionRecordInput struct {
	RevisionType string
	Title        string
	Description  string
	SessionID    string
	CreatedBy    string
	EventType    string
	EventTitle   string
}

type createAutosaveDraftRecordInput struct {
	Reason    string
	SessionID string
}

type createRevisionEventInput struct {
	EventType   string
	EntityType  string
	EntityID    string
	Title       string
	Description string
	Metadata    any
	CreatedBy   string
}

type sqlQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type sqlRunner interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (s *Store) ListRevisionsByChapter(ctx context.Context, chapterID string) ([]model.Revision, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, chapter_id, revision_type, title, description, markdown_content, editor_json, word_count, added_words, removed_words, change_summary, session_id, created_by, created_at FROM revisions WHERE chapter_id = ? ORDER BY created_at DESC, rowid DESC`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list revisions: %w", err)
	}
	defer rows.Close()

	revisions := make([]model.Revision, 0)
	for rows.Next() {
		item, err := scanRevision(rows.Scan)
		if err != nil {
			return nil, err
		}
		revisions = append(revisions, item)
	}
	return revisions, rows.Err()
}

func (s *Store) ListAutosaveDrafts(ctx context.Context, chapterID string) ([]model.AutosaveDraft, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, chapter_id, markdown_content, editor_json, reason, session_id, word_count, created_at, expires_at FROM autosave_drafts WHERE chapter_id = ? ORDER BY created_at DESC, rowid DESC`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list autosave drafts: %w", err)
	}
	defer rows.Close()

	drafts := make([]model.AutosaveDraft, 0)
	for rows.Next() {
		item, err := scanAutosaveDraft(rows.Scan)
		if err != nil {
			return nil, err
		}
		drafts = append(drafts, item)
	}
	return drafts, rows.Err()
}

func (s *Store) CreateRevision(ctx context.Context, chapterID string, input CreateRevisionInput) (model.Revision, error) {
	chapter, err := s.GetChapter(ctx, chapterID)
	if err != nil {
		return model.Revision{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Revision{}, fmt.Errorf("begin create revision tx: %w", err)
	}
	defer tx.Rollback()

	revision, err := s.createRevisionRecord(ctx, tx, chapter, createRevisionRecordInput{
		RevisionType: normalizeRevisionType(input.RevisionType),
		Title:        strings.TrimSpace(input.Title),
		Description:  strings.TrimSpace(input.Description),
		SessionID:    strings.TrimSpace(input.SessionID),
		CreatedBy:    strings.TrimSpace(input.CreatedBy),
		EventType:    eventTypeForRevisionType(input.RevisionType),
		EventTitle:   fallback(strings.TrimSpace(input.Title), "Revision erstellt"),
	})
	if err != nil {
		return model.Revision{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.Revision{}, fmt.Errorf("commit create revision tx: %w", err)
	}
	return revision, nil
}

func (s *Store) GetRevision(ctx context.Context, id string) (model.Revision, error) {
	return s.getRevision(ctx, s.db, id)
}

func (s *Store) RestoreRevision(ctx context.Context, revisionID string, input RestoreRevisionInput) (RestoreRevisionResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RestoreRevisionResult{}, fmt.Errorf("begin restore revision tx: %w", err)
	}
	defer tx.Rollback()

	targetRevision, err := s.getRevision(ctx, tx, revisionID)
	if err != nil {
		return RestoreRevisionResult{}, err
	}

	chapter, err := s.getChapter(ctx, tx, targetRevision.ChapterID)
	if err != nil {
		return RestoreRevisionResult{}, err
	}

	protectionRevision, err := s.createRevisionRecord(ctx, tx, chapter, createRevisionRecordInput{
		RevisionType: "system",
		Title:        "Sicherungsstand vor Wiederherstellung",
		Description:  "Automatisch erstellt, bevor eine Revision wiederhergestellt wurde.",
		CreatedBy:    strings.TrimSpace(input.CreatedBy),
		EventType:    "chapter_saved",
		EventTitle:   "Sicherungsstand erstellt",
	})
	if err != nil {
		return RestoreRevisionResult{}, err
	}

	chapter.MarkdownContent = targetRevision.MarkdownContent
	chapter.EditorJSON = targetRevision.EditorJSON
	chapter.UpdatedAt = nowUTC()
	if _, err := tx.ExecContext(ctx, `UPDATE chapters SET markdown_content = ?, editor_json = ?, updated_at = ? WHERE id = ?`,
		chapter.MarkdownContent, chapter.EditorJSON, chapter.UpdatedAt, chapter.ID,
	); err != nil {
		return RestoreRevisionResult{}, fmt.Errorf("restore chapter revision: %w", err)
	}

	restoredRevision, err := s.createRevisionRecord(ctx, tx, chapter, createRevisionRecordInput{
		RevisionType: "restore",
		Title:        "Wiederhergestellter Stand",
		Description:  fmt.Sprintf("Wiederhergestellt aus Revision %s.", targetRevision.ID),
		CreatedBy:    strings.TrimSpace(input.CreatedBy),
		EventType:    "restore_performed",
		EventTitle:   "Wiederherstellung ausgefuehrt",
	})
	if err != nil {
		return RestoreRevisionResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return RestoreRevisionResult{}, fmt.Errorf("commit restore revision tx: %w", err)
	}

	if err := s.syncChapterSnapshot(ctx, chapter); err != nil {
		return RestoreRevisionResult{}, err
	}

	return RestoreRevisionResult{
		Chapter:            chapter,
		ProtectionRevision: protectionRevision,
		RestoredRevision:   restoredRevision,
	}, nil
}

func (s *Store) ListMilestonesByBook(ctx context.Context, bookID string) ([]model.Milestone, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, book_id, chapter_id, revision_id, title, description, milestone_type, locked, created_by, created_at FROM milestones WHERE book_id = ? ORDER BY created_at DESC, rowid DESC`, bookID)
	if err != nil {
		return nil, fmt.Errorf("list milestones: %w", err)
	}
	defer rows.Close()

	var milestones []model.Milestone
	for rows.Next() {
		item, err := scanMilestone(rows.Scan)
		if err != nil {
			return nil, err
		}
		milestones = append(milestones, item)
	}
	return milestones, rows.Err()
}

func (s *Store) CreateMilestone(ctx context.Context, bookID string, input CreateMilestoneInput) (model.Milestone, error) {
	revisionID := strings.TrimSpace(input.RevisionID)
	if revisionID == "" {
		return model.Milestone{}, fmt.Errorf("revision_id must not be empty")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Milestone{}, fmt.Errorf("begin create milestone tx: %w", err)
	}
	defer tx.Rollback()

	revision, err := s.getRevision(ctx, tx, revisionID)
	if err != nil {
		return model.Milestone{}, err
	}
	chapter, err := s.getChapter(ctx, tx, revision.ChapterID)
	if err != nil {
		return model.Milestone{}, err
	}
	if chapter.BookID != bookID {
		return model.Milestone{}, fmt.Errorf("revision does not belong to book")
	}

	now := nowUTC()
	item := model.Milestone{
		ID:            newID(),
		BookID:        bookID,
		ChapterID:     chapter.ID,
		RevisionID:    revision.ID,
		Title:         fallback(strings.TrimSpace(input.Title), "Milestone"),
		Description:   strings.TrimSpace(input.Description),
		MilestoneType: normalizeMilestoneType(input.MilestoneType),
		Locked:        input.Locked,
		CreatedBy:     strings.TrimSpace(input.CreatedBy),
		CreatedAt:     now,
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO milestones (id, book_id, chapter_id, revision_id, title, description, milestone_type, locked, created_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.BookID, item.ChapterID, item.RevisionID, item.Title, item.Description, item.MilestoneType, boolToInt(item.Locked), item.CreatedBy, item.CreatedAt,
	); err != nil {
		return model.Milestone{}, fmt.Errorf("create milestone: %w", err)
	}

	if _, err := s.createRevisionEventRecord(ctx, tx, revision.ID, chapter.ID, createRevisionEventInput{
		EventType:   "milestone_created",
		EntityType:  "milestone",
		EntityID:    item.ID,
		Title:       "Milestone erstellt",
		Description: item.Title,
		CreatedBy:   item.CreatedBy,
		Metadata: map[string]any{
			"milestone_type": item.MilestoneType,
			"locked":         item.Locked,
		},
	}); err != nil {
		return model.Milestone{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.Milestone{}, fmt.Errorf("commit create milestone tx: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateMilestone(ctx context.Context, id string, input UpdateMilestoneInput) (model.Milestone, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Milestone{}, fmt.Errorf("begin update milestone tx: %w", err)
	}
	defer tx.Rollback()

	item, err := s.getMilestone(ctx, tx, id)
	if err != nil {
		return model.Milestone{}, err
	}

	item.Title = fallback(strings.TrimSpace(input.Title), item.Title)
	item.Description = strings.TrimSpace(input.Description)
	item.MilestoneType = normalizeMilestoneType(input.MilestoneType)
	item.Locked = input.Locked
	createdBy := strings.TrimSpace(input.CreatedBy)

	if _, err := tx.ExecContext(ctx, `UPDATE milestones SET title = ?, description = ?, milestone_type = ?, locked = ? WHERE id = ?`,
		item.Title, item.Description, item.MilestoneType, boolToInt(item.Locked), item.ID,
	); err != nil {
		return model.Milestone{}, fmt.Errorf("update milestone: %w", err)
	}

	if _, err := s.createRevisionEventRecord(ctx, tx, item.RevisionID, item.ChapterID, createRevisionEventInput{
		EventType:   "milestone_updated",
		EntityType:  "milestone",
		EntityID:    item.ID,
		Title:       "Milestone aktualisiert",
		Description: item.Title,
		CreatedBy:   createdBy,
		Metadata: map[string]any{
			"milestone_type": item.MilestoneType,
			"locked":         item.Locked,
		},
	}); err != nil {
		return model.Milestone{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.Milestone{}, fmt.Errorf("commit update milestone tx: %w", err)
	}
	return item, nil
}

func (s *Store) DeleteMilestone(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete milestone tx: %w", err)
	}
	defer tx.Rollback()

	item, err := s.getMilestone(ctx, tx, id)
	if err != nil {
		return err
	}

	if _, err := s.createRevisionEventRecord(ctx, tx, item.RevisionID, item.ChapterID, createRevisionEventInput{
		EventType:   "milestone_deleted",
		EntityType:  "milestone",
		EntityID:    item.ID,
		Title:       "Milestone geloescht",
		Description: item.Title,
		CreatedBy:   item.CreatedBy,
		Metadata: map[string]any{
			"milestone_type": item.MilestoneType,
			"locked":         item.Locked,
		},
	}); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM milestones WHERE id = ?`, item.ID)
	if err != nil {
		return fmt.Errorf("delete milestone: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete milestone rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete milestone tx: %w", err)
	}
	return nil
}

func (s *Store) ListRevisionEvents(ctx context.Context, revisionID string) ([]model.RevisionEvent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, revision_id, chapter_id, event_type, entity_type, entity_id, title, description, metadata_json, created_by, created_at FROM revision_events WHERE revision_id = ? ORDER BY created_at ASC, rowid ASC`, revisionID)
	if err != nil {
		return nil, fmt.Errorf("list revision events: %w", err)
	}
	defer rows.Close()

	events := make([]model.RevisionEvent, 0)
	for rows.Next() {
		item, err := scanRevisionEvent(rows.Scan)
		if err != nil {
			return nil, err
		}
		events = append(events, item)
	}
	return events, rows.Err()
}

func (s *Store) getMilestone(ctx context.Context, queryer sqlQueryer, id string) (model.Milestone, error) {
	row := queryer.QueryRowContext(ctx, `SELECT id, book_id, chapter_id, revision_id, title, description, milestone_type, locked, created_by, created_at FROM milestones WHERE id = ?`, id)
	item, err := scanMilestone(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Milestone{}, ErrNotFound
		}
		return model.Milestone{}, err
	}
	return item, nil
}

func (s *Store) getChapter(ctx context.Context, queryer sqlQueryer, id string) (model.Chapter, error) {
	var item model.Chapter
	var includedExport, visibleTOC, sampleContent int
	err := queryer.QueryRowContext(ctx, `SELECT id, book_id, title, summary, position, markdown_content, editor_json, section_type, status, is_included_in_export, is_visible_in_toc, is_sample_content, created_at, updated_at FROM chapters WHERE id = ?`, id).
		Scan(&item.ID, &item.BookID, &item.Title, &item.Summary, &item.Position, &item.MarkdownContent, &item.EditorJSON, &item.SectionType, &item.Status, &includedExport, &visibleTOC, &sampleContent, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Chapter{}, ErrNotFound
		}
		return model.Chapter{}, fmt.Errorf("get chapter: %w", err)
	}
	item.IsIncludedInExport = includedExport == 1
	item.IsVisibleInTOC = visibleTOC == 1
	item.IsSampleContent = sampleContent == 1
	return item, nil
}

func (s *Store) getRevision(ctx context.Context, queryer sqlQueryer, id string) (model.Revision, error) {
	row := queryer.QueryRowContext(ctx, `SELECT id, chapter_id, revision_type, title, description, markdown_content, editor_json, word_count, added_words, removed_words, change_summary, session_id, created_by, created_at FROM revisions WHERE id = ?`, id)
	item, err := scanRevision(row.Scan)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(strings.ToLower(err.Error()), "sql: no rows") {
			return model.Revision{}, ErrNotFound
		}
		return model.Revision{}, err
	}
	return item, nil
}

func (s *Store) createRevisionRecord(ctx context.Context, runner sqlRunner, chapter model.Chapter, input createRevisionRecordInput) (model.Revision, error) {
	normalizedType := normalizeRevisionType(input.RevisionType)
	wordCount := countWords(chapter.MarkdownContent)
	previousWordCount, err := s.latestRevisionWordCount(ctx, runner, chapter.ID)
	if err != nil {
		return model.Revision{}, err
	}
	now := nowUTC()
	item := model.Revision{
		ID:              newID(),
		ChapterID:       chapter.ID,
		RevisionType:    normalizedType,
		Title:           fallback(strings.TrimSpace(input.Title), defaultRevisionTitle(normalizedType, chapter.Title)),
		Description:     strings.TrimSpace(input.Description),
		MarkdownContent: chapter.MarkdownContent,
		EditorJSON:      chapter.EditorJSON,
		WordCount:       wordCount,
		AddedWords:      max(wordCount-previousWordCount, 0),
		RemovedWords:    max(previousWordCount-wordCount, 0),
		ChangeSummary:   defaultChangeSummary(normalizedType, chapter, wordCount, previousWordCount),
		SessionID:       strings.TrimSpace(input.SessionID),
		CreatedBy:       strings.TrimSpace(input.CreatedBy),
		CreatedAt:       now,
	}

	if _, err := runner.ExecContext(ctx, `INSERT INTO revisions (id, chapter_id, revision_type, title, description, markdown_content, editor_json, word_count, added_words, removed_words, change_summary, session_id, created_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.ChapterID, item.RevisionType, item.Title, item.Description, item.MarkdownContent, item.EditorJSON, item.WordCount, item.AddedWords, item.RemovedWords, item.ChangeSummary, item.SessionID, item.CreatedBy, item.CreatedAt,
	); err != nil {
		return model.Revision{}, fmt.Errorf("create revision: %w", err)
	}

	if _, err := s.createRevisionEventRecord(ctx, runner, item.ID, chapter.ID, createRevisionEventInput{
		EventType:   normalizeRevisionEventType(input.EventType),
		EntityType:  "chapter",
		EntityID:    chapter.ID,
		Title:       fallback(strings.TrimSpace(input.EventTitle), "Revision erstellt"),
		Description: item.ChangeSummary,
		Metadata: map[string]any{
			"revision_type": item.RevisionType,
			"word_count":    item.WordCount,
		},
		CreatedBy: item.CreatedBy,
	}); err != nil {
		return model.Revision{}, err
	}

	return item, nil
}

func (s *Store) createAutosaveDraftRecord(ctx context.Context, runner sqlRunner, chapter model.Chapter, input createAutosaveDraftRecordInput) (model.AutosaveDraft, error) {
	now := nowUTC()
	item := model.AutosaveDraft{
		ID:              newID(),
		ChapterID:       chapter.ID,
		MarkdownContent: chapter.MarkdownContent,
		EditorJSON:      chapter.EditorJSON,
		Reason:          normalizeAutosaveReason(input.Reason),
		SessionID:       strings.TrimSpace(input.SessionID),
		WordCount:       countWords(chapter.MarkdownContent),
		CreatedAt:       now,
		ExpiresAt:       nowUTCWithOffset(72),
	}

	if _, err := runner.ExecContext(ctx, `INSERT INTO autosave_drafts (id, chapter_id, markdown_content, editor_json, reason, session_id, word_count, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.ChapterID, item.MarkdownContent, item.EditorJSON, item.Reason, item.SessionID, item.WordCount, item.CreatedAt, item.ExpiresAt,
	); err != nil {
		return model.AutosaveDraft{}, fmt.Errorf("create autosave draft: %w", err)
	}

	return item, nil
}

func (s *Store) createRevisionEventRecord(ctx context.Context, runner sqlRunner, revisionID, chapterID string, input createRevisionEventInput) (model.RevisionEvent, error) {
	now := nowUTC()
	item := model.RevisionEvent{
		ID:          newID(),
		RevisionID:  revisionID,
		ChapterID:   chapterID,
		EventType:   normalizeRevisionEventType(input.EventType),
		EntityType:  fallback(strings.TrimSpace(input.EntityType), "chapter"),
		EntityID:    strings.TrimSpace(input.EntityID),
		Title:       fallback(strings.TrimSpace(input.Title), "Revision-Event"),
		Description: strings.TrimSpace(input.Description),
		Metadata:    encodeMetadata(input.Metadata),
		CreatedBy:   strings.TrimSpace(input.CreatedBy),
		CreatedAt:   now,
	}

	if _, err := runner.ExecContext(ctx, `INSERT INTO revision_events (id, revision_id, chapter_id, event_type, entity_type, entity_id, title, description, metadata_json, created_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.RevisionID, item.ChapterID, item.EventType, item.EntityType, item.EntityID, item.Title, item.Description, item.Metadata, item.CreatedBy, item.CreatedAt,
	); err != nil {
		return model.RevisionEvent{}, fmt.Errorf("create revision event: %w", err)
	}

	return item, nil
}

func (s *Store) latestRevisionWordCount(ctx context.Context, queryer sqlQueryer, chapterID string) (int, error) {
	var wordCount int
	err := queryer.QueryRowContext(ctx, `SELECT word_count FROM revisions WHERE chapter_id = ? ORDER BY created_at DESC, rowid DESC LIMIT 1`, chapterID).Scan(&wordCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("latest revision word count: %w", err)
	}
	return wordCount, nil
}

func scanRevision(scan func(dest ...any) error) (model.Revision, error) {
	var item model.Revision
	if err := scan(
		&item.ID,
		&item.ChapterID,
		&item.RevisionType,
		&item.Title,
		&item.Description,
		&item.MarkdownContent,
		&item.EditorJSON,
		&item.WordCount,
		&item.AddedWords,
		&item.RemovedWords,
		&item.ChangeSummary,
		&item.SessionID,
		&item.CreatedBy,
		&item.CreatedAt,
	); err != nil {
		return model.Revision{}, err
	}
	return item, nil
}

func scanAutosaveDraft(scan func(dest ...any) error) (model.AutosaveDraft, error) {
	var item model.AutosaveDraft
	if err := scan(
		&item.ID,
		&item.ChapterID,
		&item.MarkdownContent,
		&item.EditorJSON,
		&item.Reason,
		&item.SessionID,
		&item.WordCount,
		&item.CreatedAt,
		&item.ExpiresAt,
	); err != nil {
		return model.AutosaveDraft{}, fmt.Errorf("scan autosave draft: %w", err)
	}
	return item, nil
}

func scanMilestone(scan func(dest ...any) error) (model.Milestone, error) {
	var item model.Milestone
	var locked int
	if err := scan(
		&item.ID,
		&item.BookID,
		&item.ChapterID,
		&item.RevisionID,
		&item.Title,
		&item.Description,
		&item.MilestoneType,
		&locked,
		&item.CreatedBy,
		&item.CreatedAt,
	); err != nil {
		return model.Milestone{}, fmt.Errorf("scan milestone: %w", err)
	}
	item.Locked = locked == 1
	return item, nil
}

func scanRevisionEvent(scan func(dest ...any) error) (model.RevisionEvent, error) {
	var item model.RevisionEvent
	if err := scan(
		&item.ID,
		&item.RevisionID,
		&item.ChapterID,
		&item.EventType,
		&item.EntityType,
		&item.EntityID,
		&item.Title,
		&item.Description,
		&item.Metadata,
		&item.CreatedBy,
		&item.CreatedAt,
	); err != nil {
		return model.RevisionEvent{}, fmt.Errorf("scan revision event: %w", err)
	}
	return item, nil
}

func normalizeRevisionType(value string) string {
	switch strings.TrimSpace(value) {
	case "manual", "session", "restore", "before_review", "after_review", "before_export", "structure_change", "system":
		return strings.TrimSpace(value)
	default:
		return "manual"
	}
}

func normalizeSaveMode(value string) string {
	switch strings.TrimSpace(value) {
	case "manual", "autosave":
		return strings.TrimSpace(value)
	default:
		return "legacy"
	}
}

func shouldPersistAutosaveDraft(saveMode string, input UpdateChapterInput) bool {
	return saveMode != "legacy" || strings.TrimSpace(input.AutosaveReason) != "" || strings.TrimSpace(input.SessionID) != ""
}

func shouldCreateRevisionForSave(saveMode string, input UpdateChapterInput) bool {
	return input.CreateRevision || saveMode == "manual"
}

func revisionTypeForSave(saveMode, requestedType string) string {
	if strings.TrimSpace(requestedType) != "" {
		return normalizeRevisionType(requestedType)
	}
	if saveMode == "autosave" {
		return "session"
	}
	return "manual"
}

func revisionEventTitleForSave(saveMode, chapterTitle, requestedTitle string) string {
	if strings.TrimSpace(requestedTitle) != "" {
		return strings.TrimSpace(requestedTitle)
	}
	if saveMode == "manual" {
		if strings.TrimSpace(chapterTitle) == "" {
			return "Kapitel manuell gespeichert"
		}
		return "Kapitel gespeichert"
	}
	return "Schreibstand gespeichert"
}

func defaultAutosaveReason(saveMode, value string) string {
	if strings.TrimSpace(value) != "" {
		return normalizeAutosaveReason(value)
	}
	switch saveMode {
	case "manual":
		return "manual_save"
	case "autosave":
		return "idle_autosave"
	default:
		return "legacy_update"
	}
}

func normalizeAutosaveReason(value string) string {
	switch strings.TrimSpace(value) {
	case "idle_autosave", "before_navigation", "manual_save", "manual_safety_save", "legacy_update", "before_switch", "recovery_save":
		return strings.TrimSpace(value)
	default:
		return "idle_autosave"
	}
}

func normalizeMilestoneType(value string) string {
	switch strings.TrimSpace(value) {
	case "rough_draft", "before_review", "after_review", "before_export", "final", "publisher_submission", "reading_sample", "custom":
		return strings.TrimSpace(value)
	default:
		return "custom"
	}
}

func normalizeRevisionEventType(value string) string {
	switch strings.TrimSpace(value) {
	case "chapter_saved", "status_changed", "restore_performed", "milestone_created", "milestone_updated", "milestone_deleted", "review_started", "review_completed", "clipboard_inserted", "anchor_created":
		return strings.TrimSpace(value)
	default:
		return "chapter_saved"
	}
}

func eventTypeForRevisionType(value string) string {
	switch normalizeRevisionType(value) {
	case "restore":
		return "restore_performed"
	default:
		return "chapter_saved"
	}
}

func defaultRevisionTitle(revisionType, chapterTitle string) string {
	switch revisionType {
	case "session":
		return "Schreibsession"
	case "restore":
		return "Wiederhergestellter Stand"
	case "before_review":
		return "Stand vor Review"
	case "after_review":
		return "Stand nach Review"
	case "before_export":
		return "Stand vor Export"
	case "structure_change":
		return "Strukturstand"
	case "system":
		return "Systemstand"
	default:
		if strings.TrimSpace(chapterTitle) == "" {
			return "Manueller Stand"
		}
		return "Revision zu " + strings.TrimSpace(chapterTitle)
	}
}

func defaultChangeSummary(revisionType string, chapter model.Chapter, wordCount, previousWordCount int) string {
	switch revisionType {
	case "restore":
		return "Kapitelstand wurde wiederhergestellt."
	case "before_review":
		return "Kapitelstand vor einem Review gesichert."
	case "after_review":
		return "Kapitelstand nach einem Review gesichert."
	case "before_export":
		return "Kapitelstand vor dem Export gesichert."
	case "structure_change":
		return "Kapitelstand nach einer Strukturänderung gesichert."
	case "system":
		return "Technischer Sicherungsstand des Kapitels."
	case "session":
		return fmt.Sprintf("Schreibsession mit %d Woertern gesichert.", wordCount)
	default:
		if previousWordCount == 0 {
			return fmt.Sprintf("Kapitelstand mit %d Woertern gesichert.", wordCount)
		}
		delta := wordCount - previousWordCount
		if delta == 0 {
			return fmt.Sprintf("Kapitelstand unveraendert bei %d Woertern gesichert.", wordCount)
		}
		return fmt.Sprintf("Kapitelstand mit %d Woertern gesichert (%+d).", wordCount, delta)
	}
}

func countWords(value string) int {
	return len(strings.Fields(strings.TrimSpace(value)))
}

func encodeMetadata(value any) string {
	if value == nil {
		return "{}"
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	if len(encoded) == 0 {
		return "{}"
	}
	return string(encoded)
}

func nowUTCWithOffset(hours int) string {
	return time.Now().UTC().Add(time.Duration(hours) * time.Hour).Format(time.RFC3339)
}
