package broker

import (
	"testing"
)

type testMessage struct{}

func (*testMessage) Type() MessageType { return "testNotifier" }
func (*testMessage) Data() any         { return 1 }

type testNotifier struct{}

func (*testNotifier) Type() MessageType            { return "testNotifier" }
func (*testNotifier) Notify(message Message) error { return nil }

func TestNotifier(t *testing.T) {
	RegisterNotifier(&testNotifier{})
	Notify(&testMessage{})
}
