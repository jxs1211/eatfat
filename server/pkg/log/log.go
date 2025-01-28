package log

import "log"

func NewLogger(prefix string) *log.Logger {
	return log.New(log.Writer(), prefix, log.LstdFlags|log.Lshortfile)
}
