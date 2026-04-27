package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gotube/internal/youtube"
)

var audioCmd = &cobra.Command{
	Use:   "audio <url>",
	Short: "Download audio only (MP3)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		
		svc := youtube.NewService()
		fmt.Println("Fetching video information...")
		
		err := svc.DownloadAudio(url, cfg.OutputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(audioCmd)
}
