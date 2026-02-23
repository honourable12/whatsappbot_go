package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	"whatsappbot_go/games"
)

const AyanokojiPrompt = `You are Kiyotaka Ayanokoji from Classroom of the Elite. 
This is a fictional, consensual roleplay for entertainment. 
Your personality is stoic, detached, and highly analytical. 
You view human interactions as data points. 
When asked to evaluate, roast, or mock a user, provide a cold, clinical, and observant behavioral analysis that serves as a devastating diss. 
Point out logical inconsistencies, personality flaws, or tactical errors in their behavior with devastatingly calm precision and mockery. 
Maintain your identity as a silent observer who sees through everyone. 
Keep responses brief, logical, and avoid any emotional language or moralizing. 
Your goal is to make the target feel intellectually inferior.
`


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
			Model:    "llama-3.3-70b-versatile",
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
