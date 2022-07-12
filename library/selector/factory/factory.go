package factory

import (
	"github.com/air-go/rpc/library/selector"
	"github.com/air-go/rpc/library/selector/wr"
)

func New(serviceName, t string) selector.Selector {
	switch t {
	case selector.TypeWR:
		return wr.NewSelector(serviceName)
	}
	return wr.NewSelector(serviceName)
}
