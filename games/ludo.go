package games

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
)

type Piece struct {
	Pos      int  `json:"pos"`       // -1: Base, 0-51: Common, 52-57: Home stretch, 58: Finished
	InBase   bool `json:"in_base"`
	Finished bool `json:"finished"`
}

type LudoPlayer struct {
	JID    string   `json:"jid"`
	Name   string   `json:"name"`
	Color  string   `json:"color"` // Red, Blue, Yellow, Green
	Pieces [4]Piece `json:"pieces"`
}

type LudoGame struct {
	Players     []*LudoPlayer `json:"players"`
	Active      bool          `json:"active"`
	TurnIdx     int           `json:"turn_idx"`
	LastRoll    int           `json:"last_roll"`
	RollPending bool          `json:"roll_pending"`
	Host        string        `json:"host"`
}

func (g *LudoGame) Serialize() string {
	data, _ := json.Marshal(g)
	return string(data)
}

func DeserializeLudo(state string) *LudoGame {
	var g LudoGame
	json.Unmarshal([]byte(state), &g)
	return &g
}

func HandleLudo(chatID string, senderID string, senderName string, args []string) (string, []byte) {
	if len(args) == 0 {
		return "Usage: !ak ludo [make/join/start/roll/move]", nil
	}

	cmd := strings.ToLower(args[0])
	switch cmd {
	case "make":
		game := &LudoGame{
			Players: []*LudoPlayer{{JID: senderID, Name: senderName, Color: "Red", Pieces: [4]Piece{{InBase: true}, {InBase: true}, {InBase: true}, {InBase: true}}}},
			Host:    senderID,
			Active:  false,
		}
		SaveGameState(chatID, "ludo", game.Serialize())
		return fmt.Sprintf("Ludo lobby created! Host: %s (Red). Use !ak ludo join to pick a color.", senderName), nil

	case "join":
		gType, state, err := GetGameState(chatID)
		if err != nil || gType != "ludo" { return "No active Ludo lobby.", nil }
		game := DeserializeLudo(state)
		if game.Active { return "Game already started.", nil }
		if len(game.Players) >= 4 { return "Lobby full.", nil }
		for _, p := range game.Players {
			if p.JID == senderID { return "You're already in.", nil }
		}

		colors := []string{"Blue", "Yellow", "Green"}
		color := colors[len(game.Players)-1]
		game.Players = append(game.Players, &LudoPlayer{
			JID: senderID, Name: senderName, Color: color,
			Pieces: [4]Piece{{InBase: true}, {InBase: true}, {InBase: true}, {InBase: true}},
		})
		SaveGameState(chatID, "ludo", game.Serialize())
		return fmt.Sprintf("%s joined as %s!", senderName, color), nil

	case "start":
		gType, state, err := GetGameState(chatID)
		if err != nil || gType != "ludo" { return "No active Ludo lobby.", nil }
		game := DeserializeLudo(state)
		if senderID != game.Host { return "Only host can start.", nil }
		if len(game.Players) < 2 { return "Need at least 2 players.", nil }
		
		game.Active = true
		game.TurnIdx = 0
		game.RollPending = true
		SaveGameState(chatID, "ludo", game.Serialize())
		return fmt.Sprintf("Ludo started! %s's turn (Red). Use !ak ludo roll", game.Players[0].Name), nil

	case "roll":
		gType, state, err := GetGameState(chatID)
		if err != nil || gType != "ludo" { return "No active game.", nil }
		game := DeserializeLudo(state)
		if !game.Active { return "Game hasn't started.", nil }
		
		p := game.Players[game.TurnIdx]
		if p.JID != senderID { return fmt.Sprintf("It's %s's turn.", p.Name), nil }
		if !game.RollPending { return "You already rolled. Move a piece with !ak ludo move [1-4]", nil }

		roll := rand.Intn(6) + 1
		game.LastRoll = roll
		game.RollPending = false
		
		// Check if any move is possible
		canMove := false
		for _, pc := range p.Pieces {
			if pc.InBase && roll == 6 { canMove = true; break }
			if !pc.InBase && !pc.Finished { canMove = true; break }
		}

		if !canMove {
			game.RollPending = true
			game.TurnIdx = (game.TurnIdx + 1) % len(game.Players)
			SaveGameState(chatID, "ludo", game.Serialize())
			return fmt.Sprintf("You rolled a %d. No moves possible. Switching to %s.", roll, game.Players[game.TurnIdx].Name), nil
		}

		SaveGameState(chatID, "ludo", game.Serialize())
		return fmt.Sprintf("%s rolled a %d! Move a piece with !ak ludo move [1-4]", p.Name, roll), nil

	case "move":
		// Logic for moving pieces...
		return "Move logic implementation in progress. Use !ak ludo roll for now.", nil

	default:
		return "Unknown Ludo command.", nil
	}
}
