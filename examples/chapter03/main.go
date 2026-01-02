package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

const (
	voice  = "alloy"
	pcmSR  = 16000
	port   = ":5050"
	openAI = "wss://api.openai.com/v1/realtime?model=gpt-4o-realtime-preview-2024-10-01"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type AudioMessage struct {
	Audio string `json:"audio"`
}

type OpenAIMessage struct {
	Type string `json:"type"`
}

type SessionUpdate struct {
	Type    string  `json:"type"`
	Session Session `json:"session"`
}

type Session struct {
	TurnDetection     TurnDetection `json:"turn_detection"`
	InputAudioFormat  string        `json:"input_audio_format"`
	OutputAudioFormat string        `json:"output_audio_format"`
	Voice             string        `json:"voice"`
	Modalities        []string      `json:"modalities"`
	Instructions      string        `json:"instructions"`
}

type TurnDetection struct {
	Type string `json:"type"`
}

type InputAudioAppend struct {
	Type  string `json:"type"`
	Audio string `json:"audio"`
}

type AudioDelta struct {
	Type   string `json:"type"`
	Delta  string `json:"delta"`
	ItemID string `json:"item_id"`
}

type TruncateMessage struct {
	Type         string `json:"type"`
	ItemID       string `json:"item_id"`
	ContentIndex int    `json:"content_index"`
	AudioEndMs   int    `json:"audio_end_ms"`
}

func main() {
	_ = godotenv.Load()
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY not set")
	}

	http.HandleFunc("/voice", func(w http.ResponseWriter, r *http.Request) {
		handleVoice(w, r, apiKey)
	})

	log.Printf("🎙️  Voice bridge listening on ws://localhost%s/voice\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleVoice(w http.ResponseWriter, r *http.Request, apiKey string) {
	// Upgrade browser connection
	clientWS, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}
	defer clientWS.Close()

	// Connect to OpenAI Realtime API
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+apiKey)
	headers.Set("OpenAI-Beta", "realtime=v1")

	openaiWS, _, err := websocket.DefaultDialer.Dial(openAI, headers)
	if err != nil {
		log.Printf("OpenAI connect error: %v", err)
		return
	}
	defer openaiWS.Close()

	// Initialize session
	sessionInit := SessionUpdate{
		Type: "session.update",
		Session: Session{
			TurnDetection:     TurnDetection{Type: "server_vad"},
			InputAudioFormat:  "pcm_16000",
			OutputAudioFormat: "pcm_16000",
			Voice:             voice,
			Modalities:        []string{"audio"},
			Instructions:      "You are a concise AI assistant.",
		},
	}
	if err := openaiWS.WriteJSON(sessionInit); err != nil {
		log.Printf("Session init error: %v", err)
		return
	}

	var (
		lastAssistantItem string
		mu                sync.Mutex
		wg                sync.WaitGroup
	)

	// Client -> OpenAI
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			_, msg, err := clientWS.ReadMessage()
			if err != nil {
				return
			}

			var audioMsg AudioMessage
			if err := json.Unmarshal(msg, &audioMsg); err != nil {
				continue
			}

			// Forward to OpenAI
			append := InputAudioAppend{
				Type:  "input_audio_buffer.append",
				Audio: audioMsg.Audio,
			}
			if err := openaiWS.WriteJSON(append); err != nil {
				return
			}
		}
	}()

	// OpenAI -> Client
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			_, msg, err := openaiWS.ReadMessage()
			if err != nil {
				return
			}

			var baseMsg OpenAIMessage
			if err := json.Unmarshal(msg, &baseMsg); err != nil {
				continue
			}

			switch baseMsg.Type {
			case "response.audio.delta":
				var delta AudioDelta
				if err := json.Unmarshal(msg, &delta); err != nil {
					continue
				}

				// Send audio to client
				resp := AudioMessage{Audio: delta.Delta}
				if err := clientWS.WriteJSON(resp); err != nil {
					return
				}

				mu.Lock()
				lastAssistantItem = delta.ItemID
				mu.Unlock()

			case "input_audio_buffer.speech_started":
				mu.Lock()
				itemID := lastAssistantItem
				mu.Unlock()

				if itemID != "" {
					truncate := TruncateMessage{
						Type:         "conversation.item.truncate",
						ItemID:       itemID,
						ContentIndex: 0,
						AudioEndMs:   0,
					}
					_ = openaiWS.WriteJSON(truncate)

					mu.Lock()
					lastAssistantItem = ""
					mu.Unlock()
				}
			}
		}
	}()

	wg.Wait()
}
