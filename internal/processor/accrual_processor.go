package processor

import (
	"avgys-gophermat/internal/model/order"
	"avgys-gophermat/internal/model/responses"
	"avgys-gophermat/internal/service/accrualclient"
	orderrepository "avgys-gophermat/sqlc/order"
	"context"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type AccrualResult struct {
	orderNum string
	response *responses.AccrualOrder
	err      error
}

type AcrrualProcessor struct {
	workersLimit int
	pool         JobRunner

	orderService   OrderService
	accrualService accrualclient.AccrualClient
	logger         *zerolog.Logger

	currentlyProcessing []int64

	mu sync.RWMutex

	accrualPoolSleepTime atomic.Int64
}

func NewAcrrualProcessor(done context.Context, orderService OrderService, accrualService accrualclient.AccrualClient, traceLogger *zerolog.Logger) *AcrrualProcessor {

	log := traceLogger.With().Str("service_name", "job_pool").Logger()

	p := &AcrrualProcessor{
		workersLimit: 20,

		orderService:   orderService,
		accrualService: accrualService,
		logger:         &log,
	}

	p.accrualPoolSleepTime.Store(0)

	p.pool = NewPool(p.workersLimit/4, p.workersLimit, p.workersLimit)

	p.InitPolling(done)

	return p
}

func (p *AcrrualProcessor) startScan(ctx context.Context) chan int64 {

	resultCh := make(chan int64, p.workersLimit)

	go func() error {
		const interval = time.Second
		t := time.NewTicker(interval)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-t.C:

				procCount := len(p.currentlyProcessing)
				scanLimit := p.workersLimit - procCount

				orders, err := p.orderService.GetOrderUnprocessedOrders(ctx, scanLimit)

				if err != nil {
					return err
				}

				p.mu.RLock()

				newOrders := lo.Filter(orders, func(item orderrepository.Order, _ int) bool {
					return !lo.Contains(p.currentlyProcessing, item.OrderNum)
				})

				p.mu.RUnlock()

				p.MarkProcessing(newOrders)

				sleepTime := p.GetSleepTime()

				if sleepTime > 0 {
					time.Sleep(sleepTime)
				}

				for _, order := range p.currentlyProcessing {
					select {
					case <-ctx.Done():
						return nil
					case resultCh <- order:
					}
				}
			}
		}
	}()

	return resultCh
}

func (p *AcrrualProcessor) InitPolling(done context.Context) {
	orderCh := p.startScan(done)

	processedCh := p.startPolling(done, orderCh)

	p.StoreResult(done, processedCh)

	//Start job pool
	go p.pool.Run(done)
}

func (p *AcrrualProcessor) StoreResult(done context.Context, processedCh chan AccrualResult) {

	go func() {

		for {
			select {
			case <-done.Done():
				return
			case accrualRs := <-processedCh:
				if accrualRs.err != nil {
					if errors.Is(accrualRs.err, accrualclient.ErrOrderNotExists) {
						storeAccrualResponse(done, p, &responses.AccrualOrder{OrderNum: accrualRs.orderNum, Accrual: 0, Status: order.StatusName[order.StatusInvalid]})
					}

					p.logger.Err(accrualRs.err)

				} else if accrualRs.response != nil {
					storeAccrualResponse(done, p, accrualRs.response)
				}

				orderNum, err := strconv.ParseInt(accrualRs.orderNum, 10, 64)

				if err != nil {
					p.logger.Err(accrualRs.err)
					continue
				}

				p.UnmarkProcessing(orderNum)
			}
		}
	}()
}

func storeAccrualResponse(done context.Context, p *AcrrualProcessor, accrualRs *responses.AccrualOrder) {
	storeLimit := time.Second * 10
	ctxTimeout, cancel := context.WithTimeout(done, storeLimit)

	p.orderService.UpdateOrderStatus(ctxTimeout, accrualRs)
	cancel()
}

func (p *AcrrualProcessor) startPolling(done context.Context, orderCh chan int64) chan AccrualResult {
	resultCh := make(chan AccrualResult, p.workersLimit)

	go func() error {

		for {
			select {
			case <-done.Done():
				return nil
			case orderNum := <-orderCh:
				p.pool.Enqueue(done, func(c context.Context) { p.PollOrder(c, resultCh, orderNum) })
			}
		}
	}()

	return resultCh
}

func (p *AcrrualProcessor) PollOrder(c context.Context, resultCh chan AccrualResult, orderNum int64) {
	orderNumStr := strconv.FormatInt(orderNum, 10)
	response, err := p.accrualService.Send(c, orderNumStr)

	if errors.Is(err, accrualclient.ErrTooManyRequests) {
		const retryTimeLimit int64 = 60

		p.accrualPoolSleepTime.Store(retryTimeLimit)
		time.AfterFunc(time.Second*time.Duration(p.accrualPoolSleepTime.Load()), func() {
			p.accrualPoolSleepTime.Store(0)
		})

		return
	}

	resultCh <- AccrualResult{orderNum: orderNumStr, response: response, err: err}
}

func (p *AcrrualProcessor) MarkProcessing(newOrders []orderrepository.Order) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, newOrder := range newOrders {
		p.currentlyProcessing = append(p.currentlyProcessing, newOrder.OrderNum)
	}
}

func (p *AcrrualProcessor) UnmarkProcessing(processed ...int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.currentlyProcessing = lo.Filter(p.currentlyProcessing, func(item int64, _ int) bool {
		return !lo.Contains(processed, item)
	})
}

func (p *AcrrualProcessor) GetSleepTime() time.Duration {
	return time.Second * time.Duration(p.accrualPoolSleepTime.Load())
}
