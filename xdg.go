package main

import (
	"log"

	"github.com/adrg/xdg"
)

func Xdg() {
	// XDG Base Directory paths.
	log.Println("Home data directory:", xdg.DataHome)
	log.Println("Data directories:", xdg.DataDirs)
	log.Println("Home config directory:", xdg.ConfigHome)
	log.Println("Config directories:", xdg.ConfigDirs)
	log.Println("Cache directory:", xdg.CacheHome)
	log.Println("Runtime directory:", xdg.RuntimeDir)

	// Non-standard directories.
	log.Println("Home state directory:", xdg.StateHome)
	log.Println("Application directories:", xdg.ApplicationDirs)
	log.Println("Font directories:", xdg.FontDirs)

	// Obtain a suitable location for application config files.
	// ConfigFile takes one parameter which must contain the name of the file,
	// but it can also contain a set of parent directories. If the directories
	// don't exist, they will be created relative to the base config directory.
	configFilePath, err := xdg.ConfigFile("appname/config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Save the config file at:", configFilePath)

	// For other types of application files use:
	// xdg.DataFile()
	// xdg.CacheFile()
	// xdg.RuntimeFile()
	// xdg.StateFile()

	// Finding application config files.
	// SearchConfigFile takes one parameter which must contain the name of
	// the file, but it can also contain a set of parent directories relative
	// to the config search paths (xdg.ConfigHome and xdg.ConfigDirs).
	configFilePath, err = xdg.SearchConfigFile("appname/config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Config file was found at:", configFilePath)

	// For other types of application files use:
	// xdg.SearchDataFile()
	// xdg.SearchCacheFile()
	// xdg.SearchRuntimeFile()
	// xdg.SearchStateFile()
}
