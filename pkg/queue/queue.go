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
	Clipboard Clipboard
}

func NewManager(s storage.Storage, c Clipboard) *Manager {
	return &Manager{
		storage:   s,
		Clipboard: c,
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

	// Don't add duplicate if it's the same as the last one
	if len(state.Items) > 0 && state.Items[len(state.Items)-1] == item {
		return nil
	}

	state.Items = append(state.Items, item)
	if err := m.storage.Save(state); err != nil {
		return err
	}

	// In FIFO mode, the system clipboard should always contain the FIRST item to be popped.
	// After adding a new item, the system clipboard contains the NEW item.
	// We need to restore it to the first item if we have more than one.
	if !state.IsStack && len(state.Items) > 1 {
		if err := m.Clipboard.Write(state.Items[0]); err != nil {
			return fmt.Errorf("failed to restore first item to clipboard: %w", err)
		}
	}

	return nil
}

// Pop removes an item from the queue and prepares the NEXT item in the clipboard
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

	// After popping, prepare the NEXT item in the clipboard so the next paste is correct.
	if len(state.Items) > 0 {
		nextItem := state.Items[0]
		if isStack {
			nextItem = state.Items[len(state.Items)-1]
		}
		if err := m.Clipboard.Write(nextItem); err != nil {
			return "", fmt.Errorf("failed to prepare next item in clipboard: %w", err)
		}
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
		state.Items = []string{} // Clear on deactivate
	} else {
		state.Items = []string{} // Cmd+I: Start/Activate and clear
	}
	return m.storage.Save(state)
}

// SetStackMode sets whether to use LIFO (stack) or FIFO (queue)
func (m *Manager) SetStackMode(isStack bool) error {
	state, err := m.storage.Load()
	if err != nil {
		return err
	}
	state.IsStack = isStack

	// If the mode changed and we have items, we should update the clipboard to the new "next" item.
	if len(state.Items) > 0 {
		nextItem := state.Items[0]
		if isStack {
			nextItem = state.Items[len(state.Items)-1]
		}
		if err := m.Clipboard.Write(nextItem); err != nil {
			return fmt.Errorf("failed to update clipboard for new mode: %w", err)
		}
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
