package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gotube/internal/youtube"
)

var listCmd = &cobra.Command{
	Use:   "list <url>",
	Short: "List available formats for a video",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		
		svc := youtube.NewService()
		fmt.Println("Fetching video information...")
		
		err := svc.PrintFormats(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
