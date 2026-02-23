package games

import (
	"strings"
	"time"
)

type Argument struct {
	P1JID     string
	P2JID     string
	StartTime time.Time
}

func HandleArgument(chatID string, sender string, args []string) (string, []string, bool) {
	if len(args) == 0 {
		return "Usage: !ak argue @mention (to start) or !ak argue end/resign", nil, false
	}

	switch strings.ToLower(args[0]) {
	case "end", "resign":
		p1, p2, _, active, err := GetArgument(chatID)
		if err != nil || !active {
			return "No active argument in this chat.", nil, false
		}
		if sender != p1 && sender != p2 {
			return "Only participants can end the argument.", nil, false
		}
		
		// Signal to end the argument
		return "Argument ended. Generating verdict...", []string{p1, p2}, true
		
	default:
		// Check if it's a mention
		if strings.Contains(args[0], "@") {
			// This is handled in main.go to get the JID
			return "", nil, false
		}
		return "Unknown argue command.", nil, false
	}
}
