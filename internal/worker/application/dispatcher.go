package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/SalehMWS/Muse/internal/worker/domain"
)

var ErrNoHandler = errors.New("no handler registered for job type")

type Dispatcher struct {
	handlers map[domain.JobType]Handler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{handlers: map[domain.JobType]Handler{}}
}

func (d *Dispatcher) Register(jobType domain.JobType, handler Handler) {
	d.handlers[jobType] = handler
}

func (d *Dispatcher) Dispatch(ctx context.Context, job domain.Job) error {
	handler, ok := d.handlers[job.Type]
	if !ok {
		return fmt.Errorf("%w: %s", ErrNoHandler, job.Type)
	}
	return handler.Handle(ctx, job)
}
