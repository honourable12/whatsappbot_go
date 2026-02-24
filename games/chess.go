package games

import (
	"fmt"
	"math"
	"strings"

	"github.com/notnil/chess"
)

func HandleChess(chatID string, sender string, name string, args []string, opponentID string) (string, string) {
	if len(args) == 0 {
		return "Usage: !ak chess register, !ak chess challenge @user, !ak chess challenge ai, !ak chess move [e2e4], !ak chess resign", ""
	}

	cmd := strings.ToLower(args[0])

	switch cmd {
	case "register":
		err := RegisterChessPlayer(sender, name)
		if err != nil {
			return "Registration failed.", ""
		}
		return "You are now registered as a chess player.", ""

	case "challenge":
		p1, err := GetChessPlayer(sender)
		if err != nil {
			return "Register first with !ak chess register.", ""
		}

		isAI := false
		p2JID := opponentID
		if len(args) > 1 && strings.ToLower(args[1]) == "ai" {
			isAI = true
			p2JID = "AI"
		} else if p2JID == "" {
			return "Tag someone to challenge or use 'ai'.", ""
		}

		var p2Name string
		if isAI {
			p2Name = "Ayanokoji (AI)"
		} else {
			p2, err := GetChessPlayer(p2JID)
			if err != nil {
				return "Your opponent is not registered. They must use !ak chess register.", ""
			}
			p2Name = p2.Name
		}

		game := chess.NewGame()
		StartChessGame(chatID, sender, p2JID, isAI, game.FEN())
		return fmt.Sprintf("Chess Challenge Started!\nWhite: %s\nBlack: %s\nUse !ak chess move [from][to] (e.g., e2e4)", p1.Name, p2Name), game.FEN()

	case "move":
		if len(args) < 2 {
			return "Specify your move (e.g., e2e4).", ""
		}
		moveStr := strings.ToLower(args[1])
		pWhite, pBlack, isAI, fen, active, err := GetChessGame(chatID)
		if err != nil || !active {
			return "No active chess game in this chat.", ""
		}

		gameFunc, err := chess.FEN(fen)
		if err != nil {
			return "Failed to load game state.", ""
		}
		game := chess.NewGame(gameFunc)

		// Check whose turn it is
		turn := game.Position().Turn()
		if turn == chess.White && sender != pWhite {
			return "It's not your turn. Waiting for White.", fen
		}
		if turn == chess.Black && sender != pBlack {
			return "It's not your turn. Waiting for Black.", fen
		}

		// Try making move with multiple notations
		var move *chess.Move
		notations := []chess.Notation{chess.AlgebraicNotation{}, chess.UCINotation{}, chess.LongAlgebraicNotation{}}
		for _, n := range notations {
			if m, err := n.Decode(game.Position(), moveStr); err == nil {
				move = m
				break
			}
		}

		if move == nil {
			return "Invalid move. Use algebraic (e4) or coordinate (e2e4) notation.", fen
		}

		err = game.Move(move)
		if err != nil {
			return "Illegal move.", fen
		}

		UpdateChessGame(chatID, game.FEN())

		status := game.Method()
		if status != chess.NoMethod {
			return EndChessGame(chatID, game, pWhite, pBlack, isAI), game.FEN()
		}

		// If AI turn
		if isAI && game.Position().Turn() == chess.Black {
			aiMove := GetBestMove(game)
			game.Move(aiMove)
			UpdateChessGame(chatID, game.FEN())
			
			if game.Method() != chess.NoMethod {
				return EndChessGame(chatID, game, pWhite, pBlack, isAI), game.FEN()
			}
			return fmt.Sprintf("AI moved: %s\nYour turn.", aiMove.String()), game.FEN()
		}

		return "Move successful.", game.FEN()

	case "resign":
		pWhite, pBlack, isAI, fen, active, err := GetChessGame(chatID)
		if err != nil || !active {
			return "No active chess game.", ""
		}

		gameFunc, _ := chess.FEN(fen)
		game := chess.NewGame(gameFunc)
		
		winnerJID := pBlack
		if sender == pBlack {
			winnerJID = pWhite
		}
		
		return ForceEndChessGame(chatID, game, winnerJID, pWhite, pBlack, isAI, "resignation"), ""

	default:
		return "Unknown chess command.", ""
	}
}

func GetBestMove(game *chess.Game) *chess.Move {
	moves := game.ValidMoves()
	bestMove := moves[0]
	bestScore := -1000

	for _, m := range moves {
		score := 0
		if m.HasTag(chess.Check) {
			score += 5
		}
		
		// Simple piece value capture check
		p := game.Position().Board().Piece(m.S2())
		if p != chess.NoPiece {
			score += int(pieceValue(p.Type()))
		}
		
		if score > bestScore {
			bestScore = score
			bestMove = m
		}
	}
	return bestMove
}

func pieceValue(p chess.PieceType) int {
	switch p {
	case chess.Pawn: return 10
	case chess.Knight, chess.Bishop: return 30
	case chess.Rook: return 50
	case chess.Queen: return 90
	case chess.King: return 900
	default: return 0
	}
}

func EndChessGame(chatID string, game *chess.Game, pWhite, pBlack string, isAI bool) string {
	method := game.Method()
	outcome := game.Outcome()
	
	msg := fmt.Sprintf("Game Over! Method: %s\n", method.String())
	
	winnerJID := ""
	if outcome == chess.WhiteWon {
		winnerJID = pWhite
		msg += "White Wins!"
	} else if outcome == chess.BlackWon {
		winnerJID = pBlack
		msg += "Black Wins!"
	} else {
		msg += "It's a Draw!"
	}

	HandleEloUpdate(pWhite, pBlack, winnerJID, outcome == chess.Draw, isAI)
	DeleteChessGame(chatID)

	analysis := "\n\nAyanokoji Analysis: Everything was calculated from the start. "
	if winnerJID == pWhite && isAI {
		analysis += "Your tactics were surprisingly coherent."
	} else if isAI {
		analysis += "Human behavior is predictable."
	}

	return msg + analysis
}

func ForceEndChessGame(chatID string, game *chess.Game, winnerJID, pWhite, pBlack string, isAI bool, reason string) string {
	msg := fmt.Sprintf("Game Over by %s!\n", reason)
	if winnerJID != "" {
		msg += fmt.Sprintf("Winner: @%s", strings.Split(winnerJID, "@")[0])
	} else {
		msg += "It's a Draw!"
	}

	HandleEloUpdate(pWhite, pBlack, winnerJID, winnerJID == "", isAI)
	DeleteChessGame(chatID)
	return msg
}

func HandleEloUpdate(pWhite, pBlack, winnerJID string, isDraw bool, isAI bool) {
	pw, err1 := GetChessPlayer(pWhite)
	var pb *ChessPlayer
	var err2 error
	if isAI {
		pb = &ChessPlayer{JID: "AI", Elo: 1500}
	} else {
		pb, err2 = GetChessPlayer(pBlack)
	}

	if err1 != nil || err2 != nil {
		return
	}

	k := 32.0
	ea := 1.0 / (1.0 + math.Pow(10, float64(pb.Elo-pw.Elo)/400.0))
	eb := 1.0 / (1.0 + math.Pow(10, float64(pw.Elo-pb.Elo)/400.0))

	sa := 0.5
	sb := 0.5
	if !isDraw {
		if winnerJID == pWhite {
			sa = 1.0
			sb = 0.0
		} else {
			sa = 0.0
			sb = 1.0
		}
	}

	newEloW := pw.Elo + int(k*(sa-ea))
	newEloB := pb.Elo + int(k*(sb-eb))

	w, l, d := 0, 0, 0
	if isDraw { d = 1 } else if winnerJID == pWhite { w = 1 } else { l = 1 }
	UpdateChessStats(pWhite, w, l, d, newEloW)

	if !isAI {
		w, l, d = 0, 0, 0
		if isDraw { d = 1 } else if winnerJID == pBlack { w = 1 } else { l = 1 }
		UpdateChessStats(pBlack, w, l, d, newEloB)
	}
}
