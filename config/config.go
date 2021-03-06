package config

import (
	"fmt"
	"log"
	"path"

	// "github.com/mitchellh/go-homedir"
	"github.com/adrg/xdg"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

type Config struct {
	StoragePath string `yaml:"StoragePath"`
}

var (
	DefaultConfig     Config
	DefaultConfigDir  string
	DefaultConfigPath string
	DefaultStorageDir string
)

func init() {
	var err error
	DefaultConfigPath, err = xdg.ConfigFile("idict/idict.yaml")
	if err != nil {
		log.Fatal(err)
	}
	DefaultConfigDir = path.Dir(DefaultConfigPath)
	DefaultStorageDir = path.Join(xdg.DataHome, "idict")
	DefaultConfig = Config{
		StoragePath: DefaultStorageDir,
	}
	err = createDefaultFile(afero.NewOsFs())
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(DefaultStorageDir, DefaultConfigPath)
}

type initConfigErr struct {
	s string
}

func (e *initConfigErr) Error() string {
	return e.s
}

func newInitConfigErr(err error) error {
	return &initConfigErr{
		s: fmt.Sprintf("Init config error: %s", err.Error()),
	}
}

func createDefaultFile(fs afero.Fs) error {
	err := fs.MkdirAll(DefaultConfigDir, 0755)
	if err != nil {
		return err
	}
	fs.MkdirAll(DefaultStorageDir, 0755)
	if err != nil {
		return err
	}

	exist, err := afero.Exists(fs, DefaultConfigPath)
	if err != nil {
		return err
	}

	if !exist {
		handle, err := fs.Create(DefaultConfigPath)
		if err != nil {
			return err
		}
		defer handle.Close()
		err = yaml.NewEncoder(handle).Encode(&DefaultConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitConfig(fs afero.Fs, configPathOption string) (Config, error) {
	var config Config
	var configfile string

	if configPathOption == "" {
		configfile = DefaultConfigPath
	} else {
		exist, err := afero.Exists(fs, configPathOption)
		if err != nil {
			return config, newInitConfigErr(err)
		}
		if !exist {
			return config, &initConfigErr{
				s: fmt.Sprintf("Init config error: %s not exist", configPathOption),
			}
		}
		configfile = configPathOption
	}

	handle, err := fs.Open(configfile)
	if err != nil {
		return config, newInitConfigErr(err)
	}
	defer handle.Close()
	err = yaml.NewDecoder(handle).Decode(&config)
	if err != nil {
		return config, newInitConfigErr(err)
	}
	return config, nil
}
