package main

import (
	"agents-go/examples/chapter06/go/common"
	"fmt"
	"strings"
)

func main() {
	corpusRaw := []string{
		"Agent J is the fresh recruit with attitude",
		"Agent K has years of MIB experience and a cool neuralyzer",
		"The galaxy is saved by two Agents in black suits",
	}

	// Tokenize corpus
	var corpus [][]string
	for _, doc := range corpusRaw {
		corpus = append(corpus, strings.Fields(doc))
	}

	// 1. Build BM25 index
	bm25 := common.NewBM25Okapi(corpus)

	// 2. Execute retrieval for an interesting query
	queryString := "Who is a recruit?"
	query := strings.Fields(queryString)

	// Get top N=2
	topN := bm25.GetTopN(query, corpus, 2)

	fmt.Println("Query:", queryString)
	fmt.Println("Top matching lines:")
	for _, line := range topN {
		fmt.Printf(" • %s\n", line)
	}
}
