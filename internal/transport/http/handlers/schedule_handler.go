package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	scheduleusecase "example.com/taskservice/internal/usecase/schedule"
)

type ScheduleHandler struct {
	usecase scheduleusecase.Usecase
}

func NewScheduleHandler(usecase scheduleusecase.Usecase) *ScheduleHandler {
	return &ScheduleHandler{usecase: usecase}
}

func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req scheduleMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input, err := req.toCreateInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), input)
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newScheduleDTO(created))
}

func (h *ScheduleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newScheduleDTO(s))
}

func (h *ScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req scheduleMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input, err := req.toUpdateInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, input)
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newScheduleDTO(updated))
}

func (h *ScheduleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ScheduleHandler) List(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.usecase.List(r.Context())
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	response := make([]scheduleDTO, 0, len(schedules))
	for i := range schedules {
		response = append(response, newScheduleDTO(&schedules[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *ScheduleHandler) GenerateTasks(w http.ResponseWriter, r *http.Request) {
	var req generateTasksDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid date format, expected YYYY-MM-DD"))
		return
	}

	tasks, err := h.usecase.GenerateTasks(r.Context(), date)
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	response := make([]taskDTO, 0, len(tasks))
	for i := range tasks {
		response = append(response, newTaskDTO(&tasks[i]))
	}

	writeJSON(w, http.StatusCreated, response)
}

func writeScheduleUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, scheduledomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, scheduleusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

// DTOs

type scheduleMutationDTO struct {
	Title          string                        `json:"title"`
	Description    string                        `json:"description"`
	RecurrenceType scheduledomain.RecurrenceType `json:"recurrence_type"`
	EveryNDays     *int                          `json:"every_n_days,omitempty"`
	DayOfMonth     *int                          `json:"day_of_month,omitempty"`
	SpecificDates  []string                      `json:"specific_dates,omitempty"`
	StartDate      string                        `json:"start_date"`
	EndDate        *string                       `json:"end_date,omitempty"`
	IsActive       *bool                         `json:"is_active,omitempty"`
}

func (d *scheduleMutationDTO) parseDates() (startDate time.Time, endDate *time.Time, specificDates []time.Time, err error) {
	startDate, err = time.Parse("2006-01-02", d.StartDate)
	if err != nil {
		return time.Time{}, nil, nil, errors.New("invalid start_date format, expected YYYY-MM-DD")
	}

	if d.EndDate != nil {
		ed, err := time.Parse("2006-01-02", *d.EndDate)
		if err != nil {
			return time.Time{}, nil, nil, errors.New("invalid end_date format, expected YYYY-MM-DD")
		}
		endDate = &ed
	}

	for _, ds := range d.SpecificDates {
		t, err := time.Parse("2006-01-02", ds)
		if err != nil {
			return time.Time{}, nil, nil, errors.New("invalid specific_dates format, expected YYYY-MM-DD")
		}
		specificDates = append(specificDates, t)
	}

	return startDate, endDate, specificDates, nil
}

func (d *scheduleMutationDTO) toCreateInput() (scheduleusecase.CreateInput, error) {
	startDate, endDate, specificDates, err := d.parseDates()
	if err != nil {
		return scheduleusecase.CreateInput{}, err
	}

	return scheduleusecase.CreateInput{
		Title:          d.Title,
		Description:    d.Description,
		RecurrenceType: d.RecurrenceType,
		EveryNDays:     d.EveryNDays,
		DayOfMonth:     d.DayOfMonth,
		SpecificDates:  specificDates,
		StartDate:      startDate,
		EndDate:        endDate,
	}, nil
}

func (d *scheduleMutationDTO) toUpdateInput() (scheduleusecase.UpdateInput, error) {
	startDate, endDate, specificDates, err := d.parseDates()
	if err != nil {
		return scheduleusecase.UpdateInput{}, err
	}

	isActive := true
	if d.IsActive != nil {
		isActive = *d.IsActive
	}

	return scheduleusecase.UpdateInput{
		Title:          d.Title,
		Description:    d.Description,
		RecurrenceType: d.RecurrenceType,
		EveryNDays:     d.EveryNDays,
		DayOfMonth:     d.DayOfMonth,
		SpecificDates:  specificDates,
		StartDate:      startDate,
		EndDate:        endDate,
		IsActive:       isActive,
	}, nil
}

type scheduleDTO struct {
	ID             int64                         `json:"id"`
	Title          string                        `json:"title"`
	Description    string                        `json:"description"`
	RecurrenceType scheduledomain.RecurrenceType `json:"recurrence_type"`
	EveryNDays     *int                          `json:"every_n_days,omitempty"`
	DayOfMonth     *int                          `json:"day_of_month,omitempty"`
	SpecificDates  []string                      `json:"specific_dates,omitempty"`
	StartDate      string                        `json:"start_date"`
	EndDate        *string                       `json:"end_date,omitempty"`
	IsActive       bool                          `json:"is_active"`
	CreatedAt      time.Time                     `json:"created_at"`
	UpdatedAt      time.Time                     `json:"updated_at"`
}

func newScheduleDTO(s *scheduledomain.Schedule) scheduleDTO {
	dto := scheduleDTO{
		ID:             s.ID,
		Title:          s.Title,
		Description:    s.Description,
		RecurrenceType: s.RecurrenceType,
		EveryNDays:     s.EveryNDays,
		DayOfMonth:     s.DayOfMonth,
		StartDate:      s.StartDate.Format("2006-01-02"),
		IsActive:       s.IsActive,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}

	if s.EndDate != nil {
		formatted := s.EndDate.Format("2006-01-02")
		dto.EndDate = &formatted
	}

	if len(s.SpecificDates) > 0 {
		dto.SpecificDates = make([]string, len(s.SpecificDates))
		for i, d := range s.SpecificDates {
			dto.SpecificDates[i] = d.Format("2006-01-02")
		}
	}

	return dto
}

type generateTasksDTO struct {
	Date string `json:"date"`
}

func (d scheduleDTO) MarshalJSON() ([]byte, error) {
	type Alias scheduleDTO
	a := struct {
		Alias
	}{Alias: Alias(d)}
	return json.Marshal(a)
}
