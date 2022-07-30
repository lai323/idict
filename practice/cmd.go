package practice

import (
	"errors"

	"github.com/lai323/idict/config"
	"github.com/lai323/idict/db"
	"github.com/spf13/cobra"
)

type Options struct {
	Import  string
	List    bool
	Show    string
	Delete  string
	Degree  bool
	Clean   bool
	Shuffle bool
}

func Run(cfg *config.Config, options *Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if cfg.StoragePath == "" {
			return errors.New("StoragePath empty")
		}
		*cfg = mergeConfig(*cfg, *options)
		dictdb, err := db.NewBoltDictDB(cfg.DbFile())
		if err != nil {
			return err
		}
		defer dictdb.Close()

		if options.List {
			return db.CollectionList(dictdb)
		}
		if options.Show != "" {
			return db.CollectionWords(options.Show, dictdb)
		}
		if options.Import != "" {
			return db.CollectionImport(options.Import, "", dictdb)
		}
		if options.Delete != "" {
			return db.CollectionDelete(options.Delete, dictdb)
		}
		if options.Degree {
			return db.PracticeDegree(dictdb)
		}
		if options.Clean {
			return dictdb.PracticeClean()
		}

		worset := "default"
		if len(args) > 1 {
			return errors.New("Only one wordset can to practice")
		}
		if len(args) == 1 {
			worset = args[0]
		}
		Start(cfg, worset, dictdb, options.Shuffle)
		return nil
	}
}

func ValidateCli(cfg *config.Config, options Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if cfg.StoragePath == "" {
			return errors.New("StoragePath empty")
		}
		*cfg = mergeConfig(*cfg, options)
		return nil
	}
}

func mergeConfig(cfg config.Config, options Options) config.Config {
	// cfg.Sort = idictconfig.GetStringOption(*options.Sort, cfg.Sort)
	return cfg
}
