package structlistener

import (
	"testing"

	"github.com/apolloconfig/agollo/v4/agcache/memory"
	"github.com/apolloconfig/agollo/v4/extension"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/apollo/mock"
)

func Test_structChangeListener_OnChange(t *testing.T) {
	convey.Convey("Test_structChangeListener_OnChange", t, func() {
		convey.Convey("namespace not exits", func() {
			New(map[string]*NamespaceConf{}).OnChange(&storage.ChangeEvent{})
		})
		convey.Convey("namespace empty illegal", func() {
			namespace := ""
			var conf struct {
				A string `json:"a"`
			}

			l := New(map[string]*NamespaceConf{})
			l.SetConfig(namespace, conf)

			l.OnChange(&storage.ChangeEvent{
				// TODO storage.ChangeEvent doesn't support injecting Namespace at the bottom
				Changes: map[string]*storage.ConfigChange{
					"content": &storage.ConfigChange{
						NewValue: `{"a":"a"}`,
					},
				},
			})
		})
	})
}

func Test_structChangeListener_OnNewestChange(t *testing.T) {
	convey.Convey("Test_structChangeListener_OnNewestChange", t, func() {
		convey.Convey("success", func() {
			New(map[string]*NamespaceConf{}).OnNewestChange(&storage.FullChangeEvent{})
		})
	})
}

func Test_structChangeListener_InitConfig(t *testing.T) {
	convey.Convey("Test_structChangeListener_InitConfig", t, func() {
		convey.Convey("success", func() {
			var conf1 struct {
				A string `json:"a"`
			}

			n := "conf.json"

			ctl := gomock.NewController(t)
			defer ctl.Finish()

			// init default cache
			cache := &memory.DefaultCache{}
			err := cache.Set("content", `{"a":"a"}`, 10)
			assert.Nil(t, err)

			// mock cache factory
			cacheFactory := mock.NewMockCacheFactory(ctl)
			cacheFactory.EXPECT().Create().Return(cache)
			extension.SetCacheFactory(cacheFactory)

			// set config waitInit done
			c := storage.CreateNamespaceConfig(n)
			conf := c.GetConfig(n)
			conf.GetWaitInit().Done()

			// mock apollo client
			client := mock.NewMockClient(ctl)
			client.EXPECT().GetConfig(gomock.Any()).Times(1).Return(conf)

			l := New(map[string]*NamespaceConf{
				n: &NamespaceConf{FileStruct: &conf1},
			})
			l.InitConfig(client)

			value, ok := l.GetConfig(n)
			assert.Equal(t, true, ok)
			assert.NotNil(t, value)
		})
	})
}
