package structlistener

import (
	"sync"

	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/storage"

	"github.com/air-go/rpc/library/apollo/agollo/listener"
	"github.com/air-go/rpc/library/apollo/agollo/util"
)

type StructChangeListener struct {
	namespaces sync.Map
}

var _ listener.Listener = (*StructChangeListener)(nil)

func (c *StructChangeListener) OnChange(changeEvent *storage.ChangeEvent) {}

func (c *StructChangeListener) OnNewestChange(event *storage.FullChangeEvent) {
	conf, ok := c.namespaces.Load(event.Namespace)
	if !ok {
		return
	}

	value, ok := event.Changes["content"]
	if !ok {
		log.Errorf("StructChangeListener.OnNewestChange %s err: content not exists", event.Namespace)
		return
	}

	content, ok := value.(string)
	if !ok {
		log.Errorf("StructChangeListener.OnNewestChange %s err: content assert fail", event.Namespace)
		return
	}

	if err := util.ExtractConf(event.Namespace, content, conf); err != nil {
		log.Errorf("StructChangeListener.OnNewestChange %s err: %s", event.Namespace, err.Error())
		return
	}
}

func (c *StructChangeListener) InitConfig(client agollo.Client, namespaceStruct map[string]interface{}) {
	for namespace, confStruct := range namespaceStruct {
		conf := client.GetConfig(namespace)
		if conf == nil {
			panic(namespace + " conf nil")
		}

		content := conf.GetValue("content")
		if err := util.ExtractConf(namespace, content, confStruct); err != nil {
			panic(err)
		}

		c.namespaces.Store(namespace, confStruct)
	}
}

func (c *StructChangeListener) GetNamespace(namespace string) (interface{}, bool) {
	return c.namespaces.Load(namespace)
}
