package main

import (
	"context"
	"flag"
	"os"

	proxyserver "github.com/bhakiyakalimuthu/state-tracker/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Included in the build process
	_BuildVersion string
	_AppName      string
	//
	proxyTo = flag.String("PROXY_TO", "grpc.osmosis.zone:9090", "destination server address where traffic to be proxied")
)

func main() {
	l := newLogger(_AppName, _BuildVersion)
	ctx, cancel := context.WithCancel(context.Background())
	if err := proxyserver.RunServer(ctx, cancel, l, "9090", *proxyTo); err != nil {
		l.Fatal("failed to run proxy server", zap.Error(err))
	}
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
