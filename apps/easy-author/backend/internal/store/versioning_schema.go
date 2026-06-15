package store

func versioningSchemaStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS autosave_drafts (
			id TEXT PRIMARY KEY,
			chapter_id TEXT NOT NULL,
			markdown_content TEXT NOT NULL DEFAULT '',
			editor_json TEXT NOT NULL DEFAULT '',
			reason TEXT NOT NULL DEFAULT 'idle_autosave',
			session_id TEXT NOT NULL DEFAULT '',
			word_count INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			expires_at TEXT NOT NULL DEFAULT '',
			FOREIGN KEY(chapter_id) REFERENCES chapters(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS autosave_drafts_chapter_created_idx ON autosave_drafts(chapter_id, created_at DESC);`,
		`CREATE TABLE IF NOT EXISTS revisions (
			id TEXT PRIMARY KEY,
			chapter_id TEXT NOT NULL,
			revision_type TEXT NOT NULL DEFAULT 'manual',
			title TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			markdown_content TEXT NOT NULL DEFAULT '',
			editor_json TEXT NOT NULL DEFAULT '',
			word_count INTEGER NOT NULL DEFAULT 0,
			added_words INTEGER NOT NULL DEFAULT 0,
			removed_words INTEGER NOT NULL DEFAULT 0,
			change_summary TEXT NOT NULL DEFAULT '',
			session_id TEXT NOT NULL DEFAULT '',
			created_by TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			FOREIGN KEY(chapter_id) REFERENCES chapters(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS revisions_chapter_created_idx ON revisions(chapter_id, created_at DESC);`,
		`CREATE TABLE IF NOT EXISTS milestones (
			id TEXT PRIMARY KEY,
			book_id TEXT NOT NULL,
			chapter_id TEXT NOT NULL,
			revision_id TEXT NOT NULL,
			title TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			milestone_type TEXT NOT NULL DEFAULT 'custom',
			locked INTEGER NOT NULL DEFAULT 0,
			created_by TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			FOREIGN KEY(book_id) REFERENCES books(id) ON DELETE CASCADE,
			FOREIGN KEY(chapter_id) REFERENCES chapters(id) ON DELETE CASCADE,
			FOREIGN KEY(revision_id) REFERENCES revisions(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS milestones_book_created_idx ON milestones(book_id, created_at DESC);`,
		`CREATE TABLE IF NOT EXISTS revision_events (
			id TEXT PRIMARY KEY,
			revision_id TEXT NOT NULL,
			chapter_id TEXT NOT NULL,
			event_type TEXT NOT NULL DEFAULT 'chapter_saved',
			entity_type TEXT NOT NULL DEFAULT 'chapter',
			entity_id TEXT NOT NULL DEFAULT '',
			title TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			metadata_json TEXT NOT NULL DEFAULT '{}',
			created_by TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			FOREIGN KEY(revision_id) REFERENCES revisions(id) ON DELETE CASCADE,
			FOREIGN KEY(chapter_id) REFERENCES chapters(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS revision_events_revision_created_idx ON revision_events(revision_id, created_at ASC);`,
	}
}
