package domain

type RuntimeInfoDto struct {
	Status       Status `json:"status"`
	LastAccessed int64  `json:"last_accessed"`
}
