package wordset

import (
	idictconfig "github.com/lai323/idict/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func Start(config *idictconfig.Config, fs afero.Fs, wordSetImport *string, wordSetList *bool, wordSetShow *string, wordDel *string) func(*cobra.Command, []string) error {

	return func(cmd *cobra.Command, args []string) error {
		if *wordSetList {
			return WordSetManage{StoragePath: config.StoragePath}.List()
		}
		if *wordSetShow != "" {
			return WordSetManage{StoragePath: config.StoragePath}.Show(*wordSetShow)
		}
		if *wordSetImport != "" {
			return WordSetManage{StoragePath: config.StoragePath}.Import(*wordSetImport)
		}
		if *wordDel != "" {
			return WordSetManage{StoragePath: config.StoragePath}.Del(*wordDel)
		}
		cmd.Help()
		return nil
	}
}
