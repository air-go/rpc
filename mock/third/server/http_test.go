package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHTTP(t *testing.T) {
	data := map[string]interface{}{
		"a": 1,
	}
	bb, _ := json.Marshal(data)

	s, err := NewHTTP(func(server *gin.Engine) {
		server.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, data)
			c.Abort()
		})
	})
	assert.Nil(t, err)
	go func() {
		_ = s.Start()
	}()
	time.Sleep(time.Second * 1)

	resp, err := http.Get(fmt.Sprintf("http://%s%s", s.Addr(), "/test"))
	assert.Nil(t, err)
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, b, bb)
	_ = s.Stop()
}
