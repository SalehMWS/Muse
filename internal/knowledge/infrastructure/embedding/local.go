package embedding

import (
	"context"
	"hash/fnv"
	"math"
	"strings"
)

const defaultDimension = 256

type LocalEmbedder struct {
	dim int
}

func NewLocalEmbedder(dim int) *LocalEmbedder {
	if dim <= 0 {
		dim = defaultDimension
	}
	return &LocalEmbedder{dim: dim}
}

func (e *LocalEmbedder) Dimension() int {
	return e.dim
}

func (e *LocalEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, text := range texts {
		out[i] = e.vectorize(text)
	}
	return out, nil
}

func (e *LocalEmbedder) vectorize(text string) []float32 {
	vec := make([]float32, e.dim)
	for _, token := range strings.Fields(strings.ToLower(text)) {
		vec[bucket(token, e.dim)] += sign(token)
	}

	var norm float64
	for _, v := range vec {
		norm += float64(v) * float64(v)
	}
	if norm > 0 {
		inv := float32(1 / math.Sqrt(norm))
		for i := range vec {
			vec[i] *= inv
		}
	}
	return vec
}

func bucket(token string, dim int) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(token))
	return int(h.Sum32()>>1) % dim
}

func sign(token string) float32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(token + "\x00sign"))
	if h.Sum32()%2 == 0 {
		return 1
	}
	return -1
}
