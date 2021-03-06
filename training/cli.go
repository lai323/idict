package training

import (
	// "errors"
	"errors"
	"fmt"
	// "strings"

	idictconfig "github.com/lai323/idict/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type Options struct {
	RefreshInterval       *int
	Interactive           *bool
	ExtraInfoExchange     *bool
	ExtraInfoFundamentals *bool
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
		fmt.Println("Run", args)
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
	// config.RefreshInterval = idictconfig.GetRefreshInterval(*options.RefreshInterval, config.RefreshInterval)
	// config.ExtraInfoExchange = idictconfig.GetBoolOption(*options.ExtraInfoExchange, config.ExtraInfoExchange)
	// config.ExtraInfoFundamentals = idictconfig.GetBoolOption(*options.ExtraInfoFundamentals, config.ExtraInfoFundamentals)
	// config.ShowSummary = idictconfig.GetBoolOption(*options.ShowSummary, config.ShowSummary)
	// config.Sort = idictconfig.GetStringOption(*options.Sort, config.Sort)
	return config
}
