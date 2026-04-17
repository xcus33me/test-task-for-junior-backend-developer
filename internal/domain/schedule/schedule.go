package schedule

import "time"

type RecurrenceType string

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthly       RecurrenceType = "monthly"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceEvenDays      RecurrenceType = "even_days"
	RecurrenceOddDays       RecurrenceType = "odd_days"
)

func (r RecurrenceType) Valid() bool {
	switch r {
	case RecurrenceDaily, RecurrenceMonthly, RecurrenceSpecificDates, RecurrenceEvenDays, RecurrenceOddDays:
		return true
	default:
		return false
	}
}

type Schedule struct {
	ID             int64          `json:"id"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	RecurrenceType RecurrenceType `json:"recurrence_type"`
	EveryNDays     *int           `json:"every_n_days,omitempty"`
	DayOfMonth     *int           `json:"day_of_month,omitempty"`
	SpecificDates  []time.Time    `json:"specific_dates,omitempty"`
	StartDate      time.Time      `json:"start_date"`
	EndDate        *time.Time     `json:"end_date,omitempty"`
	IsActive       bool           `json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

func (s *Schedule) MatchesDate(date time.Time) bool {
	d := date.Truncate(24 * time.Hour)
	start := s.StartDate.Truncate(24 * time.Hour)

	if d.Before(start) {
		return false
	}

	if s.EndDate != nil {
		end := s.EndDate.Truncate(24 * time.Hour)
		if d.After(end) {
			return false
		}
	}

	if !s.IsActive {
		return false
	}

	switch s.RecurrenceType {
	case RecurrenceDaily:
		n := 1
		if s.EveryNDays != nil && *s.EveryNDays > 0 {
			n = *s.EveryNDays
		}
		days := int(d.Sub(start).Hours() / 24)
		return days%n == 0

	case RecurrenceMonthly:
		if s.DayOfMonth == nil {
			return false
		}
		return d.Day() == *s.DayOfMonth

	case RecurrenceSpecificDates:
		for _, sd := range s.SpecificDates {
			if sd.Truncate(24 * time.Hour).Equal(d) {
				return true
			}
		}
		return false

	case RecurrenceEvenDays:
		return d.Day()%2 == 0

	case RecurrenceOddDays:
		return d.Day()%2 != 0

	default:
		return false
	}
}
