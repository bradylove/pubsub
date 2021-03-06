package node

import (
	"math/rand"
)

type Subscription interface {
	Write(data interface{})
}

type ShardingAlgorithm interface {
	Write(data interface{}, subscriptions []Subscription)
}

type Node struct {
	children      map[string]*Node
	subscriptions map[string][]SubscriptionEnvelope
	shards        map[int64]string
}

type SubscriptionEnvelope struct {
	Subscription
	id int64
}

func New() *Node {
	return &Node{
		children:      make(map[string]*Node),
		subscriptions: make(map[string][]SubscriptionEnvelope),
		shards:        make(map[int64]string),
	}
}

func (n *Node) AddChild(key string) *Node {
	if n == nil {
		return nil
	}

	if child, ok := n.children[key]; ok {
		return child
	}

	child := New()
	n.children[key] = child
	return child
}

func (n *Node) FetchChild(key string) *Node {
	if n == nil {
		return nil
	}

	if child, ok := n.children[key]; ok {
		return child
	}

	return nil
}

func (n *Node) DeleteChild(key string) {
	if n == nil {
		return
	}

	delete(n.children, key)
}

func (n *Node) ChildLen() int {
	return len(n.children)
}

func (n *Node) AddSubscription(s Subscription, shardID string) int64 {
	if n == nil {
		return 0
	}

	id := rand.Int63()
	n.shards[id] = shardID
	n.subscriptions[shardID] = append(n.subscriptions[shardID], SubscriptionEnvelope{
		Subscription: s,
		id:           id,
	})
	return id
}

func (n *Node) DeleteSubscription(id int64) {
	if n == nil {
		return
	}

	shardID, ok := n.shards[id]
	if !ok {
		return
	}

	delete(n.shards, id)

	s := n.subscriptions[shardID]
	for i, ss := range s {
		if ss.id != id {
			continue
		}

		n.subscriptions[shardID] = append(s[:i], s[i+1:]...)
	}

	if len(n.subscriptions[shardID]) == 0 {
		delete(n.subscriptions, shardID)
	}
}

func (n *Node) SubscriptionLen() int {
	return len(n.shards)
}

func (n *Node) ForEachSubscription(f func(shardID string, s []SubscriptionEnvelope)) {
	if n == nil {
		return
	}

	for shardID, s := range n.subscriptions {
		f(shardID, s)
	}
}
