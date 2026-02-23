package games

import (
	"encoding/json"
	"fmt"
	"strings"
)

type TTTGame struct {
	Board   [3][3]string `json:"board"`
	Turn    string       `json:"turn"`
	PlayerX string       `json:"player_x"` 
	PlayerO string       `json:"player_o"` 
	Active  bool         `json:"active"`
}

func (g *TTTGame) Serialize() string {
	data, _ := json.Marshal(g)
	return string(data)
}

func DeserializeTTT(state string) *TTTGame {
	var g TTTGame
	json.Unmarshal([]byte(state), &g)
	return &g
}

func NewTTTGame(chatID string, challenger string, opponent string) string {
	game := &TTTGame{
		Board:   [3][3]string{{" ", " ", " "}, {" ", " ", " "}, {" ", " ", " "}},
		Turn:    "X",
		PlayerX: challenger,
		PlayerO: opponent,
		Active:  true,
	}
	SaveGameState(chatID, "ttt", game.Serialize())
	
	opponentName := "Ayanokoji (AI)"
	if opponent != "AI" {
		opponentName = "the tagged user"
	}
	
	return fmt.Sprintf("Tic-Tac-Toe started!\nPlayer X: You\nPlayer O: %s\nUse !ak ttt move [1-9] to play.\n%s", opponentName, game.Render())
}

func (g *TTTGame) Render() string {
	var sb strings.Builder
	sb.WriteString("```\n")
	for i, row := range g.Board {
		sb.WriteString(fmt.Sprintf(" %s | %s | %s \n", row[0], row[1], row[2]))
		if i < 2 {
			sb.WriteString("-----------\n")
		}
	}
	sb.WriteString("```")
	return sb.String()
}

func (g *TTTGame) MakeMove(sender string, pos int) (string, bool) {
	if !g.Active {
		return "No active game.", false
	}

	// Check if it's the sender's turn
	if g.Turn == "X" && sender != g.PlayerX {
		return "It's not your turn (Waiting for Player X).", false
	}
	if g.Turn == "O" && sender != g.PlayerO {
		return "It's not your turn (Waiting for Player O).", false
	}

	if pos < 1 || pos > 9 {
		return "Invalid position. Choose 1-9.", false
	}
	row, col := (pos-1)/3, (pos-1)%3
	if g.Board[row][col] != " " {
		return "Position already taken.", false
	}

	g.Board[row][col] = g.Turn
	if g.checkWin() {
		g.Active = false
		winner := "Player " + g.Turn
		if g.Turn == "O" && g.PlayerO == "AI" {
			winner = "Ayanokoji (AI)"
		}
		return fmt.Sprintf("%s wins!\n%s", winner, g.Render()), true
	}
	if g.checkDraw() {
		g.Active = false
		return fmt.Sprintf("It's a draw!\n%s", g.Render()), true
	}

	// Switch turn
	if g.Turn == "X" {
		g.Turn = "O"
	} else {
		g.Turn = "X"
	}

	// If it's now AI's turn, make a move
	if g.Active && g.Turn == "O" && g.PlayerO == "AI" {
		g.aiMove()
		if g.checkWin() {
			g.Active = false
			return fmt.Sprintf("Ayanokoji (AI) wins!\n%s", g.Render()), true
		}
		if g.checkDraw() {
			g.Active = false
			return "It's a draw!\n" + g.Render(), true
		}
		g.Turn = "X"
	}

	return g.Render(), true
}

func (g *TTTGame) aiMove() {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if g.Board[i][j] == " " {
				g.Board[i][j] = "O"
				return
			}
		}
	}
}

func (g *TTTGame) checkWin() bool {
	b := g.Board
	for i := 0; i < 3; i++ {
		// Rows
		if b[i][0] != " " && b[i][0] == b[i][1] && b[i][1] == b[i][2] {
			return true
		}
		// Columns
		if b[0][i] != " " && b[0][i] == b[1][i] && b[1][i] == b[2][i] {
			return true
		}
	}
	// Diagonals
	if b[0][0] != " " && b[0][0] == b[1][1] && b[1][1] == b[2][2] {
		return true
	}
	if b[0][2] != " " && b[0][2] == b[1][1] && b[1][1] == b[2][0] {
		return true
	}
	return false
}

func (g *TTTGame) checkDraw() bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if g.Board[i][j] == " " {
				return false
			}
		}
	}
	return true
}

func HandleTTT(chatID string, senderID string, opponentID string, args []string) string {
	if len(args) == 0 || (len(args) == 1 && strings.Contains(args[0], "@")) {
		return NewTTTGame(chatID, senderID, opponentID)
	}

	gameType, state, err := GetGameState(chatID)
	if err != nil || gameType != "ttt" {
		return "No active game in this chat. Start one with !ak ttt @mention or !ak ttt"
	}

	game := DeserializeTTT(state)
	if !game.Active {
		return "No active game. Start one with !ak ttt"
	}

	switch strings.ToLower(args[0]) {
	case "move":
		if len(args) < 2 {
			return "Specify a position 1-9."
		}
		var pos int
		fmt.Sscanf(args[1], "%d", &pos)
		res, ok := game.MakeMove(senderID, pos)
		if ok {
			if !game.Active {
				DeleteGameState(chatID)
			} else {
				SaveGameState(chatID, "ttt", game.Serialize())
			}
		}
		return res
	case "cancel":
		if senderID == game.PlayerX || senderID == game.PlayerO {
			DeleteGameState(chatID)
			return "Game cancelled."
		}
		return "Only players can cancel the game."
	default:
		return "Unknown command. Use !ak ttt move [1-9] or !ak ttt cancel"
	}
}
