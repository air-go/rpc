package structlistener

import (
	"sync"

	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/storage"

	"github.com/air-go/rpc/library/apollo/listener"
	"github.com/air-go/rpc/library/apollo/util"
)

type NamespaceConf struct {
	FileStruct interface{}
}

type structChangeListener struct {
	namespaces    sync.Map
	namespaceConf map[string]*NamespaceConf
}

var _ listener.Listener = (*structChangeListener)(nil)

type Option func(*structChangeListener)

func New(conf map[string]*NamespaceConf, opts ...Option) *structChangeListener {
	l := &structChangeListener{
		namespaceConf: conf,
	}

	for _, o := range opts {
		o(l)
	}

	return l
}

func (c *structChangeListener) OnChange(event *storage.ChangeEvent) {
	conf, ok := c.namespaces.Load(event.Namespace)
	if !ok {
		return
	}

	value, ok := event.Changes["content"]
	if !ok {
		log.Errorf("structChangeListener.OnNewestChange %s err: content not exists", event.Namespace)
		return
	}

	content, ok := value.NewValue.(string)
	if !ok {
		log.Errorf("structChangeListener.OnNewestChange %s err: content assert fail", event.Namespace)
		return
	}

	if err := util.ExtractConf(event.Namespace, content, conf); err != nil {
		log.Errorf("structChangeListener.OnNewestChange %s err: %s", event.Namespace, err.Error())
		return
	}
}

func (c *structChangeListener) OnNewestChange(event *storage.FullChangeEvent) {}

func (c *structChangeListener) InitConfig(client agollo.Client) {
	for namespace, confStruct := range c.namespaceConf {
		conf := client.GetConfig(namespace)
		if conf == nil {
			panic(namespace + " conf nil")
		}

		content := conf.GetValue("content")
		if content == "" {
			panic(namespace + " content empty")
		}

		if err := util.ExtractConf(namespace, content, confStruct); err != nil {
			panic(namespace + err.Error())
		}

		c.SetConfig(namespace, confStruct)
	}
}

func (c *structChangeListener) GetConfig(namespace string) (interface{}, bool) {
	return c.namespaces.Load(namespace)
}

func (c *structChangeListener) SetConfig(namespace string, value interface{}) {
	c.namespaces.Store(namespace, value)
}
