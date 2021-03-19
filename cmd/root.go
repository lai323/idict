package cmd

import (
	"fmt"
	"log"
	"os"

	idictconfig "github.com/lai323/idict/config"
	"github.com/lai323/idict/dict"
	"github.com/lai323/idict/training"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	configPath            string
	storagePath           string
	config                idictconfig.Config
	proxy                 string
	showSummary           bool
	sort                  string
	rootCmd               = &cobra.Command{Use: "idict"}
	transCmd              = &cobra.Command{
		Use:   "trans",
		Short: "translate one word or sentence",
		RunE: dict.Run(
			&config,
			afero.NewOsFs(),
			dict.Options{
				ShowSummary:           &showSummary,
				Proxy:                 &proxy,
				Sort:                  &sort,
			},
			dict.Start(&config)),
	}

	trainingCmd = &cobra.Command{
		Use:   "training",
		Short: "training english",
		Args: training.ValidateCli(
			&config,
			afero.NewOsFs(),
			training.Options{
				ShowSummary:           &showSummary,
				Proxy:                 &proxy,
				Sort:                  &sort,
			},
		),
		Run: training.Run(training.Start(&config)),
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", fmt.Sprintf("config file (default is %s)", idictconfig.DefaultConfigPath))
	rootCmd.PersistentFlags().StringVar(&storagePath, "storage", "", fmt.Sprintf("storage dir (default is %s)", idictconfig.DefaultStorageDir))
	rootCmd.AddCommand(trainingCmd)
	rootCmd.AddCommand(transCmd)
}

func initConfig() {
	var err error
	config, err = idictconfig.InitConfig(afero.NewOsFs(), configPath)
	if err != nil {
		log.Fatal(err)
	}
}
