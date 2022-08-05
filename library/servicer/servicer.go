//go:generate go run -mod=mod github.com/golang/mock/mockgen -package mock -source ./servicer.go -destination ./mock/servicer.go Servicer
package servicer

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

const (
	TypeRegistry uint8 = 1
	TypeIPPort   uint8 = 2
	TypeDomain   uint8 = 3
)

var (
	lock      sync.RWMutex
	servicers = make(map[string]Servicer)
)

func SetServicer(s Servicer) (err error) {
	lock.Lock()
	defer lock.Unlock()

	name := s.Name()
	if _, ok := servicers[name]; ok {
		return errors.Errorf("repeat servicer: %s", name)
	}
	servicers[name] = s

	return
}

func UpdateServicer(s Servicer) {
	lock.Lock()
	defer lock.Unlock()

	servicers[s.Name()] = s
}

func DelServicer(s Servicer) {
	lock.Lock()
	defer lock.Unlock()
	delete(servicers, s.Name())
}

func GetServicer(serviceName string) (Servicer, bool) {
	s, has := servicers[serviceName]
	return s, has
}

type Servicer interface {
	Name() string
	RegistryName() string
	Pick(ctx context.Context) (Node, error)
	All(ctx context.Context) ([]Node, error)
	Done(ctx context.Context, node Node, err error) error
	GetCaCrt() []byte
	GetClientPem() []byte
	GetClientKey() []byte
}
