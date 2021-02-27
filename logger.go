package transformer

import "fmt"

type Logger struct {
	config *Config
}

func (self *Logger) Printf(format string, args ...interface{}) {
	if self.config.Verbose {
		fmt.Printf(format, args...)
	}
}

func NewLogger(config *Config) *Logger {
	return &Logger{config}
}
