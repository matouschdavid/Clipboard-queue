package queue

import (
	"errors"
	"sync"
	"github.com/vibe-coding/cbq/pkg/storage"
)

// Clipboard interface allows mocking the system clipboard for tests
type Clipboard interface {
	Read() (string, error)
	Write(text string) error
}

// Manager handles the core business logic of the clipboard queue
type Manager struct {
	storage   storage.Storage
	Clipboard Clipboard
	mu        sync.Mutex
}

func NewManager(s storage.Storage, c Clipboard) *Manager {
	return &Manager{
		storage:   s,
		Clipboard: c,
	}
}

// Add appends a new item to the queue if it's active
func (m *Manager) Add(item string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.storage.Load()
	if err != nil {
		return err
	}

	if !state.Active {
		return nil
	}

	// Don't add duplicate if it's the same as the last one
	if len(state.Items) > 0 && state.Items[len(state.Items)-1] == item {
		return nil
	}

	state.Items = append(state.Items, item)
	return m.storage.Save(state)
}

// Pop removes an item from the queue
// if isStack is true, it uses LIFO, otherwise FIFO
func (m *Manager) Pop(isStack bool) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.storage.Load()
	if err != nil {
		return "", err
	}

	if len(state.Items) == 0 {
		return "", errors.New("queue is empty")
	}

	var item string
	if isStack {
		// LIFO: last item
		item = state.Items[len(state.Items)-1]
		state.Items = state.Items[:len(state.Items)-1]
	} else {
		// FIFO: first item
		item = state.Items[0]
		state.Items = state.Items[1:]
	}

	if err := m.storage.Save(state); err != nil {
		return "", err
	}

	return item, nil
}

// SetActive sets the collection mode state
func (m *Manager) SetActive(active bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.storage.Load()
	if err != nil {
		return err
	}
	state.Active = active
	state.Items = []string{} // Always clear on change for Cmd+I/Cmd+R
	return m.storage.Save(state)
}

// SetStackMode sets whether to use LIFO (stack) or FIFO (queue)
func (m *Manager) SetStackMode(isStack bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.storage.Load()
	if err != nil {
		return err
	}
	state.IsStack = isStack
	return m.storage.Save(state)
}

// SyncClipboard updates the system clipboard to match the current next item in the queue
func (m *Manager) SyncClipboard() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.storage.Load()
	if err != nil {
		return err
	}

	if len(state.Items) == 0 {
		return nil
	}

	nextItem := state.Items[0]
	if state.IsStack {
		nextItem = state.Items[len(state.Items)-1]
	}

	return m.Clipboard.Write(nextItem)
}

// GetStatus returns the current state
func (m *Manager) GetStatus() (*storage.State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.storage.Load()
}

// Clear empties the queue
func (m *Manager) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.storage.Clear()
}
