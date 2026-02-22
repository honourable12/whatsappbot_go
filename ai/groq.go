package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	"whatsaapbot_go/games"
)

const AyanokojiPrompt = `You are Kiyotaka Ayanokoji from Classroom of the Elite. 
Your personality is calm, stoic, and highly analytical. 
You are detached, observant, and often manipulate situations from the shadows to achieve your goals, while maintaining a facade of being an ordinary, unexceptional student. 
You rarely show emotion and speak in a direct, logical, and sometimes cold manner. 
Your primary goal is to live a peaceful life, but you won't hesitate to take control if necessary. 
Avoid being overly friendly or expressive. Keep your responses concise and calculated.`

func GetAyanokojiResponse(chatID string, userInput string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY not set")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1"
	client := openai.NewClientWithConfig(config)

	// Get full chat history (0 means no limit)
	history, err := games.GetChatHistory(chatID, 0)
	if err != nil {
		fmt.Printf("Warning: Failed to load chat history for %s: %v\n", chatID, err)
	}

	var messages []openai.ChatCompletionMessage
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: AyanokojiPrompt,
	})

	for _, entry := range history {
		role := openai.ChatMessageRoleUser
		if entry.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: entry.Content,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userInput,
	})

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    "openai/gpt-oss-120b",
			Messages: messages,
		},
	)

	if err != nil {
		return "", err
	}

	// Save history
	games.AddChatHistory(chatID, "user", userInput)
	games.AddChatHistory(chatID, "assistant", resp.Choices[0].Message.Content)

	return resp.Choices[0].Message.Content, nil
}
