package games

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type BankAccount struct {
	JID             string
	Balance         int
	BankLoan        int
	PIN             string
	ProtectionUntil time.Time
}

func RegisterBankAccount(jid string, pin string) error {
	_, err := db.Exec("INSERT INTO bank_accounts (jid, pin, balance) VALUES (?, ?, 500) ON CONFLICT(jid) DO NOTHING", jid, pin)
	return err
}

func GetBankAccount(jid string) (*BankAccount, error) {
	var a BankAccount
	var prot string
	err := db.QueryRow("SELECT jid, balance, bank_loan, pin, protection_until FROM bank_accounts WHERE jid = ?", jid).Scan(&a.JID, &a.Balance, &a.BankLoan, &a.PIN, &prot)
	if err != nil { return nil, err }
	a.ProtectionUntil, _ = time.Parse("2006-01-02 15:04:05", prot)
	return &a, nil
}

func UpdateBalance(jid string, amount int) error {
	_, err := db.Exec("UPDATE bank_accounts SET balance = balance + ? WHERE jid = ?", amount, jid)
	return err
}

func SetProtection(jid string, duration time.Duration) error {
	until := time.Now().Add(duration).Format("2006-01-02 15:04:05")
	_, err := db.Exec("UPDATE bank_accounts SET protection_until = ? WHERE jid = ?", until, jid)
	return err
}

func GetBankLoan(jid string, amount int) (string, error) {
	acc, err := GetBankAccount(jid)
	if err != nil { return "Account not found.", err }
	if acc.BankLoan > 0 { return "Outstanding loan exists.", nil }
	_, err = db.Exec("UPDATE bank_accounts SET balance = balance + ?, bank_loan = ? WHERE jid = ?", amount, amount, jid)
	return fmt.Sprintf("Loan of %d approved.", amount), err
}

func RepayBankLoan(jid string, amount int) (string, error) {
	acc, err := GetBankAccount(jid)
	if err != nil { return "Account not found.", err }
	repay := amount
	if repay > acc.BankLoan { repay = acc.BankLoan }
	_, err = db.Exec("UPDATE bank_accounts SET balance = balance - ?, bank_loan = bank_loan - ? WHERE jid = ?", repay, repay, jid)
	return fmt.Sprintf("Repaid %d.", repay), err
}

func HandleBank(jid string, args []string) string {
	if len(args) == 0 {
		acc, _ := GetBankAccount(jid)
		if acc == nil { return "Account not found. Use !ak bank register [4-digit-pin]." }
		return fmt.Sprintf("🏦 Bank Details:\nBalance: %d\nLoan: %d\nProtection: %s", acc.Balance, acc.BankLoan, acc.ProtectionUntil.Format("15:04:05"))
	}
	cmd := strings.ToLower(args[0])
	switch cmd {
	case "register":
		if len(args) < 2 { return "Usage: register [pin]" }
		RegisterBankAccount(jid, args[1])
		return "Bank account registered."
	case "loan":
		if len(args) < 2 { return "Usage: loan [amount]" }
		var amount int
		fmt.Sscanf(args[1], "%d", &amount)
		res, _ := GetBankLoan(jid, amount)
		return res
	case "repay":
		if len(args) < 2 { return "Usage: repay [amount]" }
		var amount int
		fmt.Sscanf(args[1], "%d", &amount)
		res, _ := RepayBankLoan(jid, amount)
		return res
	default:
		return "Unknown bank command."
	}
}

func RobUser(robberJID, victimJID, pinGuess string) (string, bool) {
	victim, err := GetBankAccount(victimJID)
	if err != nil { return "Account not found.", false }
	if time.Now().Before(victim.ProtectionUntil) { return "User is protected.", false }
	if victim.PIN != pinGuess {
		UpdateBalance(robberJID, -100)
		return "Wrong PIN. Fined 100.", false
	}
	amount := int(float64(victim.Balance) * (0.3 + rand.Float64()*0.4))
	UpdateBalance(victimJID, -amount)
	UpdateBalance(robberJID, amount)
	SetProtection(victimJID, 1*time.Hour)
	return fmt.Sprintf("Robbed %d.", amount), true
}
