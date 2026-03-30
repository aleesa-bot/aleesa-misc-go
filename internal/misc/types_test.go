package misc

import (
	"encoding/json"
	"testing"
)

func TestRMsgJSONDeserialization(t *testing.T) {
	tests := []struct {
		name           string
		jsonInput      string
		expectedFrom   string
		expectedChatid string
		expectedPlugin string
		expectCsignNil bool
	}{
		{
			name:           "full message with all fields",
			jsonInput:      `{"from":"telegram","chatid":"123","userid":"456","message":"hello","plugin":"telegram","mode":"private","misc":{"answer":1,"bot_nick":"aleesa","csign":"!","fwd_cnt":1,"good_morning":0,"msg_format":1,"username":"user"}}`,
			expectedFrom:   "telegram",
			expectedChatid: "123",
			expectedPlugin: "telegram",
			expectCsignNil: false,
		},
		{
			name:           "minimal message",
			jsonInput:      `{"from":"irc","chatid":"#room","userid":"nick","message":"!karma","plugin":"irc","mode":"channel"}`,
			expectedFrom:   "irc",
			expectedChatid: "#room",
			expectedPlugin: "irc",
			expectCsignNil: true,
		},
		{
			name:           "message with fwd_cnt zero",
			jsonInput:      `{"from":"jabber","chatid":"room@conference","userid":"user","message":"test","plugin":"jabber","mode":"group","misc":{"fwd_cnt":0}}`,
			expectedFrom:   "jabber",
			expectedChatid: "room@conference",
			expectedPlugin: "jabber",
			expectCsignNil: true,
		},
		{
			name:           "message with threadid",
			jsonInput:      `{"from":"telegram","chatid":"123","userid":"456","threadid":"789","message":"reply","plugin":"telegram","mode":"thread"}`,
			expectedFrom:   "telegram",
			expectedChatid: "123",
			expectedPlugin: "telegram",
			expectCsignNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg rMsg
			if err := json.Unmarshal([]byte(tt.jsonInput), &msg); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if msg.From != tt.expectedFrom {
				t.Errorf("From: expected %s, got %s", tt.expectedFrom, msg.From)
			}
			if msg.Chatid != tt.expectedChatid {
				t.Errorf("Chatid: expected %s, got %s", tt.expectedChatid, msg.Chatid)
			}
			if msg.Plugin != tt.expectedPlugin {
				t.Errorf("Plugin: expected %s, got %s", tt.expectedPlugin, msg.Plugin)
			}
			if tt.expectCsignNil && msg.Misc.Csign != "" {
				t.Errorf("Csign: expected empty, got %s", msg.Misc.Csign)
			}
		})
	}
}

func TestSMsgJSONSerialization(t *testing.T) {
	tests := []struct {
		name    string
		message sMsg
	}{
		{
			name: "full message",
			message: sMsg{
				From:     "telegram",
				Chatid:   "123",
				Userid:   "456",
				Threadid: "789",
				Message:  "Hello world",
				Plugin:   "telegram",
				Mode:     "private",
				Misc: struct {
					Answer      int64  `json:"answer"`
					Botnick     string `json:"bot_nick"`
					Csign       string `json:"csign"`
					Fwdcnt      int64  `json:"fwd_cnt"`
					GoodMorning int64  `json:"good_morning"`
					Msgformat   int64  `json:"msg_format"`
					Username    string `json:"username"`
				}{
					Answer:      1,
					Botnick:     "aleesa",
					Csign:       "!",
					Fwdcnt:      2,
					GoodMorning: 0,
					Msgformat:   1,
					Username:    "testuser",
				},
			},
		},
		{
			name: "minimal message",
			message: sMsg{
				From:    "irc",
				Chatid:  "#room",
				Userid:  "nick",
				Message: "!help",
				Plugin:  "irc",
				Mode:    "channel",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			// Verify it's valid JSON
			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("marshaled data is not valid JSON: %v", err)
			}

			// Verify key fields are present
			if _, ok := result["from"]; !ok {
				t.Error("missing 'from' field")
			}
			if _, ok := result["chatid"]; !ok {
				t.Error("missing 'chatid' field")
			}
			if _, ok := result["message"]; !ok {
				t.Error("missing 'message' field")
			}
		})
	}
}

func TestRMsgToSMsgConversion(t *testing.T) {
	fwdcnt := new(int64)
	*fwdcnt = 3

	input := rMsg{
		From:     "telegram",
		Chatid:   "123",
		Userid:   "456",
		Threadid: "789",
		Message:  "!friday",
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
			Answer:      0,
			Botnick:     "aleesa",
			Csign:       "!",
			Fwdcnt:      fwdcnt,
			GoodMorning: 1,
			Msgformat:   1,
			Username:    "testuser",
		},
	}

	// Convert
	output := convertToSMsg(input)

	// Verify all fields transferred
	if output.From != input.From {
		t.Errorf("From: expected %s, got %s", input.From, output.From)
	}
	if output.Chatid != input.Chatid {
		t.Errorf("Chatid: expected %s, got %s", input.Chatid, output.Chatid)
	}
	if output.Userid != input.Userid {
		t.Errorf("Userid: expected %s, got %s", input.Userid, output.Userid)
	}
	if output.Threadid != input.Threadid {
		t.Errorf("Threadid: expected %s, got %s", input.Threadid, output.Threadid)
	}
	if output.Message != input.Message {
		t.Errorf("Message: expected %s, got %s", input.Message, output.Message)
	}
	if output.Plugin != input.Plugin {
		t.Errorf("Plugin: expected %s, got %s", input.Plugin, output.Plugin)
	}
	if output.Mode != input.Mode {
		t.Errorf("Mode: expected %s, got %s", input.Mode, output.Mode)
	}
	if output.Misc.Csign != input.Misc.Csign {
		t.Errorf("Misc.Csign: expected %s, got %s", input.Misc.Csign, output.Misc.Csign)
	}
	if output.Misc.Fwdcnt != *input.Misc.Fwdcnt {
		t.Errorf("Misc.Fwdcnt: expected %d, got %d", *input.Misc.Fwdcnt, output.Misc.Fwdcnt)
	}
	if output.Misc.Username != input.Misc.Username {
		t.Errorf("Misc.Username: expected %s, got %s", input.Misc.Username, output.Misc.Username)
	}
	if output.Misc.Answer != input.Misc.Answer {
		t.Errorf("Misc.Answer: expected %d, got %d", input.Misc.Answer, output.Misc.Answer)
	}
	if output.Misc.Botnick != input.Misc.Botnick {
		t.Errorf("Misc.Botnick: expected %s, got %s", input.Misc.Botnick, output.Misc.Botnick)
	}
	if output.Misc.Msgformat != input.Misc.Msgformat {
		t.Errorf("Misc.Msgformat: expected %d, got %d", input.Misc.Msgformat, output.Misc.Msgformat)
	}
	if output.Misc.GoodMorning != input.Misc.GoodMorning {
		t.Errorf("Misc.GoodMorning: expected %d, got %d", input.Misc.GoodMorning, output.Misc.GoodMorning)
	}
}

// convertToSMsg simulates the conversion logic from msgParser.go
func convertToSMsg(j rMsg) sMsg {
	var message sMsg
	message.From = j.From
	message.Userid = j.Userid
	message.Chatid = j.Chatid
	message.Threadid = j.Threadid
	message.Message = j.Message
	message.Plugin = j.Plugin
	message.Mode = j.Mode
	message.Misc.Fwdcnt = *j.Misc.Fwdcnt
	message.Misc.Csign = j.Misc.Csign
	message.Misc.Username = j.Misc.Username
	message.Misc.Answer = j.Misc.Answer
	message.Misc.Botnick = j.Misc.Botnick
	message.Misc.Msgformat = j.Misc.Msgformat
	message.Misc.GoodMorning = j.Misc.GoodMorning
	return message
}

func TestJSONFieldNaming(t *testing.T) {
	// Test that sMsg uses lowercase "misc" in JSON
	msg := sMsg{
		From:    "test",
		Chatid:  "1",
		Userid:  "2",
		Message: "hello",
		Plugin:  "test",
		Mode:    "test",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Check for lowercase "misc"
	if !contains(string(data), `"misc"`) {
		t.Errorf("expected lowercase 'misc' in JSON, got: %s", string(data))
	}

	// Ensure it doesn't contain capitalized "Misc"
	if contains(string(data), `"Misc"`) {
		t.Errorf("should not contain capitalized 'Misc' in JSON, got: %s", string(data))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestRMsgFwdcntPointer(t *testing.T) {
	// Test that Fwdcnt as pointer works correctly with JSON
	jsonInput := `{"from":"test","chatid":"1","userid":"1","message":"test","plugin":"test","mode":"test","misc":{"fwd_cnt":5}}`

	var msg rMsg
	if err := json.Unmarshal([]byte(jsonInput), &msg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if msg.Misc.Fwdcnt == nil {
		t.Fatal("Fwdcnt should not be nil")
	}

	if *msg.Misc.Fwdcnt != 5 {
		t.Errorf("expected Fwdcnt=5, got %d", *msg.Misc.Fwdcnt)
	}
}

func TestRMsgFwdcntNilWhenOmitted(t *testing.T) {
	jsonInput := `{"from":"test","chatid":"1","userid":"1","message":"test","plugin":"test","mode":"test"}`

	var msg rMsg
	if err := json.Unmarshal([]byte(jsonInput), &msg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if msg.Misc.Fwdcnt != nil {
		t.Errorf("Fwdcnt should be nil when omitted, got %d", *msg.Misc.Fwdcnt)
	}
}

func TestMyConfigForwardChannels(t *testing.T) {
	cfg := myConfig{
		Server:  "localhost",
		Port:    6379,
		Channel: "test",
		Csign:   "!",
	}

	// Test that struct can hold custom values
	cfg.ForwardChannels.Games = "custom-games"
	cfg.ForwardChannels.Phrases = "custom-phrases"
	cfg.ForwardChannels.Webapp = "custom-webapp"
	cfg.ForwardChannels.WebappGo = "custom-webapp-go"
	cfg.ForwardChannels.Craniac = "custom-craniac"

	if cfg.ForwardChannels.Games != "custom-games" {
		t.Errorf("Games: expected custom-games, got %s", cfg.ForwardChannels.Games)
	}
	if cfg.ForwardChannels.Phrases != "custom-phrases" {
		t.Errorf("Phrases: expected custom-phrases, got %s", cfg.ForwardChannels.Phrases)
	}
	if cfg.ForwardChannels.Webapp != "custom-webapp" {
		t.Errorf("Webapp: expected custom-webapp, got %s", cfg.ForwardChannels.Webapp)
	}
	if cfg.ForwardChannels.WebappGo != "custom-webapp-go" {
		t.Errorf("WebappGo: expected custom-webapp-go, got %s", cfg.ForwardChannels.WebappGo)
	}
	if cfg.ForwardChannels.Craniac != "custom-craniac" {
		t.Errorf("Craniac: expected custom-craniac, got %s", cfg.ForwardChannels.Craniac)
	}
}
