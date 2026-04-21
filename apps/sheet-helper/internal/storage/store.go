package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/andygellermann/infra/apps/sheet-helper/internal/model"
	"github.com/andygellermann/infra/apps/sheet-helper/internal/pathutil"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) InitSchema(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS routes (
			domain TEXT NOT NULL,
			path TEXT NOT NULL,
			type TEXT NOT NULL,
			passphrase TEXT NOT NULL DEFAULT '',
			target TEXT NOT NULL DEFAULT '',
			title TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			list_sheet TEXT NOT NULL DEFAULT '',
			enabled INTEGER NOT NULL DEFAULT 1,
			PRIMARY KEY (domain, path)
		)`,
		`CREATE TABLE IF NOT EXISTS vcard_entries (
			domain TEXT NOT NULL,
			path TEXT NOT NULL,
			full_name TEXT NOT NULL DEFAULT '',
			organization TEXT NOT NULL DEFAULT '',
			job_title TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			phone_mobile TEXT NOT NULL DEFAULT '',
			address TEXT NOT NULL DEFAULT '',
			website TEXT NOT NULL DEFAULT '',
			image_url TEXT NOT NULL DEFAULT '',
			note TEXT NOT NULL DEFAULT '',
			enabled INTEGER NOT NULL DEFAULT 1,
			PRIMARY KEY (domain, path)
		)`,
		`CREATE TABLE IF NOT EXISTS text_entries (
			domain TEXT NOT NULL,
			path TEXT NOT NULL,
			content_type TEXT NOT NULL DEFAULT 'text/plain',
			content TEXT NOT NULL DEFAULT '',
			copy_hint TEXT NOT NULL DEFAULT '',
			expires_at TEXT NOT NULL DEFAULT '',
			enabled INTEGER NOT NULL DEFAULT 1,
			PRIMARY KEY (domain, path)
		)`,
		`CREATE TABLE IF NOT EXISTS list_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sheet_name TEXT NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			label TEXT NOT NULL DEFAULT '',
			url TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL DEFAULT '',
			password TEXT NOT NULL DEFAULT '',
			enabled INTEGER NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS click_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ts TEXT NOT NULL,
			domain TEXT NOT NULL,
			path TEXT NOT NULL,
			type TEXT NOT NULL,
			target TEXT NOT NULL DEFAULT '',
			referrer TEXT NOT NULL DEFAULT '',
			user_agent TEXT NOT NULL DEFAULT ''
		)`,
	}

	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("exec schema stmt: %w", err)
		}
	}

	return nil
}

func (s *Store) ReplaceAll(ctx context.Context, routes []model.Route, vCards []model.VCardEntry, texts []model.TextEntry, listItems []model.ListItem) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	for _, table := range []string{"routes", "vcard_entries", "text_entries", "list_items"} {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			return fmt.Errorf("clear %s: %w", table, err)
		}
	}

	for _, route := range routes {
		route.Path = pathutil.Normalize(route.Path)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO routes (domain, path, type, passphrase, target, title, description, list_sheet, enabled)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			route.Domain, route.Path, string(route.Type), route.Passphrase, route.Target, route.Title, route.Description, route.ListSheet, boolToInt(route.Enabled),
		); err != nil {
			return fmt.Errorf("insert route: %w", err)
		}
	}

	for _, entry := range vCards {
		entry.Path = pathutil.Normalize(entry.Path)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO vcard_entries (domain, path, full_name, organization, job_title, email, phone_mobile, address, website, image_url, note, enabled)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			entry.Domain, entry.Path, entry.FullName, entry.Organization, entry.JobTitle, entry.Email, entry.PhoneMobile, entry.Address, entry.Website, entry.ImageURL, entry.Note, boolToInt(entry.Enabled),
		); err != nil {
			return fmt.Errorf("insert vcard: %w", err)
		}
	}

	for _, entry := range texts {
		entry.Path = pathutil.Normalize(entry.Path)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO text_entries (domain, path, content_type, content, copy_hint, expires_at, enabled)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			entry.Domain, entry.Path, entry.ContentType, entry.Content, entry.CopyHint, entry.ExpiresAt, boolToInt(entry.Enabled),
		); err != nil {
			return fmt.Errorf("insert text: %w", err)
		}
	}

	for _, item := range listItems {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO list_items (sheet_name, sort_order, label, url, description, category, password, enabled)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			item.SheetName, item.Sort, item.Label, item.URL, item.Description, item.Category, item.Password, boolToInt(item.Enabled),
		); err != nil {
			return fmt.Errorf("insert list item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (s *Store) LookupRoute(ctx context.Context, domain, path string) (model.Route, bool, error) {
	path = pathutil.Normalize(path)
	row := s.db.QueryRowContext(ctx, `
		SELECT domain, path, type, passphrase, target, title, description, list_sheet, enabled
		FROM routes
		WHERE domain = ? AND enabled = 1
		  AND (path = ? OR rtrim(path, '/') = rtrim(?, '/'))
		ORDER BY CASE WHEN path = ? THEN 0 ELSE 1 END
		LIMIT 1`,
		domain, path, path, path,
	)

	var route model.Route
	var routeType string
	var enabled int
	if err := row.Scan(
		&route.Domain,
		&route.Path,
		&routeType,
		&route.Passphrase,
		&route.Target,
		&route.Title,
		&route.Description,
		&route.ListSheet,
		&enabled,
	); err != nil {
		if err == sql.ErrNoRows {
			return model.Route{}, false, nil
		}
		return model.Route{}, false, fmt.Errorf("scan route: %w", err)
	}

	route.Type = model.RouteType(routeType)
	route.Path = pathutil.Normalize(route.Path)
	route.Enabled = enabled == 1
	return route, true, nil
}

func (s *Store) GetVCard(ctx context.Context, domain, path string) (model.VCardEntry, bool, error) {
	path = pathutil.Normalize(path)
	row := s.db.QueryRowContext(ctx, `
		SELECT domain, path, full_name, organization, job_title, email, phone_mobile, address, website, image_url, note, enabled
		FROM vcard_entries
		WHERE domain = ? AND enabled = 1
		  AND (path = ? OR rtrim(path, '/') = rtrim(?, '/'))
		ORDER BY CASE WHEN path = ? THEN 0 ELSE 1 END
		LIMIT 1`,
		domain, path, path, path,
	)

	var entry model.VCardEntry
	var enabled int
	if err := row.Scan(
		&entry.Domain, &entry.Path, &entry.FullName, &entry.Organization, &entry.JobTitle, &entry.Email,
		&entry.PhoneMobile, &entry.Address, &entry.Website, &entry.ImageURL, &entry.Note, &enabled,
	); err != nil {
		if err == sql.ErrNoRows {
			return model.VCardEntry{}, false, nil
		}
		return model.VCardEntry{}, false, fmt.Errorf("scan vcard: %w", err)
	}
	entry.Path = pathutil.Normalize(entry.Path)
	entry.Enabled = enabled == 1
	return entry, true, nil
}

func (s *Store) GetText(ctx context.Context, domain, path string) (model.TextEntry, bool, error) {
	path = pathutil.Normalize(path)
	row := s.db.QueryRowContext(ctx, `
		SELECT domain, path, content_type, content, copy_hint, expires_at, enabled
		FROM text_entries
		WHERE domain = ? AND enabled = 1
		  AND (path = ? OR rtrim(path, '/') = rtrim(?, '/'))
		ORDER BY CASE WHEN path = ? THEN 0 ELSE 1 END
		LIMIT 1`,
		domain, path, path, path,
	)

	var entry model.TextEntry
	var enabled int
	if err := row.Scan(&entry.Domain, &entry.Path, &entry.ContentType, &entry.Content, &entry.CopyHint, &entry.ExpiresAt, &enabled); err != nil {
		if err == sql.ErrNoRows {
			return model.TextEntry{}, false, nil
		}
		return model.TextEntry{}, false, fmt.Errorf("scan text: %w", err)
	}
	entry.Path = pathutil.Normalize(entry.Path)
	entry.Enabled = enabled == 1
	return entry, true, nil
}

func (s *Store) ListItems(ctx context.Context, sheetName string) ([]model.ListItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT sheet_name, sort_order, label, url, description, category, password, enabled
		FROM list_items
		WHERE sheet_name = ? AND enabled = 1
		ORDER BY sort_order ASC, label ASC`,
		sheetName,
	)
	if err != nil {
		return nil, fmt.Errorf("query list items: %w", err)
	}
	defer rows.Close()

	var items []model.ListItem
	for rows.Next() {
		var item model.ListItem
		var enabled int
		if err := rows.Scan(&item.SheetName, &item.Sort, &item.Label, &item.URL, &item.Description, &item.Category, &item.Password, &enabled); err != nil {
			return nil, fmt.Errorf("scan list item: %w", err)
		}
		item.Enabled = enabled == 1
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate list items: %w", err)
	}
	return items, nil
}

func (s *Store) RecordClick(ctx context.Context, event model.ClickEvent) error {
	event.Path = pathutil.Normalize(event.Path)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO click_events (ts, domain, path, type, target, referrer, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		time.Now().UTC().Format(time.RFC3339), event.Domain, event.Path, string(event.Type), event.Target, event.Referrer, event.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("insert click: %w", err)
	}
	return nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
