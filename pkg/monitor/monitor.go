package monitor

import (
	"log"
	"sync"
	"time"

	hook "github.com/robotn/gohook"
	"github.com/vibe-coding/cbq/pkg/queue"
	"github.com/vibe-coding/cbq/pkg/storage"
)

// Modifier masks for gohook
const (
	maskShift = 0x0001 | 0x0010
	maskCtrl  = 0x0002 | 0x0020
	maskMeta  = 0x0004 | 0x0040 // Cmd on macOS
	maskAlt   = 0x0008 | 0x0080 // Option on macOS
)

// macOS Virtual Keycodes
const (
	keyC = 8
	keyV = 9
	keyI = 34
	keyR = 15
)

func getManager() *queue.Manager {
	path, err := storage.GetDefaultPath()
	if err != nil {
		log.Fatalf("failed to get storage path: %v", err)
	}
	s := storage.NewJSONStorage(path)
	c := &queue.SystemClipboard{}
	return queue.NewManager(s, c)
}

func Start() {
	evChan := hook.Start()
	defer hook.End()

	mgr := getManager()

	var (
		popMu     sync.Mutex
		isPopping bool
	)

	log.Println("CBQ Monitor started.")
	log.Println("Hotkeys (macOS):")
	log.Println("  Cmd+I: Start queue (Enable collection)")
	log.Println("  Cmd+R: Clear and deactivate queue")
	log.Println("  Cmd+C: Record to queue (when active)")
	log.Println("  Cmd+V: Pop from queue (when active)")

	for ev := range evChan {
		if ev.Kind != hook.KeyDown {
			continue
		}

		// Cmd key check (Meta on macOS)
		isCmd := (ev.Mask & maskMeta) != 0

		if isCmd {
			switch ev.Rawcode {
			case keyI: // Cmd+I: Start/Activate
				if err := mgr.SetActive(true); err != nil {
					log.Printf("Error activating: %v", err)
				} else {
					log.Println(">>> Queue STARTED (Active)")
				}
			case keyR: // Cmd+R: Reset/Deactivate
				if err := mgr.SetActive(false); err != nil {
					log.Printf("Error deactivating: %v", err)
				} else {
					log.Println(">>> Queue CLEARED and Deactivated")
				}
			case keyC: // Cmd+C: Copy
				state, _ := mgr.GetStatus()
				if state.Active {
					// Wait a bit for the system to update clipboard
					go func() {
						time.Sleep(50 * time.Millisecond) // Faster capture
						clipboard := &queue.SystemClipboard{}
						text, err := clipboard.Read()
						if err == nil && text != "" {
							if err := mgr.Add(text); err == nil {
								// After Add, restore the "current" item to clipboard (e.g. if in FIFO)
								if !state.IsStack {
									mgr.SyncClipboard()
								}
								log.Printf("Copied to queue: %s\n", text)
							}
						}
					}()
				}
			case keyV: // Cmd+V: Paste/Pop
				state, _ := mgr.GetStatus()
				if state.Active && len(state.Items) > 0 {
					popMu.Lock()
					if isPopping {
						popMu.Unlock()
						continue
					}
					isPopping = true
					popMu.Unlock()

					go func() {
						// Wait for OS to paste the current item before we prepare the next one
						time.Sleep(100 * time.Millisecond)
						item, err := mgr.Pop(state.IsStack) // Use stored mode
						if err == nil {
							mgr.SyncClipboard()
							log.Printf("Popped from queue (%s): %s\n", modeStr(state.IsStack), item)
						}

						popMu.Lock()
						isPopping = false
						popMu.Unlock()
					}()
				}
			}
		}
	}
}

func modeStr(isStack bool) string {
	if isStack {
		return "Stack"
	}
	return "Queue"
}
