package queue

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestQueue(t *testing.T) (*Queue, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "ccmux-queue-test")
	if err != nil {
		t.Fatal(err)
	}

	q := &Queue{
		filePath: filepath.Join(tmpDir, "queue.json"),
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return q, cleanup
}

func TestAdd_ShouldCreateQueueItem_GivenValidInput(t *testing.T) {
	// Setup.
	q, cleanup := setupTestQueue(t)
	defer cleanup()

	// Execute.
	item, err := q.Add(ItemTypeQuestion, "agent-1", "Test question", "Details here")

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID == "" {
		t.Error("expected item ID to be set")
	}
	if item.Type != ItemTypeQuestion {
		t.Errorf("expected type %s, got %s", ItemTypeQuestion, item.Type)
	}
	if item.AgentID != "agent-1" {
		t.Errorf("expected agent ID agent-1, got %s", item.AgentID)
	}
}

func TestList_ShouldReturnAllItems_GivenMultipleAdds(t *testing.T) {
	// Setup.
	q, cleanup := setupTestQueue(t)
	defer cleanup()
	q.Add(ItemTypeQuestion, "agent-1", "Question 1", "")
	q.Add(ItemTypePRReady, "agent-2", "PR ready", "")

	// Execute.
	items, err := q.List()

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestListByType_ShouldFilterByType_GivenMixedItems(t *testing.T) {
	// Setup.
	q, cleanup := setupTestQueue(t)
	defer cleanup()
	q.Add(ItemTypeQuestion, "agent-1", "Question", "")
	q.Add(ItemTypePRReady, "agent-2", "PR ready", "")
	q.Add(ItemTypeQuestion, "agent-3", "Another question", "")

	// Execute.
	items, err := q.ListByType(ItemTypeQuestion)

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 question items, got %d", len(items))
	}
}

func TestRemove_ShouldDeleteItem_GivenValidID(t *testing.T) {
	// Setup.
	q, cleanup := setupTestQueue(t)
	defer cleanup()
	item, _ := q.Add(ItemTypeQuestion, "agent-1", "Question", "")

	// Execute.
	err := q.Remove(item.ID)

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items, _ := q.List()
	if len(items) != 0 {
		t.Errorf("expected 0 items after removal, got %d", len(items))
	}
}

func TestRemoveByAgent_ShouldDeleteAllAgentItems_GivenAgentID(t *testing.T) {
	// Setup.
	q, cleanup := setupTestQueue(t)
	defer cleanup()
	q.Add(ItemTypeQuestion, "agent-1", "Question 1", "")
	q.Add(ItemTypePRReady, "agent-1", "PR ready", "")
	q.Add(ItemTypeQuestion, "agent-2", "Question 2", "")

	// Execute.
	err := q.RemoveByAgent("agent-1")

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items, _ := q.List()
	if len(items) != 1 {
		t.Errorf("expected 1 item after removal, got %d", len(items))
	}
	if items[0].AgentID != "agent-2" {
		t.Errorf("expected remaining item from agent-2, got %s", items[0].AgentID)
	}
}

func TestClear_ShouldRemoveAllItems_GivenPopulatedQueue(t *testing.T) {
	// Setup.
	q, cleanup := setupTestQueue(t)
	defer cleanup()
	q.Add(ItemTypeQuestion, "agent-1", "Question", "")
	q.Add(ItemTypePRReady, "agent-2", "PR ready", "")

	// Execute.
	err := q.Clear()

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items, _ := q.List()
	if len(items) != 0 {
		t.Errorf("expected 0 items after clear, got %d", len(items))
	}
}
