package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gotube/internal/config"
	"gotube/utils"
)

var (
	cfg     config.Config
	outputFlag string
	forceFlag  bool
)

var rootCmd = &cobra.Command{
	Use:   "gotube",
	Short: "GoTube is a fast YouTube video and audio downloader",
	Long:  `A complete, production-ready CLI application written in Go that allows users to download YouTube videos and audio efficiently.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load config and set output directory
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		}

		if outputFlag != "" {
			cfg.OutputDir = outputFlag
		}

		// Ensure output directory exists
		if err := utils.EnsureDir(cfg.OutputDir); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "output directory (default is current directory or config)")
	rootCmd.PersistentFlags().BoolVarP(&forceFlag, "force", "f", false, "force overwrite existing files")
}
