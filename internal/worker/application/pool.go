package application

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/SalehMWS/Muse/internal/shared/logger"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
	"github.com/SalehMWS/Muse/internal/shared/tracing"
	"github.com/SalehMWS/Muse/internal/worker/domain"
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

type counters struct {
	processed    atomic.Int64
	succeeded    atomic.Int64
	retried      atomic.Int64
	deadLettered atomic.Int64
	failed       atomic.Int64
}

type PoolOptions struct {
	Workers  int
	Block    time.Duration
	Queue    string
	Recorder *metrics.Worker
}

type Pool struct {
	broker     Broker
	dispatcher *Dispatcher
	logger     *zap.Logger
	workers    int
	block      time.Duration
	queue      string
	counters   counters
	recorder   *metrics.Worker
}

func NewPool(broker Broker, dispatcher *Dispatcher, log *zap.Logger, opts PoolOptions) *Pool {
	if log == nil {
		log = zap.NewNop()
	}
	if opts.Workers <= 0 {
		opts.Workers = defaultWorkers
	}
	if opts.Block <= 0 {
		opts.Block = defaultBlock
	}
	if opts.Queue == "" {
		opts.Queue = "default"
	}
	return &Pool{
		broker:     broker,
		dispatcher: dispatcher,
		logger:     log,
		workers:    opts.Workers,
		block:      opts.Block,
		queue:      opts.Queue,
		recorder:   opts.Recorder,
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
	job := delivery.Job
	jobType := string(job.Type)

	p.counters.processed.Add(1)
	p.recorder.JobStarted()

	jobCtx, scoped := p.scope(ctx, job)
	start := time.Now()
	err := p.dispatch(jobCtx, job)
	elapsed := time.Since(start)

	if err != nil {
		p.recorder.JobFinished(jobType, metrics.OutcomeFailure, elapsed)
		p.onFailure(jobCtx, scoped, delivery, err)
		return
	}

	if ackErr := p.broker.Ack(jobCtx, delivery); ackErr != nil {
		p.recorder.JobFinished(jobType, metrics.OutcomeFailure, elapsed)
		scoped.Error("worker: ack succeeded job", zap.Error(ackErr))
		return
	}

	p.counters.succeeded.Add(1)
	p.recorder.JobFinished(jobType, metrics.OutcomeSuccess, elapsed)
	scoped.Info("worker: job completed", zap.Duration("duration", elapsed))
}

func (p *Pool) dispatch(ctx context.Context, job domain.Job) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			p.recorder.PanicRecovered()
			err = fmt.Errorf("worker: handler panic: %v", recovered)
		}
	}()
	return p.dispatcher.Dispatch(ctx, job)
}

func (p *Pool) scope(ctx context.Context, job domain.Job) (context.Context, *zap.Logger) {
	ids := tracing.IDs{
		RequestID:     job.ID,
		CorrelationID: job.CorrelationID,
		TraceID:       job.TraceID,
		SpanID:        tracing.NewSpanID(),
	}
	if ids.TraceID == "" {
		ids.TraceID = tracing.NewTraceID()
	}
	if ids.CorrelationID == "" {
		ids.CorrelationID = job.ID
	}

	scoped := p.logger.With(append(ids.Fields(),
		zap.String("module", "worker"),
		zap.String("job_id", job.ID),
		zap.String("job_type", string(job.Type)),
		zap.String("queue", p.queue),
		zap.Int("attempt", job.Attempt),
	)...)

	ctx = tracing.WithIDs(ctx, ids)
	return logger.WithContext(ctx, scoped), scoped
}

func (p *Pool) onFailure(ctx context.Context, scoped *zap.Logger, delivery Delivery, cause error) {
	jobType := string(delivery.Job.Type)
	p.counters.failed.Add(1)

	if delivery.Job.HasAttemptsLeft() {
		retried := delivery.Job.NextAttempt()
		if err := p.broker.Enqueue(ctx, retried); err != nil {
			scoped.Error("worker: requeue job", zap.Error(err))
			return
		}
		if err := p.broker.Ack(ctx, delivery); err != nil {
			scoped.Error("worker: ack requeued job", zap.Error(err))
			return
		}
		p.counters.retried.Add(1)
		p.recorder.JobOutcome(jobType, metrics.OutcomeRetried)
		scoped.Warn("worker: job retried",
			zap.Int("next_attempt", retried.Attempt),
			zap.Error(cause),
		)
		return
	}

	if err := p.broker.DeadLetter(ctx, delivery.Job, cause.Error()); err != nil {
		scoped.Error("worker: dead-letter job", zap.Error(err))
		return
	}
	if err := p.broker.Ack(ctx, delivery); err != nil {
		scoped.Error("worker: ack dead-lettered job", zap.Error(err))
		return
	}
	p.counters.deadLettered.Add(1)
	p.recorder.JobOutcome(jobType, metrics.OutcomeDeadLettered)
	scoped.Error("worker: job dead-lettered", zap.Error(cause))
}

func (p *Pool) Stats() Stats {
	return Stats{
		Processed:    p.counters.processed.Load(),
		Succeeded:    p.counters.succeeded.Load(),
		Retried:      p.counters.retried.Load(),
		DeadLettered: p.counters.deadLettered.Load(),
		Failed:       p.counters.failed.Load(),
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
