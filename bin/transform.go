package main

import (
	"os"
	"runtime/pprof"

	"github.com/Velocidex/transformer"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	transform_command = app.Command("transform", "Transform the repository.")
	config_path       = transform_command.Flag("config", "The configuration file.").
				Short('c').Required().String()

	profile_flag = app.Flag(
		"profile", "Write profiling information to this file.").String()

	verbose = app.Flag("verbose", "Turns on velbose logging.").
		Short('v').Bool()
)

func doTransform() {
	config, err := transformer.LoadConfig(*config_path)
	kingpin.FatalIfError(err, "Config file")

	config.Verbose = *verbose

	if *profile_flag != "" {
		f2, err := os.Create(*profile_flag)
		kingpin.FatalIfError(err, "Profile file.")

		err = pprof.StartCPUProfile(f2)
		kingpin.FatalIfError(err, "Profile file.")
		defer pprof.StopCPUProfile()
	}

	err = transformer.NewTransformer(config).Transform()
	kingpin.FatalIfError(err, "Transform")
}

func init() {
	command_handlers = append(command_handlers, func(command string) bool {
		switch command {

		case transform_command.FullCommand():
			doTransform()

		default:
			return false
		}

		return true
	})
}
