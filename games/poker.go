package games

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Player struct {
	JID  string `json:"jid"`
	Name string `json:"name"`
	Hand []Card `json:"hand"`
	Bet  int    `json:"bet"`
}

type CardGame struct {
	Type     string    `json:"type"` // "poker" or "blackjack"
	Players  []*Player `json:"players"`
	Deck     []Card    `json:"deck"`
	Rules    string    `json:"rules"`
	Active   bool      `json:"active"`
	Host     string    `json:"host"`
	CurrentP int       `json:"current_p"`
}

func (g *CardGame) Serialize() string {
	data, _ := json.Marshal(g)
	return string(data)
}

func DeserializeCardGame(state string) *CardGame {
	var g CardGame
	json.Unmarshal([]byte(state), &g)
	return &g
}

func HandleCardGame(chatID string, senderID string, senderName string, args []string) string {
	if len(args) == 0 {
		return "Usage: !ak poker [make/join/start/rules/help]"
	}

	cmd := strings.ToLower(args[0])
	
	switch cmd {
	case "help":
		return `🎴 *Poker/Blackjack Commands*
- !ak poker make [texas/blackjack]: Create a new game table.
- !ak poker join: Join the current table.
- !ak poker rules [text]: Set custom rules (Host only).
- !ak poker start: Start the game (Host only).
- !ak poker cancel: Cancel the current table (Host only).`

	case "make":
		if len(args) < 2 { return "Choose a type: !ak poker make texas OR !ak poker make blackjack" }
		gType := strings.ToLower(args[1])
		if gType != "texas" && gType != "blackjack" { return "Invalid game type." }

		game := &CardGame{
			Type:    gType,
			Players: []*Player{{JID: senderID, Name: senderName}},
			Deck:    NewDeck(),
			Rules:   "Standard",
			Active:  false,
			Host:    senderID,
		}
		Shuffle(game.Deck)
		SaveGameState(chatID, "poker", game.Serialize())
		return fmt.Sprintf("Table created for %s! Host: %s. Others can use !ak poker join to play.", gType, senderName)

	case "join":
		gameType, state, err := GetGameState(chatID)
		if err != nil || gameType != "poker" { return "No active table. Host one with !ak poker make" }
		game := DeserializeCardGame(state)
		if game.Active { return "The game has already started." }

		for _, p := range game.Players {
			if p.JID == senderID { return "You already joined." }
		}

		game.Players = append(game.Players, &Player{JID: senderID, Name: senderName})
		SaveGameState(chatID, "poker", game.Serialize())
		return fmt.Sprintf("%s has joined the table! Total players: %d", senderName, len(game.Players))

	case "rules":
		gameType, state, err := GetGameState(chatID)
		if err != nil || gameType != "poker" { return "No active table." }
		game := DeserializeCardGame(state)
		if senderID != game.Host { return "Only the host can set rules." }
		if len(args) < 2 { return "Current rules: " + game.Rules }

		game.Rules = strings.Join(args[1:], " ")
		SaveGameState(chatID, "poker", game.Serialize())
		return "Rules updated: " + game.Rules

	case "start":
		gameType, state, err := GetGameState(chatID)
		if err != nil || gameType != "poker" { return "No active table." }
		game := DeserializeCardGame(state)
		if senderID != game.Host { return "Only the host can start." }
		if len(game.Players) < 2 && game.Type == "texas" { return "Need at least 2 players for Texas Hold'em." }

		game.Active = true
		// Initial Deal
		for _, p := range game.Players {
			p.Hand = []Card{game.Deck[0], game.Deck[1]}
			game.Deck = game.Deck[2:]
		}

		SaveGameState(chatID, "poker", game.Serialize())
		return fmt.Sprintf("Game started! Type: %s. Rules: %s.\nWait for your cards in PM.", game.Type, game.Rules)

	case "cancel":
		gameType, _, err := GetGameState(chatID)
		if err != nil || gameType != "poker" { return "No active table." }
		DeleteGameState(chatID)
		return "Table cancelled."

	default:
		return "Unknown command. Use !ak poker help"
	}
}
