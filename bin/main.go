package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

type CommandHandler func(command string) bool

var (
	app = kingpin.New("transformer",
		"Transforms a repository from one form to another.")

	command_handlers []CommandHandler
)

func main() {
	app.HelpFlag.Short('h')
	app.UsageTemplate(kingpin.CompactUsageTemplate)
	args := os.Args[1:]

	command := kingpin.MustParse(app.Parse(args))

	for _, command_handler := range command_handlers {
		if command_handler(command) {
			break
		}
	}
}
