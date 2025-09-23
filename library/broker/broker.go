package broker

import (
	"sync"
)

type MessageType string

const (
	MessageTypeDiscoverer = "Discoverer"
	MessageTypeConnPool   = "ConnPool"
)

type Message interface {
	Type() MessageType
	Data() any
}

type Notifier interface {
	Type() MessageType
	Notify(message Message) error
}

func RegisterNotifier(n Notifier) {
	brokerNotifiers.Add(n)
}

func Notify(message Message) {
	brokerNotifiers.Notify(message)
}

var brokerNotifiers = notifiers{values: map[MessageType][]Notifier{}}

type notifiers struct {
	mux    sync.RWMutex
	values map[MessageType][]Notifier
}

func (ns *notifiers) Add(n Notifier) {
	ns.mux.Lock()
	defer ns.mux.Unlock()

	ns.values[n.Type()] = append(ns.values[n.Type()], n)
}

func (ns *notifiers) Notify(message Message) {
	ns.mux.RLock()
	defer ns.mux.RUnlock()

	for _, n := range ns.values[message.Type()] {
		_ = n.Notify(message)
	}
}
