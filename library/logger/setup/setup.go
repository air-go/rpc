package setup

import (
	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/logger/nop"
)

type SetupLogger struct {
	logger logger.Logger
}

func (l *SetupLogger) SetLogger(logger logger.Logger) {
	l.logger = logger
}

func (l *SetupLogger) Logger() logger.Logger {
	return l.logger
}

func (l *SetupLogger) AutoLogger() logger.Logger {
	if l.logger != nil {
		return l.logger
	}
	return nop.Logger
}
