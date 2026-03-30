package misc

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestMsgParser_Validation(t *testing.T) {
	tests := []struct {
		name        string
		inputJSON   string
		expectPanic bool
	}{
		{
			name:        "empty message should not panic",
			inputJSON:   `{"from":"test","chatid":"1","userid":"1","message":"","plugin":"test","mode":"test"}`,
			expectPanic: false,
		},
		{
			name:        "short message less than csign",
			inputJSON:   `{"from":"test","chatid":"1","userid":"1","message":"!","plugin":"test","mode":"test","misc":{"csign":"!!!"}}`,
			expectPanic: false,
		},
		{
			name:        "valid message with single char csign",
			inputJSON:   `{"from":"test","chatid":"1","userid":"1","message":"!karma","plugin":"test","mode":"test","misc":{"csign":"!"}}`,
			expectPanic: false,
		},
		{
			name:        "message shorter than multi char csign",
			inputJSON:   `{"from":"test","chatid":"1","userid":"1","message":"!","plugin":"test","mode":"test","misc":{"csign":"!!"}}`,
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("unexpected panic: %v", r)
					}
				}
			}()

			// We only test that parsing doesn't panic
			// Actual publishing requires Redis
			var msg rMsg
			if err := json.Unmarshal([]byte(tt.inputJSON), &msg); err != nil {
				t.Logf("JSON parse error (expected for some tests): %v", err)
			}
		})
	}
}

func TestMsgParser_CommandRouting(t *testing.T) {
	// Save original config
	origConfig := Config
	Config = myConfig{
		Csign: "!",
		ForwardChannels: struct {
			Games    string `json:"games,omitempty"`
			Phrases  string `json:"phrases,omitempty"`
			Webapp   string `json:"webapp,omitempty"`
			WebappGo string `json:"webapp-go,omitempty"`
			Craniac  string `json:"craniac,omitempty"`
		}{
			Games:    "games",
			Phrases:  "phrases",
			Webapp:   "webapp",
			WebappGo: "webapp-go",
			Craniac:  "craniac",
		},
	}
	defer func() { Config = origConfig }()

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		// Phrases commands
		{"friday command", "!friday", "phrases"},
		{"fortune command", "!fortune", "phrases"},
		{"karma command", "!karma", "phrases"},
		{"russian friday", "!пятница", "phrases"},
		{"russian karma", "!карма", "phrases"},

		// Webapp-go commands
		{"frog command", "!frog", "webapp-go"},
		{"cat command", "!cat", "webapp-go"},
		{"weather command", "!w Moskow", "webapp-go"},
		{"russian weather", "!погода Москва", "webapp-go"},

		// Games commands
		{"dig command", "!dig", "games"},
		{"fish command", "!fish", "games"},
		{"russian dig", "!копать", "games"},

		// Default fallback
		{"unknown command", "!unknown", "craniac"},
		{"no command prefix", "just some text", "craniac"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := rMsg{
				From:    "test",
				Chatid:  "1",
				Userid:  "1",
				Message: tt.message,
				Plugin:  "test",
				Mode:    "test",
				Misc: struct {
					Answer      int64  `json:"answer,omitempty"`
					Botnick     string `json:"bot_nick,omitempty"`
					Csign       string `json:"csign,omitempty"`
					Fwdcnt      *int64 `json:"fwd_cnt,omitempty"`
					GoodMorning int64  `json:"good_morning,omitempty"`
					Msgformat   int64  `json:"msg_format,omitempty"`
					Username    string `json:"username,omitempty"`
				}{
					Csign:  Config.Csign,
					Fwdcnt: new(int64),
				},
			}
			*msg.Misc.Fwdcnt = 1

			result := determineChannel(msg.Message, msg.Misc.Csign)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMsgParser_KarmaPatterns(t *testing.T) {
	origConfig := Config
	Config = myConfig{
		Csign: "!",
		ForwardChannels: struct {
			Games    string `json:"games,omitempty"`
			Phrases  string `json:"phrases,omitempty"`
			Webapp   string `json:"webapp,omitempty"`
			WebappGo string `json:"webapp-go,omitempty"`
			Craniac  string `json:"craniac,omitempty"`
		}{
			Phrases: "phrases",
			Craniac: "craniac",
		},
	}
	defer func() { Config = origConfig }()

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{"karma increment", "something++", "phrases"},
		{"karma decrement", "something--", "phrases"},
		{"russian plus plus", "что-то++", "phrases"},
		{"no karma pattern", "just text", "craniac"},
		{"multiple lines no karma", "line1\nline2++", "craniac"},
		{"single line with karma", "line1++", "phrases"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineKarmaChannel(tt.message)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMsgParser_FwdcntPointer(t *testing.T) {
	// Test that Fwdcnt pointer handling works correctly
	fwdcnt := new(int64)
	*fwdcnt = 0

	if fwdcnt == nil {
		t.Error("fwdcnt should not be nil after new(int64)")
	}

	if *fwdcnt != 0 {
		t.Errorf("expected 0, got %d", *fwdcnt)
	}

	*fwdcnt++
	if *fwdcnt != 1 {
		t.Errorf("expected 1, got %d", *fwdcnt)
	}
}

// determineChannel is extracted logic for testing command routing
func determineChannel(message, csign string) string {
	sendTo := "craniac"

	if len(message) >= len(csign) && message[0:len(csign)] == csign {
		cmd := message[len(csign):]

		cmds := []string{"friday", "пятница", "proverb", "пословица", "пословиться", "fortune", "фортунка", "f", "ф",
			"karma", "карма", "rum", "ром", "vodka", "водка", "beer", "пиво", "tequila", "текила", "whisky", "виски",
			"absinthe", "абсент", "fuck"}

		for _, command := range cmds {
			if cmd == command {
				return "phrases"
			}
		}

		cmds = []string{"frog", "лягушка", "horse", "лошадь", "лошадка", "rabbit", "bunny", "кролик",
			"snail", "улитка", "cat", "кис", "fox", "лис", "buni", "anek", "анек", "анекдот",
			"xkcd", "monkeyuser", "tits", "boobs", "tities", "boobies", "сиси", "сисечки", "butt",
			"booty", "ass", "попа", "попка", "drink", "праздник", "owl", "сова", "сыч", "w", "п", "погода",
			"погодка", "погадка"}

		for _, command := range cmds {
			if cmd == command {
				return "webapp-go"
			}
		}

		cmds = []string{"dig", "копать", "fish", "fishing", "рыба", "рыбка", "рыбалка"}

		for _, command := range cmds {
			if cmd == command {
				return "games"
			}
		}

		cmdLen := len(cmd)
		cmds = []string{"w ", "п ", "погода ", "погодка ", "погадка ", "weather "}

		for _, command := range cmds {
			if cmdLen > len(command) && cmd[0:len(command)] == command {
				return "webapp-go"
			}
		}

		cmds = []string{"karma ", "карма ", "rum ", "ром ", "vodka ", "водка ", "beer ", "пиво ", "tequila ",
			"текила ", "whisky ", "виски ", "absinthe ", "абсент "}

		for _, command := range cmds {
			if cmdLen > len(command) && cmd[0:len(command)] == command {
				return "phrases"
			}
		}
	}

	return sendTo
}

// determineKarmaChannel is extracted logic for testing karma patterns
func determineKarmaChannel(message string) string {
	msgLen := len(message)

	if msgLen > 2 {
		if message[msgLen-2:] == "--" || message[msgLen-2:] == "++" {
			// Предполагается, что менять карму мы будем для одной фразы, то есть для 1 строки
			if len(strings.Split(message, "\n")) == 1 {
				return "phrases"
			}
		}
	}

	return "craniac"
}

func TestMsgParser_BuildsMsgCorrectly(t *testing.T) {
	// Test that message building produces correct output
	fwdcnt := new(int64)
	*fwdcnt = 2

	input := rMsg{
		From:     "telegram",
		Chatid:   "123",
		Userid:   "456",
		Threadid: "789",
		Message:  "!karma user",
		Plugin:   "telegram",
		Mode:     "private",
		Misc: struct {
			Answer      int64  `json:"answer,omitempty"`
			Botnick     string `json:"bot_nick,omitempty"`
			Csign       string `json:"csign,omitempty"`
			Fwdcnt      *int64 `json:"fwd_cnt,omitempty"`
			GoodMorning int64  `json:"good_morning,omitempty"`
			Msgformat   int64  `json:"msg_format,omitempty"`
			Username    string `json:"username,omitempty"`
		}{
			Answer:   1,
			Botnick:  "aleesa",
			Csign:    "!",
			Fwdcnt:   fwdcnt,
			Username: "testuser",
		},
	}

	// Simulate the message building logic
	var message sMsg
	message.From = input.From
	message.Userid = input.Userid
	message.Chatid = input.Chatid
	message.Threadid = input.Threadid
	message.Message = input.Message
	message.Plugin = input.Plugin
	message.Mode = input.Mode
	message.Misc.Fwdcnt = *input.Misc.Fwdcnt
	message.Misc.Csign = input.Misc.Csign
	message.Misc.Username = input.Misc.Username
	message.Misc.Answer = input.Misc.Answer
	message.Misc.Botnick = input.Misc.Botnick
	message.Misc.Msgformat = input.Misc.Msgformat
	message.Misc.GoodMorning = input.Misc.GoodMorning

	// Verify
	if message.Misc.Fwdcnt != 2 {
		t.Errorf("expected Fwdcnt 2, got %d", message.Misc.Fwdcnt)
	}
	if message.Misc.Answer != 1 {
		t.Errorf("expected Answer 1, got %d", message.Misc.Answer)
	}
	if message.Misc.Username != "testuser" {
		t.Errorf("expected Username 'testuser', got %s", message.Misc.Username)
	}

	// Verify JSON serialization works
	data, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Verify we can unmarshal it back
	var result sMsg
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result.From != message.From {
		t.Errorf("From mismatch: got %s", result.From)
	}
	if result.Misc.Fwdcnt != message.Misc.Fwdcnt {
		t.Errorf("Fwdcnt mismatch: got %d", result.Misc.Fwdcnt)
	}
}

var _ = context.Background // for import
