package model

type RouteType string

const (
	RouteTypeLink  RouteType = "link"
	RouteTypeText  RouteType = "text"
	RouteTypeVCard RouteType = "vcard"
	RouteTypeList  RouteType = "list"
)

type Route struct {
	Domain      string
	Path        string
	Type        RouteType
	Passphrase  string
	Target      string
	Title       string
	Description string
	ListSheet   string
	Enabled     bool
}

type VCardEntry struct {
	Domain       string
	Path         string
	FullName     string
	Organization string
	JobTitle     string
	Email        string
	PhoneMobile  string
	Address      string
	Website      string
	ImageURL     string
	Note         string
	Enabled      bool
}

type TextEntry struct {
	Domain      string
	Path        string
	ContentType string
	Content     string
	CopyHint    string
	ExpiresAt   string
	Enabled     bool
}

type ListItem struct {
	SheetName   string
	Sort        int
	Label       string
	URL         string
	Description string
	Category    string
	Password    string
	Enabled     bool
}

type ClickEvent struct {
	Domain    string
	Path      string
	Type      RouteType
	Target    string
	Referrer  string
	UserAgent string
}
