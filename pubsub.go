package pubsub

import (
	"github.com/apoydence/pubsub/internal/node"
)

type PubSub struct {
	e SubscriptionEnroller
	a DataAssigner
	n *node.Node
}

type SubscriptionEnroller interface {
	Enroll(sub Subscription, location []string) (key string, ok bool)
}

type DataAssigner interface {
	Assign(data interface{}, location []string) (keys []string)
}

func New(e SubscriptionEnroller, a DataAssigner) *PubSub {
	return &PubSub{
		e: e,
		a: a,
		n: node.New(),
	}
}

type Subscription interface {
	Write(data interface{})
}

type Unsubscriber func()

func (s *PubSub) Subscribe(subscription Subscription) Unsubscriber {
	s.n.Lock()
	return s.traverseSubscribe(subscription, s.n, nil)
}

func (s *PubSub) Publish(d interface{}) {
	s.n.RLock()
	s.traversePublish(d, s.n, nil)
}

func (s *PubSub) traversePublish(d interface{}, n *node.Node, l []string) {
	n.ForEachSubscription(func(ss node.Subscription) {
		ss.Write(d)
	})

	children := s.a.Assign(d, l)

	var childNodes []*node.Node
	for _, child := range children {
		c := n.FetchChild(child)
		childNodes = append(childNodes, c)
		c.RLock()
	}

	n.RUnlock()
	for i, child := range children {
		s.traversePublish(d, childNodes[i], append(l, child))
	}
}

func (s *PubSub) traverseSubscribe(ss Subscription, n *node.Node, l []string) Unsubscriber {
	child, ok := s.e.Enroll(ss, l)
	if !ok {
		n.AddSubscription(ss)
		n.Unlock()

		return func() {
			current := s.n
			current.Lock()

			for _, ll := range l {
				prev := current
				current = current.FetchChild(ll)
				current.Lock()
				prev.Unlock()
			}
			current.DeleteSubscription(ss)
			current.Unlock()
		}
	}

	c := n.AddChild(child)
	c.Lock()
	n.Unlock()
	return s.traverseSubscribe(ss, c, append(l, child))
}
