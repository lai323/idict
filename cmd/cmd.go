package cmd

import (
	"fmt"
	"log"
	"os"
	"syscall"

	idictconfig "github.com/lai323/idict/config"
	"github.com/lai323/idict/dict"
	"github.com/lai323/idict/practice"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	configPath  string
	storagePath string
	UnlockDb    bool
	pracOpt     practice.Options

	config  idictconfig.Config
	rootCmd = &cobra.Command{
		Use: "idict",
		RunE: func(cmd *cobra.Command, args []string) error {
			if UnlockDb {
				return unlockdb()
			}
			cmd.Help()
			return nil
		},
	}
	transCmd = &cobra.Command{
		Use:   "trans",
		Short: "translate one word or sentence",
		RunE:  dict.Run(&config),
	}

	practiceCmd = &cobra.Command{
		Use:   "prac",
		Short: "word practice",
		RunE:  practice.Run(&config, &pracOpt),
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
	rootCmd.PersistentFlags().BoolVar(&UnlockDb, "unlockdb", false, "unlock db")

	practiceCmd.PersistentFlags().StringVar(&pracOpt.Import, "import", "", "import collection file path")
	practiceCmd.PersistentFlags().BoolVar(&pracOpt.List, "list", false, "list collection")
	practiceCmd.PersistentFlags().StringVar(&pracOpt.Show, "show", "", "show collection")
	practiceCmd.PersistentFlags().StringVar(&pracOpt.Delete, "del", "", "delete word collection")
	practiceCmd.PersistentFlags().BoolVar(&pracOpt.Degree, "degree", false, "practice degree")
	practiceCmd.PersistentFlags().BoolVar(&pracOpt.Clean, "clean", false, "practice clean")
	practiceCmd.PersistentFlags().BoolVar(&pracOpt.Shuffle, "shuffle", false, "shuffle")

	rootCmd.AddCommand(transCmd)
	rootCmd.AddCommand(practiceCmd)
}

func initConfig() {
	var err error
	config, err = idictconfig.InitConfig(afero.NewOsFs(), configPath)
	if err != nil {
		log.Fatal(err)
	}
}

func unlockdb() error {
	file, err := os.Open(config.DbFile())
	if err != nil {
		return err
	}
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	if err != nil {
		return err
	}
	fmt.Printf("unlock %s\n", config.DbFile())
	return nil
}
