package algorithms

import "math"

// CosineSimilarity calculates the cosine similarity between two TF-IDF vectors.
// Returns a value between 0.0 (completely different) and 1.0 (identical).
func CosineSimilarity(vecA, vecB map[string]float64) float64 {
	var dotProduct float64
	var normA float64
	var normB float64

	// Calculate dot product and norm of A
	for word, valA := range vecA {
		normA += valA * valA
		if valB, exists := vecB[word]; exists {
			dotProduct += valA * valB
		}
	}

	// Calculate norm of B
	for _, valB := range vecB {
		normB += valB * valB
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
