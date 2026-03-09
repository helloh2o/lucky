package utils

import (
	"sync"
	"sync/atomic"
	"time"
)

type Callback[T comparable] func(T)

// DedupQueue 这是安全的去重排队queue
type DedupQueue[T comparable] struct {
	queue chan T

	exists map[T]struct{}
	mu     sync.Mutex

	cb Callback[T]

	ticker *time.Ticker
	token  chan struct{}

	stopCh  chan struct{}
	wg      sync.WaitGroup
	workers int

	length atomic.Int64
}

func NewDedupQueue[T comparable](
	size int,
	workers int,
	interval time.Duration,
	cb Callback[T],
) *DedupQueue[T] {

	q := &DedupQueue[T]{
		queue:   make(chan T, size),
		exists:  make(map[T]struct{}, size),
		cb:      cb,
		ticker:  time.NewTicker(interval / time.Duration(workers)),
		token:   make(chan struct{}, workers),
		stopCh:  make(chan struct{}),
		workers: workers,
	}

	// dispatcher
	q.wg.Add(1)
	go q.dispatch()

	// workers
	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go q.worker()
	}

	return q
}

func (q *DedupQueue[T]) dispatch() {
	defer q.wg.Done()

	for {
		select {

		case <-q.stopCh:
			return

		case <-q.ticker.C:
			select {
			case q.token <- struct{}{}:
			default:
			}
		}
	}
}

func (q *DedupQueue[T]) worker() {
	defer q.wg.Done()

	for {
		select {

		case <-q.stopCh:
			return

		case <-q.token:

			select {
			case key := <-q.queue:

				// 先调用逻辑，再删除排队
				q.safeCall(key)

				q.mu.Lock()
				delete(q.exists, key)
				q.mu.Unlock()

				q.length.Add(-1)

			default:
			}
		}
	}
}

func (q *DedupQueue[T]) safeCall(key T) {

	defer func() {
		if r := recover(); r != nil {
			// 这里可以接入日志或监控
			// log.Printf("DedupQueue callback panic: %v", r)
		}
	}()

	q.cb(key)
}

// Enqueue 入队，如果 key 已存在返回 false
func (q *DedupQueue[T]) Enqueue(key T) bool {

	q.mu.Lock()

	if _, ok := q.exists[key]; ok {
		q.mu.Unlock()
		//log.Printf("key %v is queued", key)
		return false
	}

	q.exists[key] = struct{}{}
	q.mu.Unlock()

	select {

	case q.queue <- key:
		q.length.Add(1)
		return true

	default:
		// 队列满，回滚 exists
		q.mu.Lock()
		delete(q.exists, key)
		q.mu.Unlock()

		return false
	}
}

func (q *DedupQueue[T]) Len() int64 {
	return q.length.Load()
}

func (q *DedupQueue[T]) Stop() {

	close(q.stopCh)

	q.ticker.Stop()

	q.wg.Wait()
}
