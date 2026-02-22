package queue

import (
	"testing"
	"github.com/vibe-coding/cbq/pkg/storage"
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

	// Try adding when inactive
	s.state.Active = false
	err = mgr.Add("item2")
	if err != nil {
		t.Fatalf("unexpected error when inactive: %v", err)
	}
	if len(s.state.Items) != 1 {
		t.Errorf("item was added even if manager was inactive")
	}

	// Try adding duplicate
	s.state.Active = true
	err = mgr.Add("item1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.state.Items) != 1 {
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
	if c.content != "item1" {
		t.Errorf("expected clipboard content item1, got %s", c.content)
	}
	if len(s.state.Items) != 2 || s.state.Items[0] != "item2" {
		t.Errorf("queue state incorrect after FIFO pop: %v", s.state.Items)
	}

	// Test LIFO (Stack mode)
	item, err = mgr.Pop(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item3" {
		t.Errorf("expected item3, got %s", item)
	}
	if c.content != "item3" {
		t.Errorf("expected clipboard content item3, got %s", c.content)
	}
	if len(s.state.Items) != 1 || s.state.Items[0] != "item2" {
		t.Errorf("queue state incorrect after LIFO pop: %v", s.state.Items)
	}

	// Pop the last one
	item, _ = mgr.Pop(false)
	if item != "item2" {
		t.Errorf("expected item2, got %s", item)
	}

	// Pop empty queue
	_, err = mgr.Pop(false)
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
