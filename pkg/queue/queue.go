package queue

import (
	"errors"
	"sync"

	"github.com/matouschdavid/Clipboard-queue/pkg/storage"
)

// Clipboard interface allows mocking the system clipboard for tests.
type Clipboard interface {
	Read() (string, error)
	Write(text string) error
}

// Manager handles the core business logic of the clipboard queue.
type Manager struct {
	storage   storage.Storage
	clipboard Clipboard
	mu        sync.Mutex
	state     *storage.State
}

func NewManager(s storage.Storage, c Clipboard) *Manager {
	return &Manager{
		storage:   s,
		clipboard: c,
	}
}

// load returns the cached state, loading from storage if needed.
// Must be called with m.mu held.
func (m *Manager) load() (*storage.State, error) {
	if m.state != nil {
		return m.state, nil
	}
	state, err := m.storage.Load()
	if err != nil {
		return nil, err
	}
	m.state = state
	return state, nil
}

// save persists state and keeps the cache consistent.
// Rolls back in-memory changes if the write fails.
// Must be called with m.mu held.
func (m *Manager) save(state *storage.State, rollback func()) error {
	if err := m.storage.Save(state); err != nil {
		rollback()
		return err
	}
	return nil
}

// Add appends a new item to the queue if it's active.
func (m *Manager) Add(item string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.load()
	if err != nil {
		return err
	}
	if !state.Active {
		return nil
	}
	if len(state.Items) > 0 && state.Items[len(state.Items)-1] == item {
		return nil // deduplicate consecutive copies
	}

	state.Items = append(state.Items, item)
	return m.save(state, func() { state.Items = state.Items[:len(state.Items)-1] })
}

// AddAndSync appends an item and updates the clipboard in one atomic operation.
func (m *Manager) AddAndSync(item string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.load()
	if err != nil {
		return err
	}
	if !state.Active {
		return nil
	}
	if len(state.Items) > 0 && state.Items[len(state.Items)-1] == item {
		return nil // deduplicate consecutive copies
	}

	state.Items = append(state.Items, item)
	if err := m.save(state, func() { state.Items = state.Items[:len(state.Items)-1] }); err != nil {
		return err
	}
	return m.sync(state)
}

// Pop removes an item from the queue (LIFO if isStack, else FIFO).
func (m *Manager) Pop(isStack bool) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.load()
	if err != nil {
		return "", err
	}
	if len(state.Items) == 0 {
		return "", errors.New("queue is empty")
	}

	item, prev := popItem(state, isStack)
	if err := m.save(state, func() { state.Items = prev }); err != nil {
		return "", err
	}
	return item, nil
}

// PopAndSync removes the next item, prepares the one after it on the clipboard,
// and reads isStack from the persisted state (no TOCTOU race).
func (m *Manager) PopAndSync() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.load()
	if err != nil {
		return "", err
	}
	if len(state.Items) == 0 {
		return "", errors.New("queue is empty")
	}

	item, prev := popItem(state, state.IsStack)
	if err := m.save(state, func() { state.Items = prev }); err != nil {
		return "", err
	}
	if err := m.sync(state); err != nil {
		return item, err // item was popped successfully; sync failure is non-fatal
	}
	return item, nil
}

// popItem removes the appropriate element from state.Items according to mode,
// returns the popped value and the previous Items slice for rollback.
// Must be called with m.mu held.
func popItem(state *storage.State, isStack bool) (item string, prev []string) {
	prev = state.Items
	if isStack {
		item = state.Items[len(state.Items)-1]
		state.Items = state.Items[:len(state.Items)-1]
	} else {
		item = state.Items[0]
		// Copy to a new backing array to release the memory of the old first element.
		state.Items = append([]string(nil), state.Items[1:]...)
	}
	return item, prev
}

// SetActive activates or deactivates collection, clearing the queue either way.
func (m *Manager) SetActive(active bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.load()
	if err != nil {
		return err
	}
	prev := storage.State{Active: state.Active, Items: state.Items, IsStack: state.IsStack}
	state.Active = active
	state.Items = []string{}
	return m.save(state, func() {
		state.Active = prev.Active
		state.Items = prev.Items
	})
}

// SetStackMode switches between LIFO (stack) and FIFO (queue).
func (m *Manager) SetStackMode(isStack bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.load()
	if err != nil {
		return err
	}
	prev := state.IsStack
	state.IsStack = isStack
	return m.save(state, func() { state.IsStack = prev })
}

// SyncClipboard writes the current "next" item to the system clipboard.
func (m *Manager) SyncClipboard() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.load()
	if err != nil {
		return err
	}
	return m.sync(state)
}

// sync writes the next item to the clipboard.
// Must be called with m.mu held.
func (m *Manager) sync(state *storage.State) error {
	if len(state.Items) == 0 {
		return nil
	}
	next := state.Items[0]
	if state.IsStack {
		next = state.Items[len(state.Items)-1]
	}
	return m.clipboard.Write(next)
}

// GetStatus returns the current state.
func (m *Manager) GetStatus() (*storage.State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.load()
}

// Clear empties the queue.
func (m *Manager) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.storage.Clear(); err != nil {
		return err
	}
	m.state = nil // invalidate cache
	return nil
}
