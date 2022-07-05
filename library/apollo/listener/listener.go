package listener

import (
	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/storage"
)

type Listener interface {
	storage.ChangeListener
	InitConfig(client agollo.Client)
}
