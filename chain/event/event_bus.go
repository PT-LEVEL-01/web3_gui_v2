package event

import (
	"sync"
)

const (
	defaultEventBufferSize = 10240
	ContractEventTopic     = "event_contract"
)

type EventBus interface {
	Subscriber(topic string, sub Subscriber)
	UnSubscriber(topic string, sub Subscriber)
	Publish(topic string, payload interface{})
	Close()
}
type Event struct {
	Topic   string
	Payload interface{}
}
type eventBusImpl struct {
	mu sync.RWMutex
	// topicMap 主题订阅集合
	topicMap        map[string][]Subscriber
	once            sync.Once
	topicChannelMap map[string]chan *Event
	quitC           chan struct{}
	closed          bool
}

func (bus *eventBusImpl) Subscriber(topic string, sub Subscriber) {
	bus.once.Do(func() {
		bus.handleEventLooping()
	})
	bus.mu.Lock()
	defer bus.mu.Unlock()
	if bus.closed {
		return
	}
	if subs, ok := bus.topicMap[topic]; ok {
		if SubExit(subs, sub) {
			return
		}
		bus.topicMap[topic] = append(subs, sub)
	} else {
		bus.topicMap[topic] = append([]Subscriber{}, sub)
		chanEvent := make(chan *Event, defaultEventBufferSize)
		bus.topicChannelMap[topic] = chanEvent
		go func() {
			for {
				select {
				case <-bus.quitC:
					return
				case e := <-chanEvent:
					go bus.notify(e)
				}
			}
		}()
	}
}

//判断当前主题是否已经存在该订阅者
func SubExit(subs []Subscriber, sub Subscriber) bool {
	for _, v := range subs {
		if v == sub {
			return true
		}
	}
	return false
}
func (bus *eventBusImpl) UnSubscriber(topic string, sub Subscriber) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	if bus.closed {
		return
	}
	if subs, ok := bus.topicMap[topic]; ok {
		for index, s := range subs {
			if s == sub {
				leftSubs := append(subs[:index], subs[index+1:]...)
				bus.topicMap[topic] = leftSubs
			}
		}
	}
}

//推送事件
func (bus *eventBusImpl) Publish(topic string, payload interface{}) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	if bus.closed {
		return
	}
	if c, ok := bus.topicChannelMap[topic]; ok {
		c <- &Event{
			Topic:   topic,
			Payload: payload,
		}
	}
	return
}
func (bus *eventBusImpl) Close() {
	select {
	case <-bus.quitC:
	default:
		close(bus.quitC)
		for _, c := range bus.topicChannelMap {
			close(c)
		}
	}
}
func NewEventBus() EventBus {
	return &eventBusImpl{
		mu:              sync.RWMutex{},
		topicMap:        make(map[string][]Subscriber),
		once:            sync.Once{},
		topicChannelMap: make(map[string]chan *Event),
		quitC:           make(chan struct{}),
		closed:          false,
	}
}

func (bus *eventBusImpl) handleEventLooping() {
	go func() {
		for {
			select {
			case <-bus.quitC:
				bus.mu.Lock()
				for _, subs := range bus.topicMap {
					for i := 0; i < len(subs); i++ {
						s := subs[i]
						go s.OnExit()
					}
				}
				bus.mu.Unlock()
				bus.closed = true
				return

			}
		}
	}()
}
func (bus *eventBusImpl) notify(e *Event) {
	if e.Topic == "" {
		return
	}
	if subs, ok := bus.topicMap[e.Topic]; ok {
		for _, sub := range subs {
			go sub.OnListen(e)
		}
	}
}
