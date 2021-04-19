package cmd

import (
	"fmt"
	"log"
	"os"

	idictconfig "github.com/lai323/idict/config"
	"github.com/lai323/idict/dict"
	"github.com/lai323/idict/practice"
	"github.com/lai323/idict/wordset"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	configPath    string
	storagePath   string
	wordSetImport string
	wordSetList   bool
	wordSetShow   string
	wordDel       string

	config   idictconfig.Config
	rootCmd  = &cobra.Command{Use: "idict"}
	transCmd = &cobra.Command{
		Use:   "trans",
		Short: "translate one word or sentence",
		RunE: dict.Run(
			&config,
			afero.NewOsFs(),
			dict.Options{},
			dict.Start(&config)),
	}

	practiceCmd = &cobra.Command{
		Use:   "prac",
		Short: "Word practice",
		RunE: practice.Run(
			&config,
			afero.NewOsFs(),
			practice.Options{},
			dict.Start(&config)),
	}

	wordCmd = &cobra.Command{
		Use:   "word",
		Short: "manage word set",
		RunE: wordset.Start(
			&config,
			afero.NewOsFs(),
			&wordSetImport,
			&wordSetList,
			&wordSetShow,
			&wordDel,
		),
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

	wordCmd.PersistentFlags().StringVar(&wordSetImport, "import", "", "import word set file path")
	wordCmd.PersistentFlags().BoolVar(&wordSetList, "list", false, "list all word set")
	wordCmd.PersistentFlags().StringVar(&wordSetShow, "show", "", "show word set info")
	wordCmd.PersistentFlags().StringVar(&wordDel, "del", "", "delete word set")

	rootCmd.AddCommand(transCmd)
	rootCmd.AddCommand(wordCmd)
	rootCmd.AddCommand(practiceCmd)
}

func initConfig() {
	var err error
	config, err = idictconfig.InitConfig(afero.NewOsFs(), configPath)
	if err != nil {
		log.Fatal(err)
	}
}
