package log

import (
	"context"
	nLog "log"
)

// Debugf formats its arguments according to the format, analogous to fmt.Printf,
// and records the text as a log message at Debug level. The message will be associated
// with the request linked with the provided context.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	nLog.Printf(format, args...)
}

// Infof is like Debugf, but at Info level.
func Infof(ctx context.Context, format string, args ...interface{}) {
	nLog.Printf(format, args...)
}

// Warningf is like Debugf, but at Warning level.
func Warningf(ctx context.Context, format string, args ...interface{}) {
	nLog.Printf(format, args...)
}

// Errorf is like Debugf, but at Error level.
func Errorf(ctx context.Context, format string, args ...interface{}) {
	nLog.Printf(format, args...)
}

// Criticalf is like Debugf, but at Critical level.
func Criticalf(ctx context.Context, format string, args ...interface{}) {
	nLog.Printf(format, args...)
}
