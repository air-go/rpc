package bootstrap

import (
	"github.com/air-go/rpc/library/app"
	"github.com/air-go/rpc/library/config"
)

func Init(confPath string, load func() error) (err error) {
	welcome()
	pidPrint()

	config.Init(confPath)

	if err = app.InitApp(); err != nil {
		return
	}

	if err = load(); err != nil {
		return
	}

	return
}
