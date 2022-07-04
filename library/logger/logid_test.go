package logger

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestNewObjectID(t *testing.T) {
	convey.Convey("TestNewObjectID", t, func() {
		convey.Convey("success", func() {
			id := NewObjectID()
			assert.NotEmpty(t, id)
		})
	})
}

func TestNewObjectIDWithTime(t *testing.T) {
	convey.Convey("TestNewObjectIDWithTime", t, func() {
		convey.Convey("success", func() {
			id := NewObjectIDWithTime(time.Now())
			assert.NotEmpty(t, id)
		})
	})
}

func TestNewObjectIDWithHexString(t *testing.T) {
	convey.Convey("TestNewObjectIDWithHexString", t, func() {
		convey.Convey("success", func() {
			str := hex.EncodeToString([]byte("3pjmgvlegp11"))
			id, err := NewObjectIDWithHexString(str)
			assert.Nil(t, err)
			assert.NotEmpty(t, id)
		})
		convey.Convey("len != 12", func() {
			str := hex.EncodeToString([]byte("3pjmgvlegp1"))
			id, err := NewObjectIDWithHexString(str)
			assert.NotNil(t, err)
			assert.Empty(t, id)
		})
	})
}

func TestObjectID_Hex(t *testing.T) {
	convey.Convey("TestObjectID_Hex", t, func() {
		convey.Convey("success", func() {
			str := hex.EncodeToString([]byte("3pjmgvlegp11"))
			id, err := NewObjectIDWithHexString(str)
			assert.Nil(t, err)
			assert.Equal(t, id.Hex(), "33706a6d67766c6567703131")
		})
	})
}

func TestObjectID_Time(t *testing.T) {
	convey.Convey("TestNewObjectIDWithTime", t, func() {
		convey.Convey("success", func() {
			now := time.Now()
			id := NewObjectIDWithTime(now)
			assert.Equal(t, id.Time().Unix(), now.Unix())
		})
	})
}

func TestObjectID_Machine(t *testing.T) {
	convey.Convey("TestObjectID_Machine", t, func() {
		convey.Convey("success", func() {
			str := hex.EncodeToString([]byte("3pjmgvlegp11"))
			id, err := NewObjectIDWithHexString(str)
			assert.Nil(t, err)
			assert.NotEmpty(t, id.Machine())
		})
	})
}

func TestObjectID_Pid(t *testing.T) {
	convey.Convey("TestObjectID_Pid", t, func() {
		convey.Convey("success", func() {
			str := hex.EncodeToString([]byte("3pjmgvlegp11"))
			id, err := NewObjectIDWithHexString(str)
			assert.Nil(t, err)
			assert.NotEmpty(t, id.Pid())
		})
	})
}

func TestObjectID_Counter(t *testing.T) {
	convey.Convey("TestObjectID_Counter", t, func() {
		convey.Convey("success", func() {
			str := hex.EncodeToString([]byte("3pjmgvlegp11"))
			id, err := NewObjectIDWithHexString(str)
			assert.Nil(t, err)
			assert.NotEmpty(t, id.Counter())
		})
	})
}
