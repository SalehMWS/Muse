package application

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

const (
	defaultWorkers = 4
	defaultBlock   = 2 * time.Second
)

type Stats struct {
	Processed    int64 `json:"processed"`
	Succeeded    int64 `json:"succeeded"`
	Retried      int64 `json:"retried"`
	DeadLettered int64 `json:"dead_lettered"`
	Failed       int64 `json:"failed"`
}

type metrics struct {
	processed    atomic.Int64
	succeeded    atomic.Int64
	retried      atomic.Int64
	deadLettered atomic.Int64
	failed       atomic.Int64
}

type Pool struct {
	broker     Broker
	dispatcher *Dispatcher
	logger     *zap.Logger
	workers    int
	block      time.Duration
	metrics    metrics
}

func NewPool(broker Broker, dispatcher *Dispatcher, logger *zap.Logger, workers int, block time.Duration) *Pool {
	if logger == nil {
		logger = zap.NewNop()
	}
	if workers <= 0 {
		workers = defaultWorkers
	}
	if block <= 0 {
		block = defaultBlock
	}
	return &Pool{
		broker:     broker,
		dispatcher: dispatcher,
		logger:     logger,
		workers:    workers,
		block:      block,
	}
}

func (p *Pool) Run(ctx context.Context) {
	p.logger.Info("worker pool started", zap.Int("workers", p.workers))
	var wg sync.WaitGroup
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.loop(ctx)
		}()
	}
	wg.Wait()
	p.logger.Info("worker pool stopped")
}

func (p *Pool) loop(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		deliveries, err := p.broker.Read(ctx, 1, p.block)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			p.logger.Error("worker: read from broker", zap.Error(err))
			p.sleep(ctx, time.Second)
			continue
		}
		for _, delivery := range deliveries {
			p.handle(ctx, delivery)
		}
	}
}

func (p *Pool) handle(ctx context.Context, delivery Delivery) {
	p.metrics.processed.Add(1)

	if err := p.dispatcher.Dispatch(ctx, delivery.Job); err != nil {
		p.onFailure(ctx, delivery, err)
		return
	}

	if err := p.broker.Ack(ctx, delivery); err != nil {
		p.logger.Error("worker: ack succeeded job", zap.String("job_id", delivery.Job.ID), zap.Error(err))
		return
	}
	p.metrics.succeeded.Add(1)
}

func (p *Pool) onFailure(ctx context.Context, delivery Delivery, cause error) {
	p.metrics.failed.Add(1)

	if delivery.Job.HasAttemptsLeft() {
		retried := delivery.Job.NextAttempt()
		if err := p.broker.Enqueue(ctx, retried); err != nil {
			p.logger.Error("worker: requeue job", zap.String("job_id", delivery.Job.ID), zap.Error(err))
			return
		}
		if err := p.broker.Ack(ctx, delivery); err != nil {
			p.logger.Error("worker: ack requeued job", zap.String("job_id", delivery.Job.ID), zap.Error(err))
			return
		}
		p.metrics.retried.Add(1)
		p.logger.Warn("worker: job retried",
			zap.String("job_id", delivery.Job.ID),
			zap.Int("attempt", retried.Attempt),
			zap.Error(cause),
		)
		return
	}

	if err := p.broker.DeadLetter(ctx, delivery.Job, cause.Error()); err != nil {
		p.logger.Error("worker: dead-letter job", zap.String("job_id", delivery.Job.ID), zap.Error(err))
		return
	}
	if err := p.broker.Ack(ctx, delivery); err != nil {
		p.logger.Error("worker: ack dead-lettered job", zap.String("job_id", delivery.Job.ID), zap.Error(err))
		return
	}
	p.metrics.deadLettered.Add(1)
	p.logger.Error("worker: job dead-lettered",
		zap.String("job_id", delivery.Job.ID),
		zap.String("type", string(delivery.Job.Type)),
		zap.Error(cause),
	)
}

func (p *Pool) Stats() Stats {
	return Stats{
		Processed:    p.metrics.processed.Load(),
		Succeeded:    p.metrics.succeeded.Load(),
		Retried:      p.metrics.retried.Load(),
		DeadLettered: p.metrics.deadLettered.Load(),
		Failed:       p.metrics.failed.Load(),
	}
}

func (p *Pool) sleep(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
