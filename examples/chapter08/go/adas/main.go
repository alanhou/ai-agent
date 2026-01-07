package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

// --- Types ---

type Scenario struct {
	Target      int
	Stock       int
	Incoming    int
	GroundTruth int
}

type AgentSnapshot struct {
	Prompt     string
	Fitness    float64
	Generation int
}

func float32Ptr(v float32) *float32 {
	return &v
}

// --- Task Definition ---

func generateScenarios(n int) []Scenario {
	scenarios := make([]Scenario, n)
	for i := 0; i < n; i++ {
		target := 100
		stock := rand.Intn(120)
		incoming := rand.Intn(50)
		// Policy: Order = max(0, Target - (Stock + Incoming))
		groundTruth := target - (stock + incoming)
		if groundTruth < 0 {
			groundTruth = 0
		}
		scenarios[i] = Scenario{
			Target:      target,
			Stock:       stock,
			Incoming:    incoming,
			GroundTruth: groundTruth,
		}
	}
	return scenarios
}

func evaluatePrediction(prediction string, groundTruth int) float64 {
	// Extract number from prediction string (e.g. "Order 50 units" -> 50)
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(prediction)
	if match == "" {
		return 0.0
	}
	predVal, err := strconv.Atoi(match)
	if err != nil {
		return 0.0
	}

	diff := math.Abs(float64(predVal - groundTruth))
	// Score: 1.0 if exact, linear penalty up to diff=50
	if diff > 50 {
		return 0.0
	}
	return 1.0 - (diff / 50.0)
}

// --- Main ADAS Loop ---

func main() {
	_ = godotenv.Load()
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:       "gpt-4o",
		APIKey:      os.Getenv("OPENAI_API_KEY"),
		BaseURL:     os.Getenv("OPENAI_BASE_URL"),
		Temperature: float32Ptr(0.7), // Higher temp for creative prompt generation
	})
	if err != nil {
		log.Fatalf("Failed to init chat model: %v", err)
	}

	// Validation set (small for speed in demo)
	valScenarios := generateScenarios(5)

	// Initial Agent (Heuristic / Simple Prompt)
	currentPrompt := "You are a supply chain assistant. Calculate the reorder quantity."

	fmt.Println("=== ADAS Supply Chain Optimizer (Go) ===")
	fmt.Println("Goal: Optimize the agent's System Prompt to maximize reorder accuracy.")

	var archive []AgentSnapshot

	// Evaluate Initial
	fmt.Println("\n--- Evaluating Initial Prompt ---")
	initialFitness := runEvaluation(ctx, chatModel, currentPrompt, valScenarios)
	archive = append(archive, AgentSnapshot{Prompt: currentPrompt, Fitness: initialFitness, Generation: 0})
	fmt.Printf("Initial Fitness: %.4f\n", initialFitness)

	// Meta-Learning Loop
	generations := 3
	for gen := 1; gen <= generations; gen++ {
		fmt.Printf("\n--- Generation %d ---\n", gen)

		// 1. Meta-Agent: Generate new prompt based on history
		newPrompt := generateNewPrompt(ctx, chatModel, archive)
		fmt.Printf("Generated Prompt: %q\n", newPrompt)

		// 2. Evaluate new candidate
		fitness := runEvaluation(ctx, chatModel, newPrompt, valScenarios)
		fmt.Printf("Fitness: %.4f\n", fitness)

		// 3. Update Archive
		if fitness > archive[len(archive)-1].Fitness {
			fmt.Println(">>> Improved Solution Found!")
			archive = append(archive, AgentSnapshot{Prompt: newPrompt, Fitness: fitness, Generation: gen})
		} else {
			fmt.Println("No improvement.")
			// Optionally keep it anyway or discard. We discard in this simple version.
		}
	}

	fmt.Println("\n=== Optimization Complete ===")
	best := archive[len(archive)-1]
	fmt.Printf("Best Fitness: %.4f\n", best.Fitness)
	fmt.Printf("Best Prompt: %s\n", best.Prompt)
}

// runEvaluation runs the candidate agent (with candidatePrompt) against scenarios
func runEvaluation(ctx context.Context, model *openai.ChatModel, candidatePrompt string, scenarios []Scenario) float64 {
	totalScore := 0.0

	// Create a separate client for the candidate (usually same model, possibly lower temp)
	// We reuse 'model' here but strictly we might want temp=0 for the candidate execution

	for _, sc := range scenarios {
		userMsg := fmt.Sprintf("Target: %d, Stock: %d, Incoming: %d. How much to order?",
			sc.Target, sc.Stock, sc.Incoming)

		msgs := []*schema.Message{
			schema.SystemMessage(candidatePrompt),
			schema.UserMessage(userMsg),
		}

		resp, err := model.Generate(ctx, msgs)
		if err != nil {
			log.Printf("Candidate execution error: %v", err)
			continue
		}

		score := evaluatePrediction(resp.Content, sc.GroundTruth)
		totalScore += score
	}

	if len(scenarios) == 0 {
		return 0
	}
	return totalScore / float64(len(scenarios))
}

// generateNewPrompt uses the meta-agent to propose a better system prompt
func generateNewPrompt(ctx context.Context, model *openai.ChatModel, archive []AgentSnapshot) string {
	best := archive[len(archive)-1]

	metaPrompt := fmt.Sprintf(`You are an expert Prompt Engineer for AI Agents.
Your goal is to write a System Prompt for a Supply Chain Agent that maximizes its accuracy in calculating reorder quantities.

The Problem:
- The agent receives: Target Level, Current Stock, Incoming Stock.
- The agent must calculate: Reorder Quantity.
- Ground Truth Logic: max(0, Target - (Stock + Incoming)). However, the agent doesn't know this explicit formula; it must infer or be instructed to deduce it.

History of attempts:
Generation %d:
Prompt: "%s"
Fitness: %.4f (Failed to be perfect)

Your Task:
Write a NEW, IMPROVED System Prompt that will help the agent perform better. 
Consider explicitly instructing the agent on the logic if necessary, or improving its reasoning capabilities.
Return ONLY the text of the new System Prompt. Do not include quotes or explanations.`,
		best.Generation, best.Prompt, best.Fitness)

	msgs := []*schema.Message{
		schema.SystemMessage("You are an AI optimization assistant."),
		schema.UserMessage(metaPrompt),
	}

	resp, err := model.Generate(ctx, msgs)
	if err != nil {
		log.Printf("Meta-agent error: %v", err)
		return best.Prompt // fallback
	}

	return strings.TrimSpace(resp.Content)
}
