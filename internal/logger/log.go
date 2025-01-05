package logger

import (
	"os"
	"path"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(filePath string, level *zapcore.Level) (*zap.Logger, error) {
	fileName := path.Join(filePath, time.Now().Format("2006-01-02_15:04:05")+".log")
	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	pe := zap.NewProductionEncoderConfig()
	pe.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000000")

	consoleEncoder := zapcore.NewConsoleEncoder(pe)
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(f), level),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
	)

	l := zap.New(core)

	return l, nil
}
