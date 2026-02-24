package games

import (
	"fmt"
	"math/rand"
	"strings"
)

func PlayAviator(jid string, bet int) (string, int) {
	acc, _ := GetBankAccount(jid)
	if acc == nil || acc.Balance < bet { return "Insufficient funds.", 0 }

	multiplier := 1.0 + (rand.Float64() * 3.0)
	msg := fmt.Sprintf("✈️ Aviator: The plane crashed at %.2fx!", multiplier)
	
	if multiplier > 1.5 {
		win := int(float64(bet) * 1.5)
		UpdateBalance(jid, win - bet)
		return msg + fmt.Sprintf("\nYou cashed out at 1.5x. Won %d.", win), win - bet
	} else {
		UpdateBalance(jid, -bet)
		return msg + "\nYou crashed.", -bet
	}
}

func PlaySpin(jid string, bet int, choice string) (string, int) {
	acc, _ := GetBankAccount(jid)
	if acc == nil || acc.Balance < bet { return "Insufficient funds.", 0 }
	res := rand.Intn(2); resStr := "red"; if res == 1 { resStr = "black" }
	msg := fmt.Sprintf("🎰 It's %s!", strings.ToUpper(resStr))
	if strings.ToLower(choice) == resStr {
		UpdateBalance(jid, bet); return msg + fmt.Sprintf("\nWon %d.", bet), bet
	}
	UpdateBalance(jid, -bet); return msg + "\nLost.", -bet
}

func SpinBottle(mentions []string) string {
	if len(mentions) < 2 { return "Need at least 2 mentions." }
	p1 := mentions[rand.Intn(len(mentions))]; p2 := mentions[rand.Intn(len(mentions))]
	for p1 == p2 { p2 = mentions[rand.Intn(len(mentions))] }
	return fmt.Sprintf("🍾 @%s and @%s must perform a dare!", strings.Split(p1, "@")[0], strings.Split(p2, "@")[0])
}

func HandleGambling(jid string, args []string) (string, int) {
	if len(args) < 2 { return "Usage: !ak [aviator/spin/red/black] [bet] [choice-if-any]", 0 }
	cmd := strings.ToLower(args[0])
	var bet int; fmt.Sscanf(args[1], "%d", &bet)
	switch cmd {
	case "aviator": return PlayAviator(jid, bet)
	case "spin", "red", "black":
		choice := cmd
		if cmd == "spin" { if len(args) < 3 { return "Specify red or black.", 0 }; choice = args[2] }
		return PlaySpin(jid, bet, choice)
	default: return "Unknown game.", 0
	}
}
