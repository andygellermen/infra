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
