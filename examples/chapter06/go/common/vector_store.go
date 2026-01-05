package common

import (
	"context"
	"math"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
)

// Document represents a text with its embedding and metadata
type Document struct {
	Content   string
	Metadata  map[string]string
	Embedding []float64
}

// SimpleVectorStore is an in-memory vector store
type SimpleVectorStore struct {
	Documents []Document
	Embedder  *openai.Embedder
}

// NewSimpleVectorStore creates a new store
func NewSimpleVectorStore(embedder *openai.Embedder) *SimpleVectorStore {
	return &SimpleVectorStore{
		Documents: []Document{},
		Embedder:  embedder,
	}
}

// AddDocuments embeds and stores texts
func (svs *SimpleVectorStore) AddDocuments(ctx context.Context, texts []string, metadatas []map[string]string) error {
	embeddings, err := svs.Embedder.EmbedStrings(ctx, texts)
	if err != nil {
		return err
	}

	for i, text := range texts {
		doc := Document{
			Content:   text,
			Metadata:  metadatas[i],
			Embedding: embeddings[i],
		}
		svs.Documents = append(svs.Documents, doc)
	}
	return nil
}

// SimilaritySearch returns top K similar documents
func (svs *SimpleVectorStore) SimilaritySearch(ctx context.Context, query string, k int) ([]Document, error) {
	queryEmbeddings, err := svs.Embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	queryVec := queryEmbeddings[0]

	type result struct {
		doc   Document
		score float64
	}

	var results []result
	for _, doc := range svs.Documents {
		score := cosineSimilarity(queryVec, doc.Embedding)
		results = append(results, result{doc: doc, score: score})
	}

	// Sort by score descending
	// Simple bubble sort since K and N are small
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if k > len(results) {
		k = len(results)
	}

	topDocs := make([]Document, k)
	for i := 0; i < k; i++ {
		topDocs[i] = results[i].doc
	}
	return topDocs, nil
}

// cosineSimilarity helper from semantic example
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
