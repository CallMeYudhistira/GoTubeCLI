package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gotube/internal/youtube"
)

var quality string

var downloadCmd = &cobra.Command{
	Use:   "download <url>",
	Short: "Download a YouTube video",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		
		svc := youtube.NewService()
		fmt.Println("Fetching video information...")
		
		err := svc.DownloadVideo(url, cfg.OutputDir, quality, "", forceFlag, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringVarP(&quality, "quality", "q", "", "specify video quality (e.g., 720, 1080p). Defaults to best available.")
}
