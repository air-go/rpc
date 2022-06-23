package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/air-go/rpc/library/registry"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type RegistrarOption struct {
	lease  int64
	encode registry.Encode
}

type RegistrarOptionFunc func(*RegistrarOption)

func defaultRegistrarOption() *RegistrarOption {
	return &RegistrarOption{
		lease:  5,
		encode: JSONEncode,
	}
}

// EtcdRegistrar
type EtcdRegistrar struct {
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

var _ registry.Registrar = (*EtcdRegistrar)(nil)

func WithRegistrarLease(lease int64) RegistrarOptionFunc {
	return func(o *RegistrarOption) { o.lease = lease }
}

func WithRegistrarEncode(encode registry.Encode) RegistrarOptionFunc {
	return func(o *RegistrarOption) { o.encode = encode }
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

	if r.val, err = r.opts.encode(&registry.Node{
		Host: r.host,
		Port: r.port,
	}); err != nil {
		return nil, err
	}

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
	for leaseKeepResp := range s.keepAliveChan {
		_ = leaseKeepResp
		// log.Println("续租：", leaseKeepResp)
	}
}

// Close
func (s *EtcdRegistrar) DeRegister(ctx context.Context) error {
	// 撤销租约
	if _, err := s.cli.Revoke(ctx, s.leaseID); err != nil {
		return err
	}
	return s.cli.Close()
}

func JSONEncode(node *registry.Node) (string, error) {
	val, err := json.Marshal(node)
	if err != nil {
		return "", errors.New("marshal node " + err.Error())
	}

	return string(val), nil
}
