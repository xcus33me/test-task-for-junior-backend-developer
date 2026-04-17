package worker

import (
	"context"
	"log/slog"
	"time"

	scheduleusecase "example.com/taskservice/internal/usecase/schedule"
)

type Scheduler struct {
	usecase  scheduleusecase.Usecase
	logger   *slog.Logger
	interval time.Duration
}

func NewScheduler(usecase scheduleusecase.Usecase, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		usecase:  usecase,
		logger:   logger,
		interval: 1 * time.Hour,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	s.generate(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler stopped")
			return
		case <-ticker.C:
			s.generate(ctx)
		}
	}
}

func (s *Scheduler) generate(ctx context.Context) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	tasks, err := s.usecase.GenerateTasks(ctx, today)
	if err != nil {
		s.logger.Error("scheduler: failed to generate tasks", "error", err)
		return
	}

	if len(tasks) > 0 {
		s.logger.Info("scheduler: generated tasks", "count", len(tasks), "date", today.Format("2006-01-02"))
	}
}
