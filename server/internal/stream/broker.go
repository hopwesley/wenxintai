package stream

import (
	"encoding/json"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Broker struct {
	mu           sync.Mutex
	topics       map[string]*topic
	historyLimit int
	ttl          time.Duration
	stop         chan struct{}
	once         sync.Once
}

type topic struct {
	subscribers    map[int64]chan Event
	nextSubscriber int64
	history        []Event
	seq            int64
	lastActive     time.Time
}

func NewBroker(historyLimit int, ttl time.Duration) *Broker {
	if historyLimit <= 0 {
		historyLimit = 100
	}
	if ttl <= 0 {
		ttl = 2 * time.Minute
	}
	b := &Broker{
		topics:       make(map[string]*topic),
		historyLimit: historyLimit,
		ttl:          ttl,
		stop:         make(chan struct{}),
	}
	go b.cleanupLoop()
	return b
}

func (b *Broker) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			b.cleanup()
		case <-b.stop:
			return
		}
	}
}

func (b *Broker) cleanup() {
	cutoff := time.Now().Add(-b.ttl)
	b.mu.Lock()
	defer b.mu.Unlock()
	for key, tp := range b.topics {
		if len(tp.subscribers) > 0 {
			continue
		}
		if tp.lastActive.After(cutoff) {
			continue
		}
		delete(b.topics, key)
	}
}

func (b *Broker) Stop() {
	b.once.Do(func() {
		close(b.stop)
	})
}

func (b *Broker) Publish(topicKey string, payload Payload) (Event, error) {
	raw, err := json.Marshal(payload.Data)
	if err != nil {
		return Event{}, err
	}
	b.mu.Lock()
	tp := b.ensureTopic(topicKey)
	seq := atomic.AddInt64(&tp.seq, 1)
	evt := Event{
		ID:   strconv.FormatInt(seq, 10),
		Type: payload.Type,
		Data: raw,
	}
	tp.lastActive = time.Now()
	tp.history = append(tp.history, evt)
	if len(tp.history) > b.historyLimit {
		tp.history = append([]Event(nil), tp.history[len(tp.history)-b.historyLimit:]...)
	}
	subscribers := make([]chan Event, 0, len(tp.subscribers))
	for _, ch := range tp.subscribers {
		subscribers = append(subscribers, ch)
	}
	b.mu.Unlock()

	for _, ch := range subscribers {
		safeSend(ch, evt)
	}
	return evt, nil
}

func (b *Broker) Subscribe(topicKey string, lastEventID string) (<-chan Event, func()) {
	ch := make(chan Event, 64)
	b.mu.Lock()
	tp := b.ensureTopic(topicKey)
	id := atomic.AddInt64(&tp.nextSubscriber, 1)
	tp.subscribers[id] = ch
	history := replayHistory(tp.history, lastEventID)
	b.mu.Unlock()

	for _, evt := range history {
		safeSend(ch, evt)
	}

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			b.mu.Lock()
			if tp, ok := b.topics[topicKey]; ok {
				delete(tp.subscribers, id)
			}
			b.mu.Unlock()
			close(ch)
		})
	}
	return ch, cancel
}

func (b *Broker) ensureTopic(key string) *topic {
	if tp, ok := b.topics[key]; ok {
		return tp
	}
	tp := &topic{
		subscribers: make(map[int64]chan Event),
		history:     make([]Event, 0, b.historyLimit),
		lastActive:  time.Now(),
	}
	b.topics[key] = tp
	return tp
}

func replayHistory(history []Event, lastID string) []Event {
	if lastID == "" {
		return append([]Event(nil), history...)
	}
	startSeq, err := strconv.ParseInt(lastID, 10, 64)
	if err != nil {
		return append([]Event(nil), history...)
	}
	var out []Event
	for _, evt := range history {
		seq, err := strconv.ParseInt(evt.ID, 10, 64)
		if err != nil {
			continue
		}
		if seq > startSeq {
			out = append(out, evt)
		}
	}
	return out
}

func safeSend(ch chan Event, evt Event) {
	select {
	case ch <- evt:
	default:
		go func() {
			defer func() {
				recover()
			}()
			ch <- evt
		}()
	}
}
