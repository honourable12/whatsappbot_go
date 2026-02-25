package games

import (
	"fmt"
	"math/rand"
	"strings"
)

type Stock struct {
	Symbol    string
	Name      string
	Price     int
	LastPrice int
	Volatility float64
}

func InitStocks() {
	stocks := []Stock{
		{"GPH", "GopherCorp", 100, 100, 0.05},
		{"RST", "Rustaceans Ltd", 150, 150, 0.08},
		{"PYT", "Pythonic AI", 80, 80, 0.12},
		{"JSF", "JS Frameworks Inc", 50, 50, 0.20},
		{"ZIG", "Zig Zag Systems", 120, 120, 0.06},
	}
	for _, s := range stocks {
		_, err := db.Exec("INSERT INTO stocks (symbol, name, price, last_price, volatility) VALUES (?, ?, ?, ?, ?) ON CONFLICT(symbol) DO NOTHING", s.Symbol, s.Name, s.Price, s.LastPrice, s.Volatility)
		if err != nil {
			fmt.Printf("Error initializing stock %s: %v\n", s.Symbol, err)
		}
	}
}

func UpdateStockPrices() {
	rows, err := db.Query("SELECT symbol, price, volatility FROM stocks")
	if err != nil { return }
	defer rows.Close()

	for rows.Next() {
		var symbol string
		var price int
		var volatility float64
		rows.Scan(&symbol, &price, &volatility)

		change := (rand.Float64()*2 - 1) * volatility
		newPrice := int(float64(price) * (1 + change))
		if newPrice < 1 { newPrice = 1 }

		_, err = db.Exec("UPDATE stocks SET last_price = price, price = ? WHERE symbol = ?", newPrice, symbol)
		if err != nil {
			fmt.Printf("Error updating stock price for %s: %v\n", symbol, err)
		}
	}
}

func GetStockMarket() string {
	rows, err := db.Query("SELECT symbol, name, price, last_price FROM stocks")
	if err != nil { return "Error fetching market data." }
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("📈 *Bot Stock Market*\n\n")
	for rows.Next() {
		var s, n string
		var p, lp int
		rows.Scan(&s, &n, &p, &lp)
		trend := "➡️"
		if p > lp { trend = "📈" } else if p < lp { trend = "📉" }
		sb.WriteString(fmt.Sprintf("%s *%s* (%s): %d %s\n", trend, n, s, p, trend))
	}
	sb.WriteString("\nCommands: `!ak stocks buy [symbol] [qty]`, `!ak stocks sell [symbol] [qty]`")
	return sb.String()
}

func BuyStock(jid string, symbol string, quantity int) string {
	var price int
	err := db.QueryRow("SELECT price FROM stocks WHERE symbol = ?", strings.ToUpper(symbol)).Scan(&price)
	if err != nil { return "Stock not found." }

	totalCost := price * quantity
	acc, err := GetBankAccount(jid)
	if err != nil || acc.Balance < totalCost { return "Insufficient funds." }

	_, err = db.Exec("UPDATE bank_accounts SET balance = balance - ? WHERE jid = ?", totalCost, jid)
	if err != nil { return "Transaction failed." }

	_, err = db.Exec("INSERT INTO user_stocks (jid, symbol, quantity) VALUES (?, ?, ?) ON CONFLICT(jid, symbol) DO UPDATE SET quantity = quantity + excluded.quantity", jid, strings.ToUpper(symbol), quantity)
	if err != nil { return "Stock update failed." }

	return fmt.Sprintf("Successfully bought %d %s for %d.", quantity, symbol, totalCost)
}

func SellStock(jid string, symbol string, quantity int) string {
	var heldQty int
	err := db.QueryRow("SELECT quantity FROM user_stocks WHERE jid = ? AND symbol = ?", jid, strings.ToUpper(symbol)).Scan(&heldQty)
	if err != nil || heldQty < quantity { return "You don't have enough shares." }

	var price int
	err = db.QueryRow("SELECT price FROM stocks WHERE symbol = ?", strings.ToUpper(symbol)).Scan(&price)
	if err != nil { return "Stock price unavailable." }

	totalGain := price * quantity
	_, err = db.Exec("UPDATE bank_accounts SET balance = balance + ? WHERE jid = ?", totalGain, jid)
	if err != nil { return "Transaction failed." }

	if heldQty == quantity {
		_, err = db.Exec("DELETE FROM user_stocks WHERE jid = ? AND symbol = ?", jid, strings.ToUpper(symbol))
	} else {
		_, err = db.Exec("UPDATE user_stocks SET quantity = quantity - ? WHERE jid = ? AND symbol = ?", quantity, jid, strings.ToUpper(symbol))
	}
	if err != nil { return "Stock update failed." }

	return fmt.Sprintf("Successfully sold %d %s for %d.", quantity, symbol, totalGain)
}

func GetUserStocks(jid string) string {
	rows, err := db.Query("SELECT us.symbol, s.name, us.quantity, s.price FROM user_stocks us JOIN stocks s ON us.symbol = s.symbol WHERE us.jid = ?", jid)
	if err != nil { return "Error fetching portfolio." }
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("💼 *Your Portfolio*\n\n")
	hasStocks := false
	totalValue := 0
	for rows.Next() {
		hasStocks = true
		var s, n string
		var q, p int
		rows.Scan(&s, &n, &q, &p)
		value := q * p
		totalValue += value
		sb.WriteString(fmt.Sprintf("*%s* (%s): %d shares (Value: %d)\n", n, s, q, value))
	}
	if !hasStocks { return "You don't own any stocks." }
	sb.WriteString(fmt.Sprintf("\n*Total Portfolio Value: %d*", totalValue))
	return sb.String()
}
