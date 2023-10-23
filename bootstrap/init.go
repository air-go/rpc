package bootstrap

import (
	"fmt"

	"github.com/air-go/rpc/library/app"
	"github.com/air-go/rpc/library/config"
)

func Init(confPath string, load func() error) (err error) {
	fmt.Println(welcome())

	config.Init(confPath)

	if err = app.InitApp(); err != nil {
		return
	}

	if err = load(); err != nil {
		return
	}

	return
}
