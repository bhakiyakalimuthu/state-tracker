package main

import (
	"context"
	"flag"
	"os"

	"github.com/bhakiyakalimuthu/state-tracker/client"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Included in the build process
	_BuildVersion string
	_AppName      string
	//
	serverAddrs = flag.String("SERVER_ADDRESS", "localhost:9090", "destination server address where traffic to be proxied")
)

func main() {
	l := newLogger(_AppName, _BuildVersion)
	ctx, cancel := context.WithCancel(context.Background())
	client.RunClient(ctx, cancel, l, *serverAddrs)
}

func newLogger(appName, version string) *zap.Logger {
	logLevel := zap.DebugLevel
	var zapCore zapcore.Core
	level := zap.NewAtomicLevel()
	level.SetLevel(logLevel)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoder := zapcore.NewJSONEncoder(encoderCfg)
	zapCore = zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), level)

	logger := zap.New(zapCore, zap.AddCaller(), zap.ErrorOutput(zapcore.Lock(os.Stderr)))
	logger = logger.With(zap.String("app", appName), zap.String("version", version))
	return logger
}
