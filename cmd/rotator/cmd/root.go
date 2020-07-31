package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shipa988/banner_rotator/cmd/rotator/internal/data/app"
)

var cfgFile string
var debug bool
var cfg *app.Config

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "banner-rotator",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(commands *cobra.Command, args []string) { },
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", `config\config.yaml`, "config file (default is $./config/config.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "set if you want run app in debug mode")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	cfg = &app.Config{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		log.Fatal(err)
	}

	//how can i substitute config param in windows? evsubst is absent =(
	if dsn, ok := viper.Get("DSN").(string); dsn != "" && ok {
		cfg.DB.DSN = dsn
	}
	if topic, ok := viper.Get("KAFKA_TOPIC").(string); topic != "" && ok {
		cfg.Kafka.Topic = topic
	}
	if addr, ok := viper.Get("KAFKA_ADDR").(string); addr != "" && ok {
		cfg.Kafka.Addr = addr
	}
	if port, ok := viper.Get("GRPC_PORT").(string); port != "" && ok {
		cfg.API.GRPCPort = port
	}
	if gwport, ok := viper.Get("GRPC_GW_PORT").(string); gwport != "" && ok {
		cfg.API.GRPCGWPort = gwport
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
