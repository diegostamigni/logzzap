package main

import (
	"errors"
	"os"
	"time"

	"github.com/diegostamigni/logzzap"
	"github.com/logzio/logzio-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// create a new Zap logger
	logger, _ := zap.NewProduction()

	// Initialize logz with your token and optional environment flag
	sender, err := logzio.New(
		"<TOKEN>",
		logzio.SetDebug(os.Stderr),
		logzio.SetUrl("https://listener.logz.io:8071"),
		logzio.SetDrainDuration(time.Second*5),
		logzio.SetTempDirectory("myQueue"),
		logzio.SetDrainDiskThreshold(99),
	)
	if err != nil {
		panic(err)
	}
	defer sender.Stop()

	// create a new core that sends zapcore.ErrorLevel and above messages to Logz
	logzCore := logzzap.NewLogzCore(sender, zapcore.InfoLevel)

	// Wrap a NewTee to send log messages to both your main logger and to logz
	logger = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, logzCore)
	}))

	// This message will only go to the main logger
	logger.Info("Logz Core teed up", zap.String("foo", "bar"))

	// This message will only go to the main logger
	logger.Warn("Warning message with fields", zap.String("foo", "bar"))

	// This error will go to both the main logger and to Logz. the 'foo' field will appear in logz as 'custom.foo'
	testError := errors.New("im a test error")
	logger.Error("ran into an error", zap.Error(testError), zap.String("foo", "bar"), zap.Int("some-int", 10))
}
