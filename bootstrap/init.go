package bootstrap

import (
	"log"
	"syscall"

	"github.com/air-go/rpc/library/app"
	"github.com/air-go/rpc/library/config"
)

func Init(confPath string, load func() error) (err error) {
	log.Printf("Actual pid is %d", syscall.Getpid())

	config.Init(confPath)

	if err = app.InitApp(); err != nil {
		return
	}

	if err = load(); err != nil {
		return
	}

	return
}
