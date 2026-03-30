package processor

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

type Job func(context.Context)

type Pool struct {
	inputCh chan Job

	stops   []chan struct{}
	workers errgroup.Group

	mu sync.Mutex

	min, max int
	upAt     int
	downAt   int
}

func NewPool(min, max, queueSize int) *Pool {
	return &Pool{
		inputCh: make(chan Job, queueSize),
		min:     min,
		max:     max,
		upAt:    queueSize * 3 / 4,
		downAt:  queueSize / 10,
	}
}

func (p *Pool) Enqueue(ctx context.Context, j Job) error {
	select {
	case p.inputCh <- j:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *Pool) Run(ctx context.Context) error {
	for range p.min {
		p.startWorker(ctx)
	}

	//TODO add scaler

	<-ctx.Done()
	p.stopAll()

	return p.workers.Wait()
}

func (p *Pool) startWorker(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	stopCh := make(chan struct{})
	p.stops = append(p.stops, stopCh)

	p.workers.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-stopCh:
				return nil
			case j, open := <-p.inputCh:
				if !open {
					return nil
				} else if j != nil {
					j(ctx)
				}
			}
		}
	})
}

func (p *Pool) removeWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.stops) > p.min {
		return
	}

	lastID := len(p.stops) - 1

	stopCh := p.stops[lastID]
	close(stopCh)

	p.stops = p.stops[:lastID]
}

func (p *Pool) stopAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, stop := range p.stops {
		close(stop)
	}

	p.stops = nil
}
