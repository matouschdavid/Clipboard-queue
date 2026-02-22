package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Storage struct {
	Items []string `json:"items"`
}

func getPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(home, ".cbq", "state.json")
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
	}
	return path, nil
}

func Load() ([]string, error) {
	path, err := getPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []string{}, nil
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var storage Storage
	if err := json.Unmarshal(data, &storage); err != nil {
		return nil, err
	}
	return storage.Items, nil
}

func Save(items []string) error {
	path, err := getPath()
	if err != nil {
		return err
	}
	storage := Storage{Items: items}
	data, err := json.Marshal(storage)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

func Clear() error {
	return Save([]string{})
}
