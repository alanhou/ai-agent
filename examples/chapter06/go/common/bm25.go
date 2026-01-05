package common

import (
	"math"
)

// BM25Okapi implements the BM25 retrieval algorithm.
type BM25Okapi struct {
	corpusSize int64
	avgDL      float64
	docFreqs   []map[string]int
	idf        map[string]float64
	docLengths []int64
	k1         float64
	b          float64
}

// NewBM25Okapi creates a new BM25Okapi index.
func NewBM25Okapi(corpus [][]string) *BM25Okapi {
	bm25 := &BM25Okapi{
		corpusSize: int64(len(corpus)),
		docLengths: make([]int64, len(corpus)),
		docFreqs:   make([]map[string]int, len(corpus)),
		idf:        make(map[string]float64),
		k1:         1.5,
		b:          0.75,
	}

	var totalLen int64 = 0
	nd := make(map[string]int64)

	for i, doc := range corpus {
		bm25.docLengths[i] = int64(len(doc))
		totalLen += int64(len(doc))

		freqs := make(map[string]int)
		for _, word := range doc {
			freqs[word]++
		}
		bm25.docFreqs[i] = freqs

		for word := range freqs {
			nd[word]++
		}
	}

	bm25.avgDL = float64(totalLen) / float64(bm25.corpusSize)

	// Calculate IDF
	for word, freq := range nd {
		// IDF = ln( (N - n(q) + 0.5) / (n(q) + 0.5) + 1 )
		numerator := float64(bm25.corpusSize) - float64(freq) + 0.5
		denominator := float64(freq) + 0.5
		bm25.idf[word] = math.Log(numerator/denominator + 1)
	}

	return bm25
}

// GetScores calculates BM25 scores for a query against all documents.
func (bm25 *BM25Okapi) GetScores(query []string) []float64 {
	scores := make([]float64, bm25.corpusSize)

	for i := 0; i < int(bm25.corpusSize); i++ {
		score := 0.0
		docLen := float64(bm25.docLengths[i])

		for _, q := range query {
			if freq, ok := bm25.docFreqs[i][q]; ok {
				idf := bm25.idf[q]
				// score = IDF * (f * (k1 + 1)) / (f + k1 * (1 - b + b * |D| / avgdl))
				numerator := float64(freq) * (bm25.k1 + 1)
				denominator := float64(freq) + bm25.k1*(1-bm25.b+bm25.b*(docLen/bm25.avgDL))
				score += idf * (numerator / denominator)
			}
		}
		scores[i] = score
	}
	return scores
}

// GetTopN returns the top N documents matching the query.
func (bm25 *BM25Okapi) GetTopN(query []string, corpus [][]string, n int) []string {
	scores := bm25.GetScores(query)

	type result struct {
		index int
		score float64
	}

	results := make([]result, len(scores))
	for i, s := range scores {
		results[i] = result{index: i, score: s}
	}

	// Simple selection sort for top N (since N is small)
	topResults := make([]string, 0, n)

	// Copy results to avoid mutating original if we were doing a full sort,
	// but here we just pick top N.
	// Actually, let's just sort properly descending.
	// Since standard lib sort is a bit verbose with custom types in older Go
	// or requires boilerplate, let's just pick max N times.

	used := make([]bool, len(results))
	for count := 0; count < n && count < len(results); count++ {
		bestIdx := -1
		maxScore := -1.0

		for i, r := range results {
			if !used[i] && r.score > maxScore {
				maxScore = r.score
				bestIdx = i
			}
		}

		if bestIdx != -1 {
			used[bestIdx] = true
			// Reconstruct document string from corpus
			doc := ""
			for j, word := range corpus[results[bestIdx].index] {
				if j > 0 {
					doc += " "
				}
				doc += word
			}
			topResults = append(topResults, doc)
		}
	}

	return topResults
}
