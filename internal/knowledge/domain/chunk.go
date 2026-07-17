package domain

import "strings"

const (
	DefaultChunkSize    = 200
	DefaultChunkOverlap = 40
)

func SplitIntoChunks(text string, size, overlap int) []string {
	if size <= 0 {
		size = DefaultChunkSize
	}
	if overlap < 0 || overlap >= size {
		overlap = 0
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	step := size - overlap
	chunks := make([]string, 0, len(words)/step+1)
	for start := 0; start < len(words); start += step {
		end := start + size
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[start:end], " "))
		if end == len(words) {
			break
		}
	}
	return chunks
}
