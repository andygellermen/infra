package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/model"
)

func (s *Store) GetDashboard(ctx context.Context) ([]model.DashboardProject, error) {
	projects, err := s.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	var dashboard []model.DashboardProject
	for _, project := range projects {
		_, books, err := s.GetProject(ctx, project.ID)
		if err != nil {
			return nil, err
		}
		entry := model.DashboardProject{Project: project}
		for _, book := range books {
			chapters, err := s.ListChapters(ctx, book.ID)
			if err != nil {
				return nil, err
			}
			entry.Books = append(entry.Books, model.DashboardBookSummary{
				ID:           book.ID,
				Title:        book.Title,
				WorkType:     book.WorkType,
				ChapterCount: len(chapters),
				Progress:     calculateBookProgress(chapters),
				UpdatedAt:    book.UpdatedAt,
			})
		}
		dashboard = append(dashboard, entry)
	}
	return dashboard, nil
}

func (s *Store) GetBookStructure(ctx context.Context, bookID string) (model.BookStructure, error) {
	chapters, err := s.ListChapters(ctx, bookID)
	if err != nil {
		return model.BookStructure{}, err
	}
	structure := model.BookStructure{
		BookID:      bookID,
		AllChapters: chapters,
	}
	for _, chapter := range chapters {
		switch normalizeSectionType(chapter.SectionType) {
		case "front_matter":
			structure.FrontMatter = append(structure.FrontMatter, chapter)
		case "back_matter":
			structure.BackMatter = append(structure.BackMatter, chapter)
		case "fragment":
			structure.Fragments = append(structure.Fragments, chapter)
		default:
			structure.Body = append(structure.Body, chapter)
		}
	}
	return structure, nil
}

func (s *Store) UpdateChapterOptions(ctx context.Context, id string, input UpdateChapterOptionsInput) (model.Chapter, error) {
	item, err := s.GetChapter(ctx, id)
	if err != nil {
		return model.Chapter{}, err
	}
	item.SectionType = normalizeSectionType(input.SectionType)
	item.Status = normalizeChapterStatus(input.Status)
	item.IsIncludedInExport = input.IsIncludedInExport
	item.IsVisibleInTOC = input.IsVisibleInTOC
	item.IsSampleContent = input.IsSampleContent
	item.UpdatedAt = nowUTC()

	_, err = s.db.ExecContext(ctx, `UPDATE chapters SET section_type = ?, status = ?, is_included_in_export = ?, is_visible_in_toc = ?, is_sample_content = ?, updated_at = ? WHERE id = ?`,
		item.SectionType, item.Status, boolToInt(item.IsIncludedInExport), boolToInt(item.IsVisibleInTOC), boolToInt(item.IsSampleContent), item.UpdatedAt, item.ID,
	)
	if err != nil {
		return model.Chapter{}, fmt.Errorf("update chapter options: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE books SET updated_at = ? WHERE id = ?`, item.UpdatedAt, item.BookID); err != nil {
		return model.Chapter{}, fmt.Errorf("touch book after chapter options: %w", err)
	}
	return item, nil
}

func (s *Store) ListReusableBookPages(ctx context.Context, bookID string) ([]model.ReusableBookPage, error) {
	projectID, err := s.lookupBookProjectID(ctx, bookID)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, project_id, title, page_type, content_markdown, content_json, is_global, created_at, updated_at FROM reusable_book_pages WHERE is_global = 1 OR project_id = ? ORDER BY is_global DESC, updated_at DESC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list reusable pages: %w", err)
	}
	defer rows.Close()

	items := make([]model.ReusableBookPage, 0)
	for rows.Next() {
		item, err := scanReusableBookPage(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CreateReusableBookPage(ctx context.Context, bookID string, input CreateReusableBookPageInput) (model.ReusableBookPage, error) {
	projectID, err := s.lookupBookProjectID(ctx, bookID)
	if err != nil {
		return model.ReusableBookPage{}, err
	}
	now := nowUTC()
	item := model.ReusableBookPage{
		ID:              newID(),
		ProjectID:       projectID,
		Title:           fallback(strings.TrimSpace(input.Title), "Neue Buchseite"),
		PageType:        normalizeReusablePageType(input.PageType),
		ContentMarkdown: strings.TrimSpace(input.ContentMarkdown),
		ContentJSON:     strings.TrimSpace(input.ContentJSON),
		IsGlobal:        input.IsGlobal,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO reusable_book_pages (id, project_id, title, page_type, content_markdown, content_json, is_global, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.ProjectID, item.Title, item.PageType, item.ContentMarkdown, item.ContentJSON, boolToInt(item.IsGlobal), item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return model.ReusableBookPage{}, fmt.Errorf("create reusable page: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateReusableBookPage(ctx context.Context, id string, input UpdateReusableBookPageInput) (model.ReusableBookPage, error) {
	item, err := s.getReusableBookPage(ctx, id)
	if err != nil {
		return model.ReusableBookPage{}, err
	}
	item.Title = fallback(strings.TrimSpace(input.Title), item.Title)
	item.PageType = normalizeReusablePageType(fallback(strings.TrimSpace(input.PageType), item.PageType))
	item.ContentMarkdown = strings.TrimSpace(input.ContentMarkdown)
	item.ContentJSON = strings.TrimSpace(input.ContentJSON)
	item.IsGlobal = input.IsGlobal
	item.UpdatedAt = nowUTC()
	_, err = s.db.ExecContext(ctx, `UPDATE reusable_book_pages SET title = ?, page_type = ?, content_markdown = ?, content_json = ?, is_global = ?, updated_at = ? WHERE id = ?`,
		item.Title, item.PageType, item.ContentMarkdown, item.ContentJSON, boolToInt(item.IsGlobal), item.UpdatedAt, item.ID,
	)
	if err != nil {
		return model.ReusableBookPage{}, fmt.Errorf("update reusable page: %w", err)
	}
	return item, nil
}

func (s *Store) DeleteReusableBookPage(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM reusable_book_pages WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete reusable page: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete reusable page rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListThemes(ctx context.Context) ([]model.Theme, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, description, target, settings_json, created_at, updated_at FROM themes ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list themes: %w", err)
	}
	defer rows.Close()

	items := make([]model.Theme, 0)
	for rows.Next() {
		var item model.Theme
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Target, &item.SettingsJSON, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan theme: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CreateTheme(ctx context.Context, input CreateThemeInput) (model.Theme, error) {
	now := nowUTC()
	item := model.Theme{
		ID:           newID(),
		Name:         fallback(strings.TrimSpace(input.Name), "Neues Theme"),
		Description:  strings.TrimSpace(input.Description),
		Target:       normalizeThemeTarget(input.Target),
		SettingsJSON: fallback(strings.TrimSpace(input.SettingsJSON), "{}"),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO themes (id, name, description, target, settings_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.Name, item.Description, item.Target, item.SettingsJSON, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return model.Theme{}, fmt.Errorf("create theme: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateTheme(ctx context.Context, id string, input UpdateThemeInput) (model.Theme, error) {
	item, err := s.getTheme(ctx, id)
	if err != nil {
		return model.Theme{}, err
	}
	item.Name = fallback(strings.TrimSpace(input.Name), item.Name)
	item.Description = strings.TrimSpace(input.Description)
	item.Target = normalizeThemeTarget(fallback(strings.TrimSpace(input.Target), item.Target))
	item.SettingsJSON = fallback(strings.TrimSpace(input.SettingsJSON), item.SettingsJSON)
	item.UpdatedAt = nowUTC()
	_, err = s.db.ExecContext(ctx, `UPDATE themes SET name = ?, description = ?, target = ?, settings_json = ?, updated_at = ? WHERE id = ?`,
		item.Name, item.Description, item.Target, item.SettingsJSON, item.UpdatedAt, item.ID,
	)
	if err != nil {
		return model.Theme{}, fmt.Errorf("update theme: %w", err)
	}
	return item, nil
}

func (s *Store) GetBookProofingChecks(ctx context.Context, bookID string) ([]model.ProofingCheck, error) {
	bundle, err := s.GetBook(ctx, bookID)
	if err != nil {
		return nil, err
	}
	reusablePages, err := s.ListReusableBookPages(ctx, bookID)
	if err != nil {
		return nil, err
	}

	checks := make([]model.ProofingCheck, 0)
	addCheck := func(severity, scope, title, message, entityType, entityID string) {
		checks = append(checks, model.ProofingCheck{
			ID:                newID(),
			Severity:          severity,
			Scope:             scope,
			Title:             title,
			Message:           message,
			Status:            "open",
			RelatedEntityType: entityType,
			RelatedEntityID:   entityID,
		})
	}

	if strings.TrimSpace(bundle.Book.Title) == "" {
		addCheck("error", "book", "Buchtitel fehlt", "Dem Buch fehlt ein Titel.", "book", bundle.Book.ID)
	}
	if strings.TrimSpace(bundle.Book.Author) == "" {
		addCheck("warning", "book", "Autor fehlt", "Dem Buch ist noch kein Autor zugewiesen.", "book", bundle.Book.ID)
	}
	if strings.TrimSpace(bundle.Book.CoverAssetID) == "" {
		addCheck("warning", "book", "Cover fehlt", "Fuer das Buch ist noch kein Cover hinterlegt.", "book", bundle.Book.ID)
	}
	if !hasImprint(reusablePages, bundle.Chapters) {
		addCheck("warning", "structure", "Impressum fehlt", "Es wurde weder eine Impressumsseite noch ein Impressumskapitel gefunden.", "book", bundle.Book.ID)
	}

	titleCounts := make(map[string]int)
	for _, chapter := range bundle.Chapters {
		titleKey := strings.ToLower(strings.TrimSpace(chapter.Title))
		if titleKey != "" {
			titleCounts[titleKey]++
		}
	}
	for _, chapter := range bundle.Chapters {
		if strings.TrimSpace(chapter.MarkdownContent) == "" {
			addCheck("warning", "chapter", "Kapitel ist leer", fmt.Sprintf("%s enthaelt noch keinen Inhalt.", chapter.Title), "chapter", chapter.ID)
		}
		if countWords(chapter.MarkdownContent) >= 3500 {
			addCheck("info", "chapter", "Kapitel ist sehr lang", fmt.Sprintf("%s ist fuer den MVP-Pruefpfad auffaellig lang.", chapter.Title), "chapter", chapter.ID)
		}
		if strings.Contains(strings.ToUpper(chapter.MarkdownContent), "TODO") {
			addCheck("warning", "chapter", "TODO im Kapiteltext", fmt.Sprintf("%s enthaelt noch TODO-Markierungen.", chapter.Title), "chapter", chapter.ID)
		}
		if titleCounts[strings.ToLower(strings.TrimSpace(chapter.Title))] > 1 && strings.TrimSpace(chapter.Title) != "" {
			addCheck("warning", "chapter", "Doppelter Kapiteltitel", fmt.Sprintf("Der Titel %q kommt mehrfach vor.", chapter.Title), "chapter", chapter.ID)
		}
		if !chapter.IsIncludedInExport {
			addCheck("info", "chapter", "Kapitel nicht im Export", fmt.Sprintf("%s ist aktuell vom Export ausgeschlossen.", chapter.Title), "chapter", chapter.ID)
		}
	}

	return checks, nil
}

func (s *Store) getReusableBookPage(ctx context.Context, id string) (model.ReusableBookPage, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, project_id, title, page_type, content_markdown, content_json, is_global, created_at, updated_at FROM reusable_book_pages WHERE id = ?`, id)
	item, err := scanReusableBookPage(row.Scan)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(strings.ToLower(err.Error()), "sql: no rows") {
			return model.ReusableBookPage{}, ErrNotFound
		}
		return model.ReusableBookPage{}, err
	}
	return item, nil
}

func (s *Store) getTheme(ctx context.Context, id string) (model.Theme, error) {
	var item model.Theme
	err := s.db.QueryRowContext(ctx, `SELECT id, name, description, target, settings_json, created_at, updated_at FROM themes WHERE id = ?`, id).
		Scan(&item.ID, &item.Name, &item.Description, &item.Target, &item.SettingsJSON, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Theme{}, ErrNotFound
		}
		return model.Theme{}, fmt.Errorf("get theme: %w", err)
	}
	return item, nil
}

func scanReusableBookPage(scan func(dest ...any) error) (model.ReusableBookPage, error) {
	var item model.ReusableBookPage
	var isGlobal int
	if err := scan(&item.ID, &item.ProjectID, &item.Title, &item.PageType, &item.ContentMarkdown, &item.ContentJSON, &isGlobal, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return model.ReusableBookPage{}, fmt.Errorf("scan reusable page: %w", err)
	}
	item.IsGlobal = isGlobal == 1
	return item, nil
}

func calculateBookProgress(chapters []model.Chapter) int {
	if len(chapters) == 0 {
		return 0
	}
	total := 0
	for _, chapter := range chapters {
		total += chapterProgressValue(chapter.Status)
	}
	return total / len(chapters)
}

func chapterProgressValue(status string) int {
	switch normalizeChapterStatus(status) {
	case "idea":
		return 10
	case "draft":
		return 30
	case "revision":
		return 55
	case "review":
		return 80
	case "final", "archived":
		return 100
	default:
		return 0
	}
}

func hasImprint(reusablePages []model.ReusableBookPage, chapters []model.Chapter) bool {
	for _, page := range reusablePages {
		if page.PageType == "imprint" || strings.Contains(strings.ToLower(page.Title), "impressum") {
			return true
		}
	}
	for _, chapter := range chapters {
		if normalizeSectionType(chapter.SectionType) == "back_matter" && strings.Contains(strings.ToLower(chapter.Title), "impressum") {
			return true
		}
	}
	return false
}
