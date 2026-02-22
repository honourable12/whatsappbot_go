package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	"whatsaapbot_go/ai"
	"whatsaapbot_go/games"
	"whatsaapbot_go/utils"
)

var client *whatsmeow.Client

func main() {
	// Seed random for games
	rand.Seed(time.Now().UnixNano())

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Error loading .env file, using system environment variables")
	}

	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		fmt.Println("CRITICAL: GROQ_API_KEY not set. Bot will not be able to use AI features.")
	} else {
		fmt.Println("GROQ_API_KEY is set correctly.")
	}

	// Initialize games database
	if err := games.InitDB("games.db"); err != nil {
		fmt.Printf("Failed to initialize games database: %v\n", err)
	}

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New(context.Background(), "sqlite3", "file:examplestore.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	deviceRes, err := container.GetFirstDevice(context.Background())
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client = whatsmeow.NewClient(deviceRes, clientLog)
	client.AddEventHandler(handler)

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("QR channel result:", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}

func handler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		msg := v.Message
		if msg == nil {
			return
		}

		// Get message text content safely
		var text string
		if msg.GetConversation() != "" {
			text = msg.GetConversation()
		} else if msg.ExtendedTextMessage != nil {
			text = msg.ExtendedTextMessage.GetText()
		} else if msg.ImageMessage != nil {
			text = msg.ImageMessage.GetCaption()
		} else if msg.VideoMessage != nil {
			text = msg.VideoMessage.GetCaption()
		}

		if text == "" {
			return
		}

		isGroup := v.Info.IsGroup
		isMe := v.Info.IsFromMe

		if isMe {
			return
		}

		fmt.Printf("Received message from %s (Group: %v): %s\n", v.Info.Sender.String(), isGroup, text)

		lowerText := strings.ToLower(text)

		// Features
		if strings.HasPrefix(lowerText, "!ak ping") {
			client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
				Conversation: proto.String("I'm awake."),
			}, whatsmeow.SendRequestExtra{})
			return
		}

		if strings.HasPrefix(lowerText, "!ak reset") {
			games.ResetChatHistory(v.Info.Chat.String())
			client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
				Conversation: proto.String("Conversation history has been cleared."),
			}, whatsmeow.SendRequestExtra{})
			return
		}

		if strings.HasPrefix(lowerText, "!ak help") {
			handleHelp(v)
			return
		}

		if strings.HasPrefix(lowerText, "!ak steal") {
			handleSteal(v)
		} else if strings.HasPrefix(lowerText, "!ak s") {
			handleSticker(v)
		} else if strings.HasPrefix(lowerText, "!ak img") {
			handleImage(v)
		} else if strings.HasPrefix(lowerText, "!ak ttt") {
			opponent := "AI"
			if v.Message.GetExtendedTextMessage() != nil && v.Message.GetExtendedTextMessage().GetContextInfo() != nil {
				mentions := v.Message.GetExtendedTextMessage().GetContextInfo().GetMentionedJID()
				if len(mentions) > 0 {
					opponent = mentions[0]
				}
			}
			
			args := strings.Fields(text)
			if len(args) > 1 {
				// args[0] = "!ak", args[1] = "ttt"
				res := games.HandleTTT(v.Info.Chat.String(), v.Info.Sender.String(), opponent, args[2:])
				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res)}, whatsmeow.SendRequestExtra{})
			}
		} else if strings.HasPrefix(lowerText, "!ak poker") || strings.HasPrefix(lowerText, "!ak blackjack") {
			args := strings.Fields(text)
			if len(args) >= 2 {
				name := v.Info.PushName
				if name == "" {
					name = v.Info.Sender.User
				}
				// args[0] = "!ak", args[1] = "poker"/"blackjack", args[2:] = subcommands
				res := games.HandleCardGame(v.Info.Chat.String(), v.Info.Sender.String(), name, args[2:])
				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res)}, whatsmeow.SendRequestExtra{})
				
				// Handle private card dealing if game started
				if strings.Contains(res, "Game started!") {
					_, state, _ := games.GetGameState(v.Info.Chat.String())
					game := games.DeserializeCardGame(state)
					for _, p := range game.Players {
						targetJID, _ := types.ParseJID(p.JID)
						cards := ""
						for _, c := range p.Hand {
							cards += fmt.Sprintf("[%s%s] ", c.Value, c.Suit)
						}
						client.SendMessage(context.Background(), targetJID, &waE2E.Message{
							Conversation: proto.String(fmt.Sprintf("Your cards for the %s game in %s:\n%s", game.Type, v.Info.Chat.String(), cards)),
						}, whatsmeow.SendRequestExtra{})
					}
				}
			}
		} else if strings.HasPrefix(lowerText, "!ak rps") {
			handleRPS(v)
		} else if strings.HasPrefix(lowerText, "!ak") || !isGroup {
			// Activation for AI in group or any message in private
			query := text
			if strings.HasPrefix(lowerText, "!ak") {
				query = strings.TrimSpace(text[len("!ak"):])
			}
			
			if query == "" && isGroup {
				return
			}
			if query == "" {
				query = text // In private, use the whole text
			}
			handleAI(v, query)
		}
	}
}

func handleHelp(v *events.Message) {
	helpText := `🤖 *Ayanokoji Bot Commands*

*General:*
- !ak help: Show this help message.
- !ak ping: Check if the bot is alive.
- !ak reset: Clear your chat history with Ayanokoji.

*AI:*
- !ak [your question]: Chat with Ayanokoji (AI). Full history memory.

*Media:*
- !ak steal: Reply to a view-once image/video to "steal" it.
- !ak s: Reply to an image to convert it into a sticker.
- !ak img: Reply to a sticker to convert it back to an image.

*Games:*
- !ak ttt [@user]: Play/Challenge Tic-Tac-Toe.
- !ak poker make [texas/blackjack]: Create a card game table.
- !ak poker join: Join the card game.
- !ak poker start: Start the table.
- !ak rps [rock/paper/scissors]: Play RPS with Ayanokoji.`
	client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
		Conversation: proto.String(helpText),
	}, whatsmeow.SendRequestExtra{})
}

func handleRPS(v *events.Message) {
	args := strings.Fields(v.Message.GetConversation())
	if len(args) < 3 {
		client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String("Usage: !ak rps [rock/paper/scissors]")}, whatsmeow.SendRequestExtra{})
		return
	}
	userChoice := strings.ToLower(args[2])
	options := []string{"rock", "paper", "scissors"}
	botChoice := options[rand.Intn(3)]

	winMap := map[string]string{"rock": "scissors", "paper": "rock", "scissors": "paper"}

	res := fmt.Sprintf("You: %s\nAyanokoji: %s\n", userChoice, botChoice)
	if userChoice == botChoice {
		res += "It's a tie. Calculated."
	} else if winMap[userChoice] == botChoice {
		res += "You won? Unforeseen."
	} else {
		res += "I won. Everything is within my calculations."
	}

	client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res)}, whatsmeow.SendRequestExtra{})
}

func handleAI(v *events.Message, query string) {
	fmt.Printf("Generating AI response for: %s\n", query)
	resp, err := ai.GetAyanokojiResponse(v.Info.Chat.String(), query)
	if err != nil {
		fmt.Println("Error getting AI response:", err)
		client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
			Conversation: proto.String("I'm busy right now. (API Error)"),
		}, whatsmeow.SendRequestExtra{})
		return
	}
	
	client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
		Conversation: proto.String(resp),
	}, whatsmeow.SendRequestExtra{})
}

func handleSteal(v *events.Message) {
	quotedMsg := v.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage()
	if quotedMsg == nil {
		client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String("Reply to a view-once message.")}, whatsmeow.SendRequestExtra{})
		return
	}

	var data []byte
	var err error
	var mimeType string

	if img := quotedMsg.GetImageMessage(); img != nil && img.GetViewOnce() {
		data, err = client.Download(context.Background(), img)
		mimeType = "image/jpeg"
	} else if vid := quotedMsg.GetVideoMessage(); vid != nil && vid.GetViewOnce() {
		data, err = client.Download(context.Background(), vid)
		mimeType = "video/mp4"
	} else {
		client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String("That's not a view-once message.")}, whatsmeow.SendRequestExtra{})
		return
	}

	if err != nil {
		fmt.Println("Download error:", err)
		client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String("Failed to download media.")}, whatsmeow.SendRequestExtra{})
		return
	}

	if mimeType == "image/jpeg" {
		resp, err := client.Upload(context.Background(), data, whatsmeow.MediaImage)
		if err != nil {
			return
		}
		client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:           proto.String(resp.URL),
				DirectPath:    proto.String(resp.DirectPath),
				MediaKey:      resp.MediaKey,
				Mimetype:      proto.String(mimeType),
				FileEncSHA256: resp.FileEncSHA256,
				FileSHA256:    resp.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(data))),
			},
		}, whatsmeow.SendRequestExtra{})
	} else {
		resp, err := client.Upload(context.Background(), data, whatsmeow.MediaVideo)
		if err != nil {
			return
		}
		client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				URL:           proto.String(resp.URL),
				DirectPath:    proto.String(resp.DirectPath),
				MediaKey:      resp.MediaKey,
				Mimetype:      proto.String(mimeType),
				FileEncSHA256: resp.FileEncSHA256,
				FileSHA256:    resp.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(data))),
			},
		}, whatsmeow.SendRequestExtra{})
	}
}

func handleSticker(v *events.Message) {
	msg := v.Message
	var imgData []byte
	var err error

	if img := msg.GetImageMessage(); img != nil {
		imgData, err = client.Download(context.Background(), img)
	} else if msg.GetExtendedTextMessage() != nil && msg.GetExtendedTextMessage().GetContextInfo() != nil && msg.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage().GetImageMessage() != nil {
		imgData, err = client.Download(context.Background(), msg.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage().GetImageMessage())
	} else {
		client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String("Send an image or reply to one with !Ak s")}, whatsmeow.SendRequestExtra{})
		return
	}

	if err != nil {
		return
	}

	stickerBytes, err := utils.ConvertToSticker(bytes.NewReader(imgData))
	if err != nil {
		return
	}

	resp, err := client.Upload(context.Background(), stickerBytes, whatsmeow.MediaImage)
	if err != nil {
		return
	}

	client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
		StickerMessage: &waE2E.StickerMessage{
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			MediaKey:      resp.MediaKey,
			Mimetype:      proto.String("image/webp"),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(stickerBytes))),
		},
	}, whatsmeow.SendRequestExtra{})
}

func handleImage(v *events.Message) {
	if v.Message.GetExtendedTextMessage() == nil || v.Message.GetExtendedTextMessage().GetContextInfo() == nil || v.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage().GetStickerMessage() == nil {
		client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String("Reply to a sticker with !Ak img")}, whatsmeow.SendRequestExtra{})
		return
	}
	quoted := v.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage().GetStickerMessage()

	stickerData, err := client.Download(context.Background(), quoted)
	if err != nil {
		return
	}

	imgBytes, err := utils.ConvertToImage(bytes.NewReader(stickerData))
	if err != nil {
		return
	}

	resp, err := client.Upload(context.Background(), imgBytes, whatsmeow.MediaImage)
	if err != nil {
		return
	}

	client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			MediaKey:      resp.MediaKey,
			Mimetype:      proto.String("image/png"),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(imgBytes))),
		},
	}, whatsmeow.SendRequestExtra{})
}
