# Ayanokoji AI WhatsApp Bot (Go) v2.0

An advanced WhatsApp AI agent powered by **Groq (Llama 3.3 70B)** with the clinical, stoic personality of **Kiyotaka Ayanokoji**. This bot features full conversation memory, visual rendering for board games, an integrated economy, and advanced group management tools.

## 🚀 Features

### 🧠 Intelligence & Personality
- **Ayanokoji Persona:** Every response is calculated, logical, and detached.
- **Full Conversation Memory:** Remembers the entire chat history for contextual continuity.
- **Clinical Analysis (Roast):** Provides devastatingly calm behavioral evaluations and mockery of users.
- **Argument Mode:** Moderates debates by restricting the group to participants and delivering a logical verdict.

### 💰 Economy & Banking
- **Bank System:** Personal accounts with starting balances, PIN security, and loans.
- **Robbery & Hacking:** Guess PINs to rob users or defeat them in a Battle Card game to decrypt their security.
- **Protection:** Buy hourly protection to secure your resources from theft.

### 🎮 Elite Games
- **Battle Cards:** Visual card combat system with damage bars and elemental types.
- **Chess:** Integrated chess engine with visual board rendering (coordinates included), Elo ratings, and AI opponents.
- **Ludo:** Programmatic board rendering with a lobby system (2-4 players).
- **Gambling Modules:** Aviator (crash game), Spin (Red/Black), Poker, and Blackjack with betting.
- **Social Games:** Spin the Bottle, Tic-Tac-Toe, and Rock Paper Scissors.

### 🛠️ Media Tools
- **View-Once "Steal":** Extract and save view-once images/videos by replying with `!ak steal`.
- **Sticker Converter:** Convert any image to a WhatsApp sticker with `!ak s`.
- **Image Converter:** Convert stickers back to high-quality PNGs with `!ak img`.

---

## 🛠️ Commands

| Command | Category | Description |
| :--- | :--- | :--- |
| `!ak help` / `!ak menu` | General | Display the visual command menu |
| `!ak [text]` | AI | Chat with Ayanokoji AI |
| `!ak roast @user` | AI | Get a clinical diss of a mentioned user |
| `!ak argue @user` | Utility | Start a moderated debate mode |
| `!ak bank [register]` | Economy | Manage balance, loans, and protection |
| `!ak rob @user [pin]` | Economy | Attempt to steal from another user |
| `!ak aviator [bet]` | Gambling | Play the high-risk crash game |
| `!ak battle @user` | Game | Start a visual card combat match |
| `!ak hack @user` | Game | Battle to reveal a user's bank PIN |
| `!ak chess [move/register]` | Game | Play ranked chess with visual board |
| `!ak steal` | Media | (Reply) Extract view-once media |
| `!ak s` / `!ak img` | Media | (Reply) Image <-> Sticker conversion |

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
- `main.go`: Event handling, command routing, and visual menu logic.
- `ai/`: Groq API integration and Ayanokoji behavioral prompt logic.
- `games/`: Core logic for Chess, Battle Cards, Economy, Poker, and Ludo.
- `utils/`: Visual rendering engines for boards, coordinates, and media processing.
- `games.db`: SQLite storage for Elo ratings, bank accounts, and game states.

## ⚖️ Disclaimer
This bot is for educational and entertainment purposes. It uses fictional roleplay based on the character Kiyotaka Ayanokoji. Always adhere to WhatsApp's Terms of Service.
