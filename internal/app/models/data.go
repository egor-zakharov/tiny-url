package models

// Data - struct for restore/backup mem_storage
type Data struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	IsDeleted   string `json:"is_deleted"`
}

// URL - struct original_url and deleted flag
type URL struct {
	OriginalURL string `json:"original_url"`
	IsDeleted   string `json:"is_deleted"`
}

// ShortURL - struct
type ShortURL struct {
	ShortURL map[string]URL `json:"short_url"`
}

// MemData - struct for mem_storage
type MemData struct {
	UserID map[string]ShortURL `json:"user_id"`
}

// Stats - struct for stats
type Stats struct {
	Urls  int `json:"urls"`
	Users int `json:"users"`
}
