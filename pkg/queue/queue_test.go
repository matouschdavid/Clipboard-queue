package queue

import (
	"testing"
	"github.com/matouschdavid/Clipboard-queue/pkg/storage"
)

// MockStorage implements storage.Storage for testing
type MockStorage struct {
	state *storage.State
	err   error
}

func (m *MockStorage) Load() (*storage.State, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.state, nil
}

func (m *MockStorage) Save(state *storage.State) error {
	if m.err != nil {
		return m.err
	}
	m.state = state
	return nil
}

func (m *MockStorage) Clear() error {
	if m.err != nil {
		return m.err
	}
	m.state.Items = []string{}
	return nil
}

// MockClipboard implements Clipboard for testing
type MockClipboard struct {
	content string
	err     error
}

func (m *MockClipboard) Read() (string, error) {
	return m.content, m.err
}

func (m *MockClipboard) Write(text string) error {
	if m.err != nil {
		return m.err
	}
	m.content = text
	return nil
}

func TestManager_Add(t *testing.T) {
	s := &MockStorage{state: &storage.State{Active: true, Items: []string{}}}
	c := &MockClipboard{}
	mgr := NewManager(s, c)

	err := mgr.Add("item1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s.state.Items) != 1 || s.state.Items[0] != "item1" {
		t.Errorf("item1 was not added. state: %v", s.state)
	}

	// For the first item, clipboard is already correct (item1)
	// Now add item2. In FIFO mode, it should restore item1 to clipboard.
	err = mgr.Add("item2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.state.Items) != 2 || s.state.Items[0] != "item1" {
		t.Errorf("item2 was not added correctly. state: %v", s.state)
	}

	// Must call SyncClipboard explicitly now
	mgr.SyncClipboard()
	if c.content != "item1" {
		t.Errorf("expected item1 to be restored to clipboard, got %s", c.content)
	}

	// Try adding when inactive
	s.state.Active = false
	err = mgr.Add("item3")
	if err != nil {
		t.Fatalf("unexpected error when inactive: %v", err)
	}
	if len(s.state.Items) != 2 {
		t.Errorf("item was added even if manager was inactive")
	}

	// Try adding duplicate
	s.state.Active = true
	err = mgr.Add("item2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.state.Items) != 2 {
		t.Errorf("duplicate item was added")
	}
}

func TestManager_Pop(t *testing.T) {
	s := &MockStorage{state: &storage.State{
		Active: true,
		Items:  []string{"item1", "item2", "item3"},
	}}
	c := &MockClipboard{}
	mgr := NewManager(s, c)

	// Test FIFO (Queue mode)
	item, err := mgr.Pop(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item1" {
		t.Errorf("expected item1, got %s", item)
	}
	// SyncClipboard prepares the NEXT item (item2)
	mgr.SyncClipboard()
	if c.content != "item2" {
		t.Errorf("expected clipboard content item2 (prepared next), got %s", c.content)
	}
	if len(s.state.Items) != 2 || s.state.Items[0] != "item2" {
		t.Errorf("queue state incorrect after FIFO pop: %v", s.state.Items)
	}

	// Test LIFO (Stack mode)
	// Reset and enable stack mode
	s.state.Items = []string{"item1", "item2", "item3"}
	s.state.IsStack = true

	item, err = mgr.Pop(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item3" {
		t.Errorf("expected item3, got %s", item)
	}
	// SyncClipboard prepares the NEXT item (item2)
	mgr.SyncClipboard()
	if c.content != "item2" {
		t.Errorf("expected clipboard content item2 (prepared next), got %s", c.content)
	}
	if len(s.state.Items) != 2 || s.state.Items[0] != "item1" || s.state.Items[1] != "item2" {
		t.Errorf("queue state incorrect after first LIFO pop: %v", s.state.Items)
	}

	// Pop next (still LIFO)
	item, err = mgr.Pop(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item2" {
		t.Errorf("expected item2, got %s", item)
	}
	// SyncClipboard prepares next (item1)
	mgr.SyncClipboard()
	if c.content != "item1" {
		t.Errorf("expected clipboard content item1 (prepared next), got %s", c.content)
	}
	if len(s.state.Items) != 1 || s.state.Items[0] != "item1" {
		t.Errorf("queue state incorrect after second LIFO pop: %v", s.state.Items)
	}

	// Pop last one
	item, err = mgr.Pop(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item1" {
		t.Errorf("expected item1, got %s", item)
	}
	if len(s.state.Items) != 0 {
		t.Errorf("expected empty queue, got %v", s.state.Items)
	}

	// Pop empty queue should error
	_, err = mgr.Pop(true)
	if err == nil {
		t.Error("expected error when popping empty queue")
	}
}

func TestManager_SetActive(t *testing.T) {
	s := &MockStorage{state: &storage.State{
		Active: false,
		Items:  []string{"something"},
	}}
	c := &MockClipboard{}
	mgr := NewManager(s, c)

	// Activate (should clear)
	err := mgr.SetActive(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.state.Active {
		t.Error("expected active state")
	}
	if len(s.state.Items) != 0 {
		t.Error("expected items to be cleared on activate")
	}

	// Add something and deactivate
	s.state.Items = append(s.state.Items, "item")
	err = mgr.SetActive(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.state.Active {
		t.Error("expected inactive state")
	}
	if len(s.state.Items) != 0 {
		t.Error("expected items to be cleared on deactivate")
	}
}

func TestManager_SetStackMode(t *testing.T) {
	s := &MockStorage{state: &storage.State{
		Active:  true,
		Items:   []string{"item1", "item2"},
		IsStack: false,
	}}
	c := &MockClipboard{}
	mgr := NewManager(s, c)

	// Switch to stack mode
	err := mgr.SetStackMode(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.state.IsStack {
		t.Error("expected stack mode")
	}
	// Must call SyncClipboard
	mgr.SyncClipboard()
	if c.content != "item2" {
		t.Errorf("expected item2 in clipboard, got %s", c.content)
	}

	// Switch back to queue mode
	err = mgr.SetStackMode(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.state.IsStack {
		t.Error("expected queue mode")
	}
	// Must call SyncClipboard
	mgr.SyncClipboard()
	if c.content != "item1" {
		t.Errorf("expected item1 in clipboard, got %s", c.content)
	}
}

func TestManager_AddAndSync(t *testing.T) {
	s := &MockStorage{state: &storage.State{Active: true, Items: []string{}}}
	c := &MockClipboard{}
	mgr := NewManager(s, c)

	// Add first item
	err := mgr.AddAndSync("item1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.content != "item1" {
		t.Errorf("expected item1 in clipboard, got %s", c.content)
	}

	// Add second item (Queue mode)
	err = mgr.AddAndSync("item2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// In Queue mode, clipboard should stay at "item1"
	if c.content != "item1" {
		t.Errorf("expected item1 to remain in clipboard, got %s", c.content)
	}
	if len(s.state.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(s.state.Items))
	}

	// Switch to stack mode and add third item
	s.state.IsStack = true
	err = mgr.AddAndSync("item3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// In Stack mode, clipboard should be "item3"
	if c.content != "item3" {
		t.Errorf("expected item3 in clipboard (LIFO), got %s", c.content)
	}
}

func TestManager_PopAndSync(t *testing.T) {
	s := &MockStorage{state: &storage.State{
		Active: true,
		Items:  []string{"item1", "item2", "item3"},
	}}
	c := &MockClipboard{}
	mgr := NewManager(s, c)

	// Pop first item (Queue mode)
	item, err := mgr.PopAndSync(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item1" {
		t.Errorf("expected popped item1, got %s", item)
	}
	// Should have prepared item2
	if c.content != "item2" {
		t.Errorf("expected item2 prepared in clipboard, got %s", c.content)
	}

	// Pop next (still Queue mode)
	item, err = mgr.PopAndSync(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item2" {
		t.Errorf("expected popped item2, got %s", item)
	}
	// Should have prepared item3
	if c.content != "item3" {
		t.Errorf("expected item3 prepared in clipboard, got %s", c.content)
	}

	// Reset for Stack mode
	s.state.Items = []string{"item1", "item2", "item3"}
	s.state.IsStack = true
	c.content = "item3" // Initial state for stack

	// Pop from stack
	item, err = mgr.PopAndSync(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item3" {
		t.Errorf("expected popped item3, got %s", item)
	}
	// Should have prepared item2
	if c.content != "item2" {
		t.Errorf("expected item2 prepared in clipboard (LIFO), got %s", c.content)
	}
}
