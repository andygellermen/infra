ALTER TABLE events ADD COLUMN published_at TEXT;

UPDATE events
SET published_at = updated_at
WHERE is_public = 1
  AND status <> 'draft'
  AND status <> 'archived'
  AND published_at IS NULL;
