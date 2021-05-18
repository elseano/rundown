package rpc

type SubscriptionList []Subscription

type Subscribers struct {
	subscribers []Subscription
}

func (s *Subscribers) Subscribe(receiver Subscription) {
	s.subscribers = append(s.subscribers, receiver)
}

func (s *Subscribers) Emit(data []byte) {
	EmitAll(s.subscribers, data)
}

func EmitAll(subscribers []Subscription, data []byte) {
	for _, element := range subscribers {
		if sub, ok := element.(Subscription); ok {
			abort, _ := sub.Receive(data)
			if abort {
				return
			}
		}
	}
}

type Subscription interface {
	Receive(message []byte) (bool, error)
}
