package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"agents-go/examples/chapter06/go/common"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	oai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

// State represents the conversation state
type State struct {
	Messages []*schema.Message
}

var llm *oai.ChatModel

// callModel invokes the LLM
func callModel(ctx context.Context, state *State) (*State, error) {
	resp, err := llm.Generate(ctx, state.Messages)
	if err != nil {
		return state, err
	}
	state.Messages = append(state.Messages, resp)
	return state, nil
}

func main() {
	// Load .env from project root
	_ = godotenv.Load()
	ctx := context.Background()

	// 1. Initialize Embedder (for Vector Store)
	embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
		Model:   "text-embedding-ada-002",
	})
	if err != nil {
		log.Fatalf("Failed to create embedder: %v", err)
	}

	// 2. Initialize Vector Store
	vectorStore := common.NewSimpleVectorStore(embedder)

	// 3. Add Documents to Memory using standard slide window chunking strategy simulation
	// (Here we just use predetermined texts for simplicity as per standard LangChain example)
	text1 := `Machine learning is a method of data analysis that automates analytical model building. It is a branch of artificial intelligence based on the idea that systems can learn from data, identify patterns and make decisions with minimal human intervention. Machine learning algorithms are trained on datasets that contain examples of the desired output. For example, a machine learning algorithm that is used to classify images might be trained on a dataset that contains images of cats and dogs. Once an algorithm is trained, it can be used to make predictions on new data. For example, the machine learning algorithm that is used to classify images could be used to predict whether a new image contains a cat or a dog.`
	meta1 := map[string]string{
		"title": "Introduction to Machine Learning",
		"url":   "https://learn.microsoft.com/en-us/training/modules/introduction-to-machine-learning",
	}

	text2 := `Artificial intelligence (AI) is the simulation of human intelligence in machines that are programmed to think like humans and mimic their actions. The term may also be applied to any machine that exhibits traits associated with a human mind such as learning and problem-solving. AI research has been highly successful in developing effective techniques for solving a wide range of problems, from game playing to medical diagnosis.`
	meta2 := map[string]string{
		"title": "Artificial Intelligence for Beginners",
		"url":   "https://microsoft.github.io/AI-for-Beginners",
	}

	err = vectorStore.AddDocuments(ctx, []string{text1, text2}, []map[string]string{meta1, meta2})
	if err != nil {
		log.Fatalf("Failed to add documents: %v", err)
	}

	// 4. Perform Similarity Search
	query := "What is the relationship between AI and machine learning?"
	fmt.Printf("Query: %s\n", query)

	results, err := vectorStore.SimilaritySearch(ctx, query, 3)
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	fmt.Println("\nSearch Results:")
	for _, doc := range results {
		// Truncate content for display
		displayContent := doc.Content
		if len(displayContent) > 100 {
			displayContent = displayContent[:100] + "..."
		}
		fmt.Printf("- %s (Metadata: %v)\n", displayContent, doc.Metadata)
	}

	// 5. Initialize Chat Model (for Graph)
	llm, err = oai.NewChatModel(ctx, &oai.ChatModelConfig{
		Model:   "gpt-4o-mini", // Proxy for gpt-5
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
	})
	if err != nil {
		log.Fatalf("Failed to create chat model: %v", err)
	}

	// 6. Run standard Graph interaction (Stateless for this demo part, matching python)
	fmt.Println("\n--- Graph Interaction ---")
	builder := compose.NewGraph[*State, *State]()
	builder.AddLambdaNode("call_model", compose.InvokableLambda(callModel))
	builder.AddEdge(compose.START, "call_model")
	builder.AddEdge("call_model", compose.END)

	graph, err := builder.Compile(ctx)
	if err != nil {
		log.Fatalf("Failed to compile graph: %v", err)
	}

	input := &State{
		Messages: []*schema.Message{
			schema.UserMessage("hi! I'm bob"),
		},
	}
	out, err := graph.Invoke(ctx, input)
	if err != nil {
		log.Fatalf("Graph invoke failed: %v", err)
	}

	lastMsg := out.Messages[len(out.Messages)-1]
	fmt.Printf("[%s]: %s\n", lastMsg.Role, lastMsg.Content)
}
