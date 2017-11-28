package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

const estConfigDefaultContents string = `
# Your estfile stores your tasks and estimates. Some users may want to change this to a location with automatic backup, such as Dropbox or Google Drive.
estfile = "$HOME/.estfile.toml"
`

const estConfigDefaultFileNameNoSuffix string = ".estconfig"
const estConfigDefaultFileSuffix string = ".toml"
const estConfigDefaultFileName string = estConfigDefaultFileNameNoSuffix + estConfigDefaultFileSuffix

type estConfig struct {
	Estfile string // est file name
}

// getEstconfig returns the singleton estConfig for this process.
// Creates a config file if none found.
func getEstConfig() (estConfig, error) {
	if err := createFileWithDefaultContentsIfNotExists(os.Getenv("HOME")+"/"+estConfigDefaultFileName, estConfigDefaultContents); err != nil {
		return estConfig{}, fmt.Errorf("couldn't find or create %s: %s", estConfigDefaultFileName, err)
	}

	viper.SetConfigName(estConfigDefaultFileNameNoSuffix) // .toml suffix discovered automatically
	viper.AddConfigPath("$HOME")
	if err := viper.ReadInConfig(); err != nil {
		return estConfig{}, err
	}

	c := estConfig{}
	err := viper.Unmarshal(&c)
	return c, err
}
