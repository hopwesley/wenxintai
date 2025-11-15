package comm

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

var (
	_logInstance *zerolog.Logger
	logOnce      sync.Once
)

func LogInst() *zerolog.Logger {
	logOnce.Do(func() {

		writer := diode.NewWriter(os.Stderr, 1000, 10*time.Millisecond, func(missed int) {
			fmt.Printf("Logger Dropped %d messages", missed)
		})
		out := zerolog.ConsoleWriter{Out: writer}
		out.TimeFormat = time.StampMilli

		logger := zerolog.New(out).
			Level(zerolog.DebugLevel).
			With().
			Caller().
			Timestamp().
			Logger()
		_logInstance = &logger
	})

	return _logInstance
}

/*
"trace" "debug" "info" "warn" "error" "fatal" "panic"
*/

func SetLogLevel(lStr string) {
	if len(lStr) == 0 {
		lStr = "debug"
	}
	logLvl, err := zerolog.ParseLevel(lStr)
	if err != nil {
		logLvl = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLvl)
}
