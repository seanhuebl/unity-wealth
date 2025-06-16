package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() (*zap.Logger, error) {

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeDuration = func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendInt64(d.Milliseconds()) // now every zap.Duration field is in ms
	}

	encoder := zapcore.NewJSONEncoder(encoderCfg)

	infoLevel := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= zapcore.InfoLevel && l < zapcore.ErrorLevel
	})

	errorLevel := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= zapcore.ErrorLevel
	})

	stdoutSync := zapcore.AddSync(os.Stdout)
	stderrSync := zapcore.AddSync(os.Stderr)

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, stdoutSync, infoLevel),
		zapcore.NewCore(encoder, stderrSync, errorLevel),
	)
	logger := zap.New(core, zap.AddCaller())

	return logger, nil
}
