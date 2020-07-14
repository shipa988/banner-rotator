package commands

import (
	"fmt"
	"github.com/jinzhu/gorm"
	app "github.com/shipa988/hw_otus/project/internal"
	"github.com/shipa988/hw_otus/project/internal/data/repository"
	"github.com/shipa988/hw_otus/project/internal/domain/usecase"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// slotCmd represents the slot command
var slotCmd = &cobra.Command{
	Use:   "slot",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := gorm.Open(cfg.DB.Dialect, cfg.DB.DSN)
		if err != nil {
			fmt.Print(err)
		}
		logger, _ := app.NewLogger(os.Stdout)

		db.SetLogger(logger)
		repo := repository.NewPGRepo(db)
		rotator, err := usecase.NewRotatorInteractor(repo)
		if err != nil {
			log.Fatal(err)
		}
		slot_id,err:=cmd.Flags().GetUint("id")
		if err != nil {
			log.Fatal(err)
		}
		slot_description,err:=cmd.Flags().GetString("description")
		if err != nil {
			log.Fatal(err)
		}
		page_url,err:=cmd.Flags().GetString("url")
		if err != nil {
			log.Fatal(err)
		}
		err=rotator.AddSlot(page_url,slot_id,slot_description)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	addCmd.AddCommand(slotCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// slotCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	slotCmd.Flags().UintP("id", "i", 0, "Help message for slotid")
	slotCmd.Flags().StringP("description", "d", "header", "Help message for slotdescription")
	slotCmd.Flags().StringP("url", "u", " ", "Help message for pageurl")
}
