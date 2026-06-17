package algorithms

import (
	"math"
	"strings"
	"unicode"
)

// Tokenize splits a string into lowercase words, removing punctuation.
func Tokenize(text string) []string {
	var tokens []string
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	for _, word := range words {
		// Basic filtering (e.g., skip very short words or known stop words if needed)
		lower := strings.ToLower(word)
		if len(lower) > 1 {
			tokens = append(tokens, lower)
		}
	}
	return tokens
}

// TF calculates the Term Frequency of words in a document.
func TF(tokens []string) map[string]float64 {
	tf := make(map[string]float64)
	totalTokens := float64(len(tokens))
	if totalTokens == 0 {
		return tf
	}

	for _, token := range tokens {
		tf[token]++
	}

	for token, count := range tf {
		tf[token] = count / totalTokens
	}
	return tf
}

// IDF calculates the Inverse Document Frequency for a corpus of documents.
func IDF(corpus [][]string) map[string]float64 {
	idf := make(map[string]float64)
	totalDocs := float64(len(corpus))

	// Count number of documents each word appears in
	docCount := make(map[string]float64)
	for _, doc := range corpus {
		uniqueWords := make(map[string]bool)
		for _, word := range doc {
			uniqueWords[word] = true
		}
		for word := range uniqueWords {
			docCount[word]++
		}
	}

	// Calculate IDF: log(TotalDocs / DocCount)
	for word, count := range docCount {
		idf[word] = math.Log10(totalDocs / count)
	}

	return idf
}

// TFIDF calculates the TF-IDF vector for a single document given the global IDF.
func TFIDF(tf map[string]float64, idf map[string]float64) map[string]float64 {
	tfidf := make(map[string]float64)
	for word, tfVal := range tf {
		if idfVal, exists := idf[word]; exists {
			tfidf[word] = tfVal * idfVal
		} else {
			tfidf[word] = 0 // Word not in corpus
		}
	}
	return tfidf
}

// BuildCorpusTFIDF computes the IDF for the corpus and returns the TF-IDF vectors for all documents.
func BuildCorpusTFIDF(corpus [][]string) (map[string]float64, []map[string]float64) {
	idf := IDF(corpus)
	var tfidfs []map[string]float64

	for _, doc := range corpus {
		tf := TF(doc)
		tfidf := TFIDF(tf, idf)
		tfidfs = append(tfidfs, tfidf)
	}

	return idf, tfidfs
}
