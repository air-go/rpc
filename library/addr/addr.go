package addr

import (
	"net"
	"strconv"
)

type WeightedAddr interface {
	net.Addr
	Priority() int
	Weight() int64
}

type TCPAddr struct {
	IP   net.IP
	Port int
}

func (a *TCPAddr) Network() string {
	return "tcp"
}

func (a *TCPAddr) String() string {
	return net.JoinHostPort(a.IP.String(), strconv.Itoa(a.Port))
}
