package kvx

import "log/slog"

// LogDebug emits a debug log when debug logging is enabled.
func LogDebug(logger *slog.Logger, debug bool, msg string, attrs ...any) {
	if logger == nil || !debug {
		return
	}
	logger.Debug(msg, attrs...)
}

// LogError emits an error log when a logger is available.
func LogError(logger *slog.Logger, msg string, attrs ...any) {
	if logger == nil {
		return
	}
	logger.Error(msg, attrs...)
}
