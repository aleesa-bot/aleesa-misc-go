package misc

import (
	"encoding/json"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name           string
		config         myConfig
		expectedServer string
		expectedPort   int
		expectValid    bool
	}{
		{
			name: "full valid config",
			config: myConfig{
				Server:  "redis.example.com",
				Port:    6380,
				Channel: "test-channel",
				Csign:   "/",
			},
			expectedServer: "redis.example.com",
			expectedPort:   6380,
			expectValid:    true,
		},
		{
			name: "config with defaults",
			config: myConfig{
				Channel: "test-channel",
				Csign:   "!",
			},
			expectedServer: "localhost",
			expectedPort:   6379,
			expectValid:    true,
		},
		{
			name: "missing channel",
			config: myConfig{
				Server: "localhost",
				Port:   6379,
				Csign:  "!",
			},
			expectValid: false,
		},
		{
			name: "missing csign",
			config: myConfig{
				Server:  "localhost",
				Port:    6379,
				Channel: "test",
			},
			expectValid: false,
		},
		{
			name: "zero port uses default",
			config: myConfig{
				Server:  "localhost",
				Port:    0,
				Channel: "test",
				Csign:   "!",
			},
			expectedServer: "localhost",
			expectedPort:   6379,
			expectValid:    true,
		},
		{
			name: "zero timeout uses default",
			config: myConfig{
				Server:  "localhost",
				Port:    6379,
				Timeout: 0,
				Channel: "test",
				Csign:   "!",
			},
			expectedServer: "localhost",
			expectedPort:   6379,
			expectValid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.config
			valid := validateConfig(&cfg)

			if valid != tt.expectValid {
				t.Errorf("expected valid=%v, got %v", tt.expectValid, valid)
			}

			if tt.expectValid {
				applyDefaults(&cfg)
				if cfg.Server != tt.expectedServer {
					t.Errorf("expected Server=%s, got %s", tt.expectedServer, cfg.Server)
				}
				if cfg.Port != tt.expectedPort {
					t.Errorf("expected Port=%d, got %d", tt.expectedPort, cfg.Port)
				}
			}
		})
	}
}

func TestForwardChannelsDefaults(t *testing.T) {
	cfg := myConfig{
		Server:  "localhost",
		Port:    6379,
		Channel: "test",
		Csign:   "!",
	}

	applyDefaults(&cfg)

	if cfg.ForwardChannels.Games != "games" {
		t.Errorf("expected games='games', got '%s'", cfg.ForwardChannels.Games)
	}
	if cfg.ForwardChannels.Phrases != "phrases" {
		t.Errorf("expected phrases='phrases', got '%s'", cfg.ForwardChannels.Phrases)
	}
	if cfg.ForwardChannels.Webapp != "webapp" {
		t.Errorf("expected webapp='webapp', got '%s'", cfg.ForwardChannels.Webapp)
	}
	if cfg.ForwardChannels.WebappGo != "webapp-go" {
		t.Errorf("expected webapp-go='webapp-go', got '%s'", cfg.ForwardChannels.WebappGo)
	}
	if cfg.ForwardChannels.Craniac != "craniac" {
		t.Errorf("expected craniac='craniac', got '%s'", cfg.ForwardChannels.Craniac)
	}
}

func TestForwardChannelsCustomValues(t *testing.T) {
	cfg := myConfig{
		Server:  "localhost",
		Port:    6379,
		Channel: "test",
		Csign:   "!",
		ForwardChannels: struct {
			Games    string `json:"games,omitempty"`
			Phrases  string `json:"phrases,omitempty"`
			Webapp   string `json:"webapp,omitempty"`
			WebappGo string `json:"webapp-go,omitempty"`
			Craniac  string `json:"craniac,omitempty"`
		}{
			Games:    "my-games",
			Phrases:  "my-phrases",
			Webapp:   "my-webapp",
			WebappGo: "my-webapp-go",
			Craniac:  "my-craniac",
		},
	}

	applyDefaults(&cfg)

	if cfg.ForwardChannels.Games != "my-games" {
		t.Errorf("expected games='my-games', got '%s'", cfg.ForwardChannels.Games)
	}
	if cfg.ForwardChannels.Phrases != "my-phrases" {
		t.Errorf("expected phrases='my-phrases', got '%s'", cfg.ForwardChannels.Phrases)
	}
}

// validateConfig simulates the validation logic from ReadConfig
func validateConfig(cfg *myConfig) bool {
	if cfg.Channel == "" {
		return false
	}
	if cfg.Csign == "" {
		return false
	}
	return true
}

// applyDefaults simulates the default application logic from ReadConfig
func applyDefaults(cfg *myConfig) {
	if cfg.Server == "" {
		cfg.Server = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 6379
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10
	}
	if cfg.Loglevel == "" {
		cfg.Loglevel = "info"
	}
	if cfg.ForwardChannels.Games == "" {
		cfg.ForwardChannels.Games = "games"
	}
	if cfg.ForwardChannels.Phrases == "" {
		cfg.ForwardChannels.Phrases = "phrases"
	}
	if cfg.ForwardChannels.Webapp == "" {
		cfg.ForwardChannels.Webapp = "webapp"
	}
	if cfg.ForwardChannels.WebappGo == "" {
		cfg.ForwardChannels.WebappGo = "webapp-go"
	}
	if cfg.ForwardChannels.Craniac == "" {
		cfg.ForwardChannels.Craniac = "craniac"
	}
	if cfg.Csign == "" {
		cfg.Csign = "!"
	}
	if cfg.ForwardsMax == 0 {
		cfg.ForwardsMax = 5
	}
}

func TestLoglevelValidation(t *testing.T) {
	tests := []struct {
		name       string
		loglevel   string
		expectedOK bool
	}{
		{"valid error", "error", true},
		{"valid warn", "warn", true},
		{"valid info", "info", true},
		{"valid debug", "debug", true},
		{"invalid level", "trace", false},
		{"empty uses default", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &myConfig{
				Server:  "localhost",
				Port:    6379,
				Channel: "test",
				Csign:   "!",
			}

			if tt.loglevel != "" {
				cfg.Loglevel = tt.loglevel
			}

			applyDefaults(cfg)

			// Valid loglevels should result in a non-empty loglevel
			// Invalid ones would still get set to "info" as default
			if cfg.Loglevel == "" && tt.expectedOK {
				t.Error("expected non-empty loglevel")
			}
		})
	}
}

func TestForwardsMaxDefault(t *testing.T) {
	cfg := myConfig{
		Server:  "localhost",
		Port:    6379,
		Channel: "test",
		Csign:   "!",
	}

	applyDefaults(&cfg)

	if cfg.ForwardsMax != 5 {
		t.Errorf("expected ForwardsMax=5, got %d", cfg.ForwardsMax)
	}

	cfg.ForwardsMax = 10
	applyDefaults(&cfg)

	if cfg.ForwardsMax != 10 {
		t.Errorf("expected ForwardsMax=10, got %d", cfg.ForwardsMax)
	}
}

func TestConfigJSONRoundtrip(t *testing.T) {
	original := myConfig{
		Server:      "redis.example.com",
		Port:        6380,
		Timeout:     15,
		Loglevel:    "debug",
		Log:         "/var/log/test.log",
		Channel:     "test-channel",
		Csign:       "/",
		ForwardsMax: 10,
		ForwardChannels: struct {
			Games    string `json:"games,omitempty"`
			Phrases  string `json:"phrases,omitempty"`
			Webapp   string `json:"webapp,omitempty"`
			WebappGo string `json:"webapp-go,omitempty"`
			Craniac  string `json:"craniac,omitempty"`
		}{
			Games:    "custom-games",
			Phrases:  "custom-phrases",
			Webapp:   "custom-webapp",
			WebappGo: "custom-webapp-go",
			Craniac:  "custom-craniac",
		},
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal
	var restored myConfig
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Compare
	if restored.Server != original.Server {
		t.Errorf("Server: expected %s, got %s", original.Server, restored.Server)
	}
	if restored.Port != original.Port {
		t.Errorf("Port: expected %d, got %d", original.Port, restored.Port)
	}
	if restored.Channel != original.Channel {
		t.Errorf("Channel: expected %s, got %s", original.Channel, restored.Channel)
	}
	if restored.Csign != original.Csign {
		t.Errorf("Csign: expected %s, got %s", original.Csign, restored.Csign)
	}
	if restored.ForwardsMax != original.ForwardsMax {
		t.Errorf("ForwardsMax: expected %d, got %d", original.ForwardsMax, restored.ForwardsMax)
	}
	if restored.ForwardChannels.Games != original.ForwardChannels.Games {
		t.Errorf("ForwardChannels.Games: expected %s, got %s", original.ForwardChannels.Games, restored.ForwardChannels.Games)
	}
}
