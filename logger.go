// Package diagnostic provides diagnostic logging based on Go kit logger.
package diagnostic

import (
	"fmt"

	"github.com/go-kit/kit/log"
)

// FilterLogger wraps hit and miss loggers which are used depending
// whether key/value found in a log event.
type FilterLogger struct {
	Hit   log.Logger
	Miss  log.Logger
	Key   string
	Value string
}

// Log implements the Logger interface by forwarding keyvals to hit or miss wrapped loggers.
// It panics if the wrapped loggers are nil.
func (l *FilterLogger) Log(keyvals ...interface{}) error {
	for i := 0; i < len(keyvals); i += 2 {
		if k := fmt.Sprint(keyvals[i]); k == l.Key {
			if v := fmt.Sprint(keyvals[i+1]); v == l.Value {
				return l.Hit.Log(keyvals...)
			}
		}
	}
	return l.Miss.Log(keyvals...)
}
