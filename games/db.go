package games

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB(path string) error {
	var err error
	db, err = sql.Open("sqlite3", path)
	if err != nil { return err }

	query := `
	CREATE TABLE IF NOT EXISTS game_states (chat_id TEXT PRIMARY KEY, game_type TEXT, state TEXT, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS arguments (chat_id TEXT PRIMARY KEY, p1_jid TEXT, p2_jid TEXT, start_time DATETIME DEFAULT CURRENT_TIMESTAMP, active BOOLEAN DEFAULT 1);
	CREATE TABLE IF NOT EXISTS chess_players (jid TEXT PRIMARY KEY, name TEXT, elo INTEGER DEFAULT 1200, wins INTEGER DEFAULT 0, losses INTEGER DEFAULT 0, draws INTEGER DEFAULT 0);
	CREATE TABLE IF NOT EXISTS chess_games (chat_id TEXT PRIMARY KEY, p_white_jid TEXT, p_black_jid TEXT, is_ai BOOLEAN DEFAULT 0, fen TEXT, active BOOLEAN DEFAULT 1, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS chat_history (id INTEGER PRIMARY KEY AUTOINCREMENT, chat_id TEXT, role TEXT, content TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP);`

	_, err = db.Exec(query)
	return err
}

func SaveGameState(chatID string, gameType string, state string) error {
	query := `INSERT INTO game_states (chat_id, game_type, state, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP) ON CONFLICT(chat_id) DO UPDATE SET state = excluded.state, updated_at = CURRENT_TIMESTAMP;`
	_, err := db.Exec(query, chatID, gameType, state)
	return err
}

func GetGameState(chatID string) (string, string, error) {
	var g, s string
	err := db.QueryRow("SELECT game_type, state FROM game_states WHERE chat_id = ?", chatID).Scan(&g, &s)
	return g, s, err
}

func DeleteGameState(chatID string) error {
	_, err := db.Exec("DELETE FROM game_states WHERE chat_id = ?", chatID)
	return err
}

type ChatHistoryEntry struct { Role, Content string }

func AddChatHistory(chatID, role, content string) error {
	_, err := db.Exec("INSERT INTO chat_history (chat_id, role, content) VALUES (?, ?, ?)", chatID, role, content)
	return err
}

func GetChatHistory(chatID string, limit int) ([]ChatHistoryEntry, error) {
	var rows *sql.Rows
	var err error
	if limit > 0 { rows, err = db.Query("SELECT role, content FROM (SELECT role, content, timestamp FROM chat_history WHERE chat_id = ? ORDER BY timestamp DESC LIMIT ?) ORDER BY timestamp ASC", chatID, limit) } else { rows, err = db.Query("SELECT role, content FROM chat_history WHERE chat_id = ? ORDER BY timestamp ASC", chatID) }
	if err != nil { return nil, err }
	defer rows.Close()
	var history []ChatHistoryEntry
	for rows.Next() {
		var e ChatHistoryEntry
		rows.Scan(&e.Role, &e.Content)
		history = append(history, e)
	}
	return history, nil
}

func ResetChatHistory(chatID string) error {
	_, err := db.Exec("DELETE FROM chat_history WHERE chat_id = ?", chatID)
	return err
}

func StartArgument(chatID, p1, p2 string) error {
	_, err := db.Exec("INSERT INTO arguments (chat_id, p1_jid, p2_jid, active) VALUES (?, ?, ?, 1) ON CONFLICT(chat_id) DO UPDATE SET p1_jid = excluded.p1_jid, p2_jid = excluded.p2_jid, active = 1, start_time = CURRENT_TIMESTAMP;", chatID, p1, p2)
	return err
}

func GetArgument(chatID string) (string, string, string, bool, error) {
	var p1, p2, t string; var a bool
	err := db.QueryRow("SELECT p1_jid, p2_jid, start_time, active FROM arguments WHERE chat_id = ?", chatID).Scan(&p1, &p2, &t, &a)
	return p1, p2, t, a, err
}

func EndArgument(chatID string) error {
	_, err := db.Exec("DELETE FROM arguments WHERE chat_id = ?", chatID)
	return err
}

type ChessPlayer struct { JID, Name string; Elo, Wins, Losses, Draws int }

func RegisterChessPlayer(jid, name string) error {
	_, err := db.Exec("INSERT INTO chess_players (jid, name) VALUES (?, ?) ON CONFLICT(jid) DO UPDATE SET name = excluded.name", jid, name)
	return err
}

func GetChessPlayer(jid string) (*ChessPlayer, error) {
	var p ChessPlayer
	err := db.QueryRow("SELECT jid, name, elo, wins, losses, draws FROM chess_players WHERE jid = ?", jid).Scan(&p.JID, &p.Name, &p.Elo, &p.Wins, &p.Losses, &p.Draws)
	return &p, err
}

func UpdateChessStats(jid string, w, l, d, elo int) error {
	_, err := db.Exec("UPDATE chess_players SET wins = wins + ?, losses = losses + ?, draws = draws + ?, elo = ? WHERE jid = ?", w, l, d, elo, jid)
	return err
}

func StartChessGame(chatID, pw, pb string, ai bool, fen string) error {
	_, err := db.Exec("INSERT INTO chess_games (chat_id, p_white_jid, p_black_jid, is_ai, fen, active) VALUES (?, ?, ?, ?, ?, 1) ON CONFLICT(chat_id) DO UPDATE SET p_white_jid = excluded.p_white_jid, p_black_jid = excluded.p_black_jid, is_ai = excluded.is_ai, fen = excluded.fen, active = 1, updated_at = CURRENT_TIMESTAMP;", chatID, pw, pb, ai, fen)
	return err
}

func GetChessGame(chatID string) (string, string, bool, string, bool, error) {
	var pw, pb, f string; var ai, a bool
	err := db.QueryRow("SELECT p_white_jid, p_black_jid, is_ai, fen, active FROM chess_games WHERE chat_id = ?", chatID).Scan(&pw, &pb, &ai, &f, &a)
	return pw, pb, ai, f, a, err
}

func UpdateChessGame(chatID, fen string) error {
	_, err := db.Exec("UPDATE chess_games SET fen = ?, updated_at = CURRENT_TIMESTAMP WHERE chat_id = ?", fen, chatID)
	return err
}

func DeleteChessGame(chatID string) error {
	_, err := db.Exec("DELETE FROM chess_games WHERE chat_id = ?", chatID)
	return err
}
