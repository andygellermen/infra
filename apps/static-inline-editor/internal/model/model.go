package model

type Tenant struct {
	Domain            string
	LoginDomain       string
	Aliases           []string
	StaticRoot        string
	BackupRoot        string
	RepoRoot          string
	Username          string
	PasswordHash      string
	CookieSecret      string
	MainSelector      string
	AllowedBlockTags  []string
	AllowedInlineTags []string
	StartPath         string
}
