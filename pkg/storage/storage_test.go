package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJSONStorage_SaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cbq-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "state.json")
	s := NewJSONStorage(path)

	state := &State{
		Items:  []string{"item1", "item2"},
		Active: true,
	}

	err = s.Save(state)
	if err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if !loaded.Active {
		t.Errorf("expected active to be true")
	}
	if len(loaded.Items) != 2 || loaded.Items[0] != "item1" || loaded.Items[1] != "item2" {
		t.Errorf("loaded items incorrect: %v", loaded.Items)
	}
}

func TestJSONStorage_LoadNonExistent(t *testing.T) {
	s := NewJSONStorage("/non/existent/path/state.json")
	state, err := s.Load()
	if err != nil {
		t.Fatalf("unexpected error loading non-existent file: %v", err)
	}
	if state.Active != false || len(state.Items) != 0 {
		t.Errorf("expected empty state for non-existent file, got %v", state)
	}
}

func TestJSONStorage_Clear(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cbq-test")
	defer os.RemoveAll(tmpDir)
	path := filepath.Join(tmpDir, "state.json")
	s := NewJSONStorage(path)

	s.Save(&State{Items: []string{"a", "b"}, Active: true})
	
	err := s.Clear()
	if err != nil {
		t.Fatalf("failed to clear: %v", err)
	}

	loaded, _ := s.Load()
	if len(loaded.Items) != 0 {
		t.Errorf("expected 0 items after clear, got %d", len(loaded.Items))
	}
	if !loaded.Active {
		t.Error("expected active state to be preserved after clear")
	}
}
