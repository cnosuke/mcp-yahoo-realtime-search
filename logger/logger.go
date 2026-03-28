package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger initializes the global logger.
// In stdio mode (suppressConsole=true), console output is suppressed
// to prevent interference with MCP JSON-RPC protocol communication.
func InitLogger(logLevel string, logPath string, suppressConsole bool) error {
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", logLevel, err)
	}

	var encoderConfig zapcore.EncoderConfig
	var encoder zapcore.Encoder
	if level == zapcore.DebugLevel {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	var cores []zapcore.Core

	if !suppressConsole {
		stdout := zapcore.Lock(os.Stdout)
		stderr := zapcore.Lock(os.Stderr)

		infoAndBelow := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
			return l >= level && l < zapcore.WarnLevel
		})
		warnAndAbove := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
			return l >= level && l >= zapcore.WarnLevel
		})

		cores = append(cores,
			zapcore.NewCore(encoder, stdout, infoAndBelow),
			zapcore.NewCore(encoder, stderr, warnAndAbove),
		)
	}

	if logPath != "" {
		if suppressConsole && (logPath == "stdout" || logPath == "stderr") {
			return fmt.Errorf("log path %q conflicts with stdio mode", logPath)
		}
		fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		fileSink, _, err := zap.Open(logPath)
		if err != nil {
			return fmt.Errorf("failed to open log file %q: %w", logPath, err)
		}
		cores = append(cores, zapcore.NewCore(fileEncoder, fileSink, level))
	}

	var core zapcore.Core
	if len(cores) == 0 {
		core = zapcore.NewNopCore()
	} else {
		core = zapcore.NewTee(cores...)
	}

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	zap.ReplaceGlobals(logger)
	zap.S().Infow("Logger initialized",
		"log_level", logLevel,
		"log_path", logPath,
		"suppress_console", suppressConsole)

	return nil
}

// Sync flushes any buffered log entries
func Sync() error {
	return zap.L().Sync()
}
