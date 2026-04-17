package schedule

import (
	"context"
	"fmt"
	"strings"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo     Repository
	taskRepo TaskRepository
	now      func() time.Time
}

func NewService(repo Repository, taskRepo TaskRepository) *Service {
	return &Service{
		repo:     repo,
		taskRepo: taskRepo,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*scheduledomain.Schedule, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now()
	model := &scheduledomain.Schedule{
		Title:          normalized.Title,
		Description:    normalized.Description,
		RecurrenceType: normalized.RecurrenceType,
		EveryNDays:     normalized.EveryNDays,
		DayOfMonth:     normalized.DayOfMonth,
		SpecificDates:  normalized.SpecificDates,
		StartDate:      normalized.StartDate,
		EndDate:        normalized.EndDate,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return s.repo.Create(ctx, model)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*scheduledomain.Schedule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*scheduledomain.Schedule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &scheduledomain.Schedule{
		ID:             id,
		Title:          normalized.Title,
		Description:    normalized.Description,
		RecurrenceType: normalized.RecurrenceType,
		EveryNDays:     normalized.EveryNDays,
		DayOfMonth:     normalized.DayOfMonth,
		SpecificDates:  normalized.SpecificDates,
		StartDate:      normalized.StartDate,
		EndDate:        normalized.EndDate,
		IsActive:       normalized.IsActive,
		UpdatedAt:      s.now(),
	}

	return s.repo.Update(ctx, model)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]scheduledomain.Schedule, error) {
	return s.repo.List(ctx)
}

// GenerateTasks creates tasks for all active schedules that match the given date.
// It skips schedules that already have a task for this date (unique constraint).
func (s *Service) GenerateTasks(ctx context.Context, date time.Time) ([]taskdomain.Task, error) {
	date = date.Truncate(24 * time.Hour)

	schedules, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active schedules: %w", err)
	}

	var created []taskdomain.Task

	for i := range schedules {
		sched := &schedules[i]
		if !sched.MatchesDate(date) {
			continue
		}

		now := s.now()
		task := &taskdomain.Task{
			Title:         sched.Title,
			Description:   sched.Description,
			Status:        taskdomain.StatusNew,
			ScheduleID:    &sched.ID,
			ScheduledDate: &date,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		newTask, err := s.taskRepo.Create(ctx, task)
		if err != nil {
			// If we get a unique constraint violation, the task already exists — skip it.
			if isUniqueViolation(err) {
				continue
			}
			return nil, fmt.Errorf("create task for schedule %d: %w", sched.ID, err)
		}

		created = append(created, *newTask)
	}

	return created, nil
}

func isUniqueViolation(err error) bool {
	// code 23505 is unique_violation
	return strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "unique")
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.RecurrenceType.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid recurrence_type", ErrInvalidInput)
	}

	if err := validateRecurrenceParams(input.RecurrenceType, input.EveryNDays, input.DayOfMonth, input.SpecificDates); err != nil {
		return CreateInput{}, err
	}

	if input.StartDate.IsZero() {
		return CreateInput{}, fmt.Errorf("%w: start_date is required", ErrInvalidInput)
	}

	if input.EndDate != nil && input.EndDate.Before(input.StartDate) {
		return CreateInput{}, fmt.Errorf("%w: end_date must not be before start_date", ErrInvalidInput)
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.RecurrenceType.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid recurrence_type", ErrInvalidInput)
	}

	if err := validateRecurrenceParams(input.RecurrenceType, input.EveryNDays, input.DayOfMonth, input.SpecificDates); err != nil {
		return UpdateInput{}, err
	}

	if input.StartDate.IsZero() {
		return UpdateInput{}, fmt.Errorf("%w: start_date is required", ErrInvalidInput)
	}

	if input.EndDate != nil && input.EndDate.Before(input.StartDate) {
		return UpdateInput{}, fmt.Errorf("%w: end_date must not be before start_date", ErrInvalidInput)
	}

	return input, nil
}

func validateRecurrenceParams(rt scheduledomain.RecurrenceType, everyNDays *int, dayOfMonth *int, specificDates []time.Time) error {
	switch rt {
	case scheduledomain.RecurrenceDaily:
		if everyNDays != nil && *everyNDays < 1 {
			return fmt.Errorf("%w: every_n_days must be >= 1", ErrInvalidInput)
		}

	case scheduledomain.RecurrenceMonthly:
		if dayOfMonth == nil {
			return fmt.Errorf("%w: day_of_month is required for monthly recurrence", ErrInvalidInput)
		}
		if *dayOfMonth < 1 || *dayOfMonth > 31 {
			return fmt.Errorf("%w: day_of_month must be between 1 and 31", ErrInvalidInput)
		}

	case scheduledomain.RecurrenceSpecificDates:
		if len(specificDates) == 0 {
			return fmt.Errorf("%w: specific_dates must not be empty", ErrInvalidInput)
		}

	case scheduledomain.RecurrenceEvenDays, scheduledomain.RecurrenceOddDays:
	}

	return nil
}
