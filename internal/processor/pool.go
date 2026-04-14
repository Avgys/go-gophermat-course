package processor

import (
	"context"
	"sync"
	"time"

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

	go p.scaleLoop(ctx)

	<-ctx.Done()
	p.stopAll()

	return p.workers.Wait()
}

func (p *Pool) scaleLoop(ctx context.Context) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.maybeScale(ctx)
		}
	}
}

func (p *Pool) maybeScale(ctx context.Context) {
	for {
		p.mu.Lock()
		qlen := len(p.inputCh)
		n := len(p.stops)
		scaleUp := qlen > p.upAt && n < p.max
		scaleDown := qlen <= p.downAt && n > p.min
		p.mu.Unlock()

		switch {
		case scaleUp:
			p.startWorker(ctx)
		case scaleDown:
			p.removeWorker()
		default:
			return
		}
	}
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

	if len(p.stops) <= p.min {
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
