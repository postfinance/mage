package log

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new zap logger.
//
// If MAGEFILE_DEBUG=1 the level is set to Debug, otherwise it is set to Info. If
// MAGEFILE_VERBOSE is not set only errors are logged..
//
// If there is an error in logger configuration, a noop logger and
// an error are returned.
func New() (*zap.Logger, error) {
	l := zap.New(nil) // noop logger
	atom := zap.NewAtomicLevelAt(zap.InfoLevel)
	config := zap.NewProductionConfig()
	config.DisableStacktrace = true
	config.DisableCaller = true
	config.Sampling = nil
	config.Encoding = "console"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	config.Level = atom
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	atom.SetLevel(zap.ErrorLevel)
	if mg.Verbose() {
		atom.SetLevel(zap.InfoLevel)
	}
	if mg.Debug() {
		atom.SetLevel(zap.DebugLevel)
		config.DisableCaller = false
	}
	var err error
	l, err = config.Build()
	if err != nil {
		return zap.New(nil), fmt.Errorf("could not create zap logger: %s", err)
	}
	defer l.Sync()
	return l, nil
}
