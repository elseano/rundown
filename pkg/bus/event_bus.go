package bus

import (
	"reflect"

	"github.com/elseano/rundown/pkg/util"
)

type Event interface{}
type Handler interface {
	ReceiveEvent(event Event)
}

var subscriptions = make(map[Handler]Handler)

func Subscribe(subscriber Handler) {
	util.Logger.Trace().Msg("Adding subscriber")
	subscriptions[subscriber] = subscriber
}

func Unsubscribe(subscriber Handler) {
	util.Logger.Trace().Msg("Removing subscriber")

	delete(subscriptions, subscriber)
}

func Emit(event Event) {
	if reflect.ValueOf(event).Kind() != reflect.Ptr {
		panic("Must pass pointer.")
	}

	util.Logger.Trace().Msgf("Sending message %#v to %d subscribers", event, len(subscriptions))

	for _, subscriber := range subscriptions {
		subscriber.ReceiveEvent(event)
	}
}
