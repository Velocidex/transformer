package main

import (
	"github.com/Velocidex/transformer"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	watch_command = app.Command("watch", "Run a workflow periodically.")
	watch_period  = watch_command.Flag("period", "The number of seconds between polling.").
			Default("1").Int()

	workflow = watch_command.Flag("workflow", "The workflow to run.").
			Required().String()

	watch_config_path = watch_command.Flag("config", "The configuration file.").
				Short('c').Required().String()
)

func doWatch() {
	config, err := transformer.LoadConfig(*watch_config_path)
	kingpin.FatalIfError(err, "Config file")

	config.Verbose = *verbose

	err = transformer.NewTransformer(config).Watch(*watch_period, *workflow)
	kingpin.FatalIfError(err, "Watch")
}

func init() {
	command_handlers = append(command_handlers, func(command string) bool {
		switch command {

		case watch_command.FullCommand():
			doWatch()

		default:
			return false
		}

		return true
	})
}
