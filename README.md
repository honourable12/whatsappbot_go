# Ayanokoji AI WhatsApp Bot (Go)

An advanced WhatsApp AI agent powered by **Groq (Moonshot Kimi K2)** with the clinical, stoic personality of **Kiyotaka Ayanokoji**. This bot features full conversation memory, media manipulation tools, and a suite of interactive multiplayer games.

## 🚀 Features

### 🧠 Intelligence & Personality
- **Ayanokoji Persona:** Every response is calculated, logical, and detached.
- **Full Conversation Memory:** Remembers the entire chat history for contextual continuity.
- **Clinical Analysis (Roast):** Provides devastatingly calm behavioral evaluations of users.
- **Group Integration:** Activated in groups via `!ak` or through direct private messages.

### 🛠️ Media Tools
- **View-Once "Steal":** Extract and save view-once images/videos by replying with `!ak steal`.
- **Sticker Converter:** Convert any image to a WhatsApp sticker with `!ak s`.
- **Image Converter:** Convert stickers back to high-quality PNGs with `!ak img`.

### 🎮 Games
- **Ludo:** Programmatic board image rendering with a lobby system (2-4 players).
- **Poker (Texas Hold'em):** Create tables, join, and receive your cards privately via PM.
- **Blackjack:** Multiplayer card game with persistent state.
- **Tic-Tac-Toe:** Play against the AI or challenge a friend by tagging them (`!ak ttt @user`).
- **Rock Paper Scissors:** Quick logic-based games.

---

## 🛠️ Commands

| Command | Description |
| :--- | :--- |
| `!ak help` | Display the command menu |
| `!ak ping` | Check bot connectivity |
| `!ak reset` | Wipe AI conversation history for the current chat |
| `!ak [text]` | Chat with Ayanokoji AI |
| `!ak roast @user` | Get a clinical analysis of a mentioned user |
| `!ak steal` | (Reply) Extract view-once media |
| `!ak s` | (Reply) Convert image to sticker |
| `!ak img` | (Reply) Convert sticker to image |
| `!ak ludo [make/join/start]` | Manage Ludo game |
| `!ak poker [make/join/start]` | Manage Poker/Blackjack tables |
| `!ak ttt [@user]` | Start Tic-Tac-Toe |
| `!ak ttt move [1-9]` | Make a move in TTT |

---

## ⚙️ Setup & Installation

### 1. Prerequisites
- [Go](https://golang.org/dl/) (1.20 or higher)
- A Groq API Key (Get one at [console.groq.com](https://console.groq.com/))

### 2. Environment Configuration
Create a `.env` file in the root directory:
```env
GROQ_API_KEY=your_groq_api_key_here
```

### 3. Install Dependencies
```bash
go mod tidy
```

### 4. Run the Bot
```bash
go run .
```
Scan the generated **QR Code** using your WhatsApp mobile app (Linked Devices > Link a Device).

---

## 📁 Project Structure
- `main.go`: Event handling and command routing.
- `ai/`: Groq API integration and Ayanokoji prompt logic.
- `games/`: Database-backed logic for Ludo, Poker, TTT, and more.
- `utils/`: Image processing and Ludo board rendering.
- `games.db`: SQLite storage for game states and chat history.
- `examplestore.db`: WhatsApp session storage.

## ⚖️ Disclaimer
This bot is for educational and entertainment purposes. It uses fictional roleplay based on the character Kiyotaka Ayanokoji. Always adhere to WhatsApp's Terms of Service.
