package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jademcosta/graviola/pkg/app"
	"github.com/jademcosta/graviola/pkg/config"
)

var help = flag.Bool("help", false, "Show help")
var configPath = ""

func main() {

	flag.StringVar(&configPath, "config", "config", "--config=/path/to/config/file")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if configPath == "" {
		panic("config file path should be provided")
	}

	configContent, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Errorf("error reading config file: %w", err))
	}

	conf, err := config.Parse(configContent)
	if err != nil {
		panic(fmt.Errorf("error parsing config: %w", err))
	}

	conf = conf.FillDefaults()

	err = conf.IsValid()
	if err != nil {
		panic(fmt.Errorf("error validating config: %w", err))
	}

	app := app.NewApp(conf)
	app.Start()
}
