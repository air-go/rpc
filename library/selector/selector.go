package selector

import (
	"github.com/air-go/rpc/library/servicer"
)

const (
	TypeWR   = "wr"
	TypeWrr  = "wrr"
	TypeDwrr = "dwrr"
	TypeP2C  = "p2c"
	TypeICMP = "icmp"
)

type HandleInfo struct {
	Node servicer.Node
	Err  error
}

type Selector interface {
	ServiceName() string
	AddNode(node servicer.Node) (err error)
	DeleteNode(node servicer.Node) (err error)
	GetNodes() (nodes []servicer.Node, err error)
	Select() (node servicer.Node, err error)
	AfterHandle(info HandleInfo)
}
