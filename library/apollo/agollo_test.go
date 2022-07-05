package agollo

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/apollo/listener"
	"github.com/air-go/rpc/library/apollo/mock"
)

func TestNew(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	l := mock.NewMockListener(ctl)
	l.EXPECT().InitConfig(gomock.Any()).Times(1)

	namespace := "test.json"
	err := New(context.Background(), "test", "127.0.0.1:80", "default", []string{namespace}, WithCustomListeners([]listener.Listener{l}))
	assert.Nil(t, err)
}
