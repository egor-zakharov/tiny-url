package models

type Data struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	IsDeleted   string `json:"is_deleted"`
}

type URL struct {
	OriginalURL string `json:"original_url"`
	IsDeleted   string `json:"is_deleted"`
}

type ShortURL struct {
	ShortURL map[string]URL `json:"short_url"`
}

type MemData struct {
	UserID map[string]ShortURL `json:"user_id"`
}
