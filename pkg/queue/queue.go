package queue

import (
	"errors"
	"fmt"
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
	clipboard Clipboard
}

func NewManager(s storage.Storage, c Clipboard) *Manager {
	return &Manager{
		storage:   s,
		clipboard: c,
	}
}

// Add appends a new item to the queue if it's active
func (m *Manager) Add(item string) error {
	state, err := m.storage.Load()
	if err != nil {
		return err
	}

	if !state.Active {
		return nil
	}

	// Don't add duplicate if it's the same as the last one (optional, but good for quality)
	if len(state.Items) > 0 && state.Items[len(state.Items)-1] == item {
		return nil
	}

	state.Items = append(state.Items, item)
	return m.storage.Save(state)
}

// Pop removes an item from the queue and writes it to the clipboard
// if isStack is true, it uses LIFO, otherwise FIFO
func (m *Manager) Pop(isStack bool) (string, error) {
	state, err := m.storage.Load()
	if err != nil {
		return "", err
	}

	if len(state.Items) == 0 {
		return "", errors.New("queue is empty")
	}

	var item string
	if isStack {
		// LIFO
		item = state.Items[len(state.Items)-1]
		state.Items = state.Items[:len(state.Items)-1]
	} else {
		// FIFO
		item = state.Items[0]
		state.Items = state.Items[1:]
	}

	if err := m.clipboard.Write(item); err != nil {
		return "", fmt.Errorf("failed to write to clipboard: %w", err)
	}

	if err := m.storage.Save(state); err != nil {
		return "", err
	}

	return item, nil
}

// SetActive sets the collection mode state
func (m *Manager) SetActive(active bool) error {
	state, err := m.storage.Load()
	if err != nil {
		return err
	}
	state.Active = active
	if !active {
		state.Items = []string{} // Clear on deactivate as per requirements? 
		// Actually Cmd+I starts/activates and clears. Cmd+R deactivates and clears.
	} else {
		state.Items = []string{} // Cmd+I: Start/Activate and clear
	}
	return m.storage.Save(state)
}

// GetStatus returns the current state
func (m *Manager) GetStatus() (*storage.State, error) {
	return m.storage.Load()
}

// Clear empties the queue
func (m *Manager) Clear() error {
	return m.storage.Clear()
}
