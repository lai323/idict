package practice

import (
	"errors"
	"fmt"

	idictconfig "github.com/lai323/idict/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type Options struct {
	Proxy *string
}

func Run(config *idictconfig.Config, fs afero.Fs, options Options, uistarter func(string) error) func(*cobra.Command, []string) error {

	return func(cmd *cobra.Command, args []string) error {
		worset := "default"
		if len(args) > 1 {
			return errors.New("Only one wordset can to practice")
		}
		if len(args) == 1 {
			worset = args[0]
		}
		if config.StoragePath == "" {
			return errors.New("StoragePath empty")
		}
		*config = mergeConfig(*config, options)

		err := uistarter(worset)
		if err != nil {
			return fmt.Errorf("Unable to start UI: %w", err)
		}
		return nil
	}
}

func ValidateCli(config *idictconfig.Config, fs afero.Fs, options Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if config.StoragePath == "" {
			return errors.New("StoragePath empty")
		}
		*config = mergeConfig(*config, options)
		return nil
	}
}

func mergeConfig(config idictconfig.Config, options Options) idictconfig.Config {
	// config.Sort = idictconfig.GetStringOption(*options.Sort, config.Sort)
	return config
}
