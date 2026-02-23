package queue

import (
	"testing"

	"github.com/matouschdavid/Clipboard-queue/pkg/storage"
)

// MockStorage implements storage.Storage for testing.
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

// MockClipboard implements Clipboard for testing.
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

	if err := mgr.Add("item1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.state.Items) != 1 || s.state.Items[0] != "item1" {
		t.Errorf("item1 not added: %v", s.state.Items)
	}

	if err := mgr.Add("item2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.state.Items) != 2 || s.state.Items[0] != "item1" {
		t.Errorf("state incorrect after item2: %v", s.state.Items)
	}

	// SyncClipboard should put the first item (FIFO next) onto the clipboard.
	if err := mgr.SyncClipboard(); err != nil {
		t.Fatalf("sync error: %v", err)
	}
	if c.content != "item1" {
		t.Errorf("expected clipboard=item1, got %q", c.content)
	}

	// Inactive: add should be a no-op.
	s.state.Active = false
	if err := mgr.Add("item3"); err != nil {
		t.Fatalf("unexpected error when inactive: %v", err)
	}
	if len(s.state.Items) != 2 {
		t.Errorf("item added while inactive")
	}

	// Duplicate of last item should be rejected.
	s.state.Active = true
	if err := mgr.Add("item2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.state.Items) != 2 {
		t.Errorf("duplicate was added")
	}
}

func TestManager_Pop(t *testing.T) {
	s := &MockStorage{state: &storage.State{
		Active: true,
		Items:  []string{"item1", "item2", "item3"},
	}}
	c := &MockClipboard{}
	mgr := NewManager(s, c)

	// FIFO pop.
	item, err := mgr.Pop(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item1" {
		t.Errorf("expected item1, got %q", item)
	}
	mgr.SyncClipboard()
	if c.content != "item2" {
		t.Errorf("expected clipboard=item2 after FIFO pop, got %q", c.content)
	}
	if len(s.state.Items) != 2 || s.state.Items[0] != "item2" {
		t.Errorf("wrong state after FIFO pop: %v", s.state.Items)
	}

	// LIFO pop.
	s.state.Items = []string{"item1", "item2", "item3"}
	s.state.IsStack = true
	// invalidate cache so load() picks up the reset state
	mgr.state = nil

	item, err = mgr.Pop(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item3" {
		t.Errorf("expected item3, got %q", item)
	}
	mgr.SyncClipboard()
	if c.content != "item2" {
		t.Errorf("expected clipboard=item2 after LIFO pop, got %q", c.content)
	}

	item, err = mgr.Pop(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item2" {
		t.Errorf("expected item2, got %q", item)
	}
	mgr.SyncClipboard()
	if c.content != "item1" {
		t.Errorf("expected clipboard=item1, got %q", c.content)
	}

	item, err = mgr.Pop(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item1" {
		t.Errorf("expected item1, got %q", item)
	}
	if len(s.state.Items) != 0 {
		t.Errorf("expected empty queue, got %v", s.state.Items)
	}

	// Empty queue should error.
	if _, err := mgr.Pop(true); err == nil {
		t.Error("expected error on empty pop")
	}
}

func TestManager_SetActive(t *testing.T) {
	s := &MockStorage{state: &storage.State{
		Active: false,
		Items:  []string{"something"},
	}}
	mgr := NewManager(s, &MockClipboard{})

	if err := mgr.SetActive(true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.state.Active {
		t.Error("expected active")
	}
	if len(s.state.Items) != 0 {
		t.Error("expected items cleared on activate")
	}

	s.state.Items = append(s.state.Items, "item")
	if err := mgr.SetActive(false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.state.Active {
		t.Error("expected inactive")
	}
	if len(s.state.Items) != 0 {
		t.Error("expected items cleared on deactivate")
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

	if err := mgr.SetStackMode(true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.state.IsStack {
		t.Error("expected stack mode")
	}
	mgr.SyncClipboard()
	if c.content != "item2" {
		t.Errorf("expected clipboard=item2 in stack mode, got %q", c.content)
	}

	if err := mgr.SetStackMode(false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.state.IsStack {
		t.Error("expected queue mode")
	}
	mgr.SyncClipboard()
	if c.content != "item1" {
		t.Errorf("expected clipboard=item1 in queue mode, got %q", c.content)
	}
}

func TestManager_AddAndSync(t *testing.T) {
	s := &MockStorage{state: &storage.State{Active: true, Items: []string{}}}
	c := &MockClipboard{}
	mgr := NewManager(s, c)

	if err := mgr.AddAndSync("item1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.content != "item1" {
		t.Errorf("expected clipboard=item1, got %q", c.content)
	}

	// In queue mode, clipboard should stay on first item.
	if err := mgr.AddAndSync("item2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.content != "item1" {
		t.Errorf("expected clipboard=item1 (FIFO), got %q", c.content)
	}
	if len(s.state.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(s.state.Items))
	}

	// In stack mode, clipboard should advance to newest item.
	s.state.IsStack = true
	if err := mgr.AddAndSync("item3"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.content != "item3" {
		t.Errorf("expected clipboard=item3 (LIFO), got %q", c.content)
	}
}

func TestManager_PopAndSync(t *testing.T) {
	s := &MockStorage{state: &storage.State{
		Active:  true,
		IsStack: false,
		Items:   []string{"item1", "item2", "item3"},
	}}
	c := &MockClipboard{}
	mgr := NewManager(s, c)

	// FIFO — pop item1, clipboard should be prepared with item2.
	item, err := mgr.PopAndSync()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item1" {
		t.Errorf("expected item1, got %q", item)
	}
	if c.content != "item2" {
		t.Errorf("expected clipboard=item2, got %q", c.content)
	}

	item, err = mgr.PopAndSync()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item2" {
		t.Errorf("expected item2, got %q", item)
	}
	if c.content != "item3" {
		t.Errorf("expected clipboard=item3, got %q", c.content)
	}

	// Switch to LIFO.
	s.state.Items = []string{"item1", "item2", "item3"}
	s.state.IsStack = true
	mgr.state = nil // invalidate cache

	item, err = mgr.PopAndSync()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item != "item3" {
		t.Errorf("expected item3, got %q", item)
	}
	if c.content != "item2" {
		t.Errorf("expected clipboard=item2 (LIFO), got %q", c.content)
	}
}

func TestManager_FIFOMemoryLeak(t *testing.T) {
	// Verify that repeated FIFO pops don't retain the old backing array.
	// We can't inspect the internal array directly, but we can confirm correct
	// values are returned across many pops without panicking.
	items := make([]string, 100)
	for i := range items {
		items[i] = "x"
	}
	s := &MockStorage{state: &storage.State{Active: true, Items: items}}
	mgr := NewManager(s, &MockClipboard{})

	for i := 0; i < 100; i++ {
		if _, err := mgr.Pop(false); err != nil {
			t.Fatalf("pop %d failed: %v", i, err)
		}
	}
	if len(s.state.Items) != 0 {
		t.Errorf("expected empty queue, got %d items", len(s.state.Items))
	}
}
