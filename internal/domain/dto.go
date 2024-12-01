package domain

type ProcessedImage struct {
	Path  string
	Error error
}

type ProcessedResponse struct {
	Success []string `json:"success"`
	Errors  []string `json:"errors"`
}
