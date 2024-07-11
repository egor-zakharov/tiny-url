package models

// Request - postShorten handler request
type Request struct {
	URL string `json:"url"`
}

// Response - postShorten handler response
type Response struct {
	Result string `json:"result"`
}

// ShortenBatchRequest - batch handler request
type ShortenBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	URL           string `json:"original_url"`
}

// ShortenBatchResponse - batch handler response
type ShortenBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// UserURLsResponse -  get /api/user/urls handler request
type UserURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// DeleteBatchRequest -  delete handler request
type DeleteBatchRequest []string
