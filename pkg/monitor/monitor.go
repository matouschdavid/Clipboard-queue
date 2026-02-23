package monitor

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	hook "github.com/robotn/gohook"

	"github.com/matouschdavid/Clipboard-queue/pkg/queue"
	"github.com/matouschdavid/Clipboard-queue/pkg/storage"
)

// Modifier masks for gohook.
const (
	maskMeta = 0x0004 | 0x0040 // Cmd on macOS
)

// macOS virtual keycodes.
const (
	keyC = 8
	keyV = 9
	keyI = 34
	keyR = 15
)

// pollInterval is how often the clipboard is checked for changes made outside
// of Cmd+C (e.g. browser "copy to clipboard" buttons).
const pollInterval = 250 * time.Millisecond

func newManager() *queue.Manager {
	path, err := storage.GetDefaultPath()
	if err != nil {
		log.Fatalf("failed to get storage path: %v", err)
	}
	return queue.NewManager(storage.NewJSONStorage(path), &queue.SystemClipboard{})
}

// clipboardPoller watches for clipboard changes that happen without a Cmd+C
// keypress (e.g. website "copy to clipboard" buttons) and feeds them into the
// queue while it is active.
type clipboardPoller struct {
	stop chan struct{}
}

func startPoller(mgr *queue.Manager) *clipboardPoller {
	p := &clipboardPoller{stop: make(chan struct{})}
	go p.run(mgr)
	return p
}

func (p *clipboardPoller) close() {
	close(p.stop)
}

func (p *clipboardPoller) run(mgr *queue.Manager) {
	cb := &queue.SystemClipboard{}

	// Seed with the current clipboard content so we don't immediately capture
	// whatever was on the clipboard before the queue was activated.
	var lastSeen string
	if text, err := cb.Read(); err == nil {
		lastSeen = text
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stop:
			return
		case <-ticker.C:
			text, err := cb.Read()
			if err != nil || text == "" || text == lastSeen {
				continue
			}
			lastSeen = text
			if err := mgr.AddAndSync(text); err != nil {
				log.Printf("Poller: error adding to queue: %v", err)
			} else {
				log.Printf("Captured: %q", text)
			}
		}
	}
}

func Start() {
	// Graceful shutdown on SIGINT / SIGTERM.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("Received %v, shutting down.", sig)
		hook.End()
	}()

	evChan := hook.Start()
	defer hook.End()

	mgr := newManager()

	// Sync clipboard on start in case the monitor was restarted with items in the queue.
	if err := mgr.SyncClipboard(); err != nil {
		log.Printf("Warning: initial clipboard sync failed: %v", err)
	}

	log.Println("CBQ monitor started.")
	log.Println("  Cmd+I  start (clears queue)")
	log.Println("  Cmd+R  stop  (clears queue)")
	log.Println("  Cmd+C  record copy")
	log.Println("  Cmd+V  paste & advance")

	// If the queue was left active from a previous session, resume polling.
	var poller *clipboardPoller
	if state, err := mgr.GetStatus(); err == nil && state.Active {
		log.Println("Resuming active queue from previous session.")
		poller = startPoller(mgr)
	}

	var (
		popMu     sync.Mutex
		isPopping bool
	)

	for ev := range evChan {
		if ev.Kind != hook.KeyDown {
			continue
		}
		if ev.Mask&maskMeta == 0 {
			continue
		}

		switch ev.Rawcode {

		case keyI: // Cmd+I — activate and clear
			if err := mgr.SetActive(true); err != nil {
				log.Printf("Error activating: %v", err)
				continue
			}
			if poller != nil {
				poller.close()
			}
			poller = startPoller(mgr)
			log.Println("Queue STARTED")

		case keyR: // Cmd+R — deactivate and clear
			if err := mgr.SetActive(false); err != nil {
				log.Printf("Error deactivating: %v", err)
				continue
			}
			if poller != nil {
				poller.close()
				poller = nil
			}
			log.Println("Queue STOPPED")

		case keyC: // Cmd+C — record copy immediately (low-latency path)
			state, err := mgr.GetStatus()
			if err != nil {
				log.Printf("Error reading state: %v", err)
				continue
			}
			if !state.Active {
				continue
			}
			go func() {
				// Give the OS a moment to update the clipboard after Cmd+C.
				time.Sleep(30 * time.Millisecond)
				cb := &queue.SystemClipboard{}
				text, err := cb.Read()
				if err != nil || text == "" {
					return
				}
				if err := mgr.AddAndSync(text); err != nil {
					log.Printf("Error adding to queue: %v", err)
					return
				}
				log.Printf("Added: %q", text)
			}()

		case keyV: // Cmd+V — paste current item and prepare the next
			state, err := mgr.GetStatus()
			if err != nil {
				log.Printf("Error reading state: %v", err)
				continue
			}
			if !state.Active || len(state.Items) == 0 {
				continue
			}

			// Proactively ensure the clipboard has the correct item before the OS pastes it.
			if err := mgr.SyncClipboard(); err != nil {
				log.Printf("Warning: clipboard sync failed: %v", err)
			}

			popMu.Lock()
			if isPopping {
				popMu.Unlock()
				continue
			}
			isPopping = true
			popMu.Unlock()

			go func() {
				defer func() {
					popMu.Lock()
					isPopping = false
					popMu.Unlock()
				}()
				// Wait for the OS to paste before we put the next item on the clipboard.
				time.Sleep(50 * time.Millisecond)
				item, err := mgr.PopAndSync()
				if err != nil {
					if err.Error() != "queue is empty" {
						log.Printf("Error popping: %v", err)
					}
					return
				}
				log.Printf("Popped: %q", item)
			}()
		}
	}

	log.Println("CBQ monitor stopped.")
}
