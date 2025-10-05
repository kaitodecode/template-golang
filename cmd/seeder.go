package cmd

import (
	"fmt"

	_ "template-golang/docs"
	"template-golang/internal/db"
	"template-golang/internal/seeders"
	"template-golang/pkg/config"
	utlog "template-golang/pkg/logger"

	"github.com/spf13/cobra"
)

var seederCmd = &cobra.Command{
	Use:   "seeder",
	Short: "Jalankan seeder",
	Run: func(cmd *cobra.Command, args []string) {

		cfg := config.LoadConfig()
		utlog.Init(cfg.Env)

		db, err := db.ConnectDB()

		if err != nil {
			panic(fmt.Errorf("failed to connect db: %v", err))
		}

		if err := seeders.Seed(db); err != nil {
			panic(fmt.Errorf("failed to seed db: %v", err))
		}

	},
}

func init() {
	rootCmd.AddCommand(seederCmd)
}
