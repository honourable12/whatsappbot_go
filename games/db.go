package games

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB(path string) error {
	var err error
	db, err = sql.Open("sqlite3", path)
	if err != nil {
		return err
	}

	query := `
	CREATE TABLE IF NOT EXISTS game_states (
		chat_id TEXT PRIMARY KEY,
		game_type TEXT,
		state TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS chat_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id TEXT,
		role TEXT,
		content TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(query)
	return err
}

func SaveGameState(chatID string, gameType string, state string) error {
	query := `INSERT INTO game_states (chat_id, game_type, state, updated_at) 
			  VALUES (?, ?, ?, CURRENT_TIMESTAMP)
			  ON CONFLICT(chat_id) DO UPDATE SET 
			  state = excluded.state,
			  updated_at = CURRENT_TIMESTAMP;`
	_, err := db.Exec(query, chatID, gameType, state)
	return err
}

func GetGameState(chatID string) (string, string, error) {
	var gameType, state string
	query := "SELECT game_type, state FROM game_states WHERE chat_id = ?"
	err := db.QueryRow(query, chatID).Scan(&gameType, &state)
	if err != nil {
		return "", "", err
	}
	return gameType, state, nil
}

func DeleteGameState(chatID string) error {
	query := "DELETE FROM game_states WHERE chat_id = ?"
	_, err := db.Exec(query, chatID)
	return err
}

type ChatHistoryEntry struct {
	Role    string
	Content string
}

func AddChatHistory(chatID string, role string, content string) error {
	query := "INSERT INTO chat_history (chat_id, role, content) VALUES (?, ?, ?)"
	_, err := db.Exec(query, chatID, role, content)
	return err
}

func GetChatHistory(chatID string, limit int) ([]ChatHistoryEntry, error) {
	// If limit is 0, get all history
	var query string
	var rows *sql.Rows
	var err error

	if limit > 0 {
		query = "SELECT role, content FROM (SELECT role, content, timestamp FROM chat_history WHERE chat_id = ? ORDER BY timestamp DESC LIMIT ?) ORDER BY timestamp ASC"
		rows, err = db.Query(query, chatID, limit)
	} else {
		query = "SELECT role, content FROM chat_history WHERE chat_id = ? ORDER BY timestamp ASC"
		rows, err = db.Query(query, chatID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []ChatHistoryEntry
	for rows.Next() {
		var entry ChatHistoryEntry
		if err := rows.Scan(&entry.Role, &entry.Content); err != nil {
			return nil, err
		}
		history = append(history, entry)
	}
	return history, nil
}

func ResetChatHistory(chatID string) error {
	query := "DELETE FROM chat_history WHERE chat_id = ?"
	_, err := db.Exec(query, chatID)
	return err
}
