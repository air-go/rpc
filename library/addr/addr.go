package addr

import (
	"net"
)

type WeightedAddr interface {
	net.Addr
	Priority() int
	Weight() int64
}
