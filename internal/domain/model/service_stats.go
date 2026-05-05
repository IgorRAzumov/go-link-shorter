package model

// ServiceStats — статистика сервиса сокращения ссылок.
type ServiceStats struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}
