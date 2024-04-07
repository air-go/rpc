package dns

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/why444216978/go-util/assert"
	ucontext "github.com/why444216978/go-util/context"
	"github.com/why444216978/go-util/nopanic"

	"github.com/air-go/rpc/library/addr"
	"github.com/air-go/rpc/library/discoverer"
	"github.com/air-go/rpc/library/loadbalancer"
	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/logger/setup"
)

type Node struct {
	Host    string
	Port    int
	Network string // ip/ipv4/ipv6
}

type options struct {
	logger        logger.Logger
	refreshWindow time.Duration
}

func defaultOptions() *options {
	return &options{
		refreshWindow: 3 * time.Second,
	}
}

type dnsDiscoverer struct {
	*options
	setup.SetupLogger
	serviceName string
	nodes       []Node // current idc config
	lb          loadbalancer.LoadBalancer
	startOnce   sync.Once
	stop        context.CancelFunc
}

var _ discoverer.Discoverer = (*dnsDiscoverer)(nil)

type optionFunc func(*options)

func WithLogger(l logger.Logger) optionFunc {
	return func(o *options) { o.logger = l }
}

func WithRefreshWindow(t time.Duration) optionFunc {
	return func(o *options) { o.refreshWindow = t }
}

func NewDNSDiscoverer(serviceName string, lb loadbalancer.LoadBalancer, nodes []Node, opts ...optionFunc) (*dnsDiscoverer, error) {
	if assert.IsNil(lb) {
		return nil, errors.New("new dns discoverer loadbalancer nil")
	}

	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	dd := &dnsDiscoverer{
		options:     opt,
		serviceName: serviceName,
		nodes:       nodes,
		lb:          lb,
	}

	dd.SetupLogger.SetLogger(opt.logger)

	return dd, nil
}

func (dd *dnsDiscoverer) Start(ctx context.Context) (err error) {
	dd.startOnce.Do(func() {
		ctx, dd.stop = context.WithCancel(ucontext.RemoveDeadline(ctx))

		if err = dd.discover(ctx, false); err != nil {
			return
		}

		dd.loop(ctx)
	})

	return
}

func (dd *dnsDiscoverer) Stop() error {
	dd.stop()
	return nil
}

// discover force get newest and notify loadbalancer to update nodes.
func (dd *dnsDiscoverer) discover(ctx context.Context, allowError bool) error {
	addrs, err := dd.getAddrs(ctx, allowError)
	if err != nil {
		return err
	}
	if len(addrs) == 0 {
		return nil
	}

	if err = dd.lb.SetAddrs(addrs); err != nil {
		dd.AutoLogger().Error(ctx, "dnsDiscoverSetAddressesErr",
			logger.Reflect(logger.ServiceName, dd.serviceName),
			logger.Error(err),
		)
		return err
	}

	return nil
}

func (dd *dnsDiscoverer) getAddrs(ctx context.Context, allowError bool) ([]net.Addr, error) {
	addrs := []net.Addr{}
	for _, n := range dd.nodes {
		ips, err := lookupIP(ctx, n.Network, n.Host)
		if err != nil {
			dd.AutoLogger().Error(ctx, "dnsDiscoverLookupIPErr",
				logger.Reflect(logger.ServiceName, dd.serviceName),
				logger.Error(err),
			)
			if allowError {
				continue
			}
			return addrs, err
		}
		for _, i := range ips {
			addrs = append(addrs, &addr.TCPAddr{
				IP:   i,
				Port: n.Port,
			})
		}
	}
	return addrs, nil
}

func (dd *dnsDiscoverer) loop(ctx context.Context) {
	go nopanic.GoVoid(ctx, func() {
		timer := time.NewTimer(dd.refreshWindow)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				dd.AutoLogger().Info(ctx, "discoverSetAddressesErr",
					logger.Reflect(logger.ServiceName, dd.serviceName),
					logger.Error(ctx.Err()),
				)
				return
			case <-timer.C:
				_ = dd.discover(ctx, true)
			}
		}
	})
}

func lookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}

	ips, err := net.DefaultResolver.LookupIP(ctx, network, host)
	if err != nil {
		return ips, asDNSError(err)
	}

	return ips, nil
}

func asDNSError(err error) *net.DNSError {
	var de *net.DNSError
	errors.As(err, &de)
	// de.IsNotFound
	return de
}
