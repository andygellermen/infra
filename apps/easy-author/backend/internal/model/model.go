package model

type Project struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Book struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	Author     string `json:"author"`
	Visibility string `json:"visibility"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type Chapter struct {
	ID              string `json:"id"`
	BookID          string `json:"book_id"`
	Title           string `json:"title"`
	Summary         string `json:"summary"`
	Position        int    `json:"position"`
	MarkdownContent string `json:"markdown_content"`
	EditorJSON      string `json:"editor_json"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
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
	Book          Book            `json:"book"`
	Chapters      []Chapter       `json:"chapters"`
	WorkflowBoxes []WorkflowBox   `json:"workflow_boxes"`
	Clipboard     []ClipboardItem `json:"clipboard"`
}
