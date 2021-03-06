package dict

import (
	"errors"
	"fmt"
	// "strings"

	idictconfig "github.com/lai323/idict/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type Options struct {
	RefreshInterval       *int
	Text                  *string
	Interactive           *bool
	ExtraInfoExchange     *bool
	ExtraInfoFundamentals *bool
	ShowSummary           *bool
	Proxy                 *string
	Sort                  *string
}

func Run(config *idictconfig.Config, fs afero.Fs, options Options, uistarter func(string) error) func(*cobra.Command, []string) error {

	return func(cmd *cobra.Command, args []string) error {
		fmt.Println("Run", args)
		var text string
		// 验证 options, config, args; 合并 config
		if len(args) > 1 {
			return errors.New("Only one word or sentence can be translated at a time")
		}
		if len(args) == 1 {
			text = args[0]
		}
		if config.StoragePath == "" {
			return errors.New("StoragePath empty")
		}
		*config = mergeConfig(*config, options)

		err := uistarter(text)
		if err != nil {
			return fmt.Errorf("Unable to start UI: %w", err)
		}
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
