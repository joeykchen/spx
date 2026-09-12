package spx

import (
	"context"
	"slices"
	"sync"
)

type pendingQuestion struct {
	ctx      context.Context
	text     string
	sprite   *SpriteImpl
	bubble   bool
	done     bool
	canceled bool
}

// questionQueue owns data shared by scripts and engine answer callbacks.
// Presentation state stays outside the queue on the visual synchronization path.
type questionQueue struct {
	mu     sync.Mutex
	items  []*pendingQuestion
	answer string
}

func (q *questionQueue) add(request *pendingQuestion) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, request)
}

func (q *questionQueue) remove(request *pendingQuestion) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if i := slices.Index(q.items, request); i >= 0 {
		q.items = slices.Delete(q.items, i, i+1)
	}
}

func (q *questionQueue) status(request *pendingQuestion) (done, canceled bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return request.done, request.canceled
}

func (q *questionQueue) front() *pendingQuestion {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = slices.DeleteFunc(q.items, func(request *pendingQuestion) bool { return request.ctx.Err() != nil })
	if len(q.items) == 0 {
		return nil
	}
	return q.items[0]
}

func (q *questionQueue) submit(request *pendingQuestion, answer string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.accepts(request) {
		return
	}
	q.answer = answer
	request.done = true
	q.items = slices.Delete(q.items, 0, 1)
}

func (q *questionQueue) answerText() string {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.answer
}

func (q *questionQueue) clear(resetAnswer bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, request := range q.items {
		request.canceled = true
	}
	clear(q.items)
	q.items = q.items[:0]
	if resetAnswer {
		q.answer = ""
	}
}

func (q *questionQueue) isCurrent(request *pendingQuestion) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.accepts(request)
}

// accepts requires q.mu to be held.
func (q *questionQueue) accepts(request *pendingQuestion) bool {
	return len(q.items) > 0 && q.items[0] == request && !request.done && request.ctx.Err() == nil
}
