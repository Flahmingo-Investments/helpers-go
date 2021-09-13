package flog

import (
	"errors"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ErrNotInitialized is returned when structured logging is not enabled
var ErrNotInitialized = errors.New("flog: structure logging should be initialized first")

// Zap loggers
var (
	sugar  *zap.SugaredLogger
	logger *zap.Logger
)

// Debugf is called to write debug logs, such as logging request parameter to
// see what is coming inside.
var Debugf = log.Printf

// Verbosef is called to write verbose logs, such as when a new connection is
// established correctly.
var Verbosef = log.Printf

// Infof is called to write informational logs, such as when startup has
var Infof = log.Printf

// Errorf is called to write an error log, such as when a new connection fails.
var Errorf = log.Printf

// Fatalf is called to write an error log and then exit with non-zero status code.
// It cannot be disabled.
var Fatalf = log.Fatalf

// Infow is called to write informational logs, but as key value pairs.
var Infow = log.Printf

// LogDebugToStdout updates Verbosef and Info logging to use stdout instead of stderr.
func LogDebugToStdout() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	Verbosef = logger.Printf
	Infof = logger.Printf
	Debugf = logger.Printf
	Fatalf = logger.Fatalf
}

// noop is no op
func noop(string, ...interface{}) {
	// Enjoyable activities that produce flow have a potentially negative effect:
	// while they are capable of improving the quality of existence by creating
	// order in the mind, they can become addictive, at which point the self becomes
	// a captive of a certain kind of order, and is then unwilling to cope with the
	// ambiguities of life.
	//
	//	  - Mihaly Csikszentmihalyi
	//
	//
	// Thus, this function do nothing.
}

// LogVerboseToNowhere updates Verbosef so verbose log messages are discarded
func LogVerboseToNowhere() {
	Verbosef = noop
}

// DisableLogging sets all logging levels to no-op's.
func DisableLogging() {
	Debugf = noop
	Verbosef = noop
	Infof = noop
	Errorf = noop
}

// Config configures flog structured logging.
type Config struct {
	// LogDebugStdout logs to stdout instead of stderr
	LogDebugStdout bool

	// Verbose enables verbose logging.
	Verbose bool

	// Debug enables debug logging.
	Debug bool

	// Human enable human readable logging.
	// Good for development.
	Human bool
}

// InitializeSructuredLogs replaces all logging functions with structured logging
// variants.
func InitializeSructuredLogs(c *Config) (func(), error) {
	// Configuration of zap is based on its Advanced Configuration example.
	// See: https://pkg.go.dev/go.uber.org/zap#example-package-AdvancedConfiguration

	// Define level-handling logic.
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	// Lock wraps a WriteSyncer in a mutex to make it safe for concurrent use.
	// In particular, *os.File types must be locked before use.
	consoleErrors := zapcore.Lock(os.Stderr)
	consoleDebugging := consoleErrors
	if c.LogDebugStdout {
		consoleDebugging = zapcore.Lock(os.Stdout)
	}

	var config zapcore.EncoderConfig
	if !c.Human {
		config = zap.NewProductionEncoderConfig()
	} else {
		config = zap.NewDevelopmentEncoderConfig()
	}

	// GCP stackdriver requirements
	config.LevelKey = "severity"
	config.MessageKey = "message"
	config.TimeKey = "timestamp"

	if !c.Human {
		config.EncodeLevel = zapcore.CapitalLevelEncoder
	} else {
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	var consoleEncoder zapcore.Encoder

	if !c.Human {
		consoleEncoder = zapcore.NewJSONEncoder(config)
	} else {
		consoleEncoder = zapcore.NewConsoleEncoder(config)
	}

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
	)

	// By default, caller and stacktrace are not included, so add them here
	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	sugar = logger.Sugar()

	Verbosef = sugar.Infof
	if !c.Verbose {
		Verbosef = noop
	}

	Debugf = sugar.Debugf
	if !c.Debug {
		Debugf = noop
	}

	Infof = sugar.Infof
	Errorf = sugar.Errorf
	Fatalf = sugar.Fatalf
	Infow = sugar.Infow

	return func() {
		_ = logger.Sync()
	}, nil
}

// SugaredLogger returns the initialized sugared logger
func SugaredLogger() (*zap.SugaredLogger, error) {
	if sugar == nil {
		return nil, ErrNotInitialized
	}
	return sugar, nil
}

// Logger returns the initialized logger
func Logger() (*zap.Logger, error) {
	if logger == nil {
		return nil, ErrNotInitialized
	}
	return logger, nil
}
