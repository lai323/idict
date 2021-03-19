package training

import (
	"errors"

	idictconfig "github.com/lai323/idict/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type Options struct {
	ShowSummary           *bool
	Proxy                 *string
	Sort                  *string
}

func Run(uiStartFn func() error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		// err := uiStartFn()

		// if err != nil {
		// 	fmt.Println(fmt.Errorf("Unable to start UI: %w", err).Error())
		// }
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
	// config.ShowSummary = idictconfig.GetBoolOption(*options.ShowSummary, config.ShowSummary)
	// config.Sort = idictconfig.GetStringOption(*options.Sort, config.Sort)
	return config
}
