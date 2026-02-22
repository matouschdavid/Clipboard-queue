package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type State struct {
	Items  []string `json:"items"`
	Active bool     `json:"active"`
	IsStack bool    `json:"is_stack"`
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
	return &state, nil
}

func (s *JSONStorage) Save(state *State) error {
	dir := filepath.Dir(s.Path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0644)
}

func (s *JSONStorage) Clear() error {
	state, err := s.Load()
	if err != nil {
		return err
	}
	state.Items = []string{}
	return s.Save(state)
}

// Global helper functions to maintain compatibility with existing code during refactor
func Load() (*State, error) {
	path, _ := GetDefaultPath()
	return NewJSONStorage(path).Load()
}

func Save(state *State) error {
	path, _ := GetDefaultPath()
	return NewJSONStorage(path).Save(state)
}

func Clear() error {
	path, _ := GetDefaultPath()
	return NewJSONStorage(path).Clear()
}
