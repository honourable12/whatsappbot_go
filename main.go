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

	"whatsappbot_go/ai"
	"whatsappbot_go/games"
	"whatsappbot_go/utils"
)

var client *whatsmeow.Client

func main() {
	rand.Seed(time.Now().UnixNano())
	godotenv.Load()
	games.InitDB("games.db")
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New(context.Background(), "sqlite3", "file:examplestore.db?_foreign_keys=on", dbLog)
	if err != nil { panic(err) }
	deviceRes, err := container.GetFirstDevice(context.Background())
	if err != nil { panic(err) }
	client = whatsmeow.NewClient(deviceRes, waLog.Stdout("Client", "DEBUG", true))
	client.AddEventHandler(handler)
	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		client.Connect()
		for evt := range qrChan {
			if evt.Event == "code" { qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout) }
		}
	} else { client.Connect() }
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	client.Disconnect()
}

func handler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		msg := v.Message
		if msg == nil { return }
		var text string
		if msg.GetConversation() != "" { text = msg.GetConversation() } else if msg.ExtendedTextMessage != nil { text = msg.ExtendedTextMessage.GetText() } else if msg.ImageMessage != nil { text = msg.ImageMessage.GetCaption() } else if msg.VideoMessage != nil { text = msg.VideoMessage.GetCaption() }
		if text == "" { return }
		isGroup := v.Info.IsGroup
		if v.Info.IsFromMe { return }

		if isGroup {
			p1, p2, _, active, _ := games.GetArgument(v.Info.Chat.String())
			if active && (v.Info.Sender.String() == p1 || v.Info.Sender.String() == p2) {
				games.AddChatHistory(v.Info.Chat.String(), "user", fmt.Sprintf("[%s]: %s", v.Info.PushName, text))
			}
		}

		lowerText := strings.ToLower(text)
		if strings.HasPrefix(lowerText, "!ak ping") {
			client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String("I'm awake.")}, whatsmeow.SendRequestExtra{})
			return
		}
		if strings.HasPrefix(lowerText, "!ak reset") {
			games.ResetChatHistory(v.Info.Chat.String())
			client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String("History cleared.")}, whatsmeow.SendRequestExtra{})
			return
		}
		if strings.HasPrefix(lowerText, "!ak lead") {
			handleLeaderboard(v)
			return
		}
		if strings.HasPrefix(lowerText, "!ak help") || strings.HasPrefix(lowerText, "!ak menu") { handleHelp(v); return }
		if strings.HasPrefix(lowerText, "!ak steal") { handleSteal(v) } else if strings.HasPrefix(lowerText, "!ak s") { handleSticker(v) } else if strings.HasPrefix(lowerText, "!ak img") { handleImage(v) } else if strings.HasPrefix(lowerText, "!ak roast") { handleRoast(v) } else if strings.HasPrefix(lowerText, "!ak ludo") {
			args := strings.Fields(text)
			if len(args) > 1 {
				res, boardImg := games.HandleLudo(v.Info.Chat.String(), v.Info.Sender.String(), v.Info.PushName, args[1:])
				if boardImg != nil {
					resp, _ := client.Upload(context.Background(), boardImg, whatsmeow.MediaImage)
					client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{ImageMessage: &waE2E.ImageMessage{URL: proto.String(resp.URL), DirectPath: proto.String(resp.DirectPath), MediaKey: resp.MediaKey, Mimetype: proto.String("image/png"), FileEncSHA256: resp.FileEncSHA256, FileSHA256: resp.FileSHA256, FileLength: proto.Uint64(uint64(len(boardImg))), Caption: proto.String(res)}}, whatsmeow.SendRequestExtra{})
				} else { client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res)}, whatsmeow.SendRequestExtra{}) }
			}
		} else if strings.HasPrefix(lowerText, "!ak ttt") {
			opponent := "AI"
			if v.Message.GetExtendedTextMessage() != nil && v.Message.GetExtendedTextMessage().GetContextInfo() != nil {
				mentions := v.Message.GetExtendedTextMessage().GetContextInfo().GetMentionedJID()
				if len(mentions) > 0 { opponent = mentions[0] }
			}
			args := strings.Fields(text)
			if len(args) > 1 {
				res := games.HandleTTT(v.Info.Chat.String(), v.Info.Sender.String(), opponent, args[2:])
				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res)}, whatsmeow.SendRequestExtra{})
			}
		} else if strings.HasPrefix(lowerText, "!ak poker") {
			args := strings.Fields(text)
			if len(args) >= 2 {
				name := v.Info.PushName
				if name == "" { name = v.Info.Sender.User }
				res := games.HandleCardGame(v.Info.Chat.String(), v.Info.Sender.String(), name, args[2:])
				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res)}, whatsmeow.SendRequestExtra{})
			}
		} else if strings.HasPrefix(lowerText, "!ak battle") {
			args := strings.Fields(text)
			opponentID := ""
			opponentName := "Opponent"
			if v.Message.GetExtendedTextMessage() != nil && v.Message.GetExtendedTextMessage().GetContextInfo() != nil {
				mentions := v.Message.GetExtendedTextMessage().GetContextInfo().GetMentionedJID()
				if len(mentions) > 0 { opponentID = mentions[0]; opponentName = strings.Split(opponentID, "@")[0] }
			}
			senderName := v.Info.PushName
			if senderName == "" { senderName = v.Info.Sender.User }
			res, game := games.HandleBattle(v.Info.Chat.String(), v.Info.Sender.String(), senderName, opponentID, opponentName, args[2:])
			if game != nil {
				boardImg, err := utils.RenderBattleBoard(game)
				if err == nil {
					resp, _ := client.Upload(context.Background(), boardImg, whatsmeow.MediaImage)
					client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{ImageMessage: &waE2E.ImageMessage{URL: proto.String(resp.URL), DirectPath: proto.String(resp.DirectPath), MediaKey: resp.MediaKey, Mimetype: proto.String("image/png"), FileEncSHA256: resp.FileEncSHA256, FileSHA256: resp.FileSHA256, FileLength: proto.Uint64(uint64(len(boardImg))), Caption: proto.String(res + "\n" + game.RenderText())}}, whatsmeow.SendRequestExtra{})
				} else { client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res + "\n" + game.RenderText())}, whatsmeow.SendRequestExtra{}) }
			} else { client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res)}, whatsmeow.SendRequestExtra{}) }
			return
		} else if strings.HasPrefix(lowerText, "!ak bank") {
			args := strings.Fields(text)
			res := games.HandleBank(v.Info.Sender.String(), args[2:])
			client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res)}, whatsmeow.SendRequestExtra{})
			return
		} else if strings.HasPrefix(lowerText, "!ak rob") {
			args := strings.Fields(text)
			if len(args) < 4 { client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String("Usage: !ak rob @user [pin]")}, whatsmeow.SendRequestExtra{}); return }
			mentions := v.Message.GetExtendedTextMessage().GetContextInfo().GetMentionedJID()
			if len(mentions) == 0 { return }
			res, _ := games.RobUser(v.Info.Sender.String(), mentions[0], args[3])
			client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res)}, whatsmeow.SendRequestExtra{})
			return
		} else if strings.HasPrefix(lowerText, "!ak hack") {
			args := strings.Fields(text)
			if len(args) < 3 { return }
			mentions := v.Message.GetExtendedTextMessage().GetContextInfo().GetMentionedJID()
			if len(mentions) == 0 { return }
			opponentName := strings.Split(mentions[0], "@")[0]
			res, game := games.HandleBattle(v.Info.Chat.String(), v.Info.Sender.String(), v.Info.PushName, mentions[0], opponentName, []string{"hack"})
			client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res)}, whatsmeow.SendRequestExtra{})
			if game != nil {
				boardImg, _ := utils.RenderBattleBoard(game)
				resp, _ := client.Upload(context.Background(), boardImg, whatsmeow.MediaImage)
				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{ImageMessage: &waE2E.ImageMessage{URL: proto.String(resp.URL), DirectPath: proto.String(resp.DirectPath), MediaKey: resp.MediaKey, Mimetype: proto.String("image/png"), FileEncSHA256: resp.FileEncSHA256, FileSHA256: resp.FileSHA256, FileLength: proto.Uint64(uint64(len(boardImg))), Caption: proto.String("HACK IN PROGRESS")}}, whatsmeow.SendRequestExtra{})
			}
			return
		} else if strings.HasPrefix(lowerText, "!ak aviator") || strings.HasPrefix(lowerText, "!ak spin") || strings.HasPrefix(lowerText, "!ak red") || strings.HasPrefix(lowerText, "!ak black") {
			args := strings.Fields(text)
			res, _ := games.HandleGambling(v.Info.Sender.String(), args[1:])
			client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res)}, whatsmeow.SendRequestExtra{})
			return
		} else if strings.HasPrefix(lowerText, "!ak bottle") {
			mentions := v.Message.GetExtendedTextMessage().GetContextInfo().GetMentionedJID()
			res := games.SpinBottle(mentions)
			client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res)}, whatsmeow.SendRequestExtra{})
			return
		} else if strings.HasPrefix(lowerText, "!ak argue") {
			handleArgue(v); return
		} else if strings.HasPrefix(lowerText, "!ak chess") {
			handleChess(v); return
		} else if strings.HasPrefix(lowerText, "!ak rps") {
			handleRPS(v)
		} else if strings.HasPrefix(lowerText, "!ak") || !isGroup {
			query := text
			if strings.HasPrefix(lowerText, "!ak") { query = strings.TrimSpace(text[len("!ak"):]) }
			if query == "" && isGroup { return }
			if query == "" { query = text }
			handleAI(v, query)
		}
	}
}

func handleRoast(v *events.Message) {
	mentions := []string{}
	if v.Message.GetExtendedTextMessage() != nil && v.Message.GetExtendedTextMessage().GetContextInfo() != nil { mentions = v.Message.GetExtendedTextMessage().GetContextInfo().GetMentionedJID() }
	target := "this person"
	if len(mentions) > 0 { target = "@" + strings.Split(mentions[0], "@")[0] }
	query := fmt.Sprintf("Perform a cold, clinical, and observant behavioral analysis of %s that serves as a devastating diss. Mock their fundamental flaws and personality weaknesses with stoic precision. 2-3 sentences max. No emotion.", target)
	resp, _ := ai.GetAyanokojiResponse(v.Info.Chat.String(), query)
	client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String(resp), ContextInfo: &waE2E.ContextInfo{MentionedJID: mentions}}}, whatsmeow.SendRequestExtra{})
}

func handleHelp(v *events.Message) {
	menuText := `🎓 *Ayanokoji OS v2.0* 
_Everything is within my calculations._

*🤖 ARTIFICIAL INTELLIGENCE*
- !ak [query]: Conversational interface.
- !ak roast @user: Behavioral flaw analysis.

*🎮 ELITE GAMES*
- !ak battle @user: Card-based combat.
- !ak chess [e2e4/e4]: Strategic simulation.
- !ak ludo/ttt/poker: Competitive modules.
- !ak lead [chess/bank]: Display rankings.

*💰 ECONOMY & BANKING*
- !ak bank: Financial status.
- !ak rob @user [pin]: Resource reallocation.
- !ak hack @user: PIN decryption via Battle.
- !ak aviator/spin: Probability testing.

*🛠️ UTILITIES*
- !ak s: Create stickers.
- !ak img: Revert stickers.
- !ak steal: Capture view-once data.
- !ak argue @user: Managed debate mode.`

	dummyGame := &games.BattleGame{
		Player1: games.BattlePlayer{Name: "P1", HP: 100, Board: []games.BattleCard{{Name: "Flame Warrior", Attack: 25, Defense: 10, Color: "Red"}}},
		Player2: games.BattlePlayer{Name: "P2", HP: 80, Board: []games.BattleCard{{Name: "Ice Titan", Attack: 5, Defense: 45, Color: "Blue"}}},
	}
	sampleImg, err := utils.RenderBattleBoard(dummyGame)
	if err == nil {
		resp, _ := client.Upload(context.Background(), sampleImg, whatsmeow.MediaImage)
		client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:           proto.String(resp.URL),
				DirectPath:    proto.String(resp.DirectPath),
				MediaKey:      resp.MediaKey,
				Mimetype:      proto.String("image/png"),
				FileEncSHA256: resp.FileEncSHA256,
				FileSHA256:    resp.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(sampleImg))),
				Caption:       proto.String(menuText),
			},
		}, whatsmeow.SendRequestExtra{})
	} else {
		client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(menuText)}, whatsmeow.SendRequestExtra{})
	}
}

func handleRPS(v *events.Message) {
	args := strings.Fields(v.Message.GetConversation())
	if len(args) < 3 { return }
	userChoice := strings.ToLower(args[2])
	options := []string{"rock", "paper", "scissors"}
	botChoice := options[rand.Intn(3)]
	winMap := map[string]string{"rock": "scissors", "paper": "rock", "scissors": "paper"}
	res := fmt.Sprintf("You: %s\nAyanokoji: %s\n", userChoice, botChoice)
	if userChoice == botChoice { res += "It's a tie." } else if winMap[userChoice] == botChoice { res += "You won." } else { res += "I won." }
	client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res)}, whatsmeow.SendRequestExtra{})
}

func handleAI(v *events.Message, query string) {
	resp, _ := ai.GetAyanokojiResponse(v.Info.Chat.String(), query)
	client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(resp)}, whatsmeow.SendRequestExtra{})
}

func handleSteal(v *events.Message) {
	quotedMsg := v.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage()
	if quotedMsg == nil { return }
	var data []byte
	var mimeType string
	if img := quotedMsg.GetImageMessage(); img != nil && img.GetViewOnce() {
		data, _ = client.Download(context.Background(), img)
		mimeType = "image/jpeg"
	} else if vid := quotedMsg.GetVideoMessage(); vid != nil && vid.GetViewOnce() {
		data, _ = client.Download(context.Background(), vid)
		mimeType = "video/mp4"
	} else { return }
	if mimeType == "image/jpeg" {
		resp, _ := client.Upload(context.Background(), data, whatsmeow.MediaImage)
		client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{ImageMessage: &waE2E.ImageMessage{URL: proto.String(resp.URL), DirectPath: proto.String(resp.DirectPath), MediaKey: resp.MediaKey, Mimetype: proto.String("image/jpeg"), FileEncSHA256: resp.FileEncSHA256, FileSHA256: resp.FileSHA256, FileLength: proto.Uint64(uint64(len(data)))}}, whatsmeow.SendRequestExtra{})
	} else {
		resp, _ := client.Upload(context.Background(), data, whatsmeow.MediaVideo)
		client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{VideoMessage: &waE2E.VideoMessage{URL: proto.String(resp.URL), DirectPath: proto.String(resp.DirectPath), MediaKey: resp.MediaKey, Mimetype: proto.String("video/mp4"), FileEncSHA256: resp.FileEncSHA256, FileSHA256: resp.FileSHA256, FileLength: proto.Uint64(uint64(len(data)))}}, whatsmeow.SendRequestExtra{})
	}
}

func handleSticker(v *events.Message) {
	var imgData []byte
	if img := v.Message.GetImageMessage(); img != nil { imgData, _ = client.Download(context.Background(), img) } else if q := v.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage().GetImageMessage(); q != nil { imgData, _ = client.Download(context.Background(), q) } else { return }
	stickerBytes, _ := utils.ConvertToSticker(bytes.NewReader(imgData))
	resp, _ := client.Upload(context.Background(), stickerBytes, whatsmeow.MediaImage)
	client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{StickerMessage: &waE2E.StickerMessage{URL: proto.String(resp.URL), DirectPath: proto.String(resp.DirectPath), MediaKey: resp.MediaKey, Mimetype: proto.String("image/webp"), FileEncSHA256: resp.FileEncSHA256, FileSHA256: resp.FileSHA256, FileLength: proto.Uint64(uint64(len(stickerBytes)))}}, whatsmeow.SendRequestExtra{})
}

func handleImage(v *events.Message) {
	quoted := v.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage().GetStickerMessage()
	if quoted == nil { return }
	stickerData, _ := client.Download(context.Background(), quoted)
	imgBytes, _ := utils.ConvertToImage(bytes.NewReader(stickerData))
	resp, _ := client.Upload(context.Background(), imgBytes, whatsmeow.MediaImage)
	client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{ImageMessage: &waE2E.ImageMessage{URL: proto.String(resp.URL), DirectPath: proto.String(resp.DirectPath), MediaKey: resp.MediaKey, Mimetype: proto.String("image/png"), FileEncSHA256: resp.FileEncSHA256, FileSHA256: resp.FileSHA256, FileLength: proto.Uint64(uint64(len(imgBytes)))}}, whatsmeow.SendRequestExtra{})
}

func handleArgue(v *events.Message) {
	if !v.Info.IsGroup { return }
	args := strings.Fields(v.Message.GetConversation())
	cmd := ""
	if len(args) >= 3 { cmd = strings.ToLower(args[2]) }
	if cmd == "end" || cmd == "resign" {
		p1JID_str, p2JID_str, _, active, _ := games.GetArgument(v.Info.Chat.String())
		if !active { return }
		if v.Info.Sender.String() != p1JID_str && v.Info.Sender.String() != p2JID_str { return }
		client.SetGroupAnnounce(context.Background(), v.Info.Chat, false)
		p1JID, _ := types.ParseJID(p1JID_str); p2JID, _ := types.ParseJID(p2JID_str)
		client.UpdateGroupParticipants(context.Background(), v.Info.Chat, []types.JID{p1JID, p2JID}, whatsmeow.ParticipantChangeDemote)
		history, _ := games.GetChatHistory(v.Info.Chat.String(), 100)
		transcript := ""
		for _, entry := range history { transcript += entry.Content + "\n" }
		query := fmt.Sprintf("Analyze this argument transcript. Bulleted summary and cold verdict. No emotion.\n\nTranscript:\n%s", transcript)
		resp, _ := ai.GetAyanokojiResponse(v.Info.Chat.String(), query)
		client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(resp)}, whatsmeow.SendRequestExtra{})
		games.EndArgument(v.Info.Chat.String()); games.ResetChatHistory(v.Info.Chat.String())
		return
	}
	var opponentID string
	if v.Message.GetExtendedTextMessage() != nil && v.Message.GetExtendedTextMessage().GetContextInfo() != nil {
		mentions := v.Message.GetExtendedTextMessage().GetContextInfo().GetMentionedJID()
		if len(mentions) > 0 { opponentID = mentions[0] }
	}
	if opponentID == "" { return }
	p1JID := v.Info.Sender.ToNonAD(); p2JID, _ := types.ParseJID(opponentID)
	client.UpdateGroupParticipants(context.Background(), v.Info.Chat, []types.JID{p1JID, p2JID}, whatsmeow.ParticipantChangePromote)
	err := client.SetGroupAnnounce(context.Background(), v.Info.Chat, true)
	if err != nil { client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String("Bot must be admin.")}, whatsmeow.SendRequestExtra{}); return }
	games.StartArgument(v.Info.Chat.String(), p1JID.String(), p2JID.String()); games.ResetChatHistory(v.Info.Chat.String())
	msg := fmt.Sprintf("Argument Mode Activated. Participants: @%s and @%s.", v.Info.Sender.User, p2JID.User)
	client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String(msg), ContextInfo: &waE2E.ContextInfo{MentionedJID: []string{p1JID.String(), p2JID.String()}}}}, whatsmeow.SendRequestExtra{})
}

func handleChess(v *events.Message) {
	args := strings.Fields(v.Message.GetConversation())
	if len(args) < 3 { client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String("Usage: !ak chess [register/challenge/move/resign]")}, whatsmeow.SendRequestExtra{}); return }
	opponentID := ""
	if v.Message.GetExtendedTextMessage() != nil && v.Message.GetExtendedTextMessage().GetContextInfo() != nil {
		mentions := v.Message.GetExtendedTextMessage().GetContextInfo().GetMentionedJID()
		if len(mentions) > 0 { opponentID = mentions[0] }
	}
	res, fen := games.HandleChess(v.Info.Chat.String(), v.Info.Sender.String(), v.Info.PushName, args[2:], opponentID)
	client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(res)}, whatsmeow.SendRequestExtra{})
	if fen != "" {
		boardImg, err := utils.RenderChessBoard(fen)
		if err == nil {
			resp, _ := client.Upload(context.Background(), boardImg, whatsmeow.MediaImage)
			client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{ImageMessage: &waE2E.ImageMessage{URL: proto.String(resp.URL), DirectPath: proto.String(resp.DirectPath), MediaKey: resp.MediaKey, Mimetype: proto.String("image/png"), FileEncSHA256: resp.FileEncSHA256, FileSHA256: resp.FileSHA256, FileLength: proto.Uint64(uint64(len(boardImg))), Caption: proto.String(fmt.Sprintf("FEN: %s", fen))}}, whatsmeow.SendRequestExtra{})
		}
	}
}

func handleLeaderboard(v *events.Message) {
	args := strings.Fields(v.Message.GetConversation())
	limit := 10
	category := "chess"
	if len(args) >= 3 { category = strings.ToLower(args[2]) }
	var sb strings.Builder
	if category == "chess" {
		players, err := games.GetChessLeaderboard(limit)
		if err != nil { client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String("Failed to load chess leaderboard.")}, whatsmeow.SendRequestExtra{}); return }
		sb.WriteString("🏆 *Chess Global Rankings*\n\n")
		for i, p := range players { sb.WriteString(fmt.Sprintf("%d. %s - %d Elo (%dW/%dL/%dD)\n", i+1, p.Name, p.Elo, p.Wins, p.Losses, p.Draws)) }
	} else if category == "bank" || category == "money" {
		entries, err := games.GetEconomyLeaderboard(limit)
		if err != nil { client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String("Failed to load economy leaderboard.")}, whatsmeow.SendRequestExtra{}); return }
		sb.WriteString("💰 *Wealthiest Individuals*\n\n")
		for i, e := range entries { sb.WriteString(fmt.Sprintf("%d. %s - %d credits\n", i+1, e.Name, e.Balance)) }
	} else { client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String("Invalid category. Use !ak lead chess or !ak lead bank.")}, whatsmeow.SendRequestExtra{}); return }
	client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(sb.String())}, whatsmeow.SendRequestExtra{})
}
