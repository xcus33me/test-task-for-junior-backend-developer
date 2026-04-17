package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
)

type ScheduleRepository struct {
	pool *pgxpool.Pool
}

func NewScheduleRepository(pool *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{pool: pool}
}

func (r *ScheduleRepository) Create(ctx context.Context, s *scheduledomain.Schedule) (*scheduledomain.Schedule, error) {
	const query = `
		INSERT INTO schedules (title, description, recurrence_type, every_n_days, day_of_month, specific_dates, start_date, end_date, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, title, description, recurrence_type, every_n_days, day_of_month, specific_dates, start_date, end_date, is_active, created_at, updated_at
	`

	specificDates := timesToDates(s.SpecificDates)

	row := r.pool.QueryRow(ctx, query,
		s.Title, s.Description, s.RecurrenceType,
		s.EveryNDays, s.DayOfMonth, specificDates,
		s.StartDate, s.EndDate, s.IsActive,
		s.CreatedAt, s.UpdatedAt,
	)

	return scanSchedule(row)
}

func (r *ScheduleRepository) GetByID(ctx context.Context, id int64) (*scheduledomain.Schedule, error) {
	const query = `
		SELECT id, title, description, recurrence_type, every_n_days, day_of_month, specific_dates, start_date, end_date, is_active, created_at, updated_at
		FROM schedules
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	s, err := scanSchedule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, scheduledomain.ErrNotFound
		}
		return nil, err
	}

	return s, nil
}

func (r *ScheduleRepository) Update(ctx context.Context, s *scheduledomain.Schedule) (*scheduledomain.Schedule, error) {
	const query = `
		UPDATE schedules
		SET title = $1,
			description = $2,
			recurrence_type = $3,
			every_n_days = $4,
			day_of_month = $5,
			specific_dates = $6,
			start_date = $7,
			end_date = $8,
			is_active = $9,
			updated_at = $10
		WHERE id = $11
		RETURNING id, title, description, recurrence_type, every_n_days, day_of_month, specific_dates, start_date, end_date, is_active, created_at, updated_at
	`

	specificDates := timesToDates(s.SpecificDates)

	row := r.pool.QueryRow(ctx, query,
		s.Title, s.Description, s.RecurrenceType,
		s.EveryNDays, s.DayOfMonth, specificDates,
		s.StartDate, s.EndDate, s.IsActive,
		s.UpdatedAt, s.ID,
	)

	updated, err := scanSchedule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, scheduledomain.ErrNotFound
		}
		return nil, err
	}

	return updated, nil
}

func (r *ScheduleRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM schedules WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return scheduledomain.ErrNotFound
	}

	return nil
}

func (r *ScheduleRepository) List(ctx context.Context) ([]scheduledomain.Schedule, error) {
	const query = `
		SELECT id, title, description, recurrence_type, every_n_days, day_of_month, specific_dates, start_date, end_date, is_active, created_at, updated_at
		FROM schedules
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schedules := make([]scheduledomain.Schedule, 0)
	for rows.Next() {
		s, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, *s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schedules, nil
}

func (r *ScheduleRepository) ListActive(ctx context.Context) ([]scheduledomain.Schedule, error) {
	const query = `
		SELECT id, title, description, recurrence_type, every_n_days, day_of_month, specific_dates, start_date, end_date, is_active, created_at, updated_at
		FROM schedules
		WHERE is_active = TRUE
		ORDER BY id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schedules := make([]scheduledomain.Schedule, 0)
	for rows.Next() {
		s, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, *s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schedules, nil
}

type scheduleScanner interface {
	Scan(dest ...any) error
}

func scanSchedule(scanner scheduleScanner) (*scheduledomain.Schedule, error) {
	var (
		s              scheduledomain.Schedule
		recurrenceType string
		specificDates  []time.Time
	)

	if err := scanner.Scan(
		&s.ID,
		&s.Title,
		&s.Description,
		&recurrenceType,
		&s.EveryNDays,
		&s.DayOfMonth,
		&specificDates,
		&s.StartDate,
		&s.EndDate,
		&s.IsActive,
		&s.CreatedAt,
		&s.UpdatedAt,
	); err != nil {
		return nil, err
	}

	s.RecurrenceType = scheduledomain.RecurrenceType(recurrenceType)
	s.SpecificDates = specificDates

	return &s, nil
}

func timesToDates(times []time.Time) []time.Time {
	if times == nil {
		return nil
	}
	dates := make([]time.Time, len(times))
	for i, t := range times {
		dates[i] = t.Truncate(24 * time.Hour)
	}
	return dates
}
