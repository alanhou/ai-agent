package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

var chatModel model.ChatModel

const reflexionPrompt = `You will be given the history of a past experience in which you were 
placed in an environment and given a task to complete. You were unsuccessful in 
completing the task. Do not summarize your environment, but rather think about 
the strategy and path you took to attempt to complete the task. 
Devise a concise, new plan of action that accounts for your mistake with reference 
to specific actions that you should have taken. For example, if you tried A and B but forgot C, 
then devise a plan to achieve C with environment-specific actions. You will need this 
later when you are solving the same task. Give your plan after "Plan". 

Instruction:
i am looking for dairy free and apple variety pack of chips, and price lower than 30.00 dollars
[Search]

Action: search[dairy free and apple variety pack of chips]
Observation:
[Back to Search]
Page 1 (Total results: 50)
[Next >]
[B07HRFJWP8]
Enjoy Life Foods Soft Baked Ovals, Breakfast Bars, Nut Free Bars, Soy Free, Dairy Free, Non GMO, Gluten Free, Vegan, Variety Pack, 4 Boxes (20 Bars Total)
$100.0
[B01KMHY5PG]
Enjoy Life Soft Baked Chewy Bars, Variety Pack, Nut Free Bars, Soy Free, Dairy Free, Gluten Free, 6 Boxes (30 Total Bars)
$21.49
[B008D2X8C4]
Enjoy Life Lentil Chips Variety Pack, Dairy Free Chips, Soy Free, Nut Free, Non GMO, Vegan, Gluten Free, 24 Bags (0.8 oz)
$100.0

Action: think[The first two items are not what I want. The third item is the dairy free and apple variety pack of chips I want.]
Observation: OK.

Action: click[B008D2X8C4]
Observation:
[Back to Search]
[< Prev]
flavor name [dill and sour cream][garlic & parmesan][light sea salt][margherita pizza][thai chili lime][variety pack]
size [0.8 ounce (pack of 24)][4 ounce (pack of 12)]
Price: $100.0
Rating: N.A.
[Description]
[Features]
[Reviews]
[Buy Now]

Action: think[The item has options 'variety pack' and '0.8 ounce (pack of 24)' and seems good to buy.]
Observation: OK.

Action: click[variety pack]
Observation: You have clicked variety pack.

Action: click[0.8 ounce (pack of 24)]
Observation: You have clicked 0.8 ounce (pack of 24).

Action: click[Buy Now]

STATUS: FAIL

Plan:
`

type EnvConfig struct {
	Name      string
	IsSuccess bool
	Skip      bool
	Memory    []string
}

func getCompletion(ctx context.Context, prompt string) (string, error) {
	resp, err := chatModel.Generate(ctx, []*schema.Message{
		schema.UserMessage(prompt),
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func generateReflectionQuery(trialLog string, recentReflections []string) string {
	history := strings.Join(recentReflections, "\n\n")
	return fmt.Sprintf("%s\n\n%s\n\nBased on the above, what plan would you follow next? Plan:", history, trialLog)
}

func updateMemory(ctx context.Context, trialLogPath string, envConfigs []*EnvConfig) error {
	content, err := os.ReadFile(trialLogPath)
	if err != nil {
		return err
	}
	fullLog := string(content)

	envLogs := strings.Split(fullLog, "#####\n\n#####")
	// If only one log or split logic differs somewhat, handle simple case:
	// For this simple example with one file, if split results in 1 item even if delimiter not found, we interpret it as correspond to 1 config
	if len(envLogs) != len(envConfigs) {
		// Loosen check for simple single-case example if format doesn't match perfectly
		if len(envConfigs) == 1 && len(envLogs) == 1 {
			// acceptable
		} else {
			return fmt.Errorf("bad: envLogs length %d != envConfigs length %d", len(envLogs), len(envConfigs))
		}
	}

	for i, env := range envConfigs {
		if !env.IsSuccess && !env.Skip {
			var memory []string
			if len(env.Memory) > 3 {
				memory = env.Memory[len(env.Memory)-3:]
			} else {
				memory = env.Memory
			}

			reflectionQuery := generateReflectionQuery(envLogs[i], memory)
			reflection, err := getCompletion(ctx, reflectionQuery)
			if err != nil {
				return err
			}
			envConfigs[i].Memory = append(envConfigs[i].Memory, reflection)
		}
	}
	return nil
}

func main() {
	// Load .env from project root
	_ = godotenv.Load()
	// Create context
	ctx := context.Background()

	// Init Eino ChatModel
	var err error
	chatModel, err = openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   "gpt-4o",
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
	})
	if err != nil {
		log.Fatalf("Failed to init chat model: %v", err)
	}

	// 1. Initial Prompt (Simulating the "reflexion" node)
	fmt.Println("Running initial reflexion prompt...")
	result, err := getCompletion(ctx, reflexionPrompt)
	if err != nil {
		log.Fatalf("Error getting completion: %v", err)
	}
	fmt.Printf("Initial Result:\n%s\n\n", result)

	// Determine trial log path
	// If running from root, path is examples/chapter07/python/trial_log.txt or we create local one?
	// The prompt implies "reflexion.py" logic uses `os.path.dirname(__file__)` which means it looks in `examples/chapter07/python/`
	// Let's stick to a local path relative to where we run, OR match the python behavior if intended.
	// Since user said "execute from project root", and the python script used sibling file.
	// To keep it simple and consistent with the python "create dummy if not exists", we'll just use a local path `examples/chapter07/go/trial_log.txt`
	// or just `trial_log.txt` in the CWD?
	// Let's use `examples/chapter07/go/trial_log.txt` since we are placing main.go there,
	// BUT user runs from root.

	// Use runtime.Caller to get the directory of the current file
	_, filename, _, _ := runtime.Caller(0)
	trialLogPath := filepath.Join(filepath.Dir(filename), "trial_log.txt")
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(trialLogPath), 0755); err != nil {
		log.Fatalf("Failed to create dir: %v", err)
	}

	// Create dummy log file if it doesn't exist
	if _, err := os.Stat(trialLogPath); os.IsNotExist(err) {
		dummyLog := `Action: search[dairy free and apple variety pack of chips]
Observation:
[Back to Search]
Page 1 (Total results: 50)
[Next >]
[B07HRFJWP8]
Enjoy Life Foods Soft Baked Ovals, Breakfast Bars, Nut Free Bars, Soy Free, Dairy Free, Non GMO, Gluten Free, Vegan, Variety Pack, 4 Boxes (20 Bars Total)
$100.0
Action: think[The first item is not what I want.]
STATUS: FAIL`
		err := os.WriteFile(trialLogPath, []byte(dummyLog), 0644)
		if err != nil {
			log.Printf("Warning: failed to create dummy log: %v", err)
		}
	}

	envConfigs := []*EnvConfig{
		{
			Name:      "env1",
			IsSuccess: false,
			Skip:      false,
			Memory:    []string{},
		},
	}

	fmt.Println("Updating memory...")
	err = updateMemory(ctx, trialLogPath, envConfigs)
	if err != nil {
		log.Fatalf("Error updating memory: %v", err)
	}

	fmt.Printf("Updated env_configs: %+v\n", envConfigs[0])
	fmt.Println("Reflections:", envConfigs[0].Memory)
}
