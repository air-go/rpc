package zap

import "github.com/air-go/rpc/library/logger"

var StdLogger logger.Logger

func init() {
	StdLogger, _ = NewLogger()
}
