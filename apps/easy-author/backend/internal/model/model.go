package model

type Project struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	LastOpenedAt string `json:"last_opened_at"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type Book struct {
	ID           string `json:"id"`
	ProjectID    string `json:"project_id"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Author       string `json:"author"`
	Visibility   string `json:"visibility"`
	WorkType     string `json:"work_type"`
	ThemeID      string `json:"theme_id"`
	CoverAssetID string `json:"cover_asset_id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type Chapter struct {
	ID                 string `json:"id"`
	BookID             string `json:"book_id"`
	Title              string `json:"title"`
	Summary            string `json:"summary"`
	Position           int    `json:"position"`
	MarkdownContent    string `json:"markdown_content"`
	EditorJSON         string `json:"editor_json"`
	SectionType        string `json:"section_type"`
	Status             string `json:"status"`
	IsIncludedInExport bool   `json:"is_included_in_export"`
	IsVisibleInTOC     bool   `json:"is_visible_in_toc"`
	IsSampleContent    bool   `json:"is_sample_content"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type ReusableBookPage struct {
	ID              string `json:"id"`
	ProjectID       string `json:"project_id"`
	Title           string `json:"title"`
	PageType        string `json:"page_type"`
	ContentMarkdown string `json:"content_markdown"`
	ContentJSON     string `json:"content_json"`
	IsGlobal        bool   `json:"is_global"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type Theme struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Target       string `json:"target"`
	SettingsJSON string `json:"settings_json"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type ProofingCheck struct {
	ID                string `json:"id"`
	Severity          string `json:"severity"`
	Scope             string `json:"scope"`
	Title             string `json:"title"`
	Message           string `json:"message"`
	Status            string `json:"status"`
	RelatedEntityType string `json:"related_entity_type"`
	RelatedEntityID   string `json:"related_entity_id"`
}

type DashboardBookSummary struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	WorkType     string `json:"work_type"`
	ChapterCount int    `json:"chapter_count"`
	Progress     int    `json:"progress"`
	UpdatedAt    string `json:"updated_at"`
}

type DashboardProject struct {
	Project Project                `json:"project"`
	Books   []DashboardBookSummary `json:"books"`
}

type BookStructure struct {
	BookID      string    `json:"book_id"`
	FrontMatter []Chapter `json:"front_matter"`
	Body        []Chapter `json:"body"`
	BackMatter  []Chapter `json:"back_matter"`
	Fragments   []Chapter `json:"fragments"`
	AllChapters []Chapter `json:"all_chapters"`
}

type AutosaveDraft struct {
	ID              string `json:"id"`
	ChapterID       string `json:"chapter_id"`
	MarkdownContent string `json:"markdown_content"`
	EditorJSON      string `json:"editor_json"`
	Reason          string `json:"reason"`
	SessionID       string `json:"session_id"`
	WordCount       int    `json:"word_count"`
	CreatedAt       string `json:"created_at"`
	ExpiresAt       string `json:"expires_at"`
}

type Revision struct {
	ID              string `json:"id"`
	ChapterID       string `json:"chapter_id"`
	RevisionType    string `json:"revision_type"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	MarkdownContent string `json:"markdown_content"`
	EditorJSON      string `json:"editor_json"`
	WordCount       int    `json:"word_count"`
	AddedWords      int    `json:"added_words"`
	RemovedWords    int    `json:"removed_words"`
	ChangeSummary   string `json:"change_summary"`
	SessionID       string `json:"session_id"`
	CreatedBy       string `json:"created_by"`
	CreatedAt       string `json:"created_at"`
}

type Milestone struct {
	ID            string `json:"id"`
	BookID        string `json:"book_id"`
	ChapterID     string `json:"chapter_id"`
	RevisionID    string `json:"revision_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	MilestoneType string `json:"milestone_type"`
	Locked        bool   `json:"locked"`
	CreatedBy     string `json:"created_by"`
	CreatedAt     string `json:"created_at"`
}

type RevisionEvent struct {
	ID          string `json:"id"`
	RevisionID  string `json:"revision_id"`
	ChapterID   string `json:"chapter_id"`
	EventType   string `json:"event_type"`
	EntityType  string `json:"entity_type"`
	EntityID    string `json:"entity_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Metadata    string `json:"metadata"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
}

type WorkflowBox struct {
	ID          string   `json:"id"`
	BookID      string   `json:"book_id"`
	Title       string   `json:"title"`
	Type        string   `json:"type"`
	Tags        []string `json:"tags"`
	Position    int      `json:"position"`
	IsCollapsed bool     `json:"is_collapsed"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type Anchor struct {
	ID            string `json:"id"`
	ChapterID     string `json:"chapter_id"`
	WorkflowBoxID string `json:"workflow_box_id"`
	Title         string `json:"title"`
	AnchorType    string `json:"anchor_type"`
	SelectedText  string `json:"selected_text"`
	StartOffset   int    `json:"start_offset"`
	EndOffset     int    `json:"end_offset"`
	ContextBefore string `json:"context_before"`
	ContextAfter  string `json:"context_after"`
	Note          string `json:"note"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type ReviewComment struct {
	ID            string `json:"id"`
	ChapterID     string `json:"chapter_id"`
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
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type ClipboardItem struct {
	ID             string `json:"id"`
	BookID         string `json:"book_id"`
	ChapterID      string `json:"chapter_id"`
	Content        string `json:"content"`
	ContentType    string `json:"content_type"`
	SourceAnchorID string `json:"source_anchor_id"`
	Slot           int    `json:"slot"`
	IsPinned       bool   `json:"is_pinned"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type KnowledgeItem struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"project_id"`
	Type      string   `json:"type"`
	Name      string   `json:"name"`
	Summary   string   `json:"summary"`
	Body      string   `json:"body"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type BookBundle struct {
	Book          Book               `json:"book"`
	Chapters      []Chapter          `json:"chapters"`
	WorkflowBoxes []WorkflowBox      `json:"workflow_boxes"`
	Clipboard     []ClipboardItem    `json:"clipboard"`
	ReusablePages []ReusableBookPage `json:"reusable_pages,omitempty"`
	Themes        []Theme            `json:"themes,omitempty"`
}
