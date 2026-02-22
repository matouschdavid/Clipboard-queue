package monitor

import (
	"log"
	"time"

	"github.com/atotto/clipboard"
	"github.com/vibe-coding/cbq/pkg/storage"
)

func Start() {
	lastText, err := clipboard.ReadAll()
	if err != nil {
		log.Println("Error reading clipboard:", err)
	}

	for {
		time.Sleep(1 * time.Second)
		currentText, err := clipboard.ReadAll()
		if err != nil {
			log.Println("Error reading clipboard:", err)
			continue
		}

		if currentText != lastText && currentText != "" {
			items, _ := storage.Load()
			items = append(items, currentText)
			storage.Save(items)
			lastText = currentText
			log.Printf("Copied: %s\n", currentText)
		}
	}
}
