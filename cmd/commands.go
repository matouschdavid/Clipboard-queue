package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/matouschdavid/Clipboard-queue/pkg/monitor"
	"github.com/matouschdavid/Clipboard-queue/pkg/queue"
	"github.com/matouschdavid/Clipboard-queue/pkg/storage"
)

var (
	isStack bool
	Version = "v0.1.0" // Default version, can be overridden by LDFLAGS
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

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the clipboard monitor",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting clipboard monitor...")
		monitor.Start()
	},
}

var PopCmd = &cobra.Command{
	Use:   "pop",
	Short: "Pop the next item to clipboard",
	Long: `Pops the next item from the queue and writes it to the system clipboard.
By default it uses the current mode (Queue or Stack), but you can override it with flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		mgr := getManager()
		item, err := mgr.Pop(isStack)
		if err != nil {
			if err.Error() == "queue is empty" {
				fmt.Println("Queue is empty")
				return
			}
			log.Fatal(err)
		}

		// Explicitly write the popped item to the clipboard for manual use.
		// Although Pop prepares the NEXT item, for the manual CLI we want the item we just popped.
		if err := mgr.Clipboard.Write(item); err != nil {
			log.Printf("Warning: failed to write to clipboard: %v", err)
		}

		fmt.Printf("Popped: %s\n", item)
	},
}

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current items in queue",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := getManager()
		state, err := mgr.GetStatus()
		if err != nil {
			log.Fatal(err)
		}

		if state.Active {
			fmt.Println("Status: ACTIVE (Collecting)")
		} else {
			fmt.Println("Status: INACTIVE")
		}

		if state.IsStack {
			fmt.Println("Mode: STACK (LIFO)")
		} else {
			fmt.Println("Mode: QUEUE (FIFO)")
		}

		if len(state.Items) == 0 {
			fmt.Println("Queue is empty")
			return
		}
		for i, item := range state.Items {
			fmt.Printf("%d: %s\n", i+1, item)
		}
	},
}

var ClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the queue",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := getManager()
		if err := mgr.Clear(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Queue cleared")
	},
}

var ModeCmd = &cobra.Command{
	Use:   "mode [stack|queue]",
	Short: "Set the mode to stack (LIFO) or queue (FIFO)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mgr := getManager()
		isStack := false
		if args[0] == "stack" {
			isStack = true
		} else if args[0] != "queue" {
			log.Fatalf("invalid mode: %s. Use 'stack' or 'queue'", args[0])
		}

		if err := mgr.SetStackMode(isStack); err != nil {
			log.Fatal(err)
		}

		// Update clipboard to match the new mode
		if err := mgr.SyncClipboard(); err != nil {
			log.Printf("Warning: failed to sync clipboard: %v", err)
		}

		if isStack {
			fmt.Println("Mode set to STACK (LIFO)")
		} else {
			fmt.Println("Mode set to QUEUE (FIFO)")
		}
	},
}

var RootCmd = &cobra.Command{
	Use:   "cbq",
	Short: "cbq is a clipboard queue/stack manager",
	Long: `A clipboard manager that works like a stack or queue.
Copy multiple times, then paste multiple times in order or reversed.

Use 'cbq mode [stack|queue]' to persistently change the behavior:
  - queue (FIFO): Pastes items in the same order they were copied.
  - stack (LIFO): Pastes items in reverse order.

Global Hotkeys (when 'cbq start' is running):
  Cmd+I: Activate and clear
  Cmd+C: Copy to queue
  Cmd+V: Paste from queue
  Cmd+R: Deactivate and clear`,
}

func init() {
	RootCmd.Version = Version
	RootCmd.AddCommand(StartCmd)
	RootCmd.AddCommand(PopCmd)
	RootCmd.AddCommand(StatusCmd)
	RootCmd.AddCommand(ClearCmd)
	RootCmd.AddCommand(ModeCmd)
	PopCmd.Flags().BoolVarP(&isStack, "stack", "s", false, "Pop in stack mode (LIFO)")
}
