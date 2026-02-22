package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/vibe-coding/cbq/pkg/monitor"
	"github.com/vibe-coding/cbq/pkg/queue"
	"github.com/vibe-coding/cbq/pkg/storage"
)

var (
	isStack bool
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

var RootCmd = &cobra.Command{
	Use:   "cbq",
	Short: "cbq is a clipboard queue/stack manager",
	Long: `A clipboard manager that works like a stack or queue.
          Copy multiple times, then paste multiple times in order or reversed.`,
}

func init() {
	RootCmd.AddCommand(StartCmd)
	RootCmd.AddCommand(PopCmd)
	RootCmd.AddCommand(StatusCmd)
	RootCmd.AddCommand(ClearCmd)
	PopCmd.Flags().BoolVarP(&isStack, "stack", "s", false, "Pop in stack mode (LIFO)")
}
