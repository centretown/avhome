package section

type Section struct {
	Parent   string     `json:"section"`
	Name     string     `json:"name"`
	Title    string     `json:"title"`
	Icon     string     `json:"icon"`
	Sections []*Section `json:"sections"`
}
