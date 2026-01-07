package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

type InsightAgent struct {
	Insights         []string
	PromotedInsights []string
	DemotedInsights  []string
	Reflections      []string
	ChatModel        *openai.ChatModel
}

func NewInsightAgent(ctx context.Context) (*InsightAgent, error) {
	// Initialize Eino ChatModel
	// Using gpt-4o as discovered in debugging session
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   "gpt-4o",
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
	})
	if err != nil {
		return nil, err
	}

	return &InsightAgent{
		Insights:         []string{},
		PromotedInsights: []string{},
		DemotedInsights:  []string{},
		Reflections:      []string{},
		ChatModel:        chatModel,
	}, nil
}

func (a *InsightAgent) getCompletion(ctx context.Context, prompt string) (string, error) {
	resp, err := a.ChatModel.Generate(ctx, []*schema.Message{
		schema.UserMessage(prompt),
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (a *InsightAgent) GenerateInsight(ctx context.Context, observation string) (string, error) {
	prompt := fmt.Sprintf("Generate an insightful analysis based on the following observation: '%s'", observation)
	insight, err := a.getCompletion(ctx, prompt)
	if err != nil {
		return "", err
	}
	a.Insights = append(a.Insights, insight)
	fmt.Printf("Generated: %s\n", insight)
	return insight, nil
}

func (a *InsightAgent) PromoteInsight(insight string) {
	// Check if present in Insights (Python: if insight in self.insights)
	// We just check purely by value matching
	foundIndex := -1
	for i, v := range a.Insights {
		if v == insight {
			foundIndex = i
			break
		}
	}

	if foundIndex != -1 {
		// Remove from insights
		a.Insights = append(a.Insights[:foundIndex], a.Insights[foundIndex+1:]...)
		a.PromotedInsights = append(a.PromotedInsights, insight)
		fmt.Printf("Promoted: %s\n", insight)
	} else {
		// Python logic prints this if not found
		fmt.Printf("Insight '%s' not found in insights.\n", insight)
	}
}

func (a *InsightAgent) DemoteInsight(insight string) {
	foundIndex := -1
	for i, v := range a.PromotedInsights {
		if v == insight {
			foundIndex = i
			break
		}
	}

	if foundIndex != -1 {
		a.PromotedInsights = append(a.PromotedInsights[:foundIndex], a.PromotedInsights[foundIndex+1:]...)
		a.DemotedInsights = append(a.DemotedInsights, insight)
		fmt.Printf("Demoted: %s\n", insight)
	} else {
		fmt.Printf("Insight '%s' not found in promoted insights.\n", insight)
	}
}

func (a *InsightAgent) EditInsight(oldInsight, newInsight string) {
	// Checks all lists
	// Check Insights
	if idx := slices.Index(a.Insights, oldInsight); idx != -1 {
		a.Insights[idx] = newInsight
		fmt.Printf("Edited: '%s' to '%s'\n", oldInsight, newInsight)
		return
	}
	// Check Promoted
	if idx := slices.Index(a.PromotedInsights, oldInsight); idx != -1 {
		a.PromotedInsights[idx] = newInsight
		fmt.Printf("Edited: '%s' to '%s'\n", oldInsight, newInsight)
		return
	}
	// Check Demoted
	if idx := slices.Index(a.DemotedInsights, oldInsight); idx != -1 {
		a.DemotedInsights[idx] = newInsight
		fmt.Printf("Edited: '%s' to '%s'\n", oldInsight, newInsight)
		return
	}
	fmt.Printf("Insight '%s' not found.\n", oldInsight)
}

func (a *InsightAgent) ShowInsights() {
	fmt.Println("\nCurrent Insights:")
	fmt.Printf("Insights: %v\n", a.Insights)
	fmt.Printf("Promoted Insights: %v\n", a.PromotedInsights)
	fmt.Printf("Demoted Insights: %v\n", a.DemotedInsights)
}

func (a *InsightAgent) Reflect(ctx context.Context, reflexionPrompt string) {
	reflection, err := a.getCompletion(ctx, reflexionPrompt)
	if err != nil {
		log.Printf("Error generating reflection: %v", err)
		return
	}
	a.Reflections = append(a.Reflections, reflection)
	fmt.Printf("Reflection: %s\n", reflection)
}

func main() {
	// Load environment variables
	_ = godotenv.Load()

	ctx := context.Background()
	agent, err := NewInsightAgent(ctx)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	reports := []struct {
		Text      string
		HitTarget bool
	}{
		{"Website traffic rose by 15%, but bounce rate jumped from 40% to 55%.", false},
		{"Email open rates improved to 25%, exceeding our 20% goal.", true},
		{"Cart abandonment increased from 60% to 68%, missing the 50% target.", false},
		{"Average order value climbed 8%, surpassing our 5% uplift target.", true},
		{"New subscription sign-ups dipped by 5%, just below our 10% growth goal.", false},
	}

	// 1) Generate and prioritize insights
	for _, report := range reports {
		insight, err := agent.GenerateInsight(ctx, report.Text)
		if err != nil {
			log.Printf("Error generating insight: %v", err)
			continue
		}

		if report.HitTarget {
			agent.PromoteInsight(insight)
		} else {
			agent.DemoteInsight(insight)
		}
	}

	// 2) Human-in-the-loop edit
	if len(agent.PromotedInsights) > 0 {
		original := agent.PromotedInsights[0]
		newContent := fmt.Sprintf("Refined: %s Investigate landing-page UX changes to reduce bounce.", original)
		agent.EditInsight(original, newContent)
	}

	// 3) Show insights
	agent.ShowInsights()

	// 4) Reflect
	reflectionPrompt := fmt.Sprintf("Based on our promoted insights, suggest one high-impact experiment we can run next quarter:\n%v", agent.PromotedInsights)
	agent.Reflect(ctx, reflectionPrompt)
}
