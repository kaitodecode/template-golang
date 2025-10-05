package cmd

import (
	"fmt"

	_ "template-golang/docs"
	"template-golang/pkg/config"
	utlog "template-golang/pkg/logger"

	"template-golang/internal"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Jalankan http server",
	Run: func(cmd *cobra.Command, args []string) {

		cfg := config.LoadConfig()
		utlog.Init(cfg.Env)

		app, err := internal.InitApp()
		if err != nil {
			panic(fmt.Errorf("failed to initialize app: %v", err))
		}

		app.Listen(":" + fmt.Sprintf("%d", cfg.Port))

	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
