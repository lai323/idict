package dict

import (
	"errors"

	"github.com/lai323/idict/config"
	"github.com/spf13/cobra"
)

func Run(cfg *config.Config) func(*cobra.Command, []string) error {

	return func(cmd *cobra.Command, args []string) error {
		var text string
		if len(args) > 1 {
			return errors.New("Only one word or sentence can be translated at a time")
		}
		if len(args) == 1 {
			text = args[0]
		}
		if cfg.StoragePath == "" {
			return errors.New("StoragePath empty")
		}

		Start(cfg, text)
		return nil
	}
}
