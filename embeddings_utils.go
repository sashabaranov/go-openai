package openai

// DotProduct Calculate dot product of two vectors.
func DotProduct(v1, v2 []float32) float32 {
	var result float32
	// Iterate over vectors and calculate dot product.
	for i := 0; i < len(v1); i++ {
		result += v1[i] * v2[i]
	}
	return result
}
