// Package queue manages the attention queue for agent events.
package queue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ItemType string

const (
	ItemTypeQuestion ItemType = "question"
	ItemTypePRReady  ItemType = "pr_ready"
	ItemTypeIdle ItemType = "idle"
)

type QueueItem struct {
	ID        string    `json:"id"`
	Type      ItemType  `json:"type"`
	AgentID   string    `json:"agent_id"`
	Summary   string    `json:"summary"`
	Details   string    `json:"details"`
	Timestamp time.Time `json:"timestamp"`
}

type Queue struct {
	mu       sync.Mutex
	filePath string
}

type queueData struct {
	Items   []*QueueItem `json:"items"`
	Counter int          `json:"counter"`
}

func NewQueue(sessionID string) (*Queue, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	sessionDir := filepath.Join(homeDir, ".ccmux", "sessions", sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	return &Queue{
		filePath: filepath.Join(sessionDir, "queue.json"),
	}, nil
}

func (q *Queue) load() (*queueData, error) {
	data := &queueData{
		Items: make([]*QueueItem, 0),
	}

	bytes, err := os.ReadFile(q.filePath)
	if os.IsNotExist(err) {
		return data, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read queue file: %w", err)
	}

	if err := json.Unmarshal(bytes, data); err != nil {
		return nil, fmt.Errorf("failed to parse queue file: %w", err)
	}

	return data, nil
}

func (q *Queue) save(data *queueData) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal queue: %w", err)
	}

	if err := os.WriteFile(q.filePath, bytes, 0644); err != nil {
		return fmt.Errorf("failed to write queue file: %w", err)
	}

	return nil
}

func (q *Queue) Add(itemType ItemType, agentID, summary, details string) (*QueueItem, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	data, err := q.load()
	if err != nil {
		return nil, err
	}

	data.Counter++
	item := &QueueItem{
		ID:        fmt.Sprintf("q%d", data.Counter),
		Type:      itemType,
		AgentID:   agentID,
		Summary:   summary,
		Details:   details,
		Timestamp: time.Now(),
	}

	data.Items = append(data.Items, item)

	if err := q.save(data); err != nil {
		return nil, err
	}

	return item, nil
}

func (q *Queue) List() ([]*QueueItem, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	data, err := q.load()
	if err != nil {
		return nil, err
	}

	return data.Items, nil
}

func (q *Queue) ListByType(itemType ItemType) ([]*QueueItem, error) {
	items, err := q.List()
	if err != nil {
		return nil, err
	}

	var filtered []*QueueItem
	for _, item := range items {
		if item.Type == itemType {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

func (q *Queue) Remove(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	data, err := q.load()
	if err != nil {
		return err
	}

	var newItems []*QueueItem
	found := false
	for _, item := range data.Items {
		if item.ID == id {
			found = true
			continue
		}
		newItems = append(newItems, item)
	}

	if !found {
		return fmt.Errorf("queue item with ID %s not found", id)
	}

	data.Items = newItems

	return q.save(data)
}

func (q *Queue) RemoveByAgent(agentID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	data, err := q.load()
	if err != nil {
		return err
	}

	var newItems []*QueueItem
	for _, item := range data.Items {
		if item.AgentID == agentID {
			continue
		}
		newItems = append(newItems, item)
	}

	data.Items = newItems

	return q.save(data)
}

func (q *Queue) Clear() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	data := &queueData{
		Items:   make([]*QueueItem, 0),
		Counter: 0,
	}

	return q.save(data)
}

func (q *Queue) Get(id string) (*QueueItem, error) {
	items, err := q.List()
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}

	return nil, fmt.Errorf("queue item with ID %s not found", id)
}
