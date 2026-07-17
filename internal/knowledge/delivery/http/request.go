package http

type IngestRequest struct {
	Title   string `json:"title"`
	Source  string `json:"source"`
	Content string `json:"content"`
}

type QueryRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k"`
}
