package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gotube/internal/api"
)

var port string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the GoTube web interface",
	Run: func(cmd *cobra.Command, args []string) {
		server := api.NewServer(port)
		if err := server.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run the web server on")
}
