package processor

import (
	"avgys-gophermat/internal/model"
	"avgys-gophermat/internal/model/order"
	"avgys-gophermat/internal/service/accrualclient"
	"avgys-gophermat/internal/service/orders"
	orderrepository "avgys-gophermat/sqlc/order"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type AccrualResult struct {
	orderNum string
	response *model.AccrualOrder
	err      error
}

type Processor struct {
	workersLimit int
	pool         *Pool

	orderService   *orders.OrderService
	accrualService *accrualclient.AccrualService
	logger         *zerolog.Logger

	currentlyProcessing []int64
}

func NewProcessor(done context.Context, orderService *orders.OrderService, accrualService *accrualclient.AccrualService, traceLogger *zerolog.Logger) *Processor {
	p := &Processor{
		workersLimit: 20,

		orderService:   orderService,
		accrualService: accrualService,
		logger:         traceLogger,
	}

	p.pool = NewPool(p.workersLimit/4, p.workersLimit, p.workersLimit)

	p.InitPolling(done)

	return p
}

func (p *Processor) startScan(ctx context.Context) chan int64 {

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
				orders, err := p.orderService.GetOrderUnprocessedOrders(ctx, int32(p.workersLimit))

				if err != nil {
					return err
				}

				newOrders := lo.Filter(orders, func(item orderrepository.Order, _ int) bool {
					return !lo.Contains(p.currentlyProcessing, item.OrderNum)
				})

				p.MarkProcessing(newOrders)

				for _, order := range newOrders {
					select {
					case <-ctx.Done():
						return nil
					case resultCh <- order.OrderNum:
					}
				}
			}
		}
	}()

	return resultCh
}

func (p *Processor) InitPolling(done context.Context) {
	orderCh := p.startScan(done)

	processedCh := p.startPolling(done, orderCh)

	p.StoreResult(done, processedCh)

	//Start job pool
	go p.pool.Run(done)
}

func (p *Processor) StoreResult(done context.Context, processedCh chan AccrualResult) {

	go func() {
		// const interval = time.Second * 10
		// t := time.NewTicker(interval)
		// defer t.Stop()

		for {
			select {
			case <-done.Done():
				return
			case accrualRs := <-processedCh:
				if accrualRs.err != nil {
					if errors.Is(accrualRs.err, accrualclient.ErrOrderNotExists) {
						storeAccrualResponse(done, p, &model.AccrualOrder{OrderNum: accrualRs.orderNum, Accrual: 0, Status: order.StatusName[order.StatusInvalid]})
					}
					//TODO resolve err
				} else if accrualRs.response != nil {
					storeAccrualResponse(done, p, accrualRs.response)
				}
				// TODO resolve err

				orderNum, _ := strconv.ParseInt(accrualRs.orderNum, 10, 64)

				// TODO resolve err

				p.UnmarkProcessing(orderNum)
			}
		}
	}()
}

func storeAccrualResponse(done context.Context, p *Processor, accrualRs *model.AccrualOrder) {
	storeLimit := time.Second * 10
	ctxTimeout, cancel := context.WithTimeout(done, storeLimit)

	p.orderService.UpdateStatus(ctxTimeout, accrualRs)
	cancel()
}

func (p *Processor) startPolling(done context.Context, orderCh chan int64) chan AccrualResult {
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

func (p *Processor) PollOrder(c context.Context, resultCh chan AccrualResult, orderNum int64) {
	orderNumStr := strconv.FormatInt(orderNum, 10)
	response, err := p.accrualService.Send(c, orderNumStr)

	if errors.Is(err, accrualclient.ErrTooManyRequests) {
		//TODO Add blocker
	}

	resultCh <- AccrualResult{orderNum: orderNumStr, response: response, err: err}
}

func (p *Processor) MarkProcessing(newOrders []orderrepository.Order) {
	//TODO Add mux

	for _, newOrder := range newOrders {
		p.currentlyProcessing = append(p.currentlyProcessing, newOrder.OrderNum)
	}
}

func (p *Processor) UnmarkProcessing(processed ...int64) {
	//TODO Add mux

	p.currentlyProcessing = lo.Filter(p.currentlyProcessing, func(item int64, _ int) bool {
		return !lo.Contains(processed, item)
	})
}
