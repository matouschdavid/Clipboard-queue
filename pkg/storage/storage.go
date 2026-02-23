package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type State struct {
	Items   []string `json:"items"`
	Active  bool     `json:"active"`
	IsStack bool     `json:"is_stack"`
}

type Storage interface {
	Load() (*State, error)
	Save(state *State) error
	Clear() error
}

type JSONStorage struct {
	Path string
}

func NewJSONStorage(path string) *JSONStorage {
	return &JSONStorage{Path: path}
}

func GetDefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cbq", "state.json"), nil
}

func (s *JSONStorage) Load() (*State, error) {
	if _, err := os.Stat(s.Path); os.IsNotExist(err) {
		return &State{Items: []string{}, Active: false}, nil
	}
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, err
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	if state.Items == nil {
		state.Items = []string{}
	}
	return &state, nil
}

// Save writes state atomically via a temp file + rename to prevent corruption on crash.
func (s *JSONStorage) Save(state *State) error {
	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".cbq-state-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op if rename succeeded

	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.Path)
}

func (s *JSONStorage) Clear() error {
	state, err := s.Load()
	if err != nil {
		return err
	}
	state.Items = []string{}
	return s.Save(state)
}
