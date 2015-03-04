package plugins

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// Log logs a plugin message prefixing the message with the plugin name
func Log(name string, level log.Level, args ...string) {
	msg := fmt.Sprintf("[%s] %s", name, strings.Join(args, " "))
	var logger func(string, ...interface{})
	switch level {
	case log.DebugLevel:
		logger = log.Debugf
	case log.InfoLevel:
		logger = log.Infof
	case log.WarnLevel:
		logger = log.Warnf
	case log.ErrorLevel:
		logger = log.Errorf
	case log.FatalLevel:
		logger = log.Fatalf
	case log.PanicLevel:
		logger = log.Panicf
	default:
		fmt.Printf(msg)
		return
	}

	logger(msg)
}
