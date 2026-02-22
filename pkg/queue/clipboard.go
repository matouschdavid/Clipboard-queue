package queue

import "github.com/atotto/clipboard"

// SystemClipboard implements Clipboard using the atotto/clipboard package
type SystemClipboard struct{}

func (s *SystemClipboard) Read() (string, error) {
	return clipboard.ReadAll()
}

func (s *SystemClipboard) Write(text string) error {
	return clipboard.WriteAll(text)
}
