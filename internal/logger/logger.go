package logger

import (
	"context"
	"os"

	"github.com/AkMo3/simplify/internal/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const operationIDKey contextKey = "operation_id"

var globalLogger *zap.SugaredLogger

// Init initializes the global logger based on environment
func Init() error {
	var zapLogger *zap.Logger
	var err error

	if config.IsDevelopment() {
		zapLogger, err = newDevelopmentLogger()
	} else {
		zapLogger, err = newProductionLogger()
	}

	if err != nil {
		return err
	}

	globalLogger = zapLogger.Sugar()
	return nil
}

// newDevelopmentLogger creates a human-readable logger for development
func newDevelopmentLogger() (*zap.Logger, error) {
	encorderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encorderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)), nil
}

// newProductionLogger creates a JSON logger for production
func newProductionLogger() (*zap.Logger, error) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)), nil
}

// Sync flushes any buffered log entries
func Sync() {
	if globalLogger != nil {
		globalLogger.Sync()
	}
}

// WithOperationID creates a new context with an operation ID
func WithOperationID(ctx context.Context) context.Context {
	return context.WithValue(ctx, operationIDKey, uuid.New().String())
}

// WithCustomOperationID creates a new context with a custom operation ID
func WithCustomOperationID(ctx context.Context, opID string) context.Context {
	return context.WithValue(ctx, operationIDKey, opID)
}

// OperationIDFromContext extracts operation ID from context
func OperationIDFromContext(ctx context.Context) string {
	if opID, ok := ctx.Value(operationIDKey).(string); ok {
		return opID
	}
	return ""
}

// loggerWithContext returns a logger with operation ID if present in context
func loggerWithContext(ctx context.Context) *zap.SugaredLogger {
	if ctx == nil {
		return globalLogger
	}
	if opID := OperationIDFromContext(ctx); opID != "" {
		return globalLogger.With("operation_id", opID)
	}
	return globalLogger
}

// Debug logs a debug message (development only)
func Debug(msg string, keysAndValues ...interface{}) {
	globalLogger.Debugw(msg, keysAndValues...)
}

// DebugCtx logs a debug message with context
func DebugCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	loggerWithContext(ctx).Debugw(msg, keysAndValues...)
}

// Info logs an info message
func Info(msg string, keysAndValues ...interface{}) {
	globalLogger.Infow(msg, keysAndValues...)
}

// InfoCtx logs an info message with context
func InfoCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	loggerWithContext(ctx).Infow(msg, keysAndValues...)
}

// Warn logs a warning message
func Warn(msg string, keysAndValues ...interface{}) {
	globalLogger.Warnw(msg, keysAndValues...)
}

// WarnCtx logs a warning message with context
func WarnCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	loggerWithContext(ctx).Warnw(msg, keysAndValues...)
}

// Error logs an error message
func Error(msg string, keysAndValues ...interface{}) {
	globalLogger.Errorw(msg, keysAndValues...)
}

// ErrorCtx logs an error message with context
func ErrorCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	loggerWithContext(ctx).Errorw(msg, keysAndValues...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, keysAndValues ...interface{}) {
	globalLogger.Fatalw(msg, keysAndValues...)
}

// FatalCtx logs a fatal message with context and exits
func FatalCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	loggerWithContext(ctx).Fatalw(msg, keysAndValues...)
}
