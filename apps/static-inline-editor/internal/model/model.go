package model

type Tenant struct {
	Domain            string
	LoginDomain       string
	Aliases           []string
	StaticRoot        string
	BackupRoot        string
	RepoRoot          string
	CookieSecret      string
	AllowedEmails     []string
	MainSelector      string
	AllowedBlockTags  []string
	AllowedInlineTags []string
	StartPath         string
}

type MagicLinkRequest struct {
	Email string `json:"email"`
}

type MagicLinkRequestResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type PreviewRequest struct {
	Path    string            `json:"path"`
	Regions map[string]string `json:"regions"`
}

type PreviewResponse struct {
	OK          bool   `json:"ok"`
	Message     string `json:"message,omitempty"`
	PreviewHTML string `json:"preview_html,omitempty"`
}

type SaveRequest struct {
	Path    string            `json:"path"`
	Regions map[string]string `json:"regions"`
}

type SaveResponse struct {
	OK         bool   `json:"ok"`
	Message    string `json:"message,omitempty"`
	BackupPath string `json:"backup_path,omitempty"`
	CommitHash string `json:"commit_hash,omitempty"`
}
