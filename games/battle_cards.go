package games

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
)

type BattleCard struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Attack  int    `json:"attack"`
	Defense int    `json:"defense"`
	Color   string `json:"color"` // "Red", "Blue", "Green", "Yellow"
}

type BattlePlayer struct {
	JID    string       `json:"jid"`
	Name   string       `json:"name"`
	HP     int          `json:"hp"`
	Hand   []BattleCard `json:"hand"`
	Board  []BattleCard `json:"board"`
	Energy int          `json:"energy"`
}

type BattleGame struct {
	Player1 BattlePlayer `json:"player1"`
	Player2 BattlePlayer `json:"player2"`
	Turn    int          `json:"turn"` // 1 or 2
	Active  bool         `json:"active"`
}

var CardList = []BattleCard{
	{1, "Flame Warrior", 25, 10, "Red"},
	{2, "Water Mage", 15, 20, "Blue"},
	{3, "Earth Golem", 10, 35, "Yellow"},
	{4, "Forest Archer", 20, 15, "Green"},
	{5, "Fire Dragon", 40, 5, "Red"},
	{6, "Ice Titan", 5, 45, "Blue"},
	{7, "Sun Priest", 10, 10, "Yellow"},
	{8, "Storm Hawk", 30, 10, "Green"},
}

func (g *BattleGame) Serialize() string {
	data, _ := json.Marshal(g)
	return string(data)
}

func DeserializeBattle(state string) *BattleGame {
	var g BattleGame
	json.Unmarshal([]byte(state), &g)
	return &g
}

func NewBattleGame(chatID string, p1JID string, p1Name string, p2JID string, p2Name string) *BattleGame {
	g := &BattleGame{
		Player1: BattlePlayer{JID: p1JID, Name: p1Name, HP: 100, Hand: []BattleCard{}, Board: []BattleCard{}, Energy: 1},
		Player2: BattlePlayer{JID: p2JID, Name: p2Name, HP: 100, Hand: []BattleCard{}, Board: []BattleCard{}, Energy: 1},
		Turn:    1,
		Active:  true,
	}

	// Initial Draw
	for i := 0; i < 3; i++ {
		g.Player1.Hand = append(g.Player1.Hand, CardList[rand.Intn(len(CardList))])
		g.Player2.Hand = append(g.Player2.Hand, CardList[rand.Intn(len(CardList))])
	}

	return g
}

func (g *BattleGame) PlayCard(sender string, handIdx int) (string, bool) {
	if !g.Active {
		return "No active game.", false
	}

	var p *BattlePlayer
	if g.Turn == 1 {
		p = &g.Player1
	} else {
		p = &g.Player2
	}

	if p.JID != sender {
		return "It's not your turn.", false
	}

	if p.Energy <= 0 {
		return "No energy left this turn.", false
	}

	if handIdx < 0 || handIdx >= len(p.Hand) {
		return "Invalid card index.", false
	}

	card := p.Hand[handIdx]
	p.Board = append(p.Board, card)
	p.Hand = append(p.Hand[:handIdx], p.Hand[handIdx+1:]...)
	p.Energy--

	return fmt.Sprintf("%s played %s!", p.Name, card.Name), true
}

func (g *BattleGame) Attack(sender string, boardIdx int, targetType string, targetIdx int) (string, bool) {
	if !g.Active {
		return "No active game.", false
	}

	var attacker, defender *BattlePlayer
	if g.Turn == 1 {
		attacker, defender = &g.Player1, &g.Player2
	} else {
		attacker, defender = &g.Player2, &g.Player1
	}

	if attacker.JID != sender {
		return "It's not your turn.", false
	}

	if boardIdx < 0 || boardIdx >= len(attacker.Board) {
		return "Invalid board card index.", false
	}

	card := &attacker.Board[boardIdx]
	
	msg := ""
	if targetType == "p" {
		defender.HP -= card.Attack
		msg = fmt.Sprintf("%s attacked %s directly for %d damage!", card.Name, defender.Name, card.Attack)
	} else if targetType == "c" {
		if targetIdx < 0 || targetIdx >= len(defender.Board) {
			return "Invalid target card index.", false
		}
		targetCard := &defender.Board[targetIdx]
		targetCard.Defense -= card.Attack
		msg = fmt.Sprintf("%s attacked %s's %s for %d damage!", card.Name, defender.Name, targetCard.Name, card.Attack)
		if targetCard.Defense <= 0 {
			msg += fmt.Sprintf("\n%s's %s was destroyed!", defender.Name, targetCard.Name)
			defender.Board = append(defender.Board[:targetIdx], defender.Board[targetIdx+1:]...)
		}
	} else {
		return "Invalid target type. Use 'p' for player or 'c' for card.", false
	}

	if defender.HP <= 0 {
		defender.HP = 0
		g.Active = false
		msg += fmt.Sprintf("\n%s is defeated! %s wins!", defender.Name, attacker.Name)
	}

	return msg, true
}

func (g *BattleGame) EndTurn(sender string) (string, bool) {
	if !g.Active {
		return "No active game.", false
	}

	var p *BattlePlayer
	if g.Turn == 1 {
		p = &g.Player1
	} else {
		p = &g.Player2
	}

	if p.JID != sender {
		return "It's not your turn.", false
	}

	// Switch turn
	if g.Turn == 1 {
		g.Turn = 2
	} else {
		g.Turn = 1
	}

	// New turn setup
	var nextP *BattlePlayer
	if g.Turn == 1 {
		nextP = &g.Player1
	} else {
		nextP = &g.Player2
	}

	nextP.Energy = 1
	if len(nextP.Hand) < 5 {
		card := CardList[rand.Intn(len(CardList))]
		nextP.Hand = append(nextP.Hand, card)
	}

	return fmt.Sprintf("Turn ended. Now it's %s's turn.", nextP.Name), true
}

func (g *BattleGame) RenderText() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s HP: %d\nHand: ", g.Player1.Name, g.Player1.HP))
	for i, c := range g.Player1.Hand {
		sb.WriteString(fmt.Sprintf("[%d:%s] ", i+1, c.Name))
	}
	sb.WriteString("\nBoard: ")
	for i, c := range g.Player1.Board {
		sb.WriteString(fmt.Sprintf("[%d:%s ATK:%d DEF:%d] ", i+1, c.Name, c.Attack, c.Defense))
	}
	sb.WriteString("\n-------------------\n")
	sb.WriteString(fmt.Sprintf("%s HP: %d\nHand: ", g.Player2.Name, g.Player2.HP))
	for i, c := range g.Player2.Hand {
		sb.WriteString(fmt.Sprintf("[%d:%s] ", i+1, c.Name))
	}
	sb.WriteString("\nBoard: ")
	for i, c := range g.Player2.Board {
		sb.WriteString(fmt.Sprintf("[%d:%s ATK:%d DEF:%d] ", i+1, c.Name, c.Attack, c.Defense))
	}
	sb.WriteString("\n-------------------\n")
	
	return sb.String()
}

func HandleBattle(chatID string, senderID string, senderName string, opponentID string, opponentName string, args []string) (string, *BattleGame) {
	if len(args) == 0 || (len(args) == 1 && strings.Contains(args[0], "@")) {
		g := NewBattleGame(chatID, senderID, senderName, opponentID, opponentName)
		SaveGameState(chatID, "battle", g.Serialize())
		return fmt.Sprintf("Battle Card Game started!\nPlayer 1: %s\nPlayer 2: %s\nUse !ak battle play [idx], !ak battle attack [b_idx] [p/c] [t_idx], !ak battle end", senderName, opponentName), g
	}

	gameType, state, err := GetGameState(chatID)
	if err != nil || gameType != "battle" {
		return "No active battle in this chat. Start one with !ak battle @mention", nil
	}

	g := DeserializeBattle(state)
	
	msg := ""
	ok := false
	
	switch strings.ToLower(args[0]) {
	case "play":
		if len(args) < 2 { return "Specify card index in hand.", g }
		var idx int
		fmt.Sscanf(args[1], "%d", &idx)
		msg, ok = g.PlayCard(senderID, idx-1)
	case "attack":
		if len(args) < 4 { return "Usage: attack [board_idx] [p/c] [target_idx]", g }
		var bIdx, tIdx int
		fmt.Sscanf(args[1], "%d", &bIdx)
		targetType := strings.ToLower(args[2])
		fmt.Sscanf(args[3], "%d", &tIdx)
		msg, ok = g.Attack(senderID, bIdx-1, targetType, tIdx-1)
	case "end":
		msg, ok = g.EndTurn(senderID)
	case "cancel":
		DeleteGameState(chatID)
		return "Game cancelled.", nil
	default:
		return "Unknown command. play, attack, end, cancel.", g
	}

	if ok {
		if !g.Active {
			DeleteGameState(chatID)
		} else {
			SaveGameState(chatID, "battle", g.Serialize())
		}
	}
	
	return msg, g
}
