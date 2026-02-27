package model

type AuditEvent struct {
	Timestamp int64  `json:"timestamp"`
	Action    string `json:"action"`
	UserID    string `json:"user_id,omitempty"`
	URL       string `json:"url"`
}
