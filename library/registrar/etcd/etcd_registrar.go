package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/logger/setup"
	"github.com/air-go/rpc/library/registrar"
)

type RegistrarOption struct {
	lease int64
}

type RegistrarOptionFunc func(*RegistrarOption)

func defaultRegistrarOption() *RegistrarOption {
	return &RegistrarOption{
		lease: 5,
	}
}

// EtcdRegistrar
type EtcdRegistrar struct {
	setup.SetupLogger
	opts          *RegistrarOption
	cli           *clientv3.Client
	serviceName   string
	host          string
	port          int
	leaseID       clientv3.LeaseID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
	val           string
}

var _ registrar.Registrar = (*EtcdRegistrar)(nil)

func WithRegistrarLease(lease int64) RegistrarOptionFunc {
	return func(o *RegistrarOption) { o.lease = lease }
}

// NewRegistry
func NewRegistry(cli *clientv3.Client, name, host string, port int, opts ...RegistrarOptionFunc) (*EtcdRegistrar, error) {
	var err error

	if cli == nil {
		return nil, errors.New("cli is nil")
	}

	if name = strings.TrimSpace(name); name == "" {
		return nil, errors.New("serviceName is nil")
	}

	opt := defaultRegistrarOption()
	for _, o := range opts {
		o(opt)
	}

	r := &EtcdRegistrar{
		opts:        opt,
		cli:         cli,
		serviceName: name,
		host:        host,
		port:        port,
	}

	r.key = fmt.Sprintf("%s.%s.%d", r.serviceName, r.host, r.port)

	v := map[string]interface{}{
		"host": r.host,
		"port": r.port,
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	r.val = string(b)

	return r, nil
}

func (s *EtcdRegistrar) Register(ctx context.Context) error {
	if s.cli == nil {
		return errors.New("cli is nil")
	}

	// 申请租约设置时间keepalive
	if err := s.putKeyWithRegistrarLease(ctx, s.opts.lease); err != nil {
		return err
	}

	// 监听续租相应chan
	go s.listenLeaseRespChan()

	return nil
}

// putKeyWithRegistrarLease
func (s *EtcdRegistrar) putKeyWithRegistrarLease(ctx context.Context, lease int64) error {
	// 设置租约时间
	resp, err := s.cli.Grant(ctx, lease)
	if err != nil {
		return err
	}
	// 注册服务并绑定租约
	_, err = s.cli.Put(ctx, s.key, s.val, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	// 设置续租 定期发送需求请求
	leaseRespChan, err := s.cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}
	s.leaseID = resp.ID
	s.keepAliveChan = leaseRespChan
	return nil
}

// listenLeaseRespChan
func (s *EtcdRegistrar) listenLeaseRespChan() {
	for ls := range s.keepAliveChan {
		s.AutoLogger().Info(context.Background(), "etcdKeepAliveChanReceive",
			logger.Reflect(logger.ServiceName, s.serviceName),
			logger.Reflect(logger.Response, ls),
		)
	}
}

// DeRegister
func (s *EtcdRegistrar) DeRegister(ctx context.Context) error {
	// 撤销租约
	if _, err := s.cli.Revoke(ctx, s.leaseID); err != nil {
		return err
	}
	return s.cli.Close()
}

func (s *EtcdRegistrar) SetLogger(l logger.Logger) {
	s.SetupLogger.SetLogger(l)
}
