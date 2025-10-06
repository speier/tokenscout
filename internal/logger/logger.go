package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func Init(level string, pretty bool) {
	var output io.Writer = os.Stdout

	if pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	Logger = zerolog.New(output).
		Level(logLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	log.Logger = Logger
}

func Get() *zerolog.Logger {
	return &Logger
}

func Info() *zerolog.Event {
	return Logger.Info()
}

func Error() *zerolog.Event {
	return Logger.Error()
}

func Debug() *zerolog.Event {
	return Logger.Debug()
}

func Warn() *zerolog.Event {
	return Logger.Warn()
}

func Fatal() *zerolog.Event {
	return Logger.Fatal()
}
