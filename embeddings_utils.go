package openai

import "math"

// CosineSimilarity Calculate cosine similarity.
// Note:
// We recommend cosine similarity. The choice of distance function typically doesn’t matter much.
// OpenAI embeddings are normalized to length 1, which means that:
// Cosine similarity can be computed slightly faster using just a dot product.
// Cosine similarity and Euclidean distance will result in the identical rankings.
// See: https://platform.openai.com/docs/guides/embeddings/limitations-risks
func CosineSimilarity(v1, v2 []float32) float32 {
	// Calculate dot product
	dot := DotProduct(v1, v2)
	// Calculate magnitude of v1
	v1Magnitude := math.Sqrt(float64(DotProduct(v1, v1)))
	// Calculate magnitude of v2
	v2Magnitude := math.Sqrt(float64(DotProduct(v2, v2)))
	// Calculate cosine similarity
	return float32(float64(dot) / (v1Magnitude * v2Magnitude))
}

// DotProduct Calculate dot product of two vectors.
func DotProduct(v1, v2 []float32) float32 {
	var result float32
	// Iterate over vectors and calculate dot product.
	for i := 0; i < len(v1); i++ {
		result += v1[i] * v2[i]
	}
	return result
}
