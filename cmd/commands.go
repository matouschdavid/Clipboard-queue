package cmd

import (
	"fmt"
	"log"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"github.com/vibe-coding/cbq/pkg/monitor"
	"github.com/vibe-coding/cbq/pkg/storage"
)

var (
	isStack bool
)

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
		items, err := storage.Load()
		if err != nil {
			log.Fatal(err)
		}

		if len(items) == 0 {
			fmt.Println("Queue is empty")
			return
		}

		var item string
		if isStack {
			// LIFO (Last In First Out)
			item = items[len(items)-1]
			items = items[:len(items)-1]
		} else {
			// FIFO (First In First Out)
			item = items[0]
			items = items[1:]
		}

		err = clipboard.WriteAll(item)
		if err != nil {
			log.Fatal(err)
		}
		storage.Save(items)

		fmt.Printf("Popped: %s\n", item)
	},
}

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current items in queue",
	Run: func(cmd *cobra.Command, args []string) {
		items, _ := storage.Load()
		if len(items) == 0 {
			fmt.Println("Queue is empty")
			return
		}
		for i, item := range items {
			fmt.Printf("%d: %s\n", i+1, item)
		}
	},
}

var ClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the queue",
	Run: func(cmd *cobra.Command, args []string) {
		storage.Clear()
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
