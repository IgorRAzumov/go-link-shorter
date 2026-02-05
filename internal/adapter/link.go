package adapter

type Link struct {
	UUID        string `json:"uuid"`
	ShortKey    string `json:"short_url"`
	FullURL     string `json:"original_url"`
	UserID      string `json:"user_id,omitempty"`
	DeletedFlag bool   `json:"is_deleted"`
}
