package logger

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *Logger
	Log    *zap.SugaredLogger
)

func init() {
	var err error
	logger, err = CreateLogger(LoggerConfig{
		FuncPrefix: os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
		JsonOutput: true,
	})
	if err != nil {
		panic(fmt.Sprintf("Cannot create logger: %s", err.Error()))
	}
	Log = logger.Log
	Log.Debug("Logger initialized")
}

var FuncPrefix string // Optional prefix to trim from the function caller
var RequestId string  // Optional RequestId for the lambda

// Define strings for each log level
var logLevelSeverity = map[zapcore.Level]string{
	zapcore.DebugLevel: "DEBUG",
	zapcore.InfoLevel:  "INFO",
	zapcore.WarnLevel:  "WARNING",
	zapcore.ErrorLevel: "ERROR",
}

// Set of structured logging field keys
type cloudConfig struct {
	message    string
	level      string
	caller     string
	timestamp  string
	stacktrace string
}

// Default field keys to use for each cloud provider to meet structured logging requirements
var cloudConfigDefault = map[string]cloudConfig{
	"aws": {message: "msg", level: "level", caller: "caller", timestamp: "timestamp", stacktrace: "stacktrace"},
}

// Values related to a unique logger instance. Each logger instance
// will have a single Logger structure.
type Logger struct {
	Log         *zap.SugaredLogger // Sugared logger to invoke logs (can be modified to add fields)
	originalLog *zap.SugaredLogger // Original logger to invoke logs (unmodified)

	logToConsole bool            // Log only to console instead of file
	level        zap.AtomicLevel // Setting for current log level
}

// Configuration values for logger creation
type LoggerConfig struct {
	RequestId  string // RequestId for the lambda
	TruncReqId bool   // Truncate the RequestId to first octet
	FuncPrefix string // Optional: prefix to trim from the function caller
	JsonOutput bool   // Optional: output logs in JSON format
	Cloud      string // Optional: cloud provider to use for structured logging: Default: aws
}

// Create a unique logger instance based on a passed in configuration.
// The logger structure must be preserved to invoke configuration changes.
func CreateLogger(cfg LoggerConfig) (*Logger, error) {
	var err error

	cfg.Cloud = "aws" // Default to AWS for compatability
	FuncPrefix = cfg.FuncPrefix + string(os.PathSeparator)
	logger := Logger{level: zap.NewAtomicLevelAt(zap.InfoLevel), logToConsole: true}

	logger.originalLog, err = logger.newZapLogger(cfg)
	if err != nil {
		panic(err)
	}
	logger.Log = logger.originalLog // Preserve the original logger

	return &logger, nil
}

// Function to close the logger and flush its buffer
// This must be called instead of Sync() as the error
// must be checked (linter requirement)
func (l *Logger) CloseLogger() error {
	err := l.Log.Sync()
	return err
}

// Enable the requestId logger field
func SetRequestId(ctx context.Context, truncate bool) {
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		logger.Log.Warnf("Could not get lambda context to append requestId to logger")
		return
	}
	if truncate {
		// Truncate the RequestId to first octet only if console logging is disabled
		RequestId = strings.Split(lc.AwsRequestID, "-")[0]
	} else {
		RequestId = lc.AwsRequestID
	}
	if RequestId != "" {
		// Add to the logger
		logger.Log = logger.originalLog.With(zap.String("RequestId", RequestId))
	}
}

// Remove the current requestId from the logger
func RemoveRequestId() {
	RequestId = ""
}

// Set the logging level to include debug log entries
func EnableDebugLogs() {
	logger.level.SetLevel(zap.DebugLevel)
	if logger.Log != nil {
		logger.Log.Info("Debug logs enabled")
	}
}

// Set the logging level to exclude debug log entries
func DisableDebugLogs() {
	logger.level.SetLevel(zap.InfoLevel)
	if logger.Log != nil {
		logger.Log.Info("Debug logs disabled")
	}
}

// Encode the log level with our defined log level strings
func customEncodeLevel(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + logLevelSeverity[level] + "]")
	if RequestId != "" {
		// Add the lambda request id to the log if provided
		enc.AppendString("[" + RequestId + "]")
	}
}

// Encode the log level with our defined log level strings
func customEncodeLevelJson(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(logLevelSeverity[level])
}

// The default function encoder outputs the entire package name
// which is too long for our purpose. To modify this, we will
// add the value onto the caller encoder.
// Based on: zapcore.ShortCallerEncoder
func customCallerWithFunctionEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	trimFunction := caller.Function
	if strings.HasPrefix(trimFunction, "github.com") {
		// Trim out the `github.com` prefix
		trimFunction = strings.Join(strings.Split(caller.Function, "/")[3:], "/")
		if FuncPrefix != "" {
			// Trim the requested prefix
			trimFunction = strings.TrimPrefix(trimFunction, FuncPrefix)
		}
	}
	enc.AppendString("[" + caller.TrimmedPath() + " | " + trimFunction + "]")
}

// The default function encoder outputs the entire package name
// which is too long for our purpose. To modify this, we will
// add the value onto the caller encoder.
// Based on: zapcore.ShortCallerEncoder
func customCallerWithFunctionEncoderJson(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	trimFunction := caller.Function
	if strings.HasPrefix(trimFunction, "github.com") {
		// Trim out the `github.com` prefix
		trimFunction = strings.Join(strings.Split(caller.Function, "/")[3:], "/")
		if FuncPrefix != "" {
			// Trim the requested prefix
			trimFunction = strings.TrimPrefix(trimFunction, FuncPrefix)
		}
	}
	enc.AppendString(caller.TrimmedPath() + "/" + trimFunction)
}

// Format time to the RFC3339 standard manually to remove the timezone.
func rfc3339TimeAndPidEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05Z"))
}

// Create a new Zap sugared logger using the base
// logger preset. The sugared logger is less efficient
// but allows for printf style logging.
func (l *Logger) newZapLogger(cfg LoggerConfig) (*zap.SugaredLogger, error) {
	var err error

	encCfg := zapcore.EncoderConfig{
		TimeKey:        cloudConfigDefault[cfg.Cloud].timestamp,
		LevelKey:       cloudConfigDefault[cfg.Cloud].level,
		NameKey:        "logger",
		CallerKey:      cloudConfigDefault[cfg.Cloud].caller,
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     cloudConfigDefault[cfg.Cloud].message,
		StacktraceKey:  cloudConfigDefault[cfg.Cloud].stacktrace,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    customEncodeLevel,
		EncodeTime:     rfc3339TimeAndPidEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   customCallerWithFunctionEncoder,
	}

	// Define log levels
	allLogs := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= l.level.Level()
	})

	var encoder zapcore.Encoder
	encoder = zapcore.NewConsoleEncoder(encCfg)
	if cfg.JsonOutput {
		encCfg.EncodeLevel = customEncodeLevelJson
		encCfg.EncodeCaller = customCallerWithFunctionEncoderJson
		encoder = zapcore.NewJSONEncoder(encCfg)
	}

	// Only log to console
	consoleOutput := zapcore.Lock(os.Stdout)
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, consoleOutput, allLogs),
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	return logger.Sugar(), err
}
