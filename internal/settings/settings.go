package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Prompt is a named prompt snippet that can be injected into the agent system prompt.
type Prompt struct {
	Name    string `json:"name"`
	Prompt  string `json:"prompt"`
	Default bool   `json:"default"`
}

// Settings holds global ccmux user preferences.
type Settings struct {
	Prompts []Prompt `json:"prompts,omitempty"`
}

// Store manages the ~/.ccmux/settings.json file.
type Store struct {
	mu       sync.Mutex
	filePath string
}

func NewStore() (*Store, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	ccmuxDir := filepath.Join(homeDir, ".ccmux")
	if err := os.MkdirAll(ccmuxDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create ccmux directory: %w", err)
	}

	return &Store{
		filePath: filepath.Join(ccmuxDir, "settings.json"),
	}, nil
}

func (s *Store) Load() (*Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	settings := &Settings{}

	raw, err := os.ReadFile(s.filePath)
	if os.IsNotExist(err) {
		return settings, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	if err := json.Unmarshal(raw, settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings file: %w", err)
	}

	return settings, nil
}

func (s *Store) Save(settings *Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bytes, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(s.filePath, bytes, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}
