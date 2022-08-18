package gerrors

// LogLevel is used while creating new error with NewWithLogLevel method.
// If the formatter is configured to have a logger and it's logger implements
// the proper log level interface, error will be logged at the provided level.
type LogLevel int

const (
	// LogLevelOff will ignore logging the error event if formatter's logger is set.
	LogLevelOff LogLevel = iota - 1
	LogLevelError
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
	LogLevelTrace
)

type logger interface {
	Error(err error, msg string, keysAndValues ...any)
}

type warnLogger interface {
	Warn(msg string, keysAndValues ...any)
}

type infoLogger interface {
	Info(msg string, keysAndValues ...any)
}

type debugLogger interface {
	Debug(msg string, keysAndValues ...any)
}

type traceLogger interface {
	Trace(msg string, keysAndValues ...any)
}
