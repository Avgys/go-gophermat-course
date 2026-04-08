package processor

import "context"

type JobRunner interface {
	Enqueue(ctx context.Context, j Job) error
	Run(ctx context.Context) error
}
