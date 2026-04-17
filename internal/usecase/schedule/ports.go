package schedule

import (
	"context"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, schedule *scheduledomain.Schedule) (*scheduledomain.Schedule, error)
	GetByID(ctx context.Context, id int64) (*scheduledomain.Schedule, error)
	Update(ctx context.Context, schedule *scheduledomain.Schedule) (*scheduledomain.Schedule, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]scheduledomain.Schedule, error)
	ListActive(ctx context.Context) ([]scheduledomain.Schedule, error)
}

type TaskRepository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*scheduledomain.Schedule, error)
	GetByID(ctx context.Context, id int64) (*scheduledomain.Schedule, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*scheduledomain.Schedule, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]scheduledomain.Schedule, error)
	GenerateTasks(ctx context.Context, date time.Time) ([]taskdomain.Task, error)
}

type CreateInput struct {
	Title          string
	Description    string
	RecurrenceType scheduledomain.RecurrenceType
	EveryNDays     *int
	DayOfMonth     *int
	SpecificDates  []time.Time
	StartDate      time.Time
	EndDate        *time.Time
}

type UpdateInput struct {
	Title          string
	Description    string
	RecurrenceType scheduledomain.RecurrenceType
	EveryNDays     *int
	DayOfMonth     *int
	SpecificDates  []time.Time
	StartDate      time.Time
	EndDate        *time.Time
	IsActive       bool
}
