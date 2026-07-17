package http

type CreateContentRequest struct {
	Title       string   `json:"title"`
	Caption     string   `json:"caption"`
	Language    string   `json:"language"`
	ContentType string   `json:"content_type"`
	Visibility  string   `json:"visibility"`
	Tags        []string `json:"tags"`
}

type UpdateContentRequest struct {
	Title       *string   `json:"title"`
	Caption     *string   `json:"caption"`
	Status      *string   `json:"status"`
	Language    *string   `json:"language"`
	ContentType *string   `json:"content_type"`
	Visibility  *string   `json:"visibility"`
	Tags        *[]string `json:"tags"`
}

type GenerateCaptionRequest struct {
	Prompt string `json:"prompt"`
}
