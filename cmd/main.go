package cmd

import (
	"fmt"
	"os"

	"template-golang/pkg/config"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "myapp",
		Short: "MyApp CLI",
		Run: func(cmd *cobra.Command, args []string) {
			// load config dari flag
			config.LoadConfig()

			// contoh penggunaan
			cfg := config.GetConfig()
			fmt.Println("Server Port:", cfg.Port)
		},
	}
)

func init() {
	// --config default ke "."
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
