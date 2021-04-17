package wordset

import (
	idictconfig "github.com/lai323/idict/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func Import(config *idictconfig.Config, fs afero.Fs, wordSetImport string, wordSetList bool, wordSetShow string) func(*cobra.Command, []string) error {

	return func(cmd *cobra.Command, args []string) error {
		if wordSetList {
			return nil
		}
		if wordSetShow != "" {
			return nil
		}
		if wordSetImport != "" {
			return nil
		}
		cmd.Help()
		return nil
	}
}
